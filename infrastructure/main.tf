# OmniGen Infrastructure
# AI Video Generation Pipeline - Main Terraform Configuration

# Secrets Module - Must be created manually before running terraform
module "secrets" {
  source = "./modules/secrets"

  project_name                 = var.project_name
  replicate_api_key_secret_arn = var.replicate_api_key_secret_arn
}

# Networking Module - VPC, Subnets, NAT Gateway, Security Groups
module "networking" {
  source = "./modules/networking"

  project_name        = var.project_name
  aws_region          = var.aws_region
  vpc_cidr            = var.vpc_cidr
  public_subnet_cidrs = var.public_subnet_cidrs
  availability_zones  = var.availability_zones
  private_subnet_cidr = var.private_subnet_cidr
  availability_zone   = var.availability_zone
  enable_nat_gateway  = var.enable_nat_gateway
  container_port      = local.container_port
}

# IAM Module - Roles and Policies for ECS
module "iam" {
  source = "./modules/iam"

  project_name               = var.project_name
  assets_bucket_arn          = module.storage.assets_bucket_arn
  frontend_bucket_arn        = module.storage.frontend_bucket_arn
  dynamodb_table_arn         = module.storage.dynamodb_table_arn
  dynamodb_usage_table_arn   = module.storage.dynamodb_usage_table_arn
  dynamodb_scripts_table_arn = module.storage.dynamodb_scripts_table_arn
  replicate_secret_arn       = var.replicate_api_key_secret_arn
  ecr_repository_arn         = module.compute.ecr_repository_arn
}

# Storage Module - S3 Buckets and DynamoDB Table
module "storage" {
  source = "./modules/storage"

  project_name                     = var.project_name
  assets_lifecycle_ia_days         = var.s3_assets_lifecycle_ia_days
  assets_lifecycle_glacier_days    = var.s3_assets_lifecycle_glacier_days
  assets_lifecycle_expiration_days = var.s3_assets_lifecycle_expiration_days
  dynamodb_ttl_days                = var.dynamodb_ttl_days
  dynamodb_point_in_time_recovery  = var.dynamodb_point_in_time_recovery
  cloudfront_distribution_arn      = module.cdn.cloudfront_distribution_arn
}

# Monitoring Module - CloudWatch Log Groups
module "monitoring" {
  source = "./modules/monitoring"

  project_name       = var.project_name
  log_retention_days = var.cloudwatch_log_retention_days
  ecs_log_group_name = local.ecs_log_group_name
}

# Compute Module - ECS Fargate, ECR, Task Definition, Service, Auto-Scaling
module "compute" {
  source = "./modules/compute"

  project_name                = var.project_name
  environment                 = var.environment
  vpc_id                      = module.networking.vpc_id
  private_subnet_ids          = [module.networking.private_subnet_id]
  ecs_security_group_id       = module.networking.ecs_security_group_id
  alb_target_group_arn        = module.loadbalancer.target_group_arn
  task_execution_role_arn     = module.iam.ecs_task_execution_role_arn
  task_role_arn               = module.iam.ecs_task_role_arn
  cpu                         = var.ecs_cpu
  memory                      = var.ecs_memory
  min_tasks                   = var.ecs_min_tasks
  max_tasks                   = var.ecs_max_tasks
  target_cpu_utilization      = var.ecs_target_cpu_utilization
  container_name              = local.container_name
  container_port              = local.container_port
  log_group_name              = module.monitoring.ecs_log_group_name
  aws_region                  = var.aws_region
  assets_bucket_name          = module.storage.assets_bucket_name
  dynamodb_table_name         = module.storage.dynamodb_table_name
  dynamodb_usage_table_name   = module.storage.dynamodb_usage_table_name
  dynamodb_scripts_table_name = module.storage.dynamodb_scripts_table_name
  replicate_secret_arn        = var.replicate_api_key_secret_arn
  cognito_user_pool_id        = module.auth.user_pool_id
  cognito_client_id           = module.auth.client_id
  jwt_issuer                  = module.auth.issuer_url
  cognito_domain              = module.auth.hosted_ui_domain
  cloudfront_domain           = module.cdn.cloudfront_domain_name

  depends_on = [module.monitoring, module.auth]
}

# Load Balancer Module - Application Load Balancer
module "loadbalancer" {
  source = "./modules/loadbalancer"

  project_name          = var.project_name
  vpc_id                = module.networking.vpc_id
  public_subnet_ids     = module.networking.public_subnet_ids
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
  alb_dns_name           = module.loadbalancer.alb_dns_name
  price_class            = var.cloudfront_price_class
}

# Auth Module - Cognito User Pool for Authentication
module "auth" {
  source = "./modules/auth"

  project_name      = var.project_name
  aws_region        = var.aws_region
  cloudfront_domain = module.cdn.cloudfront_domain_name
}
