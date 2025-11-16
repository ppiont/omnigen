package aws

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/sfn"
)

// NewConfig creates a new AWS SDK configuration
func NewConfig(region string) (aws.Config, error) {
	// Check for local DynamoDB endpoint
	endpointURL := os.Getenv("AWS_ENDPOINT_URL")

	var opts []func(*config.LoadOptions) error
	opts = append(opts, config.WithRegion(region))

	// If using local DynamoDB, use static credentials
	if endpointURL != "" {
		opts = append(opts, config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider("dummy", "dummy", ""),
		))
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(), opts...)
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
	StepFunctions  *sfn.Client
	Lambda         *lambda.Client
}

// Config is a placeholder for config that might be needed in the future
type Config interface {
	GetAWSRegion() string
}

// NewClients creates all AWS service clients
func NewClients(cfg aws.Config, appConfig interface{}) *Clients {
	// Check for local DynamoDB endpoint
	endpointURL := os.Getenv("AWS_ENDPOINT_URL")

	var dynamoClient *dynamodb.Client
	if endpointURL != "" {
		// Use custom endpoint for DynamoDB Local
		dynamoClient = dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
			o.BaseEndpoint = &endpointURL
		})
	} else {
		dynamoClient = dynamodb.NewFromConfig(cfg)
	}

	return &Clients{
		DynamoDB:       dynamoClient,
		S3:             s3.NewFromConfig(cfg),
		SecretsManager: secretsmanager.NewFromConfig(cfg),
		StepFunctions:  sfn.NewFromConfig(cfg),
		Lambda:         lambda.NewFromConfig(cfg),
	}
}
