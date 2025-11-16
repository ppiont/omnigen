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
)

// GeneratorInput represents input from parser or Step Functions map state
type GeneratorInput struct {
	JobID       string `json:"job_id"`
	SceneNumber int    `json:"scene_number"`
	Description string `json:"description"`
	Duration    int    `json:"duration"`
	VisualStyle string `json:"visual_style"`
}

// GeneratorOutput represents the output with video URL
type GeneratorOutput struct {
	JobID       string `json:"job_id"`
	SceneNumber int    `json:"scene_number"`
	VideoURL    string `json:"video_url"` // S3 URL
	Status      string `json:"status"`
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
	log.Printf("Generator Lambda invoked for job %s, scene %d", input.JobID, input.SceneNumber)

	// Update job progress
	updateStageProgress(ctx, input.JobID, fmt.Sprintf("Generating video for scene %d", input.SceneNumber))

	// Generate video using Kling API (via existing adapter)
	videoURL, err := generateVideoClip(ctx, input)
	if err != nil {
		log.Printf("Failed to generate video for scene %d: %v", input.SceneNumber, err)
		return GeneratorOutput{
			JobID:       input.JobID,
			SceneNumber: input.SceneNumber,
			Status:      "failed",
		}, err
	}

	output := GeneratorOutput{
		JobID:       input.JobID,
		SceneNumber: input.SceneNumber,
		VideoURL:    videoURL,
		Status:      "completed",
	}

	log.Printf("Successfully generated video for scene %d: %s", input.SceneNumber, videoURL)
	return output, nil
}

// generateVideoClip calls Kling API to generate a video clip
func generateVideoClip(ctx context.Context, input GeneratorInput) (string, error) {
	logger.Info("Generating video clip with Kling API",
		zap.String("job_id", input.JobID),
		zap.Int("scene_number", input.SceneNumber),
		zap.String("description", input.Description),
	)

	// Build video generation request
	req := &adapters.VideoGenerationRequest{
		Prompt:      input.Description,
		Duration:    input.Duration, // Kling will map to 5 or 10 seconds
		AspectRatio: "16:9",         // Default to 16:9 for now
		Style:       input.VisualStyle,
	}

	// Submit video generation request
	result, err := klingAdapter.GenerateVideo(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to submit video generation: %w", err)
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
			return "", fmt.Errorf("context cancelled while polling: %w", ctx.Err())
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

			// Download video from Replicate
			videoData, err := downloadVideo(ctx, result.VideoURL)
			if err != nil {
				return "", fmt.Errorf("failed to download video: %w", err)
			}

			// Upload to S3
			s3Key := fmt.Sprintf("%s/scene-%d.mp4", input.JobID, input.SceneNumber)
			s3URL, err := uploadToS3(ctx, s3Key, videoData)
			if err != nil {
				return "", fmt.Errorf("failed to upload to S3: %w", err)
			}

			logger.Info("Video uploaded to S3",
				zap.String("s3_url", s3URL),
			)

			return s3URL, nil

		case "failed":
			return "", fmt.Errorf("video generation failed: %s", result.Error)

		case "processing":
			// Continue polling
			continue

		default:
			logger.Warn("Unknown status", zap.String("status", result.Status))
		}
	}

	return "", fmt.Errorf("video generation timed out after %d attempts", maxAttempts)
}

// downloadVideo downloads a video from a URL
func downloadVideo(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read video data: %w", err)
	}

	logger.Info("Video downloaded",
		zap.Int("size_bytes", len(data)),
	)

	return data, nil
}

// uploadToS3 uploads video data to S3 and returns the S3 URL
func uploadToS3(ctx context.Context, key string, data []byte) (string, error) {
	_, err := s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(assetsBucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String("video/mp4"),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload to S3: %w", err)
	}

	// Return S3 URL
	s3URL := fmt.Sprintf("s3://%s/%s", assetsBucket, key)
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
