package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
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
)

// AudioGeneratorInput represents simplified input from Step Functions
type AudioGeneratorInput struct {
	JobID      string `json:"job_id"`
	Prompt     string `json:"prompt"`      // Video prompt (to derive music from)
	Duration   int    `json:"duration"`    // Total video duration in seconds
	MusicMood  string `json:"music_mood"`  // upbeat, calm, dramatic, energetic
	MusicStyle string `json:"music_style"` // electronic, acoustic, orchestral
}

// AudioGeneratorOutput represents the output with audio URL
type AudioGeneratorOutput struct {
	JobID    string `json:"job_id"`
	MusicURL string `json:"music_url"` // S3 URL to generated music
	Status   string `json:"status"`
}

var (
	dynamoClient   *dynamodb.Client
	s3Client       *s3.Client
	secretsClient  *secretsmanager.Client
	minimaxAdapter *adapters.MinimaxAdapter
	jobTable       string
	assetsBucket   string
	logger         *zap.Logger
	httpClient     *http.Client
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
		Timeout: 10 * time.Minute, // Music files can be large
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
	secretsClient = secretsmanager.NewFromConfig(cfg)

	// Fetch Replicate API key from Secrets Manager
	result, err := secretsClient.GetSecretValue(context.TODO(), &secretsmanager.GetSecretValueInput{
		SecretId: &replicateSecretArn,
	})
	if err != nil {
		log.Fatalf("Failed to fetch Replicate secret: %v", err)
	}

	var secretData map[string]string
	if err := json.Unmarshal([]byte(*result.SecretString), &secretData); err != nil {
		log.Fatalf("Failed to parse secret: %v", err)
	}

	replicateKey := secretData["api_key"]

	// Initialize Minimax adapter
	minimaxAdapter = adapters.NewMinimaxAdapter(replicateKey, logger)

	logger.Info("AudioGenerator Lambda initialized (Minimax music-1.5)",
		zap.String("job_table", jobTable),
		zap.String("assets_bucket", assetsBucket),
	)
}

func handler(ctx context.Context, input AudioGeneratorInput) (AudioGeneratorOutput, error) {
	log.Printf("AudioGenerator Lambda invoked for job %s (mood: %s, style: %s)",
		input.JobID, input.MusicMood, input.MusicStyle)

	// Update job progress
	updateStageProgress(ctx, input.JobID, "Generating background music with AI")

	// Generate music using Minimax
	musicURL, err := generateMusic(ctx, input)
	if err != nil {
		log.Printf("Failed to generate music: %v", err)
		return AudioGeneratorOutput{
			JobID:  input.JobID,
			Status: "failed",
		}, err
	}

	output := AudioGeneratorOutput{
		JobID:    input.JobID,
		MusicURL: musicURL,
		Status:   "completed",
	}

	log.Printf("Successfully generated music for job %s: %s", input.JobID, musicURL)
	return output, nil
}

// generateMusic generates music using Minimax and uploads to S3
func generateMusic(ctx context.Context, input AudioGeneratorInput) (string, error) {
	logger.Info("Generating music with Minimax",
		zap.String("job_id", input.JobID),
		zap.String("prompt", input.Prompt),
		zap.String("mood", input.MusicMood),
		zap.String("style", input.MusicStyle),
		zap.Int("duration", input.Duration),
	)

	// Submit music generation request
	req := &adapters.MusicGenerationRequest{
		Prompt:     input.Prompt,
		Duration:   input.Duration,
		MusicMood:  input.MusicMood,
		MusicStyle: input.MusicStyle,
	}

	result, err := minimaxAdapter.GenerateMusic(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to submit music generation: %w", err)
	}

	logger.Info("Music generation submitted",
		zap.String("prediction_id", result.PredictionID),
		zap.String("status", result.Status),
	)

	// Poll for completion (max 5 minutes)
	maxAttempts := 60 // 60 * 5s = 5 minutes max
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
			result, err = minimaxAdapter.GetStatus(ctx, result.PredictionID)
			if err != nil {
				logger.Warn("Failed to get status, retrying",
					zap.Error(err),
					zap.Int("attempt", attempt),
				)
				continue
			}
		}

		logger.Debug("Polling music generation status",
			zap.String("status", result.Status),
			zap.Int("attempt", attempt),
		)

		switch result.Status {
		case "succeeded":
			logger.Info("Music generation completed",
				zap.String("audio_url", result.AudioURL),
			)

			// Download music from Replicate
			tmpDir := filepath.Join("/tmp", input.JobID)
			if err := os.MkdirAll(tmpDir, 0755); err != nil {
				return "", fmt.Errorf("failed to create temp dir: %w", err)
			}
			defer os.RemoveAll(tmpDir)

			musicPath := filepath.Join(tmpDir, "music.mp3")
			if err := downloadMusicToFile(ctx, result.AudioURL, musicPath); err != nil {
				return "", fmt.Errorf("failed to download music: %w", err)
			}

			// Upload music to S3
			musicS3Key := fmt.Sprintf("music/%s/music.mp3", input.JobID)
			musicS3URL, err := uploadFileToS3(ctx, musicPath, musicS3Key, "audio/mpeg")
			if err != nil {
				return "", fmt.Errorf("failed to upload music to S3: %w", err)
			}

			logger.Info("Music uploaded to S3",
				zap.String("music_url", musicS3URL),
			)

			return musicS3URL, nil

		case "failed":
			return "", fmt.Errorf("music generation failed: %s", result.Error)

		case "processing", "starting":
			// Continue polling
			continue

		default:
			logger.Warn("Unknown status", zap.String("status", result.Status))
		}
	}

	return "", fmt.Errorf("music generation timed out after %d attempts", maxAttempts)
}

// downloadMusicToFile downloads music from a URL to a local file
func downloadMusicToFile(ctx context.Context, url, localPath string) error {
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
		return fmt.Errorf("failed to write music data: %w", err)
	}

	logger.Info("Music downloaded to file",
		zap.String("path", localPath),
		zap.Int64("size_bytes", written),
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
