terraform {
  required_version = ">= 1.5"

  backend "s3" {
    bucket = "terraform-juriscan-state"
    key    = "juriscan/infra/terraform.tfstate"
    region = "us-east-1"
  }

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.6"
    }
  }
}

provider "aws" {
  region = var.region

  default_tags {
    tags = local.tags
  }
}

data "aws_caller_identity" "current" {}

data "aws_ami" "amazon_linux_2023" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["al2023-ami-2023.*-x86_64"]
  }

  filter {
    name   = "architecture"
    values = ["x86_64"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }
}

locals {
  name_prefix = "${var.project}-${var.environment}"
  fqdn        = "${var.app_subdomain}.${var.root_domain}"

  tags = {
    Project     = var.project
    Environment = var.environment
    ManagedBy   = "terraform"
    CostProfile = "low"
  }
}
