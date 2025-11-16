# Monitoring Module - CloudWatch Log Groups

# ECS Log Group
resource "aws_cloudwatch_log_group" "ecs" {
  name              = var.ecs_log_group_name
  retention_in_days = var.log_retention_days

  tags = {
    Name = "${var.project_name}-ecs-logs"
  }
}
