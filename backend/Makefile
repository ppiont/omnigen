.PHONY: build clean deploy test local-invoke build-ParserFunction build-GeneratorFunction build-AudioGeneratorFunction build-ComposerFunction deploy-parser build-parser deploy-generator build-generator deploy-audio-generator build-audio-generator deploy-composer build-composer deploy-lambdas build-lambdas swagger help

# Default target - show help
.DEFAULT_GOAL := help

# Help target
help:
	@echo "OmniGen Backend - Available Make Targets"
	@echo ""
	@echo "Building:"
	@echo "  build               - Build all Lambda functions using SAM"
	@echo "  build-fast          - Fast local build without SAM"
	@echo "  build-lambdas       - Build all Lambda functions (no deploy)"
	@echo "  build-parser        - Build parser Lambda only"
	@echo "  build-generator     - Build generator Lambda only"
	@echo "  build-audio-generator - Build audio-generator Lambda only"
	@echo "  build-composer      - Build composer Lambda only"
	@echo ""
	@echo "Deploying:"
	@echo "  deploy              - Deploy to AWS (guided)"
	@echo "  deploy-ci           - Deploy to AWS (CI/CD mode)"
	@echo "  deploy-lambdas      - Deploy all Lambda functions"
	@echo "  deploy-parser       - Deploy parser Lambda only"
	@echo "  deploy-generator    - Deploy generator Lambda only"
	@echo "  deploy-audio-generator - Deploy audio-generator Lambda only"
	@echo "  deploy-composer     - Deploy composer Lambda only"
	@echo ""
	@echo "Testing:"
	@echo "  test                - Run all tests"
	@echo "  test-coverage       - Run tests with coverage report"
	@echo ""
	@echo "Local Development:"
	@echo "  local-api           - Start local API server"
	@echo "  local-invoke-parser - Invoke parser Lambda locally"
	@echo ""
	@echo "Documentation:"
	@echo "  swagger             - Generate/update Swagger API docs"
	@echo ""
	@echo "Maintenance:"
	@echo "  clean               - Clean build artifacts"
	@echo "  validate            - Validate SAM template"
	@echo "  package             - Package for deployment"
	@echo ""

# SAM will call these targets automatically for each function
build-ParserFunction:
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -tags lambda.norpc -o $(ARTIFACTS_DIR)/bootstrap ./cmd/lambdas/parser

build-GeneratorFunction:
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -tags lambda.norpc -o $(ARTIFACTS_DIR)/bootstrap ./cmd/lambdas/generator

build-AudioGeneratorFunction:
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -tags lambda.norpc -o $(ARTIFACTS_DIR)/bootstrap ./cmd/lambdas/audio-generator

build-ComposerFunction:
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -tags lambda.norpc -o $(ARTIFACTS_DIR)/bootstrap ./cmd/lambdas/composer

# Build all Lambda functions using SAM
build:
	sam build

# Fast local build (without SAM)
build-fast:
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -tags lambda.norpc -o build/parser/bootstrap ./cmd/lambdas/parser
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -tags lambda.norpc -o build/generator/bootstrap ./cmd/lambdas/generator
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -tags lambda.norpc -o build/audio-generator/bootstrap ./cmd/lambdas/audio-generator
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -tags lambda.norpc -o build/composer/bootstrap ./cmd/lambdas/composer

# Clean build artifacts
clean:
	rm -rf .aws-sam

# Deploy to AWS
deploy:
	sam deploy --guided

# Deploy with no prompts (CI/CD)
deploy-ci:
	sam deploy \
		--no-confirm-changeset \
		--no-fail-on-empty-changeset \
		--stack-name omnigen-backend \
		--capabilities CAPABILITY_IAM \
		--parameter-overrides Environment=production

# Run tests
test:
	go test ./...

# Test with coverage
test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

# Local invoke
local-invoke-parser:
	sam local invoke ParserFunction -e events/parser-event.json

# Start local API
local-api:
	sam local start-api

# Tail logs from deployed functions
logs-parser:
	sam logs -n ParserFunction --stack-name omnigen-backend --tail

# Validate SAM template
validate:
	sam validate

# Package for deployment
package:
	sam package \
		--output-template-file packaged.yaml \
		--s3-bucket omnigen-deploy-artifacts

# Quick build and deploy parser lambda (Terraform-managed)
deploy-parser:
	@echo "Building parser Lambda..."
	@mkdir -p build/parser
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -tags lambda.norpc -o build/parser/bootstrap ./cmd/lambdas/parser
	@echo "Packaging Lambda..."
	cd build/parser && zip -q bootstrap.zip bootstrap
	@echo "Deploying to AWS Lambda..."
	aws lambda update-function-code \
		--function-name omnigen-parser \
		--zip-file fileb://build/parser/bootstrap.zip \
		--region us-east-1 \
		--no-cli-pager
	@echo "Parser Lambda deployed successfully!"

# Build parser only (no deploy)
build-parser:
	@echo "Building parser Lambda..."
	@mkdir -p build/parser
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -tags lambda.norpc -o build/parser/bootstrap ./cmd/lambdas/parser
	cd build/parser && zip -q bootstrap.zip bootstrap
	@echo "Parser Lambda built: build/parser/bootstrap.zip"

# Quick build and deploy generator lambda (Terraform-managed)
deploy-generator:
	@echo "Building generator Lambda..."
	@mkdir -p build/generator
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -tags lambda.norpc -o build/generator/bootstrap ./cmd/lambdas/generator
	@echo "Packaging Lambda..."
	cd build/generator && zip -q bootstrap.zip bootstrap
	@echo "Deploying to AWS Lambda..."
	aws lambda update-function-code \
		--function-name omnigen-generator \
		--zip-file fileb://build/generator/bootstrap.zip \
		--region us-east-1 \
		--no-cli-pager
	@echo "Generator Lambda deployed successfully!"

# Build generator only (no deploy)
build-generator:
	@echo "Building generator Lambda..."
	@mkdir -p build/generator
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -tags lambda.norpc -o build/generator/bootstrap ./cmd/lambdas/generator
	cd build/generator && zip -q bootstrap.zip bootstrap
	@echo "Generator Lambda built: build/generator/bootstrap.zip"

# Quick build and deploy audio-generator lambda (Terraform-managed)
deploy-audio-generator:
	@echo "Building audio-generator Lambda..."
	@mkdir -p build/audio-generator
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -tags lambda.norpc -o build/audio-generator/bootstrap ./cmd/lambdas/audio-generator
	@echo "Packaging Lambda..."
	cd build/audio-generator && zip -q bootstrap.zip bootstrap
	@echo "Deploying to AWS Lambda..."
	aws lambda update-function-code \
		--function-name omnigen-audio-generator \
		--zip-file fileb://build/audio-generator/bootstrap.zip \
		--region us-east-1 \
		--no-cli-pager
	@echo "Audio-generator Lambda deployed successfully!"

# Build audio-generator only (no deploy)
build-audio-generator:
	@echo "Building audio-generator Lambda..."
	@mkdir -p build/audio-generator
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -tags lambda.norpc -o build/audio-generator/bootstrap ./cmd/lambdas/audio-generator
	cd build/audio-generator && zip -q bootstrap.zip bootstrap
	@echo "Audio-generator Lambda built: build/audio-generator/bootstrap.zip"

# Quick build and deploy composer lambda (Terraform-managed)
deploy-composer:
	@echo "Building composer Lambda..."
	@mkdir -p build/composer
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -tags lambda.norpc -o build/composer/bootstrap ./cmd/lambdas/composer
	@echo "Packaging Lambda..."
	cd build/composer && zip -q bootstrap.zip bootstrap
	@echo "Deploying to AWS Lambda..."
	aws lambda update-function-code \
		--function-name omnigen-composer \
		--zip-file fileb://build/composer/bootstrap.zip \
		--region us-east-1 \
		--no-cli-pager
	@echo "Composer Lambda deployed successfully!"

# Build composer only (no deploy)
build-composer:
	@echo "Building composer Lambda..."
	@mkdir -p build/composer
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -tags lambda.norpc -o build/composer/bootstrap ./cmd/lambdas/composer
	cd build/composer && zip -q bootstrap.zip bootstrap
	@echo "Composer Lambda built: build/composer/bootstrap.zip"

# Build all lambdas (no deploy)
build-lambdas: build-parser build-generator build-audio-generator build-composer
	@echo "All Lambda functions built successfully!"

# Deploy all lambdas
deploy-lambdas: deploy-parser deploy-generator deploy-audio-generator deploy-composer
	@echo "All Lambda functions deployed successfully!"

# Generate/update Swagger documentation
swagger:
	@echo "Generating Swagger documentation..."
	@swag init -g cmd/api/main.go --output docs
	@echo "Swagger docs generated successfully in docs/"
	@echo "  - docs/swagger.json"
	@echo "  - docs/swagger.yaml"
	@echo "  - docs/docs.go"
