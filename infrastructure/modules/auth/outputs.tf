output "user_pool_id" {
  description = "Cognito User Pool ID"
  value       = aws_cognito_user_pool.main.id
}

output "user_pool_arn" {
  description = "Cognito User Pool ARN"
  value       = aws_cognito_user_pool.main.arn
}

output "user_pool_endpoint" {
  description = "Cognito User Pool endpoint"
  value       = aws_cognito_user_pool.main.endpoint
}

output "client_id" {
  description = "Cognito User Pool Client ID"
  value       = aws_cognito_user_pool_client.api_client.id
}

output "client_secret" {
  description = "Cognito User Pool Client Secret (if applicable)"
  value       = try(aws_cognito_user_pool_client.api_client.client_secret, "")
  sensitive   = true
}

output "issuer_url" {
  description = "JWT issuer URL for token validation"
  value       = "https://cognito-idp.${var.aws_region}.amazonaws.com/${aws_cognito_user_pool.main.id}"
}

output "hosted_ui_domain" {
  description = "Cognito hosted UI domain"
  value       = "https://${aws_cognito_user_pool_domain.main.domain}.auth.${var.aws_region}.amazoncognito.com"
}

output "jwks_uri" {
  description = "JWKS URI for JWT validation"
  value       = "https://cognito-idp.${var.aws_region}.amazonaws.com/${aws_cognito_user_pool.main.id}/.well-known/jwks.json"
}
