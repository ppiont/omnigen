# OmniGen Backend API

Go-based REST API for the OmniGen AI video generation pipeline.

## Tech Stack

- **Go 1.25.4**
- **Gin** - Web framework
- **AWS SDK v2** - DynamoDB, S3, Step Functions, Secrets Manager
- **Zap** - Structured logging
- **Swagger/OpenAPI** - API documentation

## Architecture

```
backend/
├── cmd/api/          # Application entry point
├── internal/
│   ├── api/          # HTTP handlers, middleware, routing
│   ├── domain/       # Domain models
│   ├── service/      # Business logic
│   ├── repository/   # Data access layer
│   └── aws/          # AWS client wrappers
├── pkg/              # Reusable packages
└── docs/             # Generated Swagger docs
```

## Environment Variables

Required environment variables (automatically injected by ECS):

```bash
# Server
PORT=8080
ENVIRONMENT=production

# AWS
AWS_REGION=us-east-1
ASSETS_BUCKET=omnigen-assets-{account_id}
JOB_TABLE=omnigen-jobs
STEP_FUNCTIONS_ARN=arn:aws:states:us-east-1:{account_id}:stateMachine:omnigen-workflow
REPLICATE_SECRET_ARN=arn:aws:secretsmanager:us-east-1:{account_id}:secret:omnigen/replicate-api-key
```

## API Endpoints

### Health Check
```
GET /health
```

### Video Generation
```
POST /api/v1/generate
Content-Type: application/json
x-api-key: <your-api-key>

{
  "prompt": "15 second luxury watch ad with gold aesthetics",
  "duration": 15,
  "aspect_ratio": "9:16",
  "style": "luxury, minimal, elegant"
}
```

### Get Job Status
```
GET /api/v1/jobs/{job_id}
x-api-key: <your-api-key>
```

### List Jobs
```
GET /api/v1/jobs?page=1&page_size=20&status=completed
x-api-key: <your-api-key>
```

## Development

### Prerequisites

- Go 1.22+
- Docker
- AWS credentials configured

### Run Locally

```bash
# Install dependencies
go mod tidy

# Run the server
go run cmd/api/main.go

# Or with live reload (install air first)
air
```

### Generate Swagger Documentation

```bash
# Install swag
go install github.com/swaggo/swag/cmd/swag@latest

# Generate docs
swag init -g cmd/api/main.go -o docs

# View docs at http://localhost:8080/swagger/index.html
```

### Build Docker Image

```bash
# Build
docker build -t omnigen-api:latest .

# Run
docker run -p 8080:8080 \
  -e AWS_REGION=us-east-1 \
  -e ASSETS_BUCKET=omnigen-assets \
  -e JOB_TABLE=omnigen-jobs \
  omnigen-api:latest
```

## Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package
go test ./internal/service/...
```

## Deployment

Deploy using the root Makefile:

```bash
# From project root
make deploy-ecs
```

This will:
1. Build Docker image
2. Push to ECR
3. Update ECS service with new image

## API Documentation

Interactive Swagger UI available at:
- **Development**: `http://localhost:8080/swagger/index.html`
- **Production**: Disabled for security

## Project Structure

- `cmd/api/main.go` - Application entry point
- `internal/api/` - HTTP layer (handlers, middleware, routing)
- `internal/domain/` - Domain models (Job, Scene, etc.)
- `internal/service/` - Business logic (prompt parsing, scene planning, orchestration)
- `internal/repository/` - Data access (DynamoDB, S3)
- `internal/aws/` - AWS client configuration
- `pkg/logger/` - Structured logging setup
- `pkg/errors/` - Custom error types

## MVP Features

✅ Ad creative pipeline (15-60s product ads)
✅ Claude API-powered prompt parsing
✅ 3-scene video structure
✅ Step Functions workflow orchestration
✅ DynamoDB job tracking
✅ S3 video storage with presigned URLs
✅ API key authentication (Secrets Manager)
✅ Health checks for ECS/ALB
✅ Structured JSON logging
✅ OpenAPI/Swagger documentation

## License

MIT
