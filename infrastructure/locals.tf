locals {
  # Resource naming
  name_prefix = var.project_name

  # Common tags
  common_tags = {
    Project   = var.project_name
    ManagedBy = "terraform"
  }

  # Resource names
  vpc_name                     = "${local.name_prefix}-vpc"
  assets_bucket_name           = "${local.name_prefix}-assets"
  frontend_bucket_name         = "${local.name_prefix}-frontend"
  dynamodb_table_name          = "${local.name_prefix}-jobs"
  ecs_cluster_name             = local.name_prefix
  ecs_service_name             = "${local.name_prefix}-api"
  ecr_repository_name          = "${local.name_prefix}-api"
  lambda_generator_name        = "${local.name_prefix}-generator"
  lambda_composer_name         = "${local.name_prefix}-composer"
  step_functions_name          = "${local.name_prefix}-workflow"
  alb_name                     = local.name_prefix
  cloudfront_distribution_name = "${local.name_prefix}-frontend"

  # Security group names
  alb_sg_name    = "${local.name_prefix}-alb-sg"
  ecs_sg_name    = "${local.name_prefix}-ecs-sg"
  lambda_sg_name = "${local.name_prefix}-lambda-sg"

  # CloudWatch log group names
  ecs_log_group_name              = "/ecs/${local.name_prefix}"
  lambda_generator_log_group      = "/aws/lambda/${local.lambda_generator_name}"
  lambda_composer_log_group       = "/aws/lambda/${local.lambda_composer_name}"
  lambda_parser_log_group         = "/aws/lambda/${local.name_prefix}-parser"
  lambda_audio_generator_log_group = "/aws/lambda/${local.name_prefix}-audio-generator"
  step_functions_log_group        = "/aws/states/${local.step_functions_name}"

  # IAM role names
  ecs_task_execution_role_name = "${local.name_prefix}-ecs-task-execution"
  ecs_task_role_name           = "${local.name_prefix}-ecs-task"
  lambda_execution_role_name   = "${local.name_prefix}-lambda-execution"
  step_functions_role_name     = "${local.name_prefix}-step-functions"

  # Container configuration
  container_name = "${local.name_prefix}-api"
  container_port = 8080

  # DynamoDB GSI names
  user_jobs_index_name = "UserJobsIndex"
  status_index_name    = "StatusIndex"
}
