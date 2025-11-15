variable "project_name" {
  description = "Project name for resource naming"
  type        = string
  default     = "omnigen"
}

variable "aws_region" {
  description = "AWS region for resources"
  type        = string
  default     = "us-east-1"
}

variable "github_org" {
  description = "GitHub organization or user name"
  type        = string
  default     = "ppiont"
}

variable "github_repo" {
  description = "GitHub repository name"
  type        = string
  default     = "omnigen"
}
