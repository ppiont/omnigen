variable "aws_region" {
  description = "AWS region for all resources"
  type        = string
  default     = "us-east-1"

  validation {
    condition     = can(regex("^[a-z]{2}-[a-z]+-[0-9]{1}$", var.aws_region))
    error_message = "AWS region must be in format: us-east-1, eu-west-1, etc."
  }
}

variable "project_name" {
  description = "Project name used for resource naming"
  type        = string
  default     = "omnigen"

  validation {
    condition     = can(regex("^[a-z][a-z0-9-]{1,20}$", var.project_name))
    error_message = "Project name must be lowercase alphanumeric with hyphens, 2-21 chars."
  }
}

variable "availability_zone" {
  description = "Single availability zone for cost optimization"
  type        = string
  default     = "us-east-1a"
}

variable "vpc_cidr" {
  description = "CIDR block for VPC"
  type        = string
  default     = "10.0.0.0/16"
}

variable "public_subnet_cidr" {
  description = "CIDR block for public subnet"
  type        = string
  default     = "10.0.1.0/24"
}

variable "private_subnet_cidr" {
  description = "CIDR block for private subnet"
  type        = string
  default     = "10.0.10.0/24"
}

variable "replicate_api_key_secret_arn" {
  description = "ARN of AWS Secrets Manager secret containing Replicate API key"
  type        = string

  validation {
    condition     = can(regex("^arn:aws:secretsmanager:", var.replicate_api_key_secret_arn))
    error_message = "Must be a valid Secrets Manager ARN."
  }
}

variable "ecs_min_tasks" {
  description = "Minimum number of ECS tasks"
  type        = number
  default     = 1

  validation {
    condition     = var.ecs_min_tasks >= 1 && var.ecs_min_tasks <= 10
    error_message = "ECS min tasks must be between 1 and 10."
  }
}

variable "ecs_max_tasks" {
  description = "Maximum number of ECS tasks"
  type        = number
  default     = 5

  validation {
    condition     = var.ecs_max_tasks >= 1 && var.ecs_max_tasks <= 20
    error_message = "ECS max tasks must be between 1 and 20."
  }
}

variable "ecs_cpu" {
  description = "CPU units for ECS tasks (1024 = 1 vCPU)"
  type        = number
  default     = 1024

  validation {
    condition     = contains([256, 512, 1024, 2048, 4096], var.ecs_cpu)
    error_message = "ECS CPU must be 256, 512, 1024, 2048, or 4096."
  }
}

variable "ecs_memory" {
  description = "Memory for ECS tasks in MB"
  type        = number
  default     = 2048

  validation {
    condition     = contains([512, 1024, 2048, 3072, 4096, 5120, 6144, 7168, 8192], var.ecs_memory)
    error_message = "ECS memory must be valid for selected CPU."
  }
}

variable "ecs_target_cpu_utilization" {
  description = "Target CPU utilization percentage for auto-scaling"
  type        = number
  default     = 70

  validation {
    condition     = var.ecs_target_cpu_utilization >= 30 && var.ecs_target_cpu_utilization <= 90
    error_message = "Target CPU utilization must be between 30 and 90."
  }
}

variable "lambda_generator_memory" {
  description = "Memory allocation for generator Lambda in MB"
  type        = number
  default     = 2048

  validation {
    condition     = var.lambda_generator_memory >= 128 && var.lambda_generator_memory <= 10240
    error_message = "Lambda memory must be between 128 and 10240 MB."
  }
}

variable "lambda_composer_memory" {
  description = "Memory allocation for composer Lambda in MB"
  type        = number
  default     = 10240

  validation {
    condition     = var.lambda_composer_memory >= 128 && var.lambda_composer_memory <= 10240
    error_message = "Lambda memory must be between 128 and 10240 MB."
  }
}

variable "lambda_timeout" {
  description = "Timeout for Lambda functions in seconds"
  type        = number
  default     = 900

  validation {
    condition     = var.lambda_timeout >= 60 && var.lambda_timeout <= 900
    error_message = "Lambda timeout must be between 60 and 900 seconds."
  }
}

variable "lambda_generator_concurrency" {
  description = "Reserved concurrent executions for generator Lambda"
  type        = number
  default     = 10

  validation {
    condition     = var.lambda_generator_concurrency >= 1 && var.lambda_generator_concurrency <= 100
    error_message = "Concurrency must be between 1 and 100."
  }
}

variable "lambda_composer_concurrency" {
  description = "Reserved concurrent executions for composer Lambda"
  type        = number
  default     = 5

  validation {
    condition     = var.lambda_composer_concurrency >= 1 && var.lambda_composer_concurrency <= 100
    error_message = "Concurrency must be between 1 and 100."
  }
}

variable "cloudwatch_log_retention_days" {
  description = "CloudWatch Logs retention period in days"
  type        = number
  default     = 7

  validation {
    condition     = contains([1, 3, 5, 7, 14, 30, 60, 90, 120, 150, 180, 365, 400, 545, 731, 1827, 3653], var.cloudwatch_log_retention_days)
    error_message = "Must be a valid CloudWatch Logs retention period."
  }
}

variable "s3_assets_lifecycle_ia_days" {
  description = "Days before transitioning S3 assets to Infrequent Access"
  type        = number
  default     = 30
}

variable "s3_assets_lifecycle_glacier_days" {
  description = "Days before transitioning S3 assets to Glacier"
  type        = number
  default     = 90
}

variable "s3_assets_lifecycle_expiration_days" {
  description = "Days before deleting S3 assets"
  type        = number
  default     = 365
}

variable "dynamodb_ttl_days" {
  description = "Days before DynamoDB items expire via TTL"
  type        = number
  default     = 90
}

variable "dynamodb_point_in_time_recovery" {
  description = "Enable point-in-time recovery for DynamoDB"
  type        = bool
  default     = true
}

variable "cloudfront_price_class" {
  description = "CloudFront distribution price class"
  type        = string
  default     = "PriceClass_100"

  validation {
    condition     = contains(["PriceClass_All", "PriceClass_200", "PriceClass_100"], var.cloudfront_price_class)
    error_message = "Must be PriceClass_All, PriceClass_200, or PriceClass_100."
  }
}

variable "enable_nat_gateway" {
  description = "Enable NAT Gateway for private subnet internet access"
  type        = bool
  default     = true
}
