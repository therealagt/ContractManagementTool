resource "google_logging_project_bucket_config" "audit" {
  project        = var.project_id
  location       = var.region
  bucket_id      = "contract-audit-logs-${var.environment}"
  retention_days = var.log_bucket_retention_days
}
