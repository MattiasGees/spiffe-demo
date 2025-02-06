terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
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

resource "aws_iam_role_policy" "s3" {
  name = "demo-spiffe-policy"
  role = var.auth-type == "JWT" ? aws_iam_role.oidc-spire-role[0].name : aws_iam_role.x509-spire-role[0].name

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
