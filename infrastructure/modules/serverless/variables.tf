variable "project_name" {
  description = "Project name for resource naming"
  type        = string
}

variable "vpc_id" {
  description = "VPC ID"
  type        = string
}

variable "private_subnet_ids" {
  description = "List of private subnet IDs for Lambda functions"
  type        = list(string)
}

variable "lambda_security_group_id" {
  description = "Security group ID for Lambda functions"
  type        = string
}

variable "lambda_execution_role_arn" {
  description = "ARN of the Lambda execution role"
  type        = string
}

variable "step_functions_role_arn" {
  description = "ARN of the Step Functions execution role"
  type        = string
}

variable "generator_memory" {
  description = "Memory allocation for generator Lambda in MB"
  type        = number
}

variable "composer_memory" {
  description = "Memory allocation for composer Lambda in MB"
  type        = number
}

variable "timeout" {
  description = "Timeout for Lambda functions in seconds"
  type        = number
}

variable "generator_concurrency" {
  description = "Reserved concurrent executions for generator Lambda"
  type        = number
}

variable "composer_concurrency" {
  description = "Reserved concurrent executions for composer Lambda"
  type        = number
}

variable "audio_generator_memory" {
  description = "Memory allocation for audio generator Lambda in MB"
  type        = number
}

variable "audio_generator_concurrency" {
  description = "Reserved concurrent executions for audio generator Lambda"
  type        = number
}

variable "assets_bucket_name" {
  description = "Name of the assets S3 bucket"
  type        = string
}

variable "dynamodb_table_name" {
  description = "Name of the DynamoDB table"
  type        = string
}

variable "replicate_secret_arn" {
  description = "ARN of the Replicate API key secret"
  type        = string
}

variable "aws_region" {
  description = "AWS region"
  type        = string
}

variable "generator_log_group_name" {
  description = "CloudWatch log group name for generator Lambda"
  type        = string
}

variable "composer_log_group_name" {
  description = "CloudWatch log group name for composer Lambda"
  type        = string
}

variable "audio_generator_log_group_name" {
  description = "CloudWatch log group name for audio generator Lambda"
  type        = string
}

variable "step_functions_log_group_name" {
  description = "CloudWatch log group name for Step Functions"
  type        = string
}

variable "step_functions_log_group_arn" {
  description = "ARN of Step Functions CloudWatch log group"
  type        = string
}

variable "scripts_table_name" {
  description = "Name of the DynamoDB scripts table"
  type        = string
}

variable "parser_log_group_name" {
  description = "CloudWatch log group name for parser Lambda"
  type        = string
}
