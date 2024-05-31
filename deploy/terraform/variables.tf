variable "bucket-name" {
  default = "mattias-spiffe-demo"
}

variable "oidc-url" {
  default = "https://oidc-discovery.mattias-gcp.jetstacker.net"
}

variable "spiffe-id" {
  default = "spiffe://spire.internal.mattiasgees.be/ns/spiffe-demo/sa/spiffe-customer"
}

variable "aws-region"{
  default = "eu-west-2"
}
