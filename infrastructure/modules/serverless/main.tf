# Serverless Module - Lambda Functions and Step Functions

# Data sources
data "aws_caller_identity" "current" {}
data "aws_region" "current" {}

# Archive placeholder Lambda code
data "archive_file" "lambda_generator_placeholder" {
  type        = "zip"
  source_file = "${path.module}/lambda-placeholders/generator.js"
  output_path = "${path.module}/lambda-placeholders/generator.zip"
}

data "archive_file" "lambda_composer_placeholder" {
  type        = "zip"
  source_file = "${path.module}/lambda-placeholders/composer.js"
  output_path = "${path.module}/lambda-placeholders/composer.zip"
}
