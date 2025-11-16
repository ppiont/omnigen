package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"go.uber.org/zap"

	"github.com/omnigen/backend/internal/adapters"
	"github.com/omnigen/backend/internal/domain"
)

// GeneratorInput represents input from Step Functions map state
// Matches workflow.asl.json Parameters structure
type GeneratorInput struct {
	JobID string       `json:"job_id"`
	Scene domain.Scene `json:"scene"`
}

// GeneratorOutput represents the output with video URL and last frame for coherence
type GeneratorOutput struct {
	SceneNumber  int     `json:"scene_number"`
	VideoURL     string  `json:"video_url"`      // S3 URL to video
	LastFrameURL string  `json:"last_frame_url"` // S3 URL to last frame (for next scene coherence)
	Duration     float64 `json:"duration"`
	Status       string  `json:"status"`
}

var (
	dynamoClient *dynamodb.Client
	s3Client     *s3.Client
	smClient     *secretsmanager.Client
	jobTable     string
	assetsBucket string
	replicateKey string
	logger       *zap.Logger
	klingAdapter *adapters.KlingAdapter
	httpClient   *http.Client
)

func init() {
	// Initialize logger
	var err error
	logger, err = zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	// Initialize HTTP client for downloads
	httpClient = &http.Client{
		Timeout: 10 * time.Minute, // Videos can be large
	}

	jobTable = os.Getenv("JOB_TABLE")
	assetsBucket = os.Getenv("ASSETS_BUCKET")
	replicateSecretArn := os.Getenv("REPLICATE_SECRET_ARN")

	if jobTable == "" || assetsBucket == "" || replicateSecretArn == "" {
		log.Fatal("Required environment variables not set")
	}

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("Failed to load AWS config: %v", err)
	}

	dynamoClient = dynamodb.NewFromConfig(cfg)
	s3Client = s3.NewFromConfig(cfg)
	smClient = secretsmanager.NewFromConfig(cfg)

	// Fetch Replicate API key from Secrets Manager
	result, err := smClient.GetSecretValue(context.TODO(), &secretsmanager.GetSecretValueInput{
		SecretId: &replicateSecretArn,
	})
	if err != nil {
		log.Fatalf("Failed to fetch Replicate secret: %v", err)
	}

	var secretData map[string]string
	if err := json.Unmarshal([]byte(*result.SecretString), &secretData); err != nil {
		log.Fatalf("Failed to parse secret: %v", err)
	}

	replicateKey = secretData["api_key"]

	// Initialize Kling adapter
	klingAdapter = adapters.NewKlingAdapter(replicateKey, logger)
	logger.Info("Generator Lambda initialized",
		zap.String("job_table", jobTable),
		zap.String("assets_bucket", assetsBucket),
	)
}

func handler(ctx context.Context, input GeneratorInput) (GeneratorOutput, error) {
	log.Printf("Generator Lambda invoked for job %s, scene %d", input.JobID, input.Scene.SceneNumber)

	// Update job progress
	updateStageProgress(ctx, input.JobID, fmt.Sprintf("Generating video for scene %d", input.Scene.SceneNumber))

	// Generate video using Kling API and extract last frame
	videoURL, lastFrameURL, err := generateVideoClip(ctx, input)
	if err != nil {
		log.Printf("Failed to generate video for scene %d: %v", input.Scene.SceneNumber, err)
		return GeneratorOutput{
			SceneNumber: input.Scene.SceneNumber,
			Status:      "failed",
		}, err
	}

	output := GeneratorOutput{
		SceneNumber:  input.Scene.SceneNumber,
		VideoURL:     videoURL,
		LastFrameURL: lastFrameURL,
		Duration:     input.Scene.Duration,
		Status:       "completed",
	}

	log.Printf("Successfully generated video for scene %d: %s", input.Scene.SceneNumber, videoURL)
	return output, nil
}

// generateVideoClip calls Kling API to generate a video clip and extracts last frame
func generateVideoClip(ctx context.Context, input GeneratorInput) (string, string, error) {
	logger.Info("Generating video clip with Kling API",
		zap.String("job_id", input.JobID),
		zap.Int("scene_number", input.Scene.SceneNumber),
		zap.String("prompt", input.Scene.GenerationPrompt),
	)

	// Build video generation request using scene's generation_prompt
	req := &adapters.VideoGenerationRequest{
		Prompt:        input.Scene.GenerationPrompt,
		Duration:      int(input.Scene.Duration),
		AspectRatio:   "16:9",                    // Default to 16:9 for now
		StartImageURL: input.Scene.StartImageURL, // For visual coherence
	}

	// Submit video generation request
	result, err := klingAdapter.GenerateVideo(ctx, req)
	if err != nil {
		return "", "", fmt.Errorf("failed to submit video generation: %w", err)
	}

	logger.Info("Video generation submitted",
		zap.String("prediction_id", result.PredictionID),
		zap.String("status", result.Status),
	)

	// Poll for completion
	maxAttempts := 120 // 120 * 5s = 10 minutes max
	pollInterval := 5 * time.Second

	for attempt := 0; attempt < maxAttempts; attempt++ {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			return "", "", fmt.Errorf("context cancelled while polling: %w", ctx.Err())
		default:
		}

		// Wait before polling (except first attempt where we already have result)
		if attempt > 0 {
			time.Sleep(pollInterval)
			result, err = klingAdapter.GetStatus(ctx, result.PredictionID)
			if err != nil {
				logger.Warn("Failed to get status, retrying",
					zap.Error(err),
					zap.Int("attempt", attempt),
				)
				continue
			}
		}

		logger.Debug("Polling video generation status",
			zap.String("status", result.Status),
			zap.Int("attempt", attempt),
		)

		switch result.Status {
		case "completed":
			logger.Info("Video generation completed",
				zap.String("video_url", result.VideoURL),
			)

			// Create temp directory for processing
			tmpDir := filepath.Join("/tmp", input.JobID)
			if err := os.MkdirAll(tmpDir, 0755); err != nil {
				return "", "", fmt.Errorf("failed to create temp dir: %w", err)
			}
			defer os.RemoveAll(tmpDir)

			// Download video from Replicate to temp file
			videoPath := filepath.Join(tmpDir, fmt.Sprintf("scene-%d.mp4", input.Scene.SceneNumber))
			if err := downloadVideoToFile(ctx, result.VideoURL, videoPath); err != nil {
				return "", "", fmt.Errorf("failed to download video: %w", err)
			}

			// Extract last frame using ffmpeg
			framePath := filepath.Join(tmpDir, fmt.Sprintf("scene-%d-last-frame.jpg", input.Scene.SceneNumber))
			if err := extractLastFrame(ctx, videoPath, framePath); err != nil {
				return "", "", fmt.Errorf("failed to extract last frame: %w", err)
			}

			// Upload video to S3
			videoS3Key := fmt.Sprintf("scenes/%s/scene-%d.mp4", input.JobID, input.Scene.SceneNumber)
			videoS3URL, err := uploadFileToS3(ctx, videoPath, videoS3Key, "video/mp4")
			if err != nil {
				return "", "", fmt.Errorf("failed to upload video to S3: %w", err)
			}

			// Upload last frame to S3
			frameS3Key := fmt.Sprintf("scenes/%s/scene-%d-last-frame.jpg", input.JobID, input.Scene.SceneNumber)
			frameS3URL, err := uploadFileToS3(ctx, framePath, frameS3Key, "image/jpeg")
			if err != nil {
				return "", "", fmt.Errorf("failed to upload frame to S3: %w", err)
			}

			logger.Info("Video and last frame uploaded to S3",
				zap.String("video_url", videoS3URL),
				zap.String("frame_url", frameS3URL),
			)

			return videoS3URL, frameS3URL, nil

		case "failed":
			return "", "", fmt.Errorf("video generation failed: %s", result.Error)

		case "processing":
			// Continue polling
			continue

		default:
			logger.Warn("Unknown status", zap.String("status", result.Status))
		}
	}

	return "", "", fmt.Errorf("video generation timed out after %d attempts", maxAttempts)
}

// downloadVideoToFile downloads a video from a URL to a local file
func downloadVideoToFile(ctx context.Context, url, localPath string) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	file, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	written, err := io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write video data: %w", err)
	}

	logger.Info("Video downloaded to file",
		zap.String("path", localPath),
		zap.Int64("size_bytes", written),
	)

	return nil
}

// extractLastFrame uses ffmpeg to extract the last frame of a video
func extractLastFrame(ctx context.Context, videoPath, framePath string) error {
	logger.Info("Extracting last frame with ffmpeg",
		zap.String("video", videoPath),
		zap.String("output", framePath),
	)

	// ffmpeg -sseof -1 -i input.mp4 -update 1 -q:v 1 -frames:v 1 output.jpg
	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-sseof", "-1", // Seek to end of file
		"-i", videoPath, // Input file
		"-update", "1", // Update output file
		"-q:v", "1", // High quality JPEG
		"-frames:v", "1", // Extract 1 frame
		"-y", // Overwrite output
		framePath,
	)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		logger.Error("ffmpeg failed",
			zap.Error(err),
			zap.String("stderr", stderr.String()),
		)
		return fmt.Errorf("ffmpeg extract frame failed: %w (stderr: %s)", err, stderr.String())
	}

	logger.Info("Last frame extracted successfully",
		zap.String("frame_path", framePath),
	)

	return nil
}

// uploadFileToS3 uploads a local file to S3 and returns the S3 URL
func uploadFileToS3(ctx context.Context, localPath, s3Key, contentType string) (string, error) {
	file, err := os.Open(localPath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	_, err = s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(assetsBucket),
		Key:         aws.String(s3Key),
		Body:        file,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload to S3: %w", err)
	}

	// Return S3 URL
	s3URL := fmt.Sprintf("s3://%s/%s", assetsBucket, s3Key)

	logger.Info("File uploaded to S3",
		zap.String("s3_key", s3Key),
		zap.String("s3_url", s3URL),
	)

	return s3URL, nil
}

// updateStageProgress updates the current stage in DynamoDB
func updateStageProgress(ctx context.Context, jobID, stage string) {
	_, err := dynamoClient.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: &jobTable,
		Key: map[string]types.AttributeValue{
			"job_id": &types.AttributeValueMemberS{Value: jobID},
		},
		UpdateExpression: aws.String("SET current_stage = :stage"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":stage": &types.AttributeValueMemberS{Value: stage},
		},
	})
	if err != nil {
		log.Printf("Warning: Failed to update progress: %v", err)
	}
}

func main() {
	lambda.Start(handler)
}
