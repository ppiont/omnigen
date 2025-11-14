# IAM Module - Roles and Policies for ECS, Lambda, Step Functions

# Data sources for AWS account and region
data "aws_caller_identity" "current" {}
data "aws_region" "current" {}
