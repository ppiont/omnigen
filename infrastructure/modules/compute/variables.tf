variable "project_name" {
  description = "Project name for resource naming"
  type        = string
}

variable "vpc_id" {
  description = "VPC ID"
  type        = string
}

variable "private_subnet_ids" {
  description = "List of private subnet IDs for ECS tasks"
  type        = list(string)
}

variable "ecs_security_group_id" {
  description = "Security group ID for ECS tasks"
  type        = string
}

variable "alb_target_group_arn" {
  description = "ARN of the ALB target group"
  type        = string
}

variable "task_execution_role_arn" {
  description = "ARN of the ECS task execution role"
  type        = string
}

variable "task_role_arn" {
  description = "ARN of the ECS task role"
  type        = string
}

variable "cpu" {
  description = "CPU units for the task (1024 = 1 vCPU)"
  type        = number
}

variable "memory" {
  description = "Memory for the task in MB"
  type        = number
}

variable "min_tasks" {
  description = "Minimum number of tasks"
  type        = number
}

variable "max_tasks" {
  description = "Maximum number of tasks"
  type        = number
}

variable "target_cpu_utilization" {
  description = "Target CPU utilization percentage for auto-scaling"
  type        = number
}

variable "container_name" {
  description = "Name of the container"
  type        = string
}

variable "container_port" {
  description = "Port on which the container listens"
  type        = number
}

variable "log_group_name" {
  description = "CloudWatch log group name"
  type        = string
}

variable "environment" {
  description = "Environment name (development, staging, production)"
  type        = string
}

variable "aws_region" {
  description = "AWS region"
  type        = string
}

variable "assets_bucket_name" {
  description = "Name of the assets S3 bucket"
  type        = string
}

variable "dynamodb_table_name" {
  description = "Name of the DynamoDB jobs table"
  type        = string
}

variable "dynamodb_usage_table_name" {
  description = "Name of the DynamoDB usage table"
  type        = string
}

variable "replicate_secret_arn" {
  description = "ARN of the Replicate API key secret"
  type        = string
}

variable "cognito_user_pool_id" {
  description = "Cognito User Pool ID for authentication"
  type        = string
}

variable "cognito_client_id" {
  description = "Cognito User Pool Client ID"
  type        = string
}

variable "jwt_issuer" {
  description = "JWT issuer URL for token validation"
  type        = string
}

variable "cognito_domain" {
  description = "Cognito hosted UI domain for CORS configuration"
  type        = string
}

variable "cloudfront_domain" {
  description = "CloudFront domain for CORS configuration"
  type        = string
}

variable "video_adapter_type" {
  description = "Video generation adapter type (veo, kling)"
  type        = string
  default     = "veo"
}

variable "veo_generate_audio" {
  description = "Enable native audio generation for Veo (true/false)"
  type        = string
  default     = "false"
}
