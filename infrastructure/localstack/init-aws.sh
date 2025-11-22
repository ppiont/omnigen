#!/bin/bash

echo "Initializing LocalStack AWS resources..."

# Create S3 bucket for assets
awslocal s3 mb s3://omnigen-assets
awslocal s3api put-bucket-cors --bucket omnigen-assets --cors-configuration '{
  "CORSRules": [
    {
      "AllowedOrigins": ["*"],
      "AllowedMethods": ["GET", "PUT", "POST", "DELETE", "HEAD"],
      "AllowedHeaders": ["*"],
      "ExposeHeaders": ["ETag"]
    }
  ]
}'

# Create DynamoDB table for jobs
awslocal dynamodb create-table \
    --table-name omnigen-jobs \
    --attribute-definitions \
        AttributeName=job_id,AttributeType=S \
        AttributeName=user_id,AttributeType=S \
    --key-schema \
        AttributeName=job_id,KeyType=HASH \
    --global-secondary-indexes \
        '[{
            "IndexName": "user_id-index",
            "KeySchema": [{"AttributeName": "user_id", "KeyType": "HASH"}],
            "Projection": {"ProjectionType": "ALL"},
            "ProvisionedThroughput": {"ReadCapacityUnits": 5, "WriteCapacityUnits": 5}
        }]' \
    --provisioned-throughput \
        ReadCapacityUnits=5,WriteCapacityUnits=5

# Create DynamoDB table for usage tracking
awslocal dynamodb create-table \
    --table-name omnigen-usage \
    --attribute-definitions \
        AttributeName=user_id,AttributeType=S \
    --key-schema \
        AttributeName=user_id,KeyType=HASH \
    --provisioned-throughput \
        ReadCapacityUnits=5,WriteCapacityUnits=5

# Create secrets for API keys (placeholder values)
awslocal secretsmanager create-secret \
    --name omnigen/replicate-api-key \
    --secret-string '{"api_key": "placeholder"}'

awslocal secretsmanager create-secret \
    --name omnigen/openai-api-key \
    --secret-string '{"api_key": "placeholder"}'

echo "LocalStack initialization complete!"
