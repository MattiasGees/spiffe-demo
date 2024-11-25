terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

data "tls_certificate" "oidc-certificate" {
  url = var.oidc-url
}

provider "google" {
  region  = var.gcp-region
  project = var.gcp-project
}

data "google_project" "project" {
  project_id = var.gcp-project
}


resource "google_storage_bucket" "bucket" {
  name                        = var.gcp-project
  location                    = "europe-west2"
  storage_class               = "STANDARD"
  uniform_bucket_level_access = true

  versioning {
    enabled = false
  }
}

resource "google_iam_workload_identity_pool" "spire-pool" {

  workload_identity_pool_id = "spire-workload-pool"
  display_name              = "SPIRE Workload Pool"
  description               = "Workload Pool to showcase SPIRE functionality"
}


resource "google_iam_workload_identity_pool_provider" "spire-provider" {
  workload_identity_pool_id          = google_iam_workload_identity_pool.spire-pool.workload_identity_pool_id
  workload_identity_pool_provider_id = "spire-oidc-provider"
  display_name                       = "SPIRE Provider"
  description                        = "OIDC Provider for SPIRE"
  attribute_mapping = {
    "google.subject"      = "assertion.sub"
    "attribute.spiffe_id" = "assertion.sub"
  }
  oidc {
    allowed_audiences = ["//iam.googleapis.com/projects/${data.google_project.project.number}/locations/global/workloadIdentityPools/spire-workload-pool/providers/spire-oidc-provider"]
    issuer_uri        = var.oidc-url
  }
}

resource "google_service_account" "spire_storage_writer" {

  account_id   = "spire-storage-writer"
  display_name = "spire-storage-writer"
  project      = var.gcp-project
}

resource "google_project_iam_member" "spire_storage_writer_policy" {
  project = var.gcp-project
  role    = "roles/storage.objectUser"
  member  = "serviceAccount:${google_service_account.spire_storage_writer.email}"
}

resource "google_service_account_iam_binding" "spire_storage_writer_binding" {

  service_account_id = google_service_account.spire_storage_writer.name
  role               = "roles/iam.serviceAccountTokenCreator"

  members = [
    "principalSet://iam.googleapis.com/projects/${data.google_project.project.number}/locations/global/workloadIdentityPools/spire-workload-pool/attribute.spiffe_id/${var.spiffe-id}"
  ]
}

resource "google_service_account_iam_binding" "spire_storage_writer_workload_identity_user_binding" {

  service_account_id = google_service_account.spire_storage_writer.name
  role               = "roles/iam.workloadIdentityUser"

  members = [
    "principalSet://iam.googleapis.com/projects/${data.google_project.project.number}/locations/global/workloadIdentityPools/spire-workload-pool/attribute.spiffe_id/${var.spiffe-id}"
  ]
}


