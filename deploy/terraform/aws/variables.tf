variable "bucket-name" {
  default = "AWS_BUCKET_NAME"
}

variable "oidc-url" {
  default = "OIDC_HOSTNAME"
}

variable "spiffe-id" {
  default = "spiffe://spire.demo.com/ns/spiffe-demo/sa/spiffe-demo-customer"
}

variable "aws-region" {
  default = "eu-west-2"
}

variable "auth-type" {
  description = "Authentication type to use either pick JWT or X509"
  default     = "JWT"
}

variable "root-CA" {
  default = "root.pem"
}