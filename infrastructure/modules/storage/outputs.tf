output "assets_bucket_name" {
  description = "Name of the assets S3 bucket"
  value       = aws_s3_bucket.assets.id
}

output "assets_bucket_arn" {
  description = "ARN of the assets S3 bucket"
  value       = aws_s3_bucket.assets.arn
}

output "assets_bucket_regional_domain_name" {
  description = "Regional domain name of the assets S3 bucket"
  value       = aws_s3_bucket.assets.bucket_regional_domain_name
}

output "frontend_bucket_name" {
  description = "Name of the frontend S3 bucket"
  value       = aws_s3_bucket.frontend.id
}

output "frontend_bucket_id" {
  description = "ID of the frontend S3 bucket"
  value       = aws_s3_bucket.frontend.id
}

output "frontend_bucket_arn" {
  description = "ARN of the frontend S3 bucket"
  value       = aws_s3_bucket.frontend.arn
}

output "frontend_bucket_regional_domain_name" {
  description = "Regional domain name of the frontend S3 bucket"
  value       = aws_s3_bucket.frontend.bucket_regional_domain_name
}

output "dynamodb_table_name" {
  description = "Name of the DynamoDB jobs table"
  value       = aws_dynamodb_table.jobs.name
}

output "dynamodb_table_arn" {
  description = "ARN of the DynamoDB jobs table"
  value       = aws_dynamodb_table.jobs.arn
}

output "dynamodb_usage_table_name" {
  description = "Name of the DynamoDB usage table"
  value       = aws_dynamodb_table.usage.name
}

output "dynamodb_usage_table_arn" {
  description = "ARN of the DynamoDB usage table"
  value       = aws_dynamodb_table.usage.arn
}
