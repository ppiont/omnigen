# DynamoDB Table for Job Tracking

resource "aws_dynamodb_table" "jobs" {
  name         = "${var.project_name}-jobs"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "job_id"

  attribute {
    name = "job_id"
    type = "S"
  }

  attribute {
    name = "user_id"
    type = "S"
  }

  attribute {
    name = "created_at"
    type = "N"
  }

  # Global Secondary Index for querying by user
  global_secondary_index {
    name            = "UserJobsIndex"
    hash_key        = "user_id"
    range_key       = "created_at"
    projection_type = "ALL"
  }

  # Time To Live configuration
  ttl {
    attribute_name = "ttl"
    enabled        = true
  }

  # Point-in-time recovery
  point_in_time_recovery {
    enabled = var.dynamodb_point_in_time_recovery
  }

  # Server-side encryption
  server_side_encryption {
    enabled = true
  }

  tags = {
    Name = "${var.project_name}-jobs"
  }
}

# DynamoDB Table for Usage Tracking
resource "aws_dynamodb_table" "usage" {
  name         = "${var.project_name}-usage"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "user_id"
  range_key    = "period"

  attribute {
    name = "user_id"
    type = "S"
  }

  attribute {
    name = "period"
    type = "S" # Format: YYYY-MM
  }

  # Time To Live configuration (optional, for cleanup of old usage records)
  ttl {
    attribute_name = "ttl"
    enabled        = true
  }

  # Point-in-time recovery
  point_in_time_recovery {
    enabled = var.dynamodb_point_in_time_recovery
  }

  # Server-side encryption
  server_side_encryption {
    enabled = true
  }

  tags = {
    Name = "${var.project_name}-usage"
  }
}
