package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.uber.org/zap"
)

// ComposerInput represents simplified input from Step Functions
type ComposerInput struct {
	JobID      string      `json:"job_id"`
	ClipVideos []ClipVideo `json:"clip_videos"`         // Generated clips from Generator Lambda
	MusicURL   string      `json:"music_url,omitempty"` // S3 URL to background music
}

// ClipVideo represents a single generated clip
type ClipVideo struct {
	ClipNumber   int     `json:"clip_number"`
	VideoURL     string  `json:"video_url"`      // S3 URL
	LastFrameURL string  `json:"last_frame_url"` // For coherence (not used in composition)
	Duration     float64 `json:"duration"`
	Status       string  `json:"status"`
}

// ComposerOutput represents the final composed video
type ComposerOutput struct {
	JobID    string `json:"job_id"`
	VideoURL string `json:"video_url"` // S3 URL to final video
	Status   string `json:"status"`
}

var (
	dynamoClient *dynamodb.Client
	s3Client     *s3.Client
	jobTable     string
	assetsBucket string
	logger       *zap.Logger
	tmpDir       string
)

func init() {
	// Initialize logger
	var err error
	logger, err = zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	// Set up temporary directory for video processing
	tmpDir = "/tmp"

	jobTable = os.Getenv("JOB_TABLE")
	assetsBucket = os.Getenv("ASSETS_BUCKET")

	if jobTable == "" || assetsBucket == "" {
		log.Fatal("Required environment variables not set")
	}

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("Failed to load AWS config: %v", err)
	}

	dynamoClient = dynamodb.NewFromConfig(cfg)
	s3Client = s3.NewFromConfig(cfg)

	logger.Info("Composer Lambda initialized",
		zap.String("job_table", jobTable),
		zap.String("assets_bucket", assetsBucket),
		zap.String("tmp_dir", tmpDir),
	)
}

func handler(ctx context.Context, input ComposerInput) (ComposerOutput, error) {
	log.Printf("Composer Lambda invoked for job %s with %d clips", input.JobID, len(input.ClipVideos))

	// Update job progress
	updateStageProgress(ctx, input.JobID, fmt.Sprintf("Stitching %d clips into final video", len(input.ClipVideos)))

	// Compose video by stitching clips together (simple concatenation)
	videoURL, err := composeVideo(ctx, input)
	if err != nil {
		log.Printf("Failed to compose video: %v", err)
		return ComposerOutput{
			JobID:  input.JobID,
			Status: "failed",
		}, err
	}

	// Update job status to completed
	if err := updateJobComplete(ctx, input.JobID, videoURL); err != nil {
		log.Printf("Warning: Failed to update job completion: %v", err)
	}

	output := ComposerOutput{
		JobID:    input.JobID,
		VideoURL: videoURL,
		Status:   "completed",
	}

	log.Printf("Successfully composed video for job %s: %s", input.JobID, videoURL)
	return output, nil
}

// composeVideo stitches clips together using ffmpeg
func composeVideo(ctx context.Context, input ComposerInput) (string, error) {
	logger.Info("Starting video composition",
		zap.String("job_id", input.JobID),
		zap.Int("clip_count", len(input.ClipVideos)),
	)

	// Create working directory for this job
	jobDir := filepath.Join(tmpDir, input.JobID)
	if err := os.MkdirAll(jobDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create job directory: %w", err)
	}
	defer os.RemoveAll(jobDir) // Clean up after processing

	// Step 1: Download all clip videos from S3
	logger.Info("Downloading clip videos from S3")
	clipFiles := make([]string, len(input.ClipVideos))
	for i, clip := range input.ClipVideos {
		// Extract key from S3 URL (format: s3://bucket/key)
		s3Key := strings.TrimPrefix(clip.VideoURL, fmt.Sprintf("s3://%s/", assetsBucket))

		localPath := filepath.Join(jobDir, fmt.Sprintf("clip-%d.mp4", clip.ClipNumber))
		if err := downloadFromS3(ctx, s3Key, localPath); err != nil {
			return "", fmt.Errorf("failed to download clip %d: %w", clip.ClipNumber, err)
		}
		clipFiles[i] = localPath
		logger.Debug("Downloaded clip", zap.Int("clip", clip.ClipNumber), zap.String("path", localPath))
	}

	// Step 2: Concatenate clips (simple stitching)
	logger.Info("Concatenating clips")
	videoOnly := filepath.Join(jobDir, "video_only.mp4")
	if err := concatenateClips(ctx, clipFiles, videoOnly); err != nil {
		return "", fmt.Errorf("failed to concatenate clips: %w", err)
	}

	// Step 3: Add music if provided
	finalVideo := videoOnly
	if input.MusicURL != "" {
		logger.Info("Adding background music to video")

		// Download music from S3
		musicKey := strings.TrimPrefix(input.MusicURL, fmt.Sprintf("s3://%s/", assetsBucket))
		musicPath := filepath.Join(jobDir, "music.mp3")
		if err := downloadFromS3(ctx, musicKey, musicPath); err != nil {
			logger.Warn("Failed to download music, proceeding without audio", zap.Error(err))
		} else {
			// Mix audio with video
			videoWithAudio := filepath.Join(jobDir, "final_with_audio.mp4")
			if err := addAudioToVideo(ctx, videoOnly, musicPath, videoWithAudio); err != nil {
				logger.Warn("Failed to add audio, using video without audio", zap.Error(err))
			} else {
				finalVideo = videoWithAudio
			}
		}
	}

	// Step 4: Upload final video to S3
	logger.Info("Uploading final video to S3")
	videoKey := fmt.Sprintf("final-videos/%s/final.mp4", input.JobID)
	if err := uploadToS3(ctx, videoKey, finalVideo); err != nil {
		return "", fmt.Errorf("failed to upload final video: %w", err)
	}

	videoURL := fmt.Sprintf("s3://%s/%s", assetsBucket, videoKey)
	logger.Info("Video composition completed",
		zap.String("video_url", videoURL),
	)

	return videoURL, nil
}

// concatenateClips uses ffmpeg concat demuxer to stitch clips together
// This is simpler and faster than complex transitions
func concatenateClips(ctx context.Context, clipFiles []string, outputFile string) error {
	if len(clipFiles) == 0 {
		return fmt.Errorf("no clip files to concatenate")
	}

	// If only one clip, just copy it
	if len(clipFiles) == 1 {
		logger.Info("Single clip, copying directly")
		return copyFile(clipFiles[0], outputFile)
	}

	// Create concat list file for ffmpeg
	// Format: file 'clip-1.mp4'\nfile 'clip-2.mp4'\n...
	concatListPath := filepath.Join(filepath.Dir(outputFile), "concat_list.txt")
	var concatList strings.Builder
	for _, clipFile := range clipFiles {
		concatList.WriteString(fmt.Sprintf("file '%s'\n", clipFile))
	}

	if err := os.WriteFile(concatListPath, []byte(concatList.String()), 0644); err != nil {
		return fmt.Errorf("failed to create concat list: %w", err)
	}

	logger.Info("Concatenating clips with ffmpeg",
		zap.Int("clip_count", len(clipFiles)),
		zap.String("concat_list", concatListPath),
	)

	// ffmpeg -f concat -safe 0 -i concat_list.txt -c copy output.mp4
	// -c copy: Stream copy (no re-encoding, very fast!)
	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-f", "concat",
		"-safe", "0",
		"-i", concatListPath,
		"-c", "copy", // Stream copy - no re-encoding!
		"-y", // Overwrite output
		outputFile,
	)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		logger.Error("ffmpeg concatenation failed",
			zap.Error(err),
			zap.String("stderr", stderr.String()),
		)
		return fmt.Errorf("ffmpeg concat failed: %w (stderr: %s)", err, stderr.String())
	}

	logger.Info("Clips concatenated successfully")
	return nil
}

// addAudioToVideo mixes audio track with video using ffmpeg
func addAudioToVideo(ctx context.Context, videoPath, audioPath, outputPath string) error {
	logger.Info("Mixing audio with video",
		zap.String("video", videoPath),
		zap.String("audio", audioPath),
	)

	// ffmpeg -i video.mp4 -i music.mp3 -c:v copy -c:a aac -shortest -y output.mp4
	// -shortest: End when the shortest input ends (video or audio)
	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-i", videoPath,
		"-i", audioPath,
		"-c:v", "copy", // Copy video codec (no re-encoding)
		"-c:a", "aac", // Encode audio to AAC
		"-b:a", "192k", // Audio bitrate
		"-shortest", // Match video duration
		"-y",        // Overwrite output
		outputPath,
	)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		logger.Error("ffmpeg audio mixing failed",
			zap.Error(err),
			zap.String("stderr", stderr.String()),
		)
		return fmt.Errorf("ffmpeg audio mixing failed: %w (stderr: %s)", err, stderr.String())
	}

	logger.Info("Audio mixed successfully")
	return nil
}

// copyFile copies a file from src to dst (used for single-clip videos)
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// downloadFromS3 downloads a file from S3 to local disk
func downloadFromS3(ctx context.Context, s3Key, localPath string) error {
	result, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(assetsBucket),
		Key:    aws.String(s3Key),
	})
	if err != nil {
		return fmt.Errorf("S3 GetObject failed: %w", err)
	}
	defer result.Body.Close()

	file, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("failed to create local file: %w", err)
	}
	defer file.Close()

	if _, err := io.Copy(file, result.Body); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// uploadToS3 uploads a local file to S3
func uploadToS3(ctx context.Context, s3Key, localPath string) error {
	file, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	_, err = s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(assetsBucket),
		Key:         aws.String(s3Key),
		Body:        file,
		ContentType: aws.String("video/mp4"),
	})
	if err != nil {
		return fmt.Errorf("S3 PutObject failed: %w", err)
	}

	return nil
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

// updateJobComplete updates the job status to completed with video URL
func updateJobComplete(ctx context.Context, jobID, videoURL string) error {
	_, err := dynamoClient.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: &jobTable,
		Key: map[string]types.AttributeValue{
			"job_id": &types.AttributeValueMemberS{Value: jobID},
		},
		UpdateExpression: aws.String("SET #status = :status, video_key = :video_key, current_stage = :stage"),
		ExpressionAttributeNames: map[string]string{
			"#status": "status",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":status":    &types.AttributeValueMemberS{Value: "completed"},
			":video_key": &types.AttributeValueMemberS{Value: videoURL},
			":stage":     &types.AttributeValueMemberS{Value: "Completed"},
		},
	})
	return err
}

func main() {
	lambda.Start(handler)
}
