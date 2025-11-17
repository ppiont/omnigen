package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

// NewConfig creates a new AWS SDK configuration
func NewConfig(ctx context.Context, region string) (aws.Config, error) {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(region),
	)
	if err != nil {
		return aws.Config{}, err
	}
	return cfg, nil
}

// Clients holds all AWS service clients
type Clients struct {
	DynamoDB       *dynamodb.Client
	S3             *s3.Client
	SecretsManager *secretsmanager.Client
}

// NewClients creates all AWS service clients
func NewClients(cfg aws.Config) *Clients {
	return &Clients{
		DynamoDB:       dynamodb.NewFromConfig(cfg),
		S3:             s3.NewFromConfig(cfg),
		SecretsManager: secretsmanager.NewFromConfig(cfg),
	}
}
