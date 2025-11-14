package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sfn"
	"github.com/omnigen/backend/internal/domain"
	"go.uber.org/zap"
)

// StepFunctionsService handles Step Functions workflow execution
type StepFunctionsService struct {
	client           *sfn.Client
	stateMachineARN  string
	logger           *zap.Logger
}

// NewStepFunctionsService creates a new Step Functions service
func NewStepFunctionsService(
	client *sfn.Client,
	stateMachineARN string,
	logger *zap.Logger,
) *StepFunctionsService {
	return &StepFunctionsService{
		client:          client,
		stateMachineARN: stateMachineARN,
		logger:          logger,
	}
}

// StartExecution starts a new Step Functions execution
func (s *StepFunctionsService) StartExecution(ctx context.Context, input *domain.StepFunctionsInput) (string, error) {
	payload, err := json.Marshal(input)
	if err != nil {
		s.logger.Error("Failed to marshal input", zap.Error(err))
		return "", fmt.Errorf("failed to marshal input: %w", err)
	}

	s.logger.Info("Starting Step Functions execution",
		zap.String("job_id", input.JobID),
		zap.String("state_machine_arn", s.stateMachineARN),
	)

	result, err := s.client.StartExecution(ctx, &sfn.StartExecutionInput{
		StateMachineArn: aws.String(s.stateMachineARN),
		Name:            aws.String(input.JobID),
		Input:           aws.String(string(payload)),
	})
	if err != nil {
		s.logger.Error("Failed to start execution",
			zap.String("job_id", input.JobID),
			zap.Error(err),
		)
		return "", fmt.Errorf("failed to start execution: %w", err)
	}

	s.logger.Info("Step Functions execution started",
		zap.String("job_id", input.JobID),
		zap.String("execution_arn", *result.ExecutionArn),
	)

	return *result.ExecutionArn, nil
}
