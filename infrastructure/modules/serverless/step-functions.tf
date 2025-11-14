# Step Functions State Machine - Video Generation Pipeline

resource "aws_sfn_state_machine" "video_pipeline" {
  name     = "${var.project_name}-workflow"
  role_arn = var.step_functions_role_arn
  type     = "EXPRESS"

  logging_configuration {
    log_destination        = "${var.step_functions_log_group_arn}:*"
    include_execution_data = true
    level                  = "ALL"
  }

  definition = jsonencode({
    Comment = "AI Video Generation Pipeline - Express Workflow"
    StartAt = "GenerateScenes"
    States = {
      GenerateScenes = {
        Type     = "Task"
        Resource = "arn:aws:states:::lambda:invoke"
        Parameters = {
          FunctionName = aws_lambda_function.generator.arn
          Payload = {
            "job_id.$"   = "$.job_id"
            "prompt.$"   = "$.prompt"
            "duration.$" = "$.duration"
            "style.$"    = "$.style"
          }
        }
        Next = "ComposeVideo"
        Retry = [
          {
            ErrorEquals     = ["States.ALL"]
            IntervalSeconds = 2
            MaxAttempts     = 2
            BackoffRate     = 2.0
          }
        ]
        Catch = [
          {
            ErrorEquals = ["States.ALL"]
            Next        = "MarkFailed"
            ResultPath  = "$.error"
          }
        ]
      }

      ComposeVideo = {
        Type     = "Task"
        Resource = "arn:aws:states:::lambda:invoke"
        Parameters = {
          FunctionName = aws_lambda_function.composer.arn
          Payload = {
            "job_id.$" = "$.job_id"
            "scenes.$" = "$.Payload.scenes"
          }
        }
        Next = "MarkComplete"
        Retry = [
          {
            ErrorEquals     = ["States.ALL"]
            IntervalSeconds = 5
            MaxAttempts     = 2
            BackoffRate     = 2.0
          }
        ]
        Catch = [
          {
            ErrorEquals = ["States.ALL"]
            Next        = "MarkFailed"
            ResultPath  = "$.error"
          }
        ]
      }

      MarkComplete = {
        Type     = "Task"
        Resource = "arn:aws:states:::dynamodb:updateItem"
        Parameters = {
          TableName = var.dynamodb_table_name
          Key = {
            job_id = {
              "S.$" = "$.job_id"
            }
          }
          UpdateExpression = "SET #status = :status, completed_at = :completed_at, video_url = :video_url"
          ExpressionAttributeNames = {
            "#status" = "status"
          }
          ExpressionAttributeValues = {
            ":status" = {
              S = "completed"
            }
            ":completed_at" = {
              "N.$" = "States.Format('{}', $$.State.EnteredTime)"
            }
            ":video_url" = {
              "S.$" = "$.Payload.video_url"
            }
          }
        }
        End = true
      }

      MarkFailed = {
        Type     = "Task"
        Resource = "arn:aws:states:::dynamodb:updateItem"
        Parameters = {
          TableName = var.dynamodb_table_name
          Key = {
            job_id = {
              "S.$" = "$.job_id"
            }
          }
          UpdateExpression = "SET #status = :status, error_message = :error_message"
          ExpressionAttributeNames = {
            "#status" = "status"
          }
          ExpressionAttributeValues = {
            ":status" = {
              S = "failed"
            }
            ":error_message" = {
              "S.$" = "States.Format('{}', $.error.Cause)"
            }
          }
        }
        Next = "Fail"
      }

      Fail = {
        Type = "Fail"
      }
    }
  })

  tags = {
    Name = "${var.project_name}-workflow"
  }
}
