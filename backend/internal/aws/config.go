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
	// Check for LocalStack or local AWS endpoints
	awsEndpoint := os.Getenv("AWS_ENDPOINT_URL")
	dynamoEndpoint := os.Getenv("DYNAMODB_ENDPOINT")

	var cfg aws.Config
	var err error

	// Use local credentials if any local endpoint is configured
	if awsEndpoint != "" || dynamoEndpoint != "" {
		// For LocalStack/local development, use static credentials
		cfg, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(region),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("test", "test", "")),
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
	// Check for LocalStack endpoint (unified) or individual service endpoints
	awsEndpoint := os.Getenv("AWS_ENDPOINT_URL")
	dynamoEndpoint := os.Getenv("DYNAMODB_ENDPOINT")

	// Use AWS_ENDPOINT_URL if set, otherwise fall back to service-specific
	if awsEndpoint != "" && dynamoEndpoint == "" {
		dynamoEndpoint = awsEndpoint
	}

	var dynamoClient *dynamodb.Client
	if dynamoEndpoint != "" {
		// Use custom endpoint for local DynamoDB
		dynamoClient = dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
			o.BaseEndpoint = &dynamoEndpoint
		})
	} else {
		dynamoClient = dynamodb.NewFromConfig(cfg)
	}

	// S3 client with optional LocalStack endpoint
	var s3Client *s3.Client
	if awsEndpoint != "" {
		s3Client = s3.NewFromConfig(cfg, func(o *s3.Options) {
			o.BaseEndpoint = &awsEndpoint
			o.UsePathStyle = true // Required for LocalStack
		})
	} else {
		s3Client = s3.NewFromConfig(cfg)
	}

	// SecretsManager client with optional LocalStack endpoint
	var smClient *secretsmanager.Client
	if awsEndpoint != "" {
		smClient = secretsmanager.NewFromConfig(cfg, func(o *secretsmanager.Options) {
			o.BaseEndpoint = &awsEndpoint
		})
	} else {
		smClient = secretsmanager.NewFromConfig(cfg)
	}

	return &Clients{
		DynamoDB:       dynamoClient,
		S3:             s3Client,
		SecretsManager: smClient,
	}
}
