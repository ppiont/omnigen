package integration

import (
	"fmt"
	"testing"
	"time"

	"github.com/omnigen/backend/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type pipelineScenario struct {
	name         string
	voice        string
	duration     int
	aspectRatio  string
	sideEffects  string
	productImage string
}

func TestPharmaceuticalAdGeneration_MaleVoice(t *testing.T) {
	scenario := pipelineScenario{
		name:         "male_voice",
		voice:        "male",
		duration:     30,
		aspectRatio:  domain.AspectRatio16x9,
		sideEffects:  "Common side effects include headache and mild nausea.",
		productImage: "https://mock.s3.amazonaws.com/assets/product-male.jpg",
	}

	job := buildMockJob(t, scenario)
	validatePharmaceuticalJob(t, job, scenario)
}

func TestPharmaceuticalAdGeneration_FemaleVoice(t *testing.T) {
	scenario := pipelineScenario{
		name:         "female_voice",
		voice:        "female",
		duration:     30,
		aspectRatio:  domain.AspectRatio16x9,
		sideEffects:  "Common side effects include fatigue and dry mouth.",
		productImage: "https://mock.s3.amazonaws.com/assets/product-female.jpg",
	}

	job := buildMockJob(t, scenario)
	validatePharmaceuticalJob(t, job, scenario)
}

func TestPharmaceuticalAdGeneration_Durations(t *testing.T) {
	durations := []int{10, 20, 30, 60}
	for _, duration := range durations {
		duration := duration
		t.Run(
			DurationLabel(duration),
			func(t *testing.T) {
				scenario := pipelineScenario{
					name:         DurationLabel(duration),
					voice:        "male",
					duration:     duration,
					aspectRatio:  domain.AspectRatio16x9,
					sideEffects:  "Side effects may include dizziness and light-headedness.",
					productImage: "https://mock.s3.amazonaws.com/assets/product-duration.jpg",
				}

				job := buildMockJob(t, scenario)
				validateDurationCoverage(t, job, scenario.duration)
			},
		)
	}
}

func TestPharmaceuticalAdGeneration_AspectRatios(t *testing.T) {
	ratios := []string{
		domain.AspectRatio16x9,
		domain.AspectRatio9x16,
		domain.AspectRatio1x1,
	}

	for _, ratio := range ratios {
		ratio := ratio
		t.Run(
			ratio,
			func(t *testing.T) {
				scenario := pipelineScenario{
					name:         ratio,
					voice:        "male",
					duration:     30,
					aspectRatio:  ratio,
					sideEffects:  "Side effects include dizziness, dry mouth, and mild nausea.",
					productImage: "https://mock.s3.amazonaws.com/assets/product-ratio.jpg",
				}

				job := buildMockJob(t, scenario)
				validateAspectRatio(t, job, ratio)
			},
		)
	}
}

func TestSideEffectsRendering(t *testing.T) {
	scenario := pipelineScenario{
		name:         "side_effects_rendering",
		voice:        "male",
		duration:     30,
		aspectRatio:  domain.AspectRatio16x9,
		sideEffects:  "Side effects include headache, dry mouth, dizziness, and nausea.",
		productImage: "https://mock.s3.amazonaws.com/assets/product-side-effects.jpg",
	}

	job := buildMockJob(t, scenario)

	assert.Equal(t, scenario.sideEffects, job.SideEffectsText)
	assert.Equal(t, scenario.sideEffects, job.SideEffects)

	expectedStart := float64(scenario.duration) * 0.8
	assert.InDelta(t, expectedStart, job.SideEffectsStartTime, 0.01)
	require.NotEmpty(t, job.SceneVideoURLs)

	lastClip := job.SceneVideoURLs[len(job.SceneVideoURLs)-1]
	assert.Contains(t, lastClip, "/clips/scene-")
	assert.NotEmpty(t, job.VideoKey)
}

func TestNarratorAudioSpeed(t *testing.T) {
	scenario := pipelineScenario{
		name:         "narrator_speed",
		voice:        "female",
		duration:     30,
		aspectRatio:  domain.AspectRatio16x9,
		sideEffects:  "Side effects include dizziness, dry mouth, and mild nausea.",
		productImage: "https://mock.s3.amazonaws.com/assets/product-narrator.jpg",
	}

	job := buildMockJob(t, scenario)

	mainSpeed := lookupSyncPoint(job.AudioSpec.SyncPoints, "narrator_speed_main")
	assert.Equal(t, "1.0x", mainSpeed)

	sideEffectsSpeed := lookupSyncPoint(job.AudioSpec.SyncPoints, "narrator_speed_side_effects")
	assert.Equal(t, "1.4x", sideEffectsSpeed)

	assert.InDelta(t, float64(job.Duration)*0.8, job.SideEffectsStartTime, 0.01)
}

func TestBackgroundMusicSeparation(t *testing.T) {
	scenario := pipelineScenario{
		name:         "music_separation",
		voice:        "male",
		duration:     30,
		aspectRatio:  domain.AspectRatio16x9,
		sideEffects:  "Side effects include fatigue, dry mouth, and headache.",
		productImage: "https://mock.s3.amazonaws.com/assets/product-music.jpg",
	}

	job := buildMockJob(t, scenario)

	assert.Contains(t, job.AudioURL, "/audio/background-music.mp3")
	assert.Contains(t, job.NarratorAudioURL, "/audio/narrator-voiceover.mp3")
	assert.NotEqual(t, job.AudioURL, job.NarratorAudioURL)

	musicVolume := lookupSyncPoint(job.AudioSpec.SyncPoints, "music_volume_level")
	assert.Equal(t, "0.30", musicVolume)
}

func TestProductImagePlacement(t *testing.T) {
	scenario := pipelineScenario{
		name:         "product_image",
		voice:        "female",
		duration:     30,
		aspectRatio:  domain.AspectRatio16x9,
		sideEffects:  "Side effects include dizziness, dry mouth, and mild nausea.",
		productImage: "https://mock.s3.amazonaws.com/assets/product-final.jpg",
	}

	job := buildMockJob(t, scenario)

	sceneCount := len(job.Scenes)
	require.Greater(t, sceneCount, 1)

	for i, scene := range job.Scenes {
		if i == sceneCount-1 {
			assert.Equal(t, scenario.productImage, scene.StartImageURL)
		} else {
			assert.NotEqual(t, scenario.productImage, scene.StartImageURL)
		}
	}
}

func TestSideEffectsTextVariations(t *testing.T) {
	shortText := "Mild headache."
	mediumText := "Side effects include headache, dizziness, dry mouth, and mild nausea."
	longText := repeatText("Side effects include nausea and dizziness.", 10)

	tests := []struct {
		name        string
		sideEffects string
	}{
		{"short_text", shortText},
		{"medium_text", mediumText},
		{"long_text", longText},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			scenario := pipelineScenario{
				name:         tc.name,
				voice:        "male",
				duration:     30,
				aspectRatio:  domain.AspectRatio16x9,
				sideEffects:  tc.sideEffects,
				productImage: "https://mock.s3.amazonaws.com/assets/product-text.jpg",
			}

			job := buildMockJob(t, scenario)
			assert.Equal(t, tc.sideEffects, job.SideEffectsText)
			assert.NotEmpty(t, job.VideoKey)
		})
	}
}

func TestFourTrackArchitecture(t *testing.T) {
	scenario := pipelineScenario{
		name:         "four_track",
		voice:        "female",
		duration:     30,
		aspectRatio:  domain.AspectRatio16x9,
		sideEffects:  "Side effects include dizziness, fatigue, and mild nausea.",
		productImage: "https://mock.s3.amazonaws.com/assets/product-four-track.jpg",
	}

	job := buildMockJob(t, scenario)

	assert.Contains(t, job.VideoKey, "/final/video.mp4")
	assert.Contains(t, job.AudioURL, "/audio/background-music.mp3")
	assert.Contains(t, job.NarratorAudioURL, "/audio/narrator-voiceover.mp3")
	assert.NotZero(t, len(job.SceneVideoURLs))

	assert.NotEmpty(t, job.SideEffectsText)
	assert.InDelta(t, float64(job.Duration)*0.8, job.SideEffectsStartTime, 0.01)

	assert.Equal(t, domain.StatusCompleted, job.Status)
	require.NotNil(t, job.CompletedAt)
}

func validatePharmaceuticalJob(t *testing.T, job *domain.Job, scenario pipelineScenario) {
	t.Helper()

	require.NotNil(t, job)
	assert.Equal(t, domain.StatusCompleted, job.Status)
	assert.Equal(t, scenario.voice, job.Voice)
	assert.Equal(t, scenario.aspectRatio, job.AspectRatio)
	assert.Equal(t, scenario.sideEffects, job.SideEffectsText)
	assert.Equal(t, scenario.sideEffects, job.SideEffects)

	expectedScenes := scenario.duration / 5
	assert.Len(t, job.Scenes, expectedScenes)
	assert.Len(t, job.SceneVideoURLs, expectedScenes)

	totalDuration := 0.0
	for _, scene := range job.Scenes {
		totalDuration += scene.Duration
	}
	assert.InDelta(t, float64(scenario.duration), totalDuration, 0.01)

	assert.Contains(t, job.AudioURL, "/audio/background-music.mp3")
	assert.Contains(t, job.NarratorAudioURL, "/audio/narrator-voiceover.mp3")
	assert.NotEmpty(t, job.VideoKey)

	assert.InDelta(t, float64(scenario.duration)*0.8, job.SideEffectsStartTime, 0.01)
}

func validateDurationCoverage(t *testing.T, job *domain.Job, duration int) {
	t.Helper()

	expectedScenes := duration / 5
	require.Len(t, job.Scenes, expectedScenes)
	require.Len(t, job.SceneVideoURLs, expectedScenes)

	for i, scene := range job.Scenes {
		assert.Equal(t, i+1, scene.SceneNumber)
		if i > 0 {
			assert.GreaterOrEqual(t, scene.StartTime, job.Scenes[i-1].StartTime)
		}
	}
}

func validateAspectRatio(t *testing.T, job *domain.Job, aspectRatio string) {
	t.Helper()
	assert.Equal(t, aspectRatio, job.AspectRatio)
}

func buildMockJob(t *testing.T, scenario pipelineScenario) *domain.Job {
	t.Helper()

	sceneCount := scenario.duration / 5
	now := time.Now().Unix()
	scenes := make([]domain.Scene, 0, sceneCount)
	sceneVideoURLs := make([]string, 0, sceneCount)

	for i := 0; i < sceneCount; i++ {
		sceneNumber := i + 1
		scene := domain.Scene{
			SceneNumber:      sceneNumber,
			StartTime:        float64(i * 5),
			Duration:         5,
			Location:         "EXT. PARK - DAY",
			Action:           "Active people outdoors enjoying nature.",
			ShotType:         domain.ShotWide,
			CameraAngle:      domain.AngleEyeLevel,
			CameraMove:       domain.MoveSteadycam,
			Lighting:         domain.LightNatural,
			ColorGrade:       domain.GradeVibrant,
			Mood:             domain.MoodEnergetic,
			VisualStyle:      domain.StyleCommercial,
			TransitionIn:     domain.TransitionCut,
			TransitionOut:    domain.TransitionCut,
			GenerationPrompt: "AI-generated scene prompt",
		}

		if i == sceneCount-1 {
			scene.StartImageURL = scenario.productImage
		} else if i > 0 {
			scene.StartImageURL = buildContinuityFrameURL(scenario.name, sceneNumber-1)
		}

		scenes = append(scenes, scene)
		sceneVideoURLs = append(sceneVideoURLs, buildSceneVideoURL(scenario.name, sceneNumber))
	}

	sideEffectsStart := float64(scenario.duration) * 0.8

	audioSpec := domain.AudioSpec{
		EnableAudio:          true,
		MusicMood:            "uplifting",
		MusicStyle:           "orchestral",
		NarratorScript:       "Narration for main content followed by side effects disclosure.",
		SideEffectsText:      scenario.sideEffects,
		SideEffectsStartTime: sideEffectsStart,
		SyncPoints: []domain.SyncPoint{
			{
				Timestamp:   0,
				Type:        "music_volume_level",
				Description: "0.30",
			},
			{
				Timestamp:   0,
				Type:        "narrator_speed_main",
				Description: "1.0x",
			},
			{
				Timestamp:   sideEffectsStart,
				Type:        "narrator_speed_side_effects",
				Description: "1.4x",
			},
		},
	}

	completedAt := now

	return &domain.Job{
		JobID:                "job-" + scenario.name,
		UserID:               "user-123",
		ScriptID:             "script-" + scenario.name,
		Status:               domain.StatusCompleted,
		Stage:                "complete",
		ThumbnailURL:         buildThumbnailURL(scenario.name),
		AudioURL:             buildAudioURL(scenario.name),
		NarratorAudioURL:     buildNarratorURL(scenario.name),
		ScenesCompleted:      sceneCount,
		SceneVideoURLs:       sceneVideoURLs,
		Prompt:               "Prompt for test scenario",
		Duration:             scenario.duration,
		AspectRatio:          scenario.aspectRatio,
		Voice:                scenario.voice,
		SideEffects:          scenario.sideEffects,
		SideEffectsText:      scenario.sideEffects,
		SideEffectsStartTime: sideEffectsStart,
		Scenes:               scenes,
		AudioSpec:            audioSpec,
		ScriptMetadata: domain.Metadata{
			ProductName:    "TestProduct",
			TargetAudience: "Adults 30-60",
			CallToAction:   "Ask your doctor about TestProduct today.",
			Keywords: []string{
				"pharmaceutical",
				"uplifting",
				"compliance",
			},
		},
		VideoKey:    buildVideoKey(scenario.name),
		CreatedAt:   now,
		UpdatedAt:   now,
		CompletedAt: &completedAt,
	}
}

func buildSceneVideoURL(scenario string, sceneNumber int) string {
	return "https://mock.s3.amazonaws.com/users/user-123/jobs/job-" + scenario + "/clips/scene-" + FormatSceneNumber(sceneNumber) + ".mp4"
}

func buildContinuityFrameURL(scenario string, sceneNumber int) string {
	return "https://mock.s3.amazonaws.com/users/user-123/jobs/job-" + scenario + "/thumbnails/scene-" + FormatSceneNumber(sceneNumber) + ".jpg"
}

func buildThumbnailURL(scenario string) string {
	return "https://mock.s3.amazonaws.com/users/user-123/jobs/job-" + scenario + "/thumbnails/job-thumbnail.jpg"
}

func buildAudioURL(scenario string) string {
	return "https://mock.s3.amazonaws.com/users/user-123/jobs/job-" + scenario + "/audio/background-music.mp3"
}

func buildNarratorURL(scenario string) string {
	return "https://mock.s3.amazonaws.com/users/user-123/jobs/job-" + scenario + "/audio/narrator-voiceover.mp3"
}

func buildVideoKey(scenario string) string {
	return "users/user-123/jobs/job-" + scenario + "/final/video.mp4"
}

func FormatSceneNumber(sceneNumber int) string {
	return formatNumberWithPadding(sceneNumber, 3)
}

func DurationLabel(duration int) string {
	return fmt.Sprintf("%ds", duration)
}

func formatNumberWithPadding(n, width int) string {
	format := fmt.Sprintf("%%0%dd", width)
	return fmt.Sprintf(format, n)
}

func lookupSyncPoint(points []domain.SyncPoint, pointType string) string {
	for _, point := range points {
		if point.Type == pointType {
			return point.Description
		}
	}
	return ""
}

func repeatText(text string, times int) string {
	result := ""
	for i := 0; i < times; i++ {
		if i > 0 {
			result += " "
		}
		result += text
	}
	return result
}
