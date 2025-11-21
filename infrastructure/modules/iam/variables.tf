variable "project_name" {
  description = "Project name for resource naming"
  type        = string
}

variable "assets_bucket_arn" {
  description = "ARN of the assets S3 bucket"
  type        = string
}

variable "frontend_bucket_arn" {
  description = "ARN of the frontend S3 bucket"
  type        = string
}

variable "dynamodb_table_arn" {
  description = "ARN of the DynamoDB jobs table"
  type        = string
}

variable "dynamodb_usage_table_arn" {
  description = "ARN of the DynamoDB usage table"
  type        = string
}

variable "replicate_secret_arn" {
  description = "ARN of the Replicate API key secret"
  type        = string
}

variable "openai_secret_arn" {
  description = "ARN of the OpenAI API key secret (optional)"
  type        = string
  default     = ""
}

variable "ecr_repository_arn" {
  description = "ARN of the ECR repository"
  type        = string
}
