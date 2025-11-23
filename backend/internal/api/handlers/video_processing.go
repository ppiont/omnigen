package handlers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/omnigen/backend/internal/repository"
	"go.uber.org/zap"
)

// processVideoCommon is a shared function for downloading, processing, and uploading video clips
func processVideoCommon(
	ctx context.Context,
	s3Service *repository.S3AssetRepository,
	assetsBucket string,
	logger *zap.Logger,
	userID string,
	jobID string,
	clipNumber int,
	videoURL string,
) (string, string, error) {
	// Create temp directory
	tmpDir := filepath.Join("/tmp", jobID, fmt.Sprintf("clip-%d", clipNumber))
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		return "", "", fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Download video from Replicate URL
	logger.Info("Downloading video from Replicate",
		zap.String("job_id", jobID),
		zap.Int("clip", clipNumber),
		zap.String("url", videoURL),
	)
	videoPath := filepath.Join(tmpDir, "video.mp4")
	if err := downloadFileCommon(ctx, videoURL, videoPath); err != nil {
		return "", "", fmt.Errorf("failed to download video: %w", err)
	}

	// Extract last frame using ffmpeg
	logger.Info("Extracting last frame with ffmpeg",
		zap.String("job_id", jobID),
		zap.Int("clip", clipNumber),
	)
	lastFramePath := filepath.Join(tmpDir, "last_frame.jpg")
	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-sseof", "-1",
		"-i", videoPath,
		"-update", "1",
		"-q:v", "2",
		"-y", lastFramePath,
	)
	if err := cmd.Run(); err != nil {
		logger.Warn("Failed to extract last frame, continuing without it",
			zap.String("job_id", jobID),
			zap.Int("clip", clipNumber),
			zap.Error(err),
		)
		lastFramePath = "" // Continue without last frame
	}

	// Upload video to S3
	logger.Info("Uploading video to S3",
		zap.String("job_id", jobID),
		zap.Int("clip", clipNumber),
	)
	videoS3Key := buildSceneClipKey(userID, jobID, clipNumber)
	videoS3URL, err := s3Service.UploadFile(ctx, assetsBucket, videoS3Key, videoPath, "video/mp4")
	if err != nil {
		return "", "", fmt.Errorf("failed to upload video to S3: %w", err)
	}

	// Upload last frame to S3 (if extracted)
	var lastFrameS3URL string
	if lastFramePath != "" {
		lastFrameS3Key := buildSceneThumbnailKey(userID, jobID, clipNumber)
		_, err = s3Service.UploadFile(ctx, assetsBucket, lastFrameS3Key, lastFramePath, "image/jpeg")
		if err != nil {
			logger.Warn("Failed to upload last frame, continuing",
				zap.String("job_id", jobID),
				zap.Error(err),
			)
			lastFrameS3URL = "" // Continue without last frame URL
		} else {
			// Generate presigned URL for Veo API access (valid for 1 hour)
			lastFrameS3URL, err = s3Service.GetPresignedURL(ctx, lastFrameS3Key, 1*time.Hour)
			if err != nil {
				logger.Warn("Failed to generate presigned URL for last frame, continuing",
					zap.String("job_id", jobID),
					zap.Error(err),
				)
				lastFrameS3URL = "" // Continue without last frame URL
			}
		}
	}

	logger.Info("Video processed and uploaded",
		zap.String("job_id", jobID),
		zap.Int("clip", clipNumber),
		zap.String("s3_url", videoS3URL),
	)

	return videoS3URL, lastFrameS3URL, nil
}

// downloadFileCommon downloads a file from URL to local path
func downloadFileCommon(ctx context.Context, url string, destPath string) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// composeVideoCommon is a shared function for composing final video from clips
// Returns: (mp4Key, webmKey, error) - webmKey may be empty if WebM encoding fails
func composeVideoCommon(
	ctx context.Context,
	s3Service *repository.S3AssetRepository,
	assetsBucket string,
	logger *zap.Logger,
	userID string,
	jobID string,
	clips []ClipVideo,
	sideEffectsText string,
	sideEffectsStartTime float64,
) (string, string, error) {
	trimmedText := strings.TrimSpace(sideEffectsText)
	var totalDuration float64
	for _, clip := range clips {
		totalDuration += clip.Duration
	}

	logger.Info("Composing final video",
		zap.String("job_id", jobID),
		zap.Int("num_clips", len(clips)),
		zap.Bool("has_side_effects", trimmedText != ""),
		zap.Float64("side_effects_start_time", sideEffectsStartTime),
		zap.Float64("video_duration_estimate", totalDuration),
	)

	if sideEffectsStartTime > 0 && trimmedText == "" {
		return "", "", fmt.Errorf("side effects text is required when sideEffectsStartTime is provided")
	}

	tmpDir := filepath.Join("/tmp", jobID, "composition")
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		return "", "", fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Download all clips from S3
	logger.Info("Downloading clips from S3 for composition",
		zap.String("job_id", jobID),
		zap.Int("num_clips", len(clips)),
	)
	var clipPaths []string
	for i, clip := range clips {
		clipPath := filepath.Join(tmpDir, fmt.Sprintf("clip-%d.mp4", i+1))
		if err := s3Service.DownloadFile(ctx, assetsBucket, extractS3Key(clip.VideoURL), clipPath); err != nil {
			return "", "", fmt.Errorf("failed to download clip %d: %w", i+1, err)
		}
		clipPaths = append(clipPaths, clipPath)
	}

	// Create concat file for ffmpeg
	concatFile := filepath.Join(tmpDir, "concat.txt")
	f, err := os.Create(concatFile)
	if err != nil {
		return "", "", fmt.Errorf("failed to create concat file: %w", err)
	}
	for _, path := range clipPaths {
		fmt.Fprintf(f, "file '%s'\n", path)
	}
	if err := f.Close(); err != nil {
		return "", "", fmt.Errorf("failed to close concat file: %w", err)
	}

	// Concatenate clips (video track only)
	logger.Info("Concatenating video clips (video track only)",
		zap.String("job_id", jobID),
		zap.Int("num_clips", len(clipPaths)),
	)
	finalVideo := filepath.Join(tmpDir, "final.mp4")
	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-f", "concat",
		"-safe", "0",
		"-i", concatFile,
		"-c:v", "copy",
		"-an", // Explicitly drop audio streams
		"-y", finalVideo,
	)
	if output, err := cmd.CombinedOutput(); err != nil {
		logger.Error("ffmpeg concat failed",
			zap.String("job_id", jobID),
			zap.String("output", string(output)),
			zap.Error(err),
		)
		return "", "", fmt.Errorf("ffmpeg concat failed: %w", err)
	}

	if trimmedText != "" && totalDuration > 0 {
		videoWidth, videoHeight, err := probeVideoDimensions(finalVideo)
		if err != nil {
			logger.Warn("Failed to probe video dimensions, using defaults",
				zap.String("job_id", jobID),
				zap.Error(err),
			)
			videoWidth = 1920
			videoHeight = 1080
		}

		config, err := buildDrawtextConfig(logger, trimmedText, sideEffectsStartTime, totalDuration, videoWidth, videoHeight)
		if err != nil {
			return "", "", err
		}
		if config != nil {
			logger.Info("Applying side effects text overlay",
				zap.String("job_id", jobID),
				zap.Float64("overlay_start", config.OverlayStart),
				zap.Float64("overlay_end", config.OverlayEnd),
			)

			videoWithText := filepath.Join(tmpDir, "video_with_text.mp4")
			cmd = exec.CommandContext(ctx, "ffmpeg",
				"-i", finalVideo,
				"-vf", config.Filter,
				"-c:v", "libx264",
				"-preset", "medium",
				"-crf", "21",
				"-y", videoWithText,
			)
			if output, err := cmd.CombinedOutput(); err != nil {
				logger.Error("ffmpeg text overlay failed",
					zap.String("job_id", jobID),
					zap.String("output", string(output)),
					zap.Error(err),
				)
				return "", "", fmt.Errorf("ffmpeg text overlay failed: %w", err)
			}

			logger.Info("Text overlay applied successfully", zap.String("job_id", jobID))
			finalVideo = videoWithText
		}
	}

	// Upload final MP4 video to S3
	logger.Info("Uploading final MP4 video to S3",
		zap.String("job_id", jobID),
	)
	mp4S3Key := buildFinalVideoKey(userID, jobID)
	_, err = s3Service.UploadFile(ctx, assetsBucket, mp4S3Key, finalVideo, "video/mp4")
	if err != nil {
		return "", "", fmt.Errorf("failed to upload MP4 video: %w", err)
	}

	// Transcode to WebM (VP9) for web-optimized delivery
	logger.Info("Transcoding to WebM format",
		zap.String("job_id", jobID),
	)
	webmVideo := filepath.Join(tmpDir, "final.webm")
	cmd = exec.CommandContext(ctx, "ffmpeg",
		"-i", finalVideo,
		"-c:v", "libvpx-vp9",
		"-crf", "30",
		"-b:v", "0",
		"-row-mt", "1",
		"-an", // No audio in video file
		"-y", webmVideo,
	)

	var webmS3Key string
	if output, err := cmd.CombinedOutput(); err != nil {
		logger.Warn("WebM transcode failed, MP4 still available",
			zap.String("job_id", jobID),
			zap.String("output", string(output)),
			zap.Error(err),
		)
		// Don't fail - MP4 is still available
	} else {
		// Upload WebM to S3
		webmS3Key = buildFinalWebMKey(userID, jobID)
		_, err = s3Service.UploadFile(ctx, assetsBucket, webmS3Key, webmVideo, "video/webm")
		if err != nil {
			logger.Warn("Failed to upload WebM, MP4 still available",
				zap.String("job_id", jobID),
				zap.Error(err),
			)
			webmS3Key = "" // Clear key since upload failed
		} else {
			logger.Info("WebM uploaded successfully",
				zap.String("job_id", jobID),
				zap.String("webm_key", webmS3Key),
			)
		}
	}

	logger.Info("Video composition complete",
		zap.String("job_id", jobID),
		zap.String("mp4_key", mp4S3Key),
		zap.String("webm_key", webmS3Key),
	)

	return mp4S3Key, webmS3Key, nil
}
