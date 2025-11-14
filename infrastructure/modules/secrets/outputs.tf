output "replicate_secret_arn" {
  description = "ARN of the Replicate API key secret"
  value       = data.aws_secretsmanager_secret.replicate_api_key.arn
}

output "replicate_secret_name" {
  description = "Name of the Replicate API key secret"
  value       = data.aws_secretsmanager_secret.replicate_api_key.name
}
