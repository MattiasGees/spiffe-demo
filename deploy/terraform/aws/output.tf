output "role_arn" {
  value = var.auth-type == "JWT" ? aws_iam_role.oidc-spire-role[0].arn : aws_iam_role.x509-spire-role[0].arn
}

output "trust_anchor_arn" {
  value = var.auth-type == "X509" ? aws_rolesanywhere_trust_anchor.x509-spire[0].arn : ""
}

output "profile_arn" {
  value = var.auth-type == "X509" ? aws_rolesanywhere_profile.x509-spire[0].arn : ""
}