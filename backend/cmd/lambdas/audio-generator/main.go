package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"go.uber.org/zap"

	"github.com/omnigen/backend/internal/domain"
)

// AudioGeneratorInput represents input from Step Functions
// Matches workflow.asl.json Parameters structure
type AudioGeneratorInput struct {
	JobID         string           `json:"job_id"`
	TotalDuration int              `json:"total_duration"` // Total video duration in seconds
	AudioSpec     domain.AudioSpec `json:"audio_spec"`
}

// AudioGeneratorOutput represents the output with audio URLs
type AudioGeneratorOutput struct {
	JobID      string     `json:"job_id"`
	AudioFiles AudioFiles `json:"audio_files"`
	Status     string     `json:"status"`
}

// AudioFiles contains S3 URLs for audio tracks
type AudioFiles struct {
	Music     string `json:"music,omitempty"`
	Voiceover string `json:"voiceover,omitempty"`
}

var (
	dynamoClient *dynamodb.Client
	jobTable     string
	assetsBucket string
	logger       *zap.Logger
)

// Music library mapping: mood + style -> S3 key
// In production, these would be pre-uploaded royalty-free music tracks
var musicLibrary = map[string]string{
	// Upbeat tracks
	"upbeat-electronic": "music-library/upbeat-electronic-1.mp3",
	"upbeat-acoustic":   "music-library/upbeat-acoustic-1.mp3",
	"upbeat-orchestral": "music-library/upbeat-orchestral-1.mp3",
	"upbeat-pop":        "music-library/upbeat-pop-1.mp3",
	"upbeat-rock":       "music-library/upbeat-rock-1.mp3",

	// Calm tracks
	"calm-electronic": "music-library/calm-electronic-1.mp3",
	"calm-acoustic":   "music-library/calm-acoustic-1.mp3",
	"calm-orchestral": "music-library/calm-orchestral-1.mp3",
	"calm-ambient":    "music-library/calm-ambient-1.mp3",
	"calm-piano":      "music-library/calm-piano-1.mp3",

	// Dramatic tracks
	"dramatic-electronic": "music-library/dramatic-electronic-1.mp3",
	"dramatic-orchestral": "music-library/dramatic-orchestral-1.mp3",
	"dramatic-cinematic":  "music-library/dramatic-cinematic-1.mp3",
	"dramatic-epic":       "music-library/dramatic-epic-1.mp3",

	// Energetic tracks
	"energetic-electronic": "music-library/energetic-electronic-1.mp3",
	"energetic-rock":       "music-library/energetic-rock-1.mp3",
	"energetic-hip-hop":    "music-library/energetic-hip-hop-1.mp3",

	// Corporate/Professional tracks
	"professional-electronic": "music-library/professional-electronic-1.mp3",
	"professional-acoustic":   "music-library/professional-acoustic-1.mp3",
	"professional-corporate":  "music-library/professional-corporate-1.mp3",

	// Default fallback
	"default": "music-library/default-background-1.mp3",
}

func init() {
	// Initialize logger
	var err error
	logger, err = zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

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

	logger.Info("AudioGenerator Lambda initialized (music library mode)",
		zap.String("job_table", jobTable),
		zap.String("assets_bucket", assetsBucket),
		zap.Int("library_size", len(musicLibrary)),
	)
}

func handler(ctx context.Context, input AudioGeneratorInput) (AudioGeneratorOutput, error) {
	log.Printf("AudioGenerator Lambda invoked for job %s", input.JobID)

	// Update job progress
	updateStageProgress(ctx, input.JobID, "Selecting background music from library")

	// Select background music from library
	musicURL := ""
	if input.AudioSpec.EnableAudio {
		var err error
		musicURL, err = selectMusicFromLibrary(input.AudioSpec, input.TotalDuration)
		if err != nil {
			log.Printf("Failed to select music: %v", err)
			return AudioGeneratorOutput{
				JobID:  input.JobID,
				Status: "failed",
			}, err
		}
	}

	// Voiceover is optional and not implemented for MVP
	// In production, this would call a TTS API like ElevenLabs or Bark
	var voiceoverURL string
	if input.AudioSpec.VoiceoverText != "" {
		log.Printf("Voiceover requested but not implemented for MVP: %s", input.AudioSpec.VoiceoverText)
		// voiceoverURL would be set here in production
	}

	output := AudioGeneratorOutput{
		JobID: input.JobID,
		AudioFiles: AudioFiles{
			Music:     musicURL,
			Voiceover: voiceoverURL,
		},
		Status: "completed",
	}

	log.Printf("Successfully selected audio for job %s: music=%s", input.JobID, musicURL)
	return output, nil
}

// selectMusicFromLibrary selects a music track from the pre-uploaded library
// based on mood and style preferences
func selectMusicFromLibrary(audioSpec domain.AudioSpec, duration int) (string, error) {
	logger.Info("Selecting music from library",
		zap.String("mood", audioSpec.MusicMood),
		zap.String("style", audioSpec.MusicStyle),
		zap.Int("duration", duration),
	)

	// Normalize mood and style to lowercase
	mood := strings.ToLower(strings.TrimSpace(audioSpec.MusicMood))
	style := strings.ToLower(strings.TrimSpace(audioSpec.MusicStyle))

	// Build lookup key
	lookupKey := fmt.Sprintf("%s-%s", mood, style)

	// Try exact match first
	s3Key, found := musicLibrary[lookupKey]
	if !found {
		// Try mood-only fallback
		for key := range musicLibrary {
			if strings.HasPrefix(key, mood+"-") {
				s3Key = musicLibrary[key]
				found = true
				logger.Info("Using mood fallback", zap.String("key", key))
				break
			}
		}
	}

	// Use default if no match found
	if !found {
		s3Key = musicLibrary["default"]
		logger.Warn("No matching music found, using default",
			zap.String("requested_mood", mood),
			zap.String("requested_style", style),
		)
	}

	// Return S3 URL
	// Note: In production, we would check if the file exists in S3
	// For MVP, we assume all library files are pre-uploaded
	s3URL := fmt.Sprintf("s3://%s/%s", assetsBucket, s3Key)

	logger.Info("Music selected from library",
		zap.String("lookup_key", lookupKey),
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
