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

variable "step_functions_arn" {
  description = "ARN of the Step Functions state machine"
  type        = string
}

variable "replicate_secret_arn" {
  description = "ARN of the Replicate API key secret"
  type        = string
}

variable "ecr_repository_arn" {
  description = "ARN of the ECR repository"
  type        = string
}

variable "lambda_generator_function_arn" {
  description = "ARN of the generator Lambda function"
  type        = string
}

variable "lambda_composer_function_arn" {
  description = "ARN of the composer Lambda function"
  type        = string
}
