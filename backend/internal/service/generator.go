package service

import (
	"context"
	"fmt"

	"github.com/omnigen/backend/internal/domain"
	"github.com/omnigen/backend/internal/repository"
	"go.uber.org/zap"
)

// GeneratorService is a stub - DEPRECATED: Video generation now handled by goroutines in GenerateHandler
// This service is kept for backwards compatibility but should not be used
type GeneratorService struct {
	jobRepo *repository.DynamoDBRepository
	logger  *zap.Logger
}

// NewGeneratorService creates a new generator service stub
func NewGeneratorService(
	jobRepo *repository.DynamoDBRepository,
	logger *zap.Logger,
) *GeneratorService {
	return &GeneratorService{
		jobRepo: jobRepo,
		logger:  logger,
	}
}

// GenerateVideo is a stub - will be refactored to work with scripts
func (s *GeneratorService) GenerateVideo(ctx context.Context, req *domain.GenerateRequest) (*domain.Job, error) {
	return nil, fmt.Errorf("GenerateVideo is being refactored - use script-based flow")
}
