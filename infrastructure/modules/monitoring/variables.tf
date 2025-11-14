variable "project_name" {
  description = "Project name for resource naming"
  type        = string
}

variable "log_retention_days" {
  description = "Number of days to retain logs in CloudWatch"
  type        = number
  default     = 7
}

variable "ecs_log_group_name" {
  description = "Name for ECS log group"
  type        = string
}

variable "lambda_generator_log_group" {
  description = "Name for Lambda generator log group"
  type        = string
}

variable "lambda_composer_log_group" {
  description = "Name for Lambda composer log group"
  type        = string
}

variable "step_functions_log_group" {
  description = "Name for Step Functions log group"
  type        = string
}
