# S3 Lifecycle Rules for Assets Bucket

resource "aws_s3_bucket_lifecycle_configuration" "assets" {
  bucket = aws_s3_bucket.assets.id

  rule {
    id     = "transition-to-ia"
    status = "Enabled"

    transition {
      days          = var.assets_lifecycle_ia_days
      storage_class = "STANDARD_IA"
    }
  }

  rule {
    id     = "transition-to-glacier"
    status = "Enabled"

    transition {
      days          = var.assets_lifecycle_glacier_days
      storage_class = "GLACIER_IR"
    }
  }

  rule {
    id     = "expire-old-assets"
    status = "Enabled"

    expiration {
      days = var.assets_lifecycle_expiration_days
    }
  }

  rule {
    id     = "delete-incomplete-uploads"
    status = "Enabled"

    abort_incomplete_multipart_upload {
      days_after_initiation = 7
    }
  }
}
