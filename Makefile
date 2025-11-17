.PHONY: help init plan apply destroy format validate outputs docker-build docker-push deploy-ecs deploy-frontend logs-ecs clean swagger

# Default target - show help when running 'make' with no arguments
.DEFAULT_GOAL := help

# Variables
AWS_REGION ?= us-east-1
PROJECT_NAME ?= omnigen
INFRA_DIR := infrastructure

# Colors for output
CYAN := \033[0;36m
GREEN := \033[0;32m
YELLOW := \033[1;33m
RED := \033[0;31m
NC := \033[0m # No Color

help: ## Show this help message
	@printf '${CYAN}╔══════════════════════════════════════════╗${NC}\n'
	@printf '${CYAN}║${NC}  ${YELLOW}OmniGen AI Video Generation Pipeline${NC}  ${CYAN}║${NC}\n'
	@printf '${CYAN}╚══════════════════════════════════════════╝${NC}\n'
	@printf '\n'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "${GREEN}%-20s${NC} %s\n", $$1, $$2}'
	@printf '\n'
	@printf 'Usage: ${CYAN}make <target>${NC}\n'
	@printf 'Example: ${CYAN}make init${NC} or ${CYAN}make deploy-all${NC}\n'

# Terraform commands

init: ## Initialize Terraform
	@printf '${CYAN}Initializing Terraform...${NC}\n'
	@cd $(INFRA_DIR) && terraform init

plan: ## Run terraform plan
	@printf '${CYAN}Planning infrastructure changes...${NC}\n'
	@cd $(INFRA_DIR) && terraform plan -out=tfplan

apply: ## Apply terraform changes
	@printf '${CYAN}Applying infrastructure changes...${NC}\n'
	@cd $(INFRA_DIR) && terraform apply tfplan
	@printf '${GREEN}Infrastructure deployed successfully!${NC}\n'
	@printf '\n'
	@make outputs

destroy: ## Destroy all infrastructure (WARNING: destructive)
	@printf '${RED}WARNING: This will destroy all infrastructure!${NC}\n'
	@printf 'Press Ctrl+C to cancel, or wait 5 seconds to continue...\n'
	@sleep 5
	@printf '${YELLOW}Emptying S3 buckets first...${NC}\n'
	-@aws s3 rm s3://${PROJECT_NAME}-assets-$$(aws sts get-caller-identity --query Account --output text) --recursive 2>/dev/null || true
	-@aws s3 rm s3://${PROJECT_NAME}-frontend-$$(aws sts get-caller-identity --query Account --output text) --recursive 2>/dev/null || true
	@cd $(INFRA_DIR) && terraform destroy

format: ## Format Terraform files
	@printf '${CYAN}Formatting Terraform files...${NC}\n'
	@cd $(INFRA_DIR) && terraform fmt -recursive
	@printf '${GREEN}Files formatted successfully!${NC}\n'

validate: ## Validate Terraform configuration
	@printf '${CYAN}Validating Terraform configuration...${NC}\n'
	@cd $(INFRA_DIR) && terraform validate
	@printf '${GREEN}Configuration is valid!${NC}\n'

outputs: ## Show terraform outputs
	@printf '${CYAN}Terraform Outputs:${NC}\n'
	@cd $(INFRA_DIR) && terraform output

# Docker commands

docker-build: ## Build Docker image for ECS (ARM64 Graviton)
	@printf '${CYAN}Building Docker image for linux/arm64 (Graviton)...${NC}\n'
	@cd backend && docker buildx build --platform linux/arm64 -t ${PROJECT_NAME}-api:latest --load .
	@printf '${GREEN}Docker image built successfully!${NC}\n'

docker-push: ## Build and push Docker image to ECR (ARM64 Graviton)
	@printf '${CYAN}Building and pushing Docker image to ECR (linux/arm64 Graviton)...${NC}\n'
	$(eval ECR_URL := $(shell cd $(INFRA_DIR) && terraform output -raw ecr_repository_url 2>/dev/null))
	@if [ -z "$(ECR_URL)" ]; then \
		printf '${RED}Error: Could not get ECR repository URL. Run terraform apply first.${NC}\n'; \
		exit 1; \
	fi
	@printf "ECR Repository: $(ECR_URL)\n"
	@printf '${CYAN}Logging in to ECR...${NC}\n'
	@aws ecr get-login-password --region ${AWS_REGION} | docker login --username AWS --password-stdin $(ECR_URL)
	@printf '${CYAN}Building image for linux/arm64 (Graviton)...${NC}\n'
	@cd backend && docker buildx build --platform linux/arm64 -t ${PROJECT_NAME}-api:latest --load .
	@printf '${CYAN}Tagging image...${NC}\n'
	@docker tag ${PROJECT_NAME}-api:latest $(ECR_URL):latest
	@printf '${CYAN}Pushing image...${NC}\n'
	@docker push $(ECR_URL):latest
	@printf '${GREEN}Docker image pushed successfully!${NC}\n'

deploy-ecs: docker-push ## Deploy new version to ECS (builds and pushes Docker image)
	@printf '${CYAN}Deploying new version to ECS...${NC}\n'
	$(eval CLUSTER := $(shell cd $(INFRA_DIR) && terraform output -raw ecs_cluster_name 2>/dev/null))
	$(eval SERVICE := $(shell cd $(INFRA_DIR) && terraform output -raw ecs_service_name 2>/dev/null))
	@aws ecs update-service \
		--cluster $(CLUSTER) \
		--service $(SERVICE) \
		--force-new-deployment \
		--region ${AWS_REGION} \
		> /dev/null
	@printf '${GREEN}ECS service update initiated!${NC}\n'
	@printf 'Monitor deployment with: make logs-ecs\n'

# Frontend commands

frontend-install: ## Install frontend dependencies with Bun
	@printf '${CYAN}Installing frontend dependencies with Bun...${NC}\n'
	@cd frontend && bun install
	@printf '${GREEN}Dependencies installed successfully!${NC}\n'

frontend-env: ## Generate frontend .env file with production values from Terraform
	@printf '${CYAN}Generating frontend .env file from Terraform outputs...${NC}\n'
	$(eval CF_DOMAIN := $(shell cd $(INFRA_DIR) && terraform output -raw cloudfront_domain_name 2>/dev/null))
	$(eval USER_POOL := $(shell cd $(INFRA_DIR) && terraform output -raw auth_user_pool_id 2>/dev/null))
	$(eval CLIENT_ID := $(shell cd $(INFRA_DIR) && terraform output -raw auth_client_id 2>/dev/null))
	$(eval COGNITO_DOMAIN := $(shell cd $(INFRA_DIR) && terraform output -raw auth_hosted_ui_domain 2>/dev/null))
	@if [ -z "$(CF_DOMAIN)" ] || [ -z "$(USER_POOL)" ] || [ -z "$(CLIENT_ID)" ] || [ -z "$(COGNITO_DOMAIN)" ]; then \
		printf '${RED}Error: Could not get all required values from Terraform. Run terraform apply first.${NC}\n'; \
		exit 1; \
	fi
	@printf '# Production environment variables (auto-generated from Terraform)\n' > frontend/.env
	@printf 'VITE_API_URL=https://$(CF_DOMAIN)\n' >> frontend/.env
	@printf 'VITE_COGNITO_USER_POOL_ID=$(USER_POOL)\n' >> frontend/.env
	@printf 'VITE_COGNITO_CLIENT_ID=$(CLIENT_ID)\n' >> frontend/.env
	@printf 'VITE_COGNITO_DOMAIN=https://$(COGNITO_DOMAIN)\n' >> frontend/.env
	@printf '${GREEN}Frontend .env file generated successfully!${NC}\n'

frontend-dev: ## Run frontend development server
	@printf '${CYAN}Starting frontend dev server...${NC}\n'
	@cd frontend && bun run dev

frontend-build: ## Build frontend for production
	@printf '${CYAN}Building frontend...${NC}\n'
	$(eval CF_DOMAIN := $(shell cd $(INFRA_DIR) && terraform output -raw cloudfront_domain_name 2>/dev/null))
	$(eval USER_POOL := $(shell cd $(INFRA_DIR) && terraform output -raw auth_user_pool_id 2>/dev/null))
	$(eval CLIENT_ID := $(shell cd $(INFRA_DIR) && terraform output -raw auth_client_id 2>/dev/null))
	$(eval COGNITO_DOMAIN := $(shell cd $(INFRA_DIR) && terraform output -raw auth_hosted_ui_domain 2>/dev/null))
	@if [ -z "$(CF_DOMAIN)" ]; then \
		printf '${RED}Error: Could not get CloudFront domain. Run terraform apply first.${NC}\n'; \
		exit 1; \
	fi
	@cd frontend && \
		VITE_API_URL=https://$(CF_DOMAIN) \
		VITE_COGNITO_USER_POOL_ID=$(USER_POOL) \
		VITE_COGNITO_CLIENT_ID=$(CLIENT_ID) \
		VITE_COGNITO_DOMAIN=$(COGNITO_DOMAIN) \
		bun run build
	@printf '${GREEN}Frontend built successfully!${NC}\n'

deploy-frontend: frontend-build ## Build and deploy frontend to S3 and invalidate CloudFront
	@printf '${CYAN}Deploying frontend...${NC}\n'
	$(eval BUCKET := $(shell cd $(INFRA_DIR) && terraform output -raw frontend_bucket_name 2>/dev/null))
	$(eval CF_ID := $(shell cd $(INFRA_DIR) && terraform output -raw cloudfront_distribution_id 2>/dev/null))
	@if [ -z "$(BUCKET)" ] || [ -z "$(CF_ID)" ]; then \
		printf '${RED}Error: Could not get bucket/CloudFront info. Run terraform apply first.${NC}\n'; \
		exit 1; \
	fi
	@printf '${CYAN}Syncing to S3...${NC}\n'
	@aws s3 sync frontend/dist/ s3://$(BUCKET)/ --delete
	@printf '${CYAN}Invalidating CloudFront cache...${NC}\n'
	@aws cloudfront create-invalidation \
		--distribution-id $(CF_ID) \
		--paths "/*" \
		> /dev/null
	@printf '${GREEN}Frontend deployed successfully!${NC}\n'
	@printf 'URL: https://%s\n' "$$(cd $(INFRA_DIR) && terraform output -raw cloudfront_domain_name)"

# Logging commands

logs-ecs: ## Tail ECS logs
	@printf '${CYAN}Tailing ECS logs (Ctrl+C to stop)...${NC}\n'
	$(eval LOG_GROUP := $(shell cd $(INFRA_DIR) && terraform output -raw ecs_log_group_name 2>/dev/null))
	@aws logs tail $(LOG_GROUP) --follow --region ${AWS_REGION}

# Health check

health: ## Check API health
	@printf '${CYAN}Checking API health...${NC}\n'
	$(eval API_URL := $(shell cd $(INFRA_DIR) && terraform output -raw alb_dns_name 2>/dev/null))
	@curl -s http://$(API_URL)/health | jq . || printf '${RED}API not responding${NC}\n'

# Documentation

swagger: ## Generate Swagger API documentation
	@printf '${CYAN}Generating Swagger documentation...${NC}\n'
	@cd backend && swag init -g cmd/api/main.go --output docs
	@printf '${GREEN}Swagger docs generated successfully in backend/docs/${NC}\n'

# Cleanup

clean: ## Clean up local Terraform files
	@printf '${CYAN}Cleaning up local files...${NC}\n'
	@cd $(INFRA_DIR) && rm -rf .terraform
	@cd $(INFRA_DIR) && rm -f .terraform.lock.hcl
	@cd $(INFRA_DIR) && rm -f tfplan
	@cd $(INFRA_DIR) && rm -f terraform.tfstate.backup
	@cd $(INFRA_DIR) && find . -name "*.zip" -type f -delete
	@printf '${GREEN}Cleanup complete!${NC}\n'

# Quick setup

setup: ## Initial setup (copy example config)
	@if [ ! -f $(INFRA_DIR)/terraform.tfvars ]; then \
		printf '${CYAN}Creating terraform.tfvars from example...${NC}\n'; \
		cp $(INFRA_DIR)/terraform.tfvars.example $(INFRA_DIR)/terraform.tfvars; \
		printf '${YELLOW}Please edit $(INFRA_DIR)/terraform.tfvars with your values${NC}\n'; \
	else \
		printf '${GREEN}terraform.tfvars already exists${NC}\n'; \
	fi

# Full deployment

deploy-all: apply docker-push deploy-ecs deploy-frontend ## Deploy infrastructure, backend, and frontend
	@printf '${GREEN}Full deployment complete!${NC}\n'
	@printf '\n'
	@make outputs
