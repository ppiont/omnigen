variable "project_name" {
  description = "Project name for resource naming"
  type        = string
}

variable "assets_lifecycle_ia_days" {
  description = "Days before transitioning assets to Infrequent Access"
  type        = number
  default     = 30
}

variable "assets_lifecycle_glacier_days" {
  description = "Days before transitioning assets to Glacier Instant Retrieval"
  type        = number
  default     = 90
}

variable "assets_lifecycle_expiration_days" {
  description = "Days before expiring/deleting assets"
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

variable "cloudfront_distribution_arn" {
  description = "ARN of CloudFront distribution for S3 bucket policy"
  type        = string
}
