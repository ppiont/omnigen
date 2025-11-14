# Secrets Module - AWS Secrets Manager

# Note: The secret must be created manually before running Terraform
# This module only references the existing secret for use in other resources

data "aws_secretsmanager_secret" "replicate_api_key" {
  arn = var.replicate_api_key_secret_arn
}

data "aws_secretsmanager_secret_version" "replicate_api_key" {
  secret_id = data.aws_secretsmanager_secret.replicate_api_key.id
}
