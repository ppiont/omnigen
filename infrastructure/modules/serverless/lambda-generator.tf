# Lambda Function - Generator (Scene generation and Replicate API orchestration)

resource "aws_lambda_function" "generator" {
  filename         = data.archive_file.lambda_generator_placeholder.output_path
  function_name    = "${var.project_name}-generator"
  role             = var.lambda_execution_role_arn
  handler          = "generator.handler"
  source_code_hash = data.archive_file.lambda_generator_placeholder.output_base64sha256
  runtime          = "nodejs20.x"
  timeout          = var.timeout
  memory_size      = var.generator_memory

  reserved_concurrent_executions = var.generator_concurrency

  vpc_config {
    subnet_ids         = var.private_subnet_ids
    security_group_ids = [var.lambda_security_group_id]
  }

  environment {
    variables = {
      AWS_REGION           = var.aws_region
      ASSETS_BUCKET        = var.assets_bucket_name
      JOB_TABLE            = var.dynamodb_table_name
      REPLICATE_SECRET_ARN = var.replicate_secret_arn
    }
  }

  logging_config {
    log_format = "JSON"
    log_group  = var.generator_log_group_name
  }

  tags = {
    Name = "${var.project_name}-generator"
  }

  depends_on = [var.generator_log_group_name]
}

# Lambda Permission for Step Functions
resource "aws_lambda_permission" "generator_step_functions" {
  statement_id  = "AllowExecutionFromStepFunctions"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.generator.function_name
  principal     = "states.amazonaws.com"
  source_arn    = aws_sfn_state_machine.video_pipeline.arn
}
