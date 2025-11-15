variable "project_name" {
  description = "Project name for resource naming"
  type        = string
}

variable "cloudfront_domain" {
  description = "CloudFront domain name for OAuth callback URLs"
  type        = string
}
