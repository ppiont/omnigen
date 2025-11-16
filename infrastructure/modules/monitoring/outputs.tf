output "ecs_log_group_name" {
  description = "Name of the ECS CloudWatch log group"
  value       = aws_cloudwatch_log_group.ecs.name
}

output "ecs_log_group_arn" {
  description = "ARN of the ECS CloudWatch log group"
  value       = aws_cloudwatch_log_group.ecs.arn
}
