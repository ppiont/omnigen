package aws

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

// NewConfig creates a new AWS SDK configuration
func NewConfig(ctx context.Context, region string) (aws.Config, error) {
	// Check if using local DynamoDB
	dynamoEndpoint := os.Getenv("DYNAMODB_ENDPOINT")

	var cfg aws.Config
	var err error

	if dynamoEndpoint != "" {
		// For local DynamoDB, use dummy credentials
		cfg, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(region),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("dummy", "dummy", "")),
		)
	} else {
		// For production, use default credentials
		cfg, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(region),
		)
	}

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
	// Check for local DynamoDB endpoint
	dynamoEndpoint := os.Getenv("DYNAMODB_ENDPOINT")

	var dynamoClient *dynamodb.Client
	if dynamoEndpoint != "" {
		// Use custom endpoint for local DynamoDB
		dynamoClient = dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
			o.BaseEndpoint = &dynamoEndpoint
		})
	} else {
		dynamoClient = dynamodb.NewFromConfig(cfg)
	}

	return &Clients{
		DynamoDB:       dynamoClient,
		S3:             s3.NewFromConfig(cfg),
		SecretsManager: secretsmanager.NewFromConfig(cfg),
	}
}
