variable "bucket-name" {
  default = "mattias-spiffe-demo"
}

variable "oidc-url" {
  default = "https://oidc-discovery.mattias-gcp.jetstacker.net"
}

variable "spiffe-id" {
  default = "spiffe://spire.internal.mattiasgees.be/ns/spiffe-demo/sa/spiffe-demo-customer"
}

variable "gcp-region" {
  default = "europe-west1"
}

variable "gcp-project" {
  default = "jetstack-mattias"
}
