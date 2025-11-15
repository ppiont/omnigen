package service

import (
	"fmt"
	"time"

	"github.com/omnigen/backend/internal/domain"
)

// MockService provides mock data for frontend development
type MockService struct {
	jobs map[string]*MockJob
}

// MockJob represents a mock job with state transitions
type MockJob struct {
	Job       *domain.Job
	CreatedAt time.Time
	Stages    []MockStage
}

// MockStage represents a stage in the mock job lifecycle
type MockStage struct {
	Name      string
	Status    string
	Progress  int
	StartTime time.Duration // Time after job creation when this stage starts
	Duration  time.Duration // How long this stage takes
}

// NewMockService creates a new mock service
func NewMockService() *MockService {
	return &MockService{
		jobs: make(map[string]*MockJob),
	}
}

// CreateMockJob creates a mock job with realistic state transitions
func (m *MockService) CreateMockJob(userID, prompt string, duration int, aspectRatio string) *domain.Job {
	jobID := fmt.Sprintf("job-mock-%d", time.Now().Unix())

	job := &domain.Job{
		JobID:       jobID,
		UserID:      userID,
		Status:      "pending",
		Prompt:      prompt,
		Duration:    duration,
		AspectRatio: aspectRatio,
		CreatedAt:   time.Now().Unix(),
	}

	// Define mock stages with realistic timing
	mockJob := &MockJob{
		Job:       job,
		CreatedAt: time.Now(),
		Stages: []MockStage{
			{
				Name:      "pending",
				Status:    "pending",
				Progress:  0,
				StartTime: 0,
				Duration:  5 * time.Second,
			},
			{
				Name:      "parsing",
				Status:    "parsing",
				Progress:  20,
				StartTime: 5 * time.Second,
				Duration:  10 * time.Second,
			},
			{
				Name:      "generating_videos",
				Status:    "generating_videos",
				Progress:  40,
				StartTime: 15 * time.Second,
				Duration:  60 * time.Second,
			},
			{
				Name:      "generating_audio",
				Status:    "generating_audio",
				Progress:  60,
				StartTime: 75 * time.Second,
				Duration:  45 * time.Second,
			},
			{
				Name:      "composing",
				Status:    "composing",
				Progress:  80,
				StartTime: 120 * time.Second,
				Duration:  60 * time.Second,
			},
			{
				Name:      "completed",
				Status:    "completed",
				Progress:  100,
				StartTime: 180 * time.Second,
				Duration:  0,
			},
		},
	}

	m.jobs[jobID] = mockJob
	return job
}

// GetMockJobStatus returns the current status of a mock job based on elapsed time
func (m *MockService) GetMockJobStatus(jobID string) *domain.Job {
	mockJob, exists := m.jobs[jobID]
	if !exists {
		return nil
	}

	elapsed := time.Since(mockJob.CreatedAt)

	// Find current stage based on elapsed time
	currentStage := mockJob.Stages[0]
	for _, stage := range mockJob.Stages {
		if elapsed >= stage.StartTime && elapsed < stage.StartTime+stage.Duration {
			currentStage = stage
			break
		}
		if elapsed >= stage.StartTime+stage.Duration {
			currentStage = stage
		}
	}

	// Update job status
	mockJob.Job.Status = currentStage.Status

	// If completed, add mock video URL
	if currentStage.Status == "completed" {
		mockJob.Job.VideoKey = fmt.Sprintf("%s/final.mp4", jobID)
		completedAt := time.Now().Unix()
		mockJob.Job.CompletedAt = &completedAt
	}

	return mockJob.Job
}

// GetMockProgress returns detailed progress information
func (m *MockService) GetMockProgress(jobID string) *MockProgress {
	mockJob, exists := m.jobs[jobID]
	if !exists {
		return nil
	}

	elapsed := time.Since(mockJob.CreatedAt)

	// Find current stage
	currentStageIndex := 0
	currentStage := mockJob.Stages[0]
	for i, stage := range mockJob.Stages {
		if elapsed >= stage.StartTime && elapsed < stage.StartTime+stage.Duration {
			currentStage = stage
			currentStageIndex = i
			break
		}
		if elapsed >= stage.StartTime+stage.Duration {
			currentStage = stage
			currentStageIndex = i
		}
	}

	// Calculate stages completed
	stagesCompleted := make([]string, 0)
	stagesPending := make([]string, 0)

	for i, stage := range mockJob.Stages {
		if i < currentStageIndex {
			stagesCompleted = append(stagesCompleted, stage.Name)
		} else if i > currentStageIndex {
			stagesPending = append(stagesPending, stage.Name)
		}
	}

	// Calculate estimated time remaining
	totalDuration := time.Duration(0)
	for _, stage := range mockJob.Stages {
		totalDuration += stage.Duration
	}
	remainingDuration := totalDuration - elapsed
	if remainingDuration < 0 {
		remainingDuration = 0
	}

	return &MockProgress{
		JobID:                  jobID,
		Status:                 currentStage.Status,
		Progress:               currentStage.Progress,
		CurrentStage:           currentStage.Name,
		StagesCompleted:        stagesCompleted,
		StagesPending:          stagesPending,
		EstimatedTimeRemaining: int(remainingDuration.Seconds()),
	}
}

// MockProgress represents detailed progress information
type MockProgress struct {
	JobID                  string   `json:"job_id"`
	Status                 string   `json:"status"`
	Progress               int      `json:"progress"`
	CurrentStage           string   `json:"current_stage"`
	StagesCompleted        []string `json:"stages_completed"`
	StagesPending          []string `json:"stages_pending"`
	EstimatedTimeRemaining int      `json:"estimated_time_remaining"`
}

// GetMockPresets returns a list of brand presets
func (m *MockService) GetMockPresets() []BrandPreset {
	return []BrandPreset{
		{
			ID:          "luxury-gold",
			Name:        "Luxury Gold",
			Description: "Elegant gold aesthetics for high-end products",
			Style:       "cinematic, elegant, gold accents, shallow depth of field",
			ColorPalette: []string{"#D4AF37", "#000000", "#FFFFFF", "#8B7355"},
			MusicMood:   "sophisticated, elegant, uplifting",
		},
		{
			ID:          "modern-tech",
			Name:        "Modern Tech",
			Description: "Clean, minimalist design for tech products",
			Style:       "modern, minimalist, white background, sharp focus",
			ColorPalette: []string{"#FFFFFF", "#2C3E50", "#3498DB", "#ECF0F1"},
			MusicMood:   "energetic, futuristic, innovative",
		},
		{
			ID:          "vibrant-lifestyle",
			Name:        "Vibrant Lifestyle",
			Description: "Colorful, energetic style for lifestyle brands",
			Style:       "vibrant, colorful, lifestyle photography, natural lighting",
			ColorPalette: []string{"#FF6B6B", "#4ECDC4", "#45B7D1", "#FFA07A"},
			MusicMood:   "upbeat, cheerful, motivational",
		},
		{
			ID:          "professional-corporate",
			Name:        "Professional Corporate",
			Description: "Clean, professional look for B2B products",
			Style:       "professional, corporate, clean lines, business setting",
			ColorPalette: []string{"#34495E", "#FFFFFF", "#3498DB", "#7F8C8D"},
			MusicMood:   "confident, professional, trustworthy",
		},
		{
			ID:          "vintage-retro",
			Name:        "Vintage Retro",
			Description: "Nostalgic, retro-inspired aesthetic",
			Style:       "vintage, retro, film grain, warm tones, nostalgic",
			ColorPalette: []string{"#E8C547", "#C25B56", "#6B5B3D", "#F4E8D0"},
			MusicMood:   "nostalgic, warm, classic",
		},
	}
}

// BrandPreset represents a brand style preset
type BrandPreset struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Style        string   `json:"style"`
	ColorPalette []string `json:"color_palette"`
	MusicMood    string   `json:"music_mood"`
}
