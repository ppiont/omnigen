# Infrastructure Outputs

# API Endpoints
output "api_url" {
  description = "Application Load Balancer DNS name for API access"
  value       = "http://${module.loadbalancer.alb_dns_name}"
}

output "alb_dns_name" {
  description = "ALB DNS name (without protocol)"
  value       = module.loadbalancer.alb_dns_name
}

# Frontend URL
output "frontend_url" {
  description = "CloudFront distribution URL for frontend access"
  value       = "https://${module.cdn.cloudfront_domain_name}"
}

output "cloudfront_domain_name" {
  description = "CloudFront domain name (without protocol)"
  value       = module.cdn.cloudfront_domain_name
}

output "cloudfront_distribution_id" {
  description = "CloudFront distribution ID for cache invalidation"
  value       = module.cdn.cloudfront_distribution_id
}

# Storage
output "assets_bucket_name" {
  description = "S3 bucket name for video assets"
  value       = module.storage.assets_bucket_name
}

output "frontend_bucket_name" {
  description = "S3 bucket name for frontend static files"
  value       = module.storage.frontend_bucket_name
}

output "dynamodb_table_name" {
  description = "DynamoDB table name for job tracking"
  value       = module.storage.dynamodb_table_name
}

# Compute
output "ecs_cluster_name" {
  description = "ECS cluster name"
  value       = module.compute.ecs_cluster_name
}

output "ecs_service_name" {
  description = "ECS service name"
  value       = module.compute.ecs_service_name
}

output "ecr_repository_url" {
  description = "ECR repository URL for Docker image push"
  value       = module.compute.ecr_repository_url
}

# Networking
output "vpc_id" {
  description = "VPC ID"
  value       = module.networking.vpc_id
}

output "public_subnet_ids" {
  description = "List of public subnet IDs (multi-AZ for ALB)"
  value       = module.networking.public_subnet_ids
}

output "public_subnet_id" {
  description = "First public subnet ID (backwards compatibility)"
  value       = module.networking.public_subnet_id
}

output "private_subnet_id" {
  description = "Private subnet ID"
  value       = module.networking.private_subnet_id
}

output "nat_gateway_id" {
  description = "NAT Gateway ID"
  value       = module.networking.nat_gateway_id
}

# IAM Roles
output "ecs_task_role_arn" {
  description = "ECS task role ARN"
  value       = module.iam.ecs_task_role_arn
}

# CloudWatch
output "ecs_log_group_name" {
  description = "ECS CloudWatch log group name"
  value       = module.monitoring.ecs_log_group_name
}

# Authentication (Cognito)
output "auth_user_pool_id" {
  description = "Cognito User Pool ID for frontend configuration"
  value       = module.auth.user_pool_id
}

output "auth_client_id" {
  description = "Cognito User Pool Client ID for frontend configuration"
  value       = module.auth.client_id
}

output "auth_hosted_ui_domain" {
  description = "Cognito Hosted UI domain URL"
  value       = module.auth.hosted_ui_domain
}

output "auth_issuer_url" {
  description = "JWT issuer URL for token validation"
  value       = module.auth.issuer_url
}

output "auth_jwks_uri" {
  description = "JWKS URI for JWT validation"
  value       = module.auth.jwks_uri
}

# Frontend Environment Variables
output "frontend_env_vars" {
  description = "Environment variables for frontend .env file"
  value       = <<-EOT

    📋 Frontend Environment Variables (.env)
    Copy these values to frontend/.env:

    VITE_API_URL=http://${module.loadbalancer.alb_dns_name}
    VITE_COGNITO_USER_POOL_ID=${module.auth.user_pool_id}
    VITE_COGNITO_CLIENT_ID=${module.auth.client_id}
    VITE_COGNITO_DOMAIN=${module.auth.hosted_ui_domain}

    For production, replace VITE_API_URL with your CloudFront domain:
    VITE_API_URL=https://${module.cdn.cloudfront_domain_name}
  EOT
}

# Quick Start Commands
output "quick_start_commands" {
  description = "Quick start commands for deployment"
  value       = <<-EOT

    🚀 OmniGen Infrastructure Deployed Successfully!

    📋 Next Steps:

    1. Build and push Docker image:
       aws ecr get-login-password --region ${var.aws_region} | docker login --username AWS --password-stdin ${module.compute.ecr_repository_url}
       docker build -t ${local.name_prefix}-api ../backend
       docker tag ${local.name_prefix}-api:latest ${module.compute.ecr_repository_url}:latest
       docker push ${module.compute.ecr_repository_url}:latest

    2. Deploy ECS service (after Docker push):
       aws ecs update-service --cluster ${module.compute.ecs_cluster_name} --service ${module.compute.ecs_service_name} --force-new-deployment --region ${var.aws_region}

    3. Build and deploy frontend:
       cd ../frontend
       npm run build
       aws s3 sync dist/ s3://${module.storage.frontend_bucket_name}/
       aws cloudfront create-invalidation --distribution-id ${module.cdn.cloudfront_distribution_id} --paths "/*"

    4. Test API:
       curl ${module.loadbalancer.alb_dns_name}/health

    5. Access frontend:
       open https://${module.cdn.cloudfront_domain_name}

    📊 Resources:
    - API URL: http://${module.loadbalancer.alb_dns_name}
    - Frontend URL: https://${module.cdn.cloudfront_domain_name}
    - Assets Bucket: s3://${module.storage.assets_bucket_name}
    - Frontend Bucket: s3://${module.storage.frontend_bucket_name}

    📝 View Logs:
    - ECS: aws logs tail ${module.monitoring.ecs_log_group_name} --follow
  EOT
}
