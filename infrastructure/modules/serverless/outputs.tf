output "lambda_generator_arn" {
  description = "ARN of the generator Lambda function"
  value       = aws_lambda_function.generator.arn
}

output "lambda_generator_name" {
  description = "Name of the generator Lambda function"
  value       = aws_lambda_function.generator.function_name
}

output "lambda_composer_arn" {
  description = "ARN of the composer Lambda function"
  value       = aws_lambda_function.composer.arn
}

output "lambda_composer_name" {
  description = "Name of the composer Lambda function"
  value       = aws_lambda_function.composer.function_name
}

output "step_functions_arn" {
  description = "ARN of the Step Functions state machine"
  value       = aws_sfn_state_machine.video_pipeline.arn
}

output "step_functions_name" {
  description = "Name of the Step Functions state machine"
  value       = aws_sfn_state_machine.video_pipeline.name
}
