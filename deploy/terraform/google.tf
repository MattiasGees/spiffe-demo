resource "google_iam_workload_identity_pool" "spire-pool" {
  count = var.enable-gcp ? 1 : 0
  workload_identity_pool_id = "spire-workload-pool"
  display_name              = "SPIRE Workload Pool"
  description               = "Workload Pool to showcase SPIRE functionality"
}


resource "google_iam_workload_identity_pool_provider" "spire-provider" {
  count = var.enable-gcp ? 1 : 0
  workload_identity_pool_id          = google_iam_workload_identity_pool.spire-pool.workload_identity_pool_id
  workload_identity_pool_provider_id = "spire-oidc-provider"
  display_name                       = "SPIRE Provider"
  description                        = "OIDC Provider for SPIRE"
  attribute_condition                = "true"
  attribute_mapping                  = {
    "google.subject"                  = "assertion.sub"
    "attribute.spiffe_id"                   = "assertion.sub"
  }
  oidc {
    allowed_audiences = ["//iam.googleapis.com/projects/<PROJECT_NUMBER>/locations/global/workloadIdentityPools/spire-workload-pool/providers/spire-workload-provider"]
    issuer_uri       = var.oidc-url
  }
}
