#!/bin/bash

# Setup local DynamoDB tables for development
# Run this after starting DynamoDB Local with: docker run -p 8000:8000 amazon/dynamodb-local

ENDPOINT="http://localhost:8000"
REGION="us-east-1"

# Set dummy credentials for DynamoDB Local
export AWS_ACCESS_KEY_ID=dummy
export AWS_SECRET_ACCESS_KEY=dummy

echo "Setting up local DynamoDB tables..."

# Create Jobs Table
echo "Creating omnigen-jobs-local table..."
aws dynamodb create-table \
    --table-name omnigen-jobs-local \
    --attribute-definitions \
        AttributeName=job_id,AttributeType=S \
        AttributeName=user_id,AttributeType=S \
        AttributeName=created_at,AttributeType=N \
    --key-schema \
        AttributeName=job_id,KeyType=HASH \
    --provisioned-throughput \
        ReadCapacityUnits=5,WriteCapacityUnits=5 \
    --global-secondary-indexes \
        "[
            {
                \"IndexName\": \"user-jobs-index\",
                \"KeySchema\": [
                    {\"AttributeName\":\"user_id\",\"KeyType\":\"HASH\"},
                    {\"AttributeName\":\"created_at\",\"KeyType\":\"RANGE\"}
                ],
                \"Projection\": {\"ProjectionType\":\"ALL\"},
                \"ProvisionedThroughput\": {\"ReadCapacityUnits\":5,\"WriteCapacityUnits\":5}
            }
        ]" \
    --endpoint-url $ENDPOINT \
    --region $REGION \
    --no-cli-pager 2>&1 | grep -v "ResourceInUseException" || echo "  ✓ Table already exists"

# Create Scripts Table
echo "Creating omnigen-scripts-local table..."
aws dynamodb create-table \
    --table-name omnigen-scripts-local \
    --attribute-definitions \
        AttributeName=script_id,AttributeType=S \
        AttributeName=user_id,AttributeType=S \
        AttributeName=created_at,AttributeType=N \
    --key-schema \
        AttributeName=script_id,KeyType=HASH \
    --provisioned-throughput \
        ReadCapacityUnits=5,WriteCapacityUnits=5 \
    --global-secondary-indexes \
        "[
            {
                \"IndexName\": \"user-scripts-index\",
                \"KeySchema\": [
                    {\"AttributeName\":\"user_id\",\"KeyType\":\"HASH\"},
                    {\"AttributeName\":\"created_at\",\"KeyType\":\"RANGE\"}
                ],
                \"Projection\": {\"ProjectionType\":\"ALL\"},
                \"ProvisionedThroughput\": {\"ReadCapacityUnits\":5,\"WriteCapacityUnits\":5}
            }
        ]" \
    --endpoint-url $ENDPOINT \
    --region $REGION \
    --no-cli-pager 2>&1 | grep -v "ResourceInUseException" || echo "  ✓ Table already exists"

# Create Usage Table
echo "Creating omnigen-usage-local table..."
aws dynamodb create-table \
    --table-name omnigen-usage-local \
    --attribute-definitions \
        AttributeName=user_id,AttributeType=S \
        AttributeName=timestamp,AttributeType=N \
    --key-schema \
        AttributeName=user_id,KeyType=HASH \
        AttributeName=timestamp,KeyType=RANGE \
    --provisioned-throughput \
        ReadCapacityUnits=5,WriteCapacityUnits=5 \
    --endpoint-url $ENDPOINT \
    --region $REGION \
    --no-cli-pager 2>&1 | grep -v "ResourceInUseException" || echo "  ✓ Table already exists"

echo ""
echo "✅ Local DynamoDB setup complete!"
echo ""
echo "List tables:"
aws dynamodb list-tables --endpoint-url $ENDPOINT --region $REGION --no-cli-pager
