resource "google_artifact_registry_repository" "main" {
  project       = var.project_id
  location      = var.region
  repository_id = "contract-mgmt-${var.environment}"
  format        = "DOCKER"
  description   = "Contract Management Tool images (${var.environment})"
}
