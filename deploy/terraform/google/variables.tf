variable "bucket-name" {
  default = "GCP_BUCKET_NAME"
}

variable "oidc-url" {
  default = "https://OIDC_HOSTNAME"
}

variable "spiffe-id" {
  default = "spiffe://spire.demo.com/ns/spiffe-demo/sa/spiffe-demo-customer"
}

variable "gcp-region" {
  default = "europe-west1"
}

variable "gcp-project" {
  default = "jetstack-mattias"
}
