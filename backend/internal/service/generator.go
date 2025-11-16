package service

import (
	"context"
	"fmt"

	"github.com/omnigen/backend/internal/domain"
	"github.com/omnigen/backend/internal/repository"
	"go.uber.org/zap"
)

// GeneratorService is a stub - being refactored to work with script-based flow
// TODO: Update to accept script_id instead of raw prompts
type GeneratorService struct {
	jobRepo       *repository.DynamoDBRepository
	stepFunctions *StepFunctionsService
	// promptParser  *PromptParser
	// scenePlanner  *ScenePlanner
	logger *zap.Logger
}

// NewGeneratorService creates a new generator service stub
func NewGeneratorService(
	jobRepo *repository.DynamoDBRepository,
	stepFunctions *StepFunctionsService,
	// promptParser *PromptParser,
	// scenePlanner *ScenePlanner,
	logger *zap.Logger,
) *GeneratorService {
	return &GeneratorService{
		jobRepo:       jobRepo,
		stepFunctions: stepFunctions,
		// promptParser:  promptParser,
		// scenePlanner:  scenePlanner,
		logger: logger,
	}
}

// GenerateVideo is a stub - will be refactored to work with scripts
func (s *GeneratorService) GenerateVideo(ctx context.Context, req *domain.GenerateRequest) (*domain.Job, error) {
	return nil, fmt.Errorf("GenerateVideo is being refactored - use script-based flow")
}
