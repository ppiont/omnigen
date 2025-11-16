# Monitoring Module - CloudWatch Log Groups

# ECS Log Group
resource "aws_cloudwatch_log_group" "ecs" {
  name              = var.ecs_log_group_name
  retention_in_days = var.log_retention_days

  tags = {
    Name = "${var.project_name}-ecs-logs"
  }
}

# Lambda Generator Log Group
resource "aws_cloudwatch_log_group" "lambda_generator" {
  name              = var.lambda_generator_log_group
  retention_in_days = var.log_retention_days

  tags = {
    Name = "${var.project_name}-lambda-generator-logs"
  }
}

# Lambda Composer Log Group
resource "aws_cloudwatch_log_group" "lambda_composer" {
  name              = var.lambda_composer_log_group
  retention_in_days = var.log_retention_days

  tags = {
    Name = "${var.project_name}-lambda-composer-logs"
  }
}

# Lambda Parser Log Group
resource "aws_cloudwatch_log_group" "lambda_parser" {
  name              = var.lambda_parser_log_group
  retention_in_days = var.log_retention_days

  tags = {
    Name = "${var.project_name}-lambda-parser-logs"
  }
}

# Lambda Audio Generator Log Group
resource "aws_cloudwatch_log_group" "lambda_audio_generator" {
  name              = var.lambda_audio_generator_log_group
  retention_in_days = var.log_retention_days

  tags = {
    Name = "${var.project_name}-lambda-audio-generator-logs"
  }
}

# Step Functions Log Group
resource "aws_cloudwatch_log_group" "step_functions" {
  name              = var.step_functions_log_group
  retention_in_days = var.log_retention_days

  tags = {
    Name = "${var.project_name}-step-functions-logs"
  }
}
