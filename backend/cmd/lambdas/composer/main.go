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

	"github.com/omnigen/backend/internal/domain"
)

// ComposerInput represents input from Step Functions
// Matches workflow output structure
type ComposerInput struct {
	JobID       string         `json:"job_id"`
	Scenes      []domain.Scene `json:"scenes"`       // Scene definitions with transitions
	SceneVideos []SceneVideo   `json:"scene_videos"` // Generated scene videos from Generator Lambda
	AudioFiles  AudioFiles     `json:"audio_files"`  // Audio from Audio Generator Lambda
}

// SceneVideo represents a single scene's generated video
type SceneVideo struct {
	SceneNumber  int     `json:"scene_number"`
	VideoURL     string  `json:"video_url"`      // S3 URL
	LastFrameURL string  `json:"last_frame_url"` // For coherence (not used in composition)
	Duration     float64 `json:"duration"`
	Status       string  `json:"status"`
}

// AudioFiles contains audio track S3 URLs
type AudioFiles struct {
	Music     string `json:"music,omitempty"`
	Voiceover string `json:"voiceover,omitempty"`
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

const (
	// Default transition duration in seconds
	defaultTransitionDuration = 0.5
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
	log.Printf("Composer Lambda invoked for job %s with %d scenes", input.JobID, len(input.SceneVideos))

	// Update job progress
	updateStageProgress(ctx, input.JobID, "Composing final video with transitions and audio")

	// Compose video using ffmpeg with crossfade transitions and audio
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

// composeVideo stitches scenes together with crossfade transitions and audio using ffmpeg
func composeVideo(ctx context.Context, input ComposerInput) (string, error) {
	logger.Info("Starting video composition with transitions and audio",
		zap.String("job_id", input.JobID),
		zap.Int("scene_count", len(input.SceneVideos)),
	)

	// Create working directory for this job
	jobDir := filepath.Join(tmpDir, input.JobID)
	if err := os.MkdirAll(jobDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create job directory: %w", err)
	}
	defer os.RemoveAll(jobDir) // Clean up after processing

	// Step 1: Download all scene videos from S3
	logger.Info("Downloading scene videos from S3")
	sceneFiles := make([]string, len(input.SceneVideos))
	for i, scene := range input.SceneVideos {
		// Extract key from S3 URL (format: s3://bucket/key)
		s3Key := strings.TrimPrefix(scene.VideoURL, fmt.Sprintf("s3://%s/", assetsBucket))

		localPath := filepath.Join(jobDir, fmt.Sprintf("scene-%d.mp4", scene.SceneNumber))
		if err := downloadFromS3(ctx, s3Key, localPath); err != nil {
			return "", fmt.Errorf("failed to download scene %d: %w", scene.SceneNumber, err)
		}
		sceneFiles[i] = localPath
		logger.Debug("Downloaded scene", zap.Int("scene", scene.SceneNumber), zap.String("path", localPath))
	}

	// Step 2: Apply crossfade transitions between scenes
	logger.Info("Applying crossfade transitions between scenes")
	transitionedVideo := filepath.Join(jobDir, "with_transitions.mp4")
	if err := applyTransitions(ctx, sceneFiles, input.Scenes, transitionedVideo); err != nil {
		return "", fmt.Errorf("failed to apply transitions: %w", err)
	}

	// Step 3: Add audio track if provided
	finalVideo := transitionedVideo
	if input.AudioFiles.Music != "" {
		logger.Info("Adding audio track to video")

		// Download music file from S3
		musicKey := strings.TrimPrefix(input.AudioFiles.Music, fmt.Sprintf("s3://%s/", assetsBucket))
		musicPath := filepath.Join(jobDir, "music.mp3")
		if err := downloadFromS3(ctx, musicKey, musicPath); err != nil {
			logger.Warn("Failed to download music, proceeding without audio",
				zap.Error(err),
				zap.String("music_url", input.AudioFiles.Music),
			)
		} else {
			// Mix audio with video
			videoWithAudio := filepath.Join(jobDir, "final_with_audio.mp4")
			if err := addAudioToVideo(ctx, transitionedVideo, musicPath, videoWithAudio); err != nil {
				logger.Warn("Failed to add audio, using video without audio",
					zap.Error(err),
				)
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

// applyTransitions applies crossfade transitions between scenes using ffmpeg xfade filter
func applyTransitions(ctx context.Context, sceneFiles []string, scenes []domain.Scene, outputFile string) error {
	if len(sceneFiles) == 0 {
		return fmt.Errorf("no scene files to compose")
	}

	// If only one scene, no transitions needed
	if len(sceneFiles) == 1 {
		logger.Info("Single scene, copying without transitions")
		return copyFile(sceneFiles[0], outputFile)
	}

	// Build ffmpeg filter_complex for crossfade transitions
	// Strategy: Chain xfade filters for each transition
	//
	// Example for 3 videos with crossfade:
	// -filter_complex "[0:v][1:v]xfade=transition=fade:duration=0.5:offset=4.5[v01];
	//                  [v01][2:v]xfade=transition=fade:duration=0.5:offset=9.0[vout]"
	// -map "[vout]"

	var filterParts []string
	var currentOffset float64 = 0

	// Calculate offsets for each transition
	// offset = sum of previous video durations - transition_duration
	for i := 0; i < len(sceneFiles)-1; i++ {
		// Determine transition type
		transitionType := getFFmpegTransition(scenes[i+1].TransitionIn)

		// Calculate offset (when the transition should start)
		// This is at the end of the current video minus the transition duration
		if i == 0 {
			currentOffset = scenes[i].Duration - defaultTransitionDuration
		} else {
			currentOffset += scenes[i].Duration - defaultTransitionDuration
		}

		var inputLeft, inputRight string
		if i == 0 {
			inputLeft = "[0:v]"
			inputRight = "[1:v]"
		} else {
			inputLeft = fmt.Sprintf("[v%d]", i-1)
			inputRight = fmt.Sprintf("[%d:v]", i+1)
		}

		outputLabel := fmt.Sprintf("[v%d]", i)
		if i == len(sceneFiles)-2 {
			// Last transition outputs to final label
			outputLabel = "[vout]"
		}

		filterPart := fmt.Sprintf("%s%sxfade=transition=%s:duration=%.2f:offset=%.2f%s",
			inputLeft, inputRight, transitionType, defaultTransitionDuration, currentOffset, outputLabel)
		filterParts = append(filterParts, filterPart)
	}

	filterComplex := strings.Join(filterParts, ";")

	// Build ffmpeg command
	args := []string{}
	for _, sceneFile := range sceneFiles {
		args = append(args, "-i", sceneFile)
	}
	args = append(args,
		"-filter_complex", filterComplex,
		"-map", "[vout]",
		"-c:v", "libx264",
		"-preset", "fast",
		"-crf", "23",
		"-y", // Overwrite output
		outputFile,
	)

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	logger.Info("Running ffmpeg with transitions",
		zap.String("filter_complex", filterComplex),
	)

	if err := cmd.Run(); err != nil {
		logger.Error("ffmpeg transitions failed",
			zap.Error(err),
			zap.String("stderr", stderr.String()),
		)
		return fmt.Errorf("ffmpeg xfade failed: %w (stderr: %s)", err, stderr.String())
	}

	return nil
}

// getFFmpegTransition maps domain.Transition to ffmpeg xfade transition type
func getFFmpegTransition(transition domain.Transition) string {
	switch transition {
	case domain.TransitionFade:
		return "fade"
	case domain.TransitionCrossFade:
		return "fade" // ffmpeg xfade uses "fade" for crossfade
	case domain.TransitionWipeLeft:
		return "wipeleft"
	case domain.TransitionWipeRight:
		return "wiperight"
	case domain.TransitionZoom:
		return "fadegrays" // Closest approximation
	case domain.TransitionCut, domain.TransitionNone:
		return "fade" // Default to fade for cut/none
	default:
		return "fade" // Default fallback
	}
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

// copyFile copies a file from src to dst (used for single-scene videos)
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
		UpdateExpression: aws.String("SET #status = :status, video_url = :video_url, current_stage = :stage"),
		ExpressionAttributeNames: map[string]string{
			"#status": "status",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":status":    &types.AttributeValueMemberS{Value: "completed"},
			":video_url": &types.AttributeValueMemberS{Value: videoURL},
			":stage":     &types.AttributeValueMemberS{Value: "Completed"},
		},
	})
	return err
}

func main() {
	lambda.Start(handler)
}
