data "tls_certificate" "oidc-certificate" {
  count = var.auth-type == "JWT" ? 1 : 0
  url   = "https://${var.oidc-url}"
}

resource "aws_iam_openid_connect_provider" "oidc-spire" {
  count = var.auth-type == "JWT" ? 1 : 0
  url = "https://${var.oidc-url}"

  client_id_list = [
    "demo",
  ]

  thumbprint_list = [data.tls_certificate.oidc-certificate.certificates[0].sha1_fingerprint]
}

resource "aws_iam_role" "oidc-spire-role" {
  count = var.auth-type == "JWT" ? 1 : 0
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