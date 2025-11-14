# OmniGen Infrastructure
# AI Video Generation Pipeline - Main Terraform Configuration

# Secrets Module - Must be created manually before running terraform
module "secrets" {
  source = "./modules/secrets"

  project_name                  = var.project_name
  replicate_api_key_secret_arn = var.replicate_api_key_secret_arn
}

# Networking Module - VPC, Subnets, NAT Gateway, Security Groups
module "networking" {
  source = "./modules/networking"

  project_name         = var.project_name
  vpc_cidr             = var.vpc_cidr
  public_subnet_cidr   = var.public_subnet_cidr
  private_subnet_cidr  = var.private_subnet_cidr
  availability_zone    = var.availability_zone
  enable_nat_gateway   = var.enable_nat_gateway
  container_port       = local.container_port
}

# IAM Module - Roles and Policies for ECS, Lambda, Step Functions
module "iam" {
  source = "./modules/iam"

  project_name                   = var.project_name
  assets_bucket_arn              = module.storage.assets_bucket_arn
  frontend_bucket_arn            = module.storage.frontend_bucket_arn
  dynamodb_table_arn             = module.storage.dynamodb_table_arn
  step_functions_arn             = module.serverless.step_functions_arn
  replicate_secret_arn           = var.replicate_api_key_secret_arn
  ecr_repository_arn             = module.compute.ecr_repository_arn
  lambda_generator_function_arn  = module.serverless.lambda_generator_arn
  lambda_composer_function_arn   = module.serverless.lambda_composer_arn
}

# Storage Module - S3 Buckets and DynamoDB Table
module "storage" {
  source = "./modules/storage"

  project_name                      = var.project_name
  assets_lifecycle_ia_days          = var.s3_assets_lifecycle_ia_days
  assets_lifecycle_glacier_days     = var.s3_assets_lifecycle_glacier_days
  assets_lifecycle_expiration_days  = var.s3_assets_lifecycle_expiration_days
  dynamodb_ttl_days                 = var.dynamodb_ttl_days
  dynamodb_point_in_time_recovery   = var.dynamodb_point_in_time_recovery
  cloudfront_distribution_arn       = module.cdn.cloudfront_distribution_arn
}

# Monitoring Module - CloudWatch Log Groups
module "monitoring" {
  source = "./modules/monitoring"

  project_name         = var.project_name
  log_retention_days   = var.cloudwatch_log_retention_days
  ecs_log_group_name   = local.ecs_log_group_name
  lambda_generator_log_group = local.lambda_generator_log_group
  lambda_composer_log_group  = local.lambda_composer_log_group
  step_functions_log_group   = local.step_functions_log_group
}

# Compute Module - ECS Fargate, ECR, Task Definition, Service, Auto-Scaling
module "compute" {
  source = "./modules/compute"

  project_name              = var.project_name
  vpc_id                    = module.networking.vpc_id
  private_subnet_ids        = [module.networking.private_subnet_id]
  ecs_security_group_id     = module.networking.ecs_security_group_id
  alb_target_group_arn      = module.loadbalancer.target_group_arn
  task_execution_role_arn   = module.iam.ecs_task_execution_role_arn
  task_role_arn             = module.iam.ecs_task_role_arn
  cpu                       = var.ecs_cpu
  memory                    = var.ecs_memory
  min_tasks                 = var.ecs_min_tasks
  max_tasks                 = var.ecs_max_tasks
  target_cpu_utilization    = var.ecs_target_cpu_utilization
  container_name            = local.container_name
  container_port            = local.container_port
  log_group_name            = module.monitoring.ecs_log_group_name
  aws_region                = var.aws_region
  assets_bucket_name        = module.storage.assets_bucket_name
  dynamodb_table_name       = module.storage.dynamodb_table_name
  step_functions_arn        = module.serverless.step_functions_arn
  replicate_secret_arn      = var.replicate_api_key_secret_arn

  depends_on = [module.monitoring]
}

# Serverless Module - Lambda Functions and Step Functions
module "serverless" {
  source = "./modules/serverless"

  project_name                = var.project_name
  vpc_id                      = module.networking.vpc_id
  private_subnet_ids          = [module.networking.private_subnet_id]
  lambda_security_group_id    = module.networking.lambda_security_group_id
  lambda_execution_role_arn   = module.iam.lambda_execution_role_arn
  step_functions_role_arn     = module.iam.step_functions_role_arn
  generator_memory            = var.lambda_generator_memory
  composer_memory             = var.lambda_composer_memory
  timeout                     = var.lambda_timeout
  generator_concurrency       = var.lambda_generator_concurrency
  composer_concurrency        = var.lambda_composer_concurrency
  assets_bucket_name          = module.storage.assets_bucket_name
  dynamodb_table_name         = module.storage.dynamodb_table_name
  replicate_secret_arn        = var.replicate_api_key_secret_arn
  aws_region                    = var.aws_region
  generator_log_group_name      = module.monitoring.lambda_generator_log_group_name
  composer_log_group_name       = module.monitoring.lambda_composer_log_group_name
  step_functions_log_group_name = module.monitoring.step_functions_log_group_name
  step_functions_log_group_arn  = module.monitoring.step_functions_log_group_arn

  depends_on = [module.monitoring]
}

# Load Balancer Module - Application Load Balancer
module "loadbalancer" {
  source = "./modules/loadbalancer"

  project_name          = var.project_name
  vpc_id                = module.networking.vpc_id
  public_subnet_ids     = [module.networking.public_subnet_id]
  alb_security_group_id = module.networking.alb_security_group_id
  container_port        = local.container_port
}

# CDN Module - CloudFront Distribution for Frontend
module "cdn" {
  source = "./modules/cdn"

  project_name           = var.project_name
  frontend_bucket_id     = module.storage.frontend_bucket_id
  frontend_bucket_arn    = module.storage.frontend_bucket_arn
  frontend_bucket_domain = module.storage.frontend_bucket_regional_domain_name
  price_class            = var.cloudfront_price_class
}
