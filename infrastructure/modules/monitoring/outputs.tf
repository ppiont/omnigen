output "ecs_log_group_name" {
  description = "Name of the ECS CloudWatch log group"
  value       = aws_cloudwatch_log_group.ecs.name
}

output "ecs_log_group_arn" {
  description = "ARN of the ECS CloudWatch log group"
  value       = aws_cloudwatch_log_group.ecs.arn
}

output "lambda_generator_log_group_name" {
  description = "Name of the Lambda generator CloudWatch log group"
  value       = aws_cloudwatch_log_group.lambda_generator.name
}

output "lambda_generator_log_group_arn" {
  description = "ARN of the Lambda generator CloudWatch log group"
  value       = aws_cloudwatch_log_group.lambda_generator.arn
}

output "lambda_composer_log_group_name" {
  description = "Name of the Lambda composer CloudWatch log group"
  value       = aws_cloudwatch_log_group.lambda_composer.name
}

output "lambda_composer_log_group_arn" {
  description = "ARN of the Lambda composer CloudWatch log group"
  value       = aws_cloudwatch_log_group.lambda_composer.arn
}

output "lambda_parser_log_group_name" {
  description = "Name of the Lambda parser CloudWatch log group"
  value       = aws_cloudwatch_log_group.lambda_parser.name
}

output "lambda_parser_log_group_arn" {
  description = "ARN of the Lambda parser CloudWatch log group"
  value       = aws_cloudwatch_log_group.lambda_parser.arn
}

output "lambda_audio_generator_log_group_name" {
  description = "Name of the Lambda audio generator CloudWatch log group"
  value       = aws_cloudwatch_log_group.lambda_audio_generator.name
}

output "lambda_audio_generator_log_group_arn" {
  description = "ARN of the Lambda audio generator CloudWatch log group"
  value       = aws_cloudwatch_log_group.lambda_audio_generator.arn
}

output "step_functions_log_group_name" {
  description = "Name of the Step Functions CloudWatch log group"
  value       = aws_cloudwatch_log_group.step_functions.name
}

output "step_functions_log_group_arn" {
  description = "ARN of the Step Functions CloudWatch log group"
  value       = aws_cloudwatch_log_group.step_functions.arn
}
