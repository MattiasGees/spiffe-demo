resource "aws_rolesanywhere_trust_anchor" "x509-spire" {
  count = var.auth-type == "X509" ? 1 : 0
  name    = "spire-root-ca"
  enabled = true
  source {
    source_data {
      x509_certificate_data = file(var.root-CA)
    }
    source_type = "CERTIFICATE_BUNDLE"
  }
}

resource "aws_rolesanywhere_profile" "x509-spire" {
  count = var.auth-type == "X509" ? 1 : 0
  name           = "spire-x509-profile"
  enabled        = true
  role_arns      = [aws_iam_role.otterize-credentials-operator.arn]
}

resource "aws_iam_role" "x509-spire-role" {
  count = var.auth-type == "X509" ? 1 : 0
  name = "demo-spiffe-role-x509"

  assume_role_policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Action = ["sts:AssumeRole", "sts:TagSession", "sts:SetSourceIdentity"]
        Effect = "Allow",
        Principal = {
          Service = "rolesanywhere.amazonaws.com",
        },
        Condition = {
          StringLike = {
            "aws:PrincipalTag/x509SAN/URI" = "${var.spiffe-id}",
          }
          ArnEquals = {
            "aws:SourceArn" = aws_rolesanywhere_trust_anchor.x509-spire.arn
          }
        }
      },
    ],
  })
}