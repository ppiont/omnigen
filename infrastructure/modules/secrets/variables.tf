variable "project_name" {
  description = "Project name for resource naming"
  type        = string
}

variable "replicate_api_key_secret_arn" {
  description = "ARN of the Replicate API key secret in Secrets Manager"
  type        = string
}
