variable "bucket-name" {
  default = "BUCKET_NAME"
}

variable "oidc-url" {
  default = "https://OIDC_HOSTNAME"
}

variable "spiffe-id" {
  default = "spiffe://spire.demo.com/ns/spiffe-demo/sa/spiffe-demo-customer"
}

variable "aws-region" {
  default = "eu-west-2"
}