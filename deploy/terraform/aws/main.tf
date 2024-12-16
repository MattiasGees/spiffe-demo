terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

data "tls_certificate" "oidc-certificate" {
  url = "https://${var.oidc-url}"
}

provider "aws" {
  region = var.aws-region
}

resource "aws_s3_bucket" "oidc-test" {
  bucket = var.bucket-name

  tags = {
    Name        = var.bucket-name
    Environment = "demo"
  }
}

resource "aws_iam_openid_connect_provider" "oidc-spire" {
  url = "https://${var.oidc-url}"

  client_id_list = [
    "demo",
  ]

  thumbprint_list = [data.tls_certificate.oidc-certificate.certificates[0].sha1_fingerprint]
}

resource "aws_iam_role" "oidc-spire-role" {
  name = "demo-spiffe-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Action = "sts:AssumeRoleWithWebIdentity",
        Effect = "Allow",
        Principal = {
          Federated = aws_iam_openid_connect_provider.oidc-spire.arn,
        },
        Condition = {
          StringEquals = {
            "${var.oidc-url}:aud" = "demo",
            "${var.oidc-url}:sub" = "${var.spiffe-id}"
          }
        }
      },
    ],
  })
}

resource "aws_iam_role_policy" "s3" {
  name = "demo-spiffe-policy"
  role = aws_iam_role.oidc-spire-role.name

  policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "VisualEditor0",
            "Effect": "Allow",
            "Action": [
                "s3:PutAccountPublicAccessBlock",
                "s3:GetAccountPublicAccessBlock",
                "s3:ListAllMyBuckets",
                "s3:ListJobs",
                "s3:CreateJob",
                "s3:HeadBucket"
            ],
            "Resource": "*"
        },
        {
            "Sid": "VisualEditor1",
            "Effect": "Allow",
            "Action": "s3:*",
            "Resource": [
                "arn:aws:s3:::${aws_s3_bucket.oidc-test.bucket}",
                "arn:aws:s3:::${aws_s3_bucket.oidc-test.bucket}/*",
                "arn:aws:s3:*:*:job/*"
            ]
        }
    ]
}
EOF
}
