terraform {
  required_version = ">= 1.13.5"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.21"
    }
  }

  # S3 Backend Configuration
  # Note: Bucket name must be created first with format: omnigen-terraform-state-<account-id>
  # Run: aws sts get-caller-identity --query Account --output text
  # Then create bucket: aws s3 mb s3://omnigen-terraform-state-<account-id>
  backend "s3" {
    bucket       = "omnigen-terraform-state-971422717446"
    key          = "infrastructure/terraform.tfstate"
    region       = "us-east-1"
    encrypt      = true
    use_lockfile = true
  }
}

provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Project   = "omnigen"
      ManagedBy = "terraform"
    }
  }
}
