# CDN Module - CloudFront Distribution for Frontend

# Origin Access Control for S3
resource "aws_cloudfront_origin_access_control" "frontend" {
  name                              = "${var.project_name}-frontend-oac"
  description                       = "Origin Access Control for ${var.project_name} frontend bucket"
  origin_access_control_origin_type = "s3"
  signing_behavior                  = "always"
  signing_protocol                  = "sigv4"
}

# CloudFront Distribution
resource "aws_cloudfront_distribution" "frontend" {
  enabled             = true
  is_ipv6_enabled     = true
  comment             = "${var.project_name} frontend distribution"
  default_root_object = "index.html"
  price_class         = var.price_class

  origin {
    domain_name              = var.frontend_bucket_domain
    origin_id                = "S3-${var.frontend_bucket_id}"
    origin_access_control_id = aws_cloudfront_origin_access_control.frontend.id
  }

  default_cache_behavior {
    allowed_methods        = ["GET", "HEAD", "OPTIONS"]
    cached_methods         = ["GET", "HEAD"]
    target_origin_id       = "S3-${var.frontend_bucket_id}"
    viewer_protocol_policy = "redirect-to-https"
    compress               = true

    # Use AWS managed cache policy for caching optimized
    cache_policy_id = "658327ea-f89d-4fab-a63d-7e88639e58f6"  # Managed-CachingOptimized
  }

  # Cache behavior for static assets (immutable caching)
  ordered_cache_behavior {
    path_pattern           = "/assets/*"
    allowed_methods        = ["GET", "HEAD", "OPTIONS"]
    cached_methods         = ["GET", "HEAD"]
    target_origin_id       = "S3-${var.frontend_bucket_id}"
    viewer_protocol_policy = "redirect-to-https"
    compress               = true

    # Use AWS managed cache policy with long TTL for immutable assets
    cache_policy_id = "658327ea-f89d-4fab-a63d-7e88639e58f6"  # Managed-CachingOptimized (1 year TTL)
  }

  # Cache behavior for index.html (no caching for SPA)
  ordered_cache_behavior {
    path_pattern           = "/index.html"
    allowed_methods        = ["GET", "HEAD", "OPTIONS"]
    cached_methods         = ["GET", "HEAD"]
    target_origin_id       = "S3-${var.frontend_bucket_id}"
    viewer_protocol_policy = "redirect-to-https"
    compress               = true

    # Use AWS managed cache policy with no caching for SPA entry point
    cache_policy_id = "4135ea2d-6df8-44a3-9df3-4b5a84be39ad"  # Managed-CachingDisabled
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
    minimum_protocol_version       = "TLSv1.2_2021"
  }

  # Custom error responses for SPA routing
  custom_error_response {
    error_code         = 403
    response_code      = 200
    response_page_path = "/index.html"
  }

  custom_error_response {
    error_code         = 404
    response_code      = 200
    response_page_path = "/index.html"
  }

  tags = {
    Name = "${var.project_name}-cloudfront"
  }
}
