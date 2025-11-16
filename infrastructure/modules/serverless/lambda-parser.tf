# Lambda Function - Parser (Script generation with LLM)

resource "aws_lambda_function" "parser" {
  filename         = data.archive_file.lambda_parser_placeholder.output_path
  function_name    = "${var.project_name}-parser"
  role             = var.lambda_execution_role_arn
  handler          = "bootstrap"
  source_code_hash = data.archive_file.lambda_parser_placeholder.output_base64sha256
  runtime          = "provided.al2023"
  timeout          = 120  # 2 minutes for LLM calls
  memory_size      = 1024 # 1GB for LLM processing
  architectures    = ["arm64"]

  reserved_concurrent_executions = 10

  environment {
    variables = {
      SCRIPTS_TABLE        = var.scripts_table_name
      REPLICATE_SECRET_ARN = var.replicate_secret_arn
    }
  }

  logging_config {
    log_format = "JSON"
    log_group  = var.parser_log_group_name
  }

  tags = {
    Name = "${var.project_name}-parser"
  }

  depends_on = [var.parser_log_group_name]
}

# Lambda Permission for API Gateway (for async invocation from /parse endpoint)
resource "aws_lambda_permission" "parser_api" {
  statement_id  = "AllowExecutionFromAPI"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.parser.function_name
  principal     = "apigateway.amazonaws.com"
}
