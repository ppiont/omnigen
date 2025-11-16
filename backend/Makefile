.PHONY: build clean deploy test local-invoke build-ParserFunction build-GeneratorFunction build-AudioGeneratorFunction build-ComposerFunction deploy-parser build-parser

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
		--region us-east-1
	@echo "Parser Lambda deployed successfully!"

# Build parser only (no deploy)
build-parser:
	@echo "Building parser Lambda..."
	@mkdir -p build/parser
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -tags lambda.norpc -o build/parser/bootstrap ./cmd/lambdas/parser
	cd build/parser && zip -q bootstrap.zip bootstrap
	@echo "Parser Lambda built: build/parser/bootstrap.zip"
