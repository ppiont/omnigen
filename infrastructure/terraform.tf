terraform {
  required_version = ">= 1.13.5"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.21"
    }
  }

  # Local backend (default)
  # Uncomment below for S3 backend after initial setup
  # backend "s3" {
  #   bucket         = "omnigen-terraform-state"
  #   key            = "infrastructure/terraform.tfstate"
  #   region         = "us-east-1"
  #   encrypt        = true
  #   dynamodb_table = "omnigen-terraform-locks"
  # }
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
