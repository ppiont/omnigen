# Step Functions State Machine - Video Generation Pipeline

resource "aws_sfn_state_machine" "video_pipeline" {
  name     = "${var.project_name}-workflow"
  role_arn = var.step_functions_role_arn
  type     = "STANDARD" # Changed from EXPRESS to support longer executions

  logging_configuration {
    log_destination        = "${var.step_functions_log_group_arn}:*"
    include_execution_data = true
    level                  = "ALL"
  }

  definition = jsonencode({
    Comment = "AI Video Generation Pipeline - Sequential Clips with Coherence"
    StartAt = "InitializeClipGeneration"
    States = {

      # Initialize state with clip counter
      InitializeClipGeneration = {
        Type = "Pass"
        Parameters = {
          "job_id.$"       = "$.job_id"
          "prompt.$"       = "$.prompt"
          "duration.$"     = "$.duration"
          "aspect_ratio.$" = "$.aspect_ratio"
          "start_image.$"  = "$.start_image"
          "num_clips.$"    = "$.num_clips"
          "music_mood.$"   = "$.music_mood"
          "music_style.$"  = "$.music_style"
          "clip_counter"   = 1
          "clip_videos"    = []
          "last_frame_url" = ""
        }
        Next = "GenerateClipLoop"
      }

      # Loop to generate clips sequentially
      GenerateClipLoop = {
        Type = "Choice"
        Choices = [
          {
            # Continue if clip_counter <= num_clips
            Variable      = "$.clip_counter"
            NumericLessThanEqualsPath = "$.num_clips"
            Next          = "GenerateSingleClip"
          }
        ]
        Default = "GenerateAudio" # All clips done, move to audio
      }

      # Generate a single clip
      GenerateSingleClip = {
        Type     = "Task"
        Resource = "arn:aws:states:::lambda:invoke"
        Parameters = {
          FunctionName = aws_lambda_function.generator.arn
          Payload = {
            "job_id.$"         = "$.job_id"
            "prompt.$"         = "$.prompt"
            "duration"         = 10 # Each clip is 10 seconds
            "aspect_ratio.$"   = "$.aspect_ratio"
            "start_image_url.$" = "$.last_frame_url" # Visual coherence!
            "clip_number.$"    = "$.clip_counter"
          }
        }
        ResultPath = "$.current_clip_result"
        Next       = "ProcessClipResult"
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

      # Process the clip result and update state
      ProcessClipResult = {
        Type = "Pass"
        Parameters = {
          "job_id.$"       = "$.job_id"
          "prompt.$"       = "$.prompt"
          "duration.$"     = "$.duration"
          "aspect_ratio.$" = "$.aspect_ratio"
          "start_image.$"  = "$.start_image"
          "num_clips.$"    = "$.num_clips"
          "music_mood.$"   = "$.music_mood"
          "music_style.$"  = "$.music_style"
          "clip_counter.$" = "States.MathAdd($.clip_counter, 1)"
          # Append current clip to array
          "clip_videos.$"  = "States.Array($.clip_videos[*], $.current_clip_result.Payload)"
          # Save last frame for next clip's coherence
          "last_frame_url.$" = "$.current_clip_result.Payload.last_frame_url"
        }
        Next = "GenerateClipLoop"
      }

      # Generate background music
      GenerateAudio = {
        Type     = "Task"
        Resource = "arn:aws:states:::lambda:invoke"
        Parameters = {
          FunctionName = aws_lambda_function.audio_generator.arn
          Payload = {
            "job_id.$"      = "$.job_id"
            "prompt.$"      = "$.prompt"
            "duration.$"    = "$.duration"
            "music_mood.$"  = "$.music_mood"
            "music_style.$" = "$.music_style"
          }
        }
        ResultPath = "$.audio_result"
        Next       = "ComposeVideo"
        Retry = [
          {
            ErrorEquals     = ["States.ALL"]
            IntervalSeconds = 10
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

      # Compose final video with all clips + audio
      ComposeVideo = {
        Type     = "Task"
        Resource = "arn:aws:states:::lambda:invoke"
        Parameters = {
          FunctionName = aws_lambda_function.composer.arn
          Payload = {
            "job_id.$"      = "$.job_id"
            "clip_videos.$" = "$.clip_videos"
            "music_url.$"   = "$.audio_result.Payload.music_url"
          }
        }
        ResultPath = "$.compose_result"
        Next       = "MarkComplete"
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

      # Mark job as completed
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
          UpdateExpression = "SET #status = :status, video_key = :video_key, current_stage = :stage"
          ExpressionAttributeNames = {
            "#status" = "status"
          }
          ExpressionAttributeValues = {
            ":status" = {
              S = "completed"
            }
            ":video_key" = {
              "S.$" = "$.compose_result.Payload.video_url"
            }
            ":stage" = {
              S = "Completed"
            }
          }
        }
        End = true
      }

      # Mark job as failed
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
          UpdateExpression = "SET #status = :status, error_message = :error_message, current_stage = :stage"
          ExpressionAttributeNames = {
            "#status" = "status"
          }
          ExpressionAttributeValues = {
            ":status" = {
              S = "failed"
            }
            ":error_message" = {
              "S.$" = "States.Format('Generation failed: {}', $.error.Cause)"
            }
            ":stage" = {
              S = "Failed"
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
