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
