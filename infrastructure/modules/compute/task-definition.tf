# ECS Task Definition

resource "aws_ecs_task_definition" "api" {
  family                   = "${var.project_name}-api"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = var.cpu
  memory                   = var.memory
  execution_role_arn       = var.task_execution_role_arn
  task_role_arn            = var.task_role_arn

  # Use ARM64 Graviton processors for better price/performance
  runtime_platform {
    operating_system_family = "LINUX"
    cpu_architecture        = "ARM64"
  }

  container_definitions = jsonencode([
    {
      name      = var.container_name
      image     = "${aws_ecr_repository.api.repository_url}:latest"
      essential = true

      portMappings = [
        {
          containerPort = var.container_port
          protocol      = "tcp"
        }
      ]

      environment = [
        {
          name  = "ENVIRONMENT"
          value = var.environment
        },
        {
          name  = "AWS_REGION"
          value = var.aws_region
        },
        {
          name  = "ASSETS_BUCKET"
          value = var.assets_bucket_name
        },
        {
          name  = "JOB_TABLE"
          value = var.dynamodb_table_name
        },
        {
          name  = "USAGE_TABLE"
          value = var.dynamodb_usage_table_name
        },
        {
          name  = "REPLICATE_SECRET_ARN"
          value = var.replicate_secret_arn
        },
        {
          name  = "COGNITO_USER_POOL_ID"
          value = var.cognito_user_pool_id
        },
        {
          name  = "COGNITO_CLIENT_ID"
          value = var.cognito_client_id
        },
        {
          name  = "JWT_ISSUER"
          value = var.jwt_issuer
        },
        {
          name  = "COGNITO_DOMAIN"
          value = var.cognito_domain
        },
        {
          name  = "CLOUDFRONT_DOMAIN"
          value = var.cloudfront_domain
        },
        {
          name  = "VIDEO_ADAPTER_TYPE"
          value = var.video_adapter_type
        },
        {
          name  = "VEO_GENERATE_AUDIO"
          value = var.veo_generate_audio
        }
      ]

      logConfiguration = {
        logDriver = "awslogs"
        options = {
          "awslogs-group"         = var.log_group_name
          "awslogs-region"        = var.aws_region
          "awslogs-stream-prefix" = "ecs"
        }
      }

      healthCheck = {
        command     = ["CMD-SHELL", "curl -f http://localhost:${var.container_port}/health || exit 1"]
        interval    = 30
        timeout     = 5
        retries     = 3
        startPeriod = 60
      }
    }
  ])

  tags = {
    Name = "${var.project_name}-api-task"
  }
}
