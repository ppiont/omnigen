package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/omnigen/backend/internal/domain"
	"github.com/omnigen/backend/internal/repository"
	"go.uber.org/zap"
)

// GeneratorService orchestrates video generation
type GeneratorService struct {
	jobRepo              *repository.DynamoDBRepository
	stepFunctionsService *StepFunctionsService
	promptParser         *PromptParser
	scenePlanner         *ScenePlanner
	logger               *zap.Logger
}

// NewGeneratorService creates a new generator service
func NewGeneratorService(
	jobRepo *repository.DynamoDBRepository,
	stepFunctionsService *StepFunctionsService,
	promptParser *PromptParser,
	scenePlanner *ScenePlanner,
	logger *zap.Logger,
) *GeneratorService {
	return &GeneratorService{
		jobRepo:              jobRepo,
		stepFunctionsService: stepFunctionsService,
		promptParser:         promptParser,
		scenePlanner:         scenePlanner,
		logger:               logger,
	}
}

// GenerateVideo orchestrates the entire video generation process
func (g *GeneratorService) GenerateVideo(ctx context.Context, req *domain.GenerateRequest) (*domain.Job, error) {
	// Generate unique job ID
	jobID := uuid.New().String()

	g.logger.Info("Starting video generation",
		zap.String("job_id", jobID),
		zap.String("prompt", req.Prompt),
		zap.Int("duration", req.Duration),
	)

	// Step 1: Parse the prompt
	parsed, err := g.promptParser.ParsePrompt(ctx, req)
	if err != nil {
		g.logger.Error("Failed to parse prompt",
			zap.String("job_id", jobID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to parse prompt: %w", err)
	}

	g.logger.Info("Prompt parsed",
		zap.String("job_id", jobID),
		zap.String("product_type", parsed.ProductType),
		zap.Strings("visual_style", parsed.VisualStyle),
	)

	// Step 2: Plan the scenes
	scenes, err := g.scenePlanner.PlanScenes(parsed)
	if err != nil {
		g.logger.Error("Failed to plan scenes",
			zap.String("job_id", jobID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to plan scenes: %w", err)
	}

	g.logger.Info("Scenes planned",
		zap.String("job_id", jobID),
		zap.Int("scene_count", len(scenes)),
	)

	// Step 3: Create job in DynamoDB
	now := time.Now().Unix()
	ttl := time.Now().Add(90 * 24 * time.Hour).Unix() // 90 days TTL

	job := &domain.Job{
		JobID:       jobID,
		UserID:      "system", // TODO: Get from auth context
		Status:      domain.StatusPending,
		Prompt:      req.Prompt,
		Duration:    req.Duration,
		Style:       req.Style,
		AspectRatio: req.AspectRatio,
		CreatedAt:   now,
		TTL:         ttl,
	}

	if err := g.jobRepo.CreateJob(ctx, job); err != nil {
		g.logger.Error("Failed to create job",
			zap.String("job_id", jobID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to create job: %w", err)
	}

	g.logger.Info("Job created in DynamoDB",
		zap.String("job_id", jobID),
	)

	// Step 4: Start Step Functions workflow
	sfnInput := &domain.StepFunctionsInput{
		JobID:    jobID,
		Prompt:   req.Prompt,
		Duration: req.Duration,
		Style:    req.Style,
		Scenes:   scenes,
	}

	executionARN, err := g.stepFunctionsService.StartExecution(ctx, sfnInput)
	if err != nil {
		g.logger.Error("Failed to start Step Functions execution",
			zap.String("job_id", jobID),
			zap.Error(err),
		)

		// Update job status to failed
		_ = g.jobRepo.UpdateJobStatus(ctx, jobID, domain.StatusFailed)

		return nil, fmt.Errorf("failed to start workflow: %w", err)
	}

	g.logger.Info("Step Functions execution started",
		zap.String("job_id", jobID),
		zap.String("execution_arn", executionARN),
	)

	// Update job status to processing
	if err := g.jobRepo.UpdateJobStatus(ctx, jobID, domain.StatusProcessing); err != nil {
		g.logger.Error("Failed to update job status",
			zap.String("job_id", jobID),
			zap.Error(err),
		)
		// Not critical - the workflow will update status eventually
	}

	job.Status = domain.StatusProcessing

	g.logger.Info("Video generation initiated successfully",
		zap.String("job_id", jobID),
		zap.Int("scene_count", len(scenes)),
	)

	return job, nil
}
