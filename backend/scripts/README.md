# Local Development Scripts

This directory contains scripts to help with local development and testing of the OmniGen application.

## Prerequisites

- Docker (for running DynamoDB Local)
- AWS CLI (install via Homebrew: `brew install awscli`)
- Go backend running with `MOCK_MODE=true` in `.env`

## Quick Start

### 1. Start DynamoDB Local

```bash
docker run -d -p 8000:8000 --name dynamodb-local amazon/dynamodb-local
```

Or if the container already exists:

```bash
docker start dynamodb-local
```

### 2. Create DynamoDB Tables

```bash
cd backend
bash scripts/setup-local-dynamodb.sh
```

This creates three tables:
- `omnigen-jobs-local` - Video generation jobs
- `omnigen-scripts-local` - AI-generated video scripts
- `omnigen-usage-local` - Usage tracking

### 3. Seed Sample Videos

```bash
bash scripts/seed-sample-video.sh        # Adds 1 completed video
bash scripts/seed-multiple-videos.sh     # Adds 8 videos with different statuses
```

The multi-video seed script creates:
- 4 completed videos (various ages from 2h to 24h ago)
- 2 processing videos (in progress)
- 1 pending video (not yet started)
- 1 failed video

### 4. View Videos

Open your browser and navigate to:
- Frontend: http://localhost:5173/library
- API: http://localhost:8080/api/v1/jobs

## Scripts Reference

### `setup-local-dynamodb.sh`

Creates local DynamoDB tables with correct schema and indexes.

**Usage:**
```bash
bash scripts/setup-local-dynamodb.sh
```

**What it does:**
- Connects to DynamoDB Local on port 8000
- Creates tables with GlobalSecondaryIndexes
- Uses dummy AWS credentials (required by AWS CLI)

### `seed-sample-video.sh`

Creates a single completed sample video in the database.

**Usage:**
```bash
bash scripts/seed-sample-video.sh
```

**Generated video:**
- Title: "Mountain Sunrise"
- Status: completed
- Duration: 30 seconds
- User: dev-user-123 (matches dev mode auth)

### `seed-multiple-videos.sh`

Creates multiple sample videos with different statuses and ages for testing.

**Usage:**
```bash
bash scripts/seed-multiple-videos.sh
```

**Generated videos:**
1. Mountain Sunrise (completed, 2h ago)
2. Coffee Commercial (completed, 5h ago)
3. Tech Product Launch (completed, 12h ago)
4. Fitness Motivation (completed, 24h ago)
5. Travel Adventure (processing, 0h ago)
6. Food Recipe (processing, 0h ago)
7. Fashion Lookbook (pending, 0h ago)
8. Test Video (failed, 1h ago)

All videos are assigned to user `dev-user-123` which matches the mock user in dev mode.

## Troubleshooting

### Issue: "aws: command not found"

**Solution:** Install AWS CLI via Homebrew
```bash
brew install awscli
```

### Issue: "Could not connect to the endpoint URL"

**Solution:** Make sure DynamoDB Local is running
```bash
docker ps  # Check if dynamodb-local is running
docker start dynamodb-local  # Start if stopped
```

### Issue: "The table does not have the specified index: UserJobsIndex"

**Solution:** Delete and recreate the tables
```bash
# Delete existing tables
export AWS_ACCESS_KEY_ID=dummy AWS_SECRET_ACCESS_KEY=dummy
aws dynamodb delete-table --table-name omnigen-jobs-local --endpoint-url http://localhost:8000 --region us-east-1

# Recreate with correct index
bash scripts/setup-local-dynamodb.sh
```

### Issue: API returns empty array `{"jobs": []}`

**Solution:** Make sure:
1. Tables are created with `setup-local-dynamodb.sh`
2. Sample data is seeded with `seed-multiple-videos.sh`
3. Backend is running with correct DYNAMODB_ENDPOINT in `.env`
4. Auth bypass is enabled in dev mode

## Environment Configuration

Make sure your `backend/.env` includes:

```env
# Required for local development
MOCK_MODE=true
DYNAMODB_ENDPOINT=http://localhost:8000
ASSETS_BUCKET=omnigen-assets-local
JOB_TABLE=omnigen-jobs-local
USAGE_TABLE=omnigen-usage-local
SCRIPTS_TABLE=omnigen-scripts-local

# Mock AWS credentials work fine for local DynamoDB
AWS_REGION=us-east-1
```

## Data Persistence

**Note:** DynamoDB Local stores data in-memory by default. If you restart the Docker container, all data will be lost. To persist data between restarts, run:

```bash
docker run -d -p 8000:8000 -v /tmp/dynamodb:/data --name dynamodb-local amazon/dynamodb-local -jar DynamoDBLocal.jar -sharedDb -dbPath /data
```

## Cleanup

To start fresh:

```bash
# Stop and remove DynamoDB container
docker stop dynamodb-local
docker rm dynamodb-local

# Remove all data
rm -rf /tmp/dynamodb

# Start fresh
docker run -d -p 8000:8000 --name dynamodb-local amazon/dynamodb-local
bash scripts/setup-local-dynamodb.sh
bash scripts/seed-multiple-videos.sh
```
