# Lambda Function - Composer (FFmpeg video stitching and finalization)

resource "aws_lambda_function" "composer" {
  filename         = data.archive_file.lambda_composer_placeholder.output_path
  function_name    = "${var.project_name}-composer"
  role             = var.lambda_execution_role_arn
  handler          = "bootstrap"
  source_code_hash = data.archive_file.lambda_composer_placeholder.output_base64sha256
  runtime          = "provided.al2023"
  timeout          = var.timeout
  memory_size      = var.composer_memory
  architectures    = ["arm64"]

  reserved_concurrent_executions = var.composer_concurrency

  ephemeral_storage {
    size = 10240 # 10 GB for video processing
  }

  vpc_config {
    subnet_ids         = var.private_subnet_ids
    security_group_ids = [var.lambda_security_group_id]
  }

  environment {
    variables = {
      ASSETS_BUCKET        = var.assets_bucket_name
      JOB_TABLE            = var.dynamodb_table_name
      REPLICATE_SECRET_ARN = var.replicate_secret_arn
    }
  }

  logging_config {
    log_format = "JSON"
    log_group  = var.composer_log_group_name
  }

  tags = {
    Name = "${var.project_name}-composer"
  }

  depends_on = [var.composer_log_group_name]
}

# Lambda Permission for Step Functions
resource "aws_lambda_permission" "composer_step_functions" {
  statement_id  = "AllowExecutionFromStepFunctions"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.composer.function_name
  principal     = "states.amazonaws.com"
  source_arn    = aws_sfn_state_machine.video_pipeline.arn
}
