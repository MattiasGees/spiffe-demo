terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

data "tls_certificate" "oidc-certificate" {
  url = var.oidc-url
}
