resource "google_logging_project_bucket_config" "audit" {
  project        = var.project_id
  location       = var.region
  bucket_id      = "contract-audit-logs-${var.environment}"
  retention_days = var.log_bucket_retention_days
}

resource "google_logging_project_sink" "audit" {
  name        = "contract-audit-sink-${var.environment}"
  project     = var.project_id
  destination = "logging.googleapis.com/projects/${var.project_id}/locations/${var.region}/buckets/${google_logging_project_bucket_config.audit.bucket_id}"

  filter                 = "logName:\"cloudaudit.googleapis.com\""
  unique_writer_identity = true
}

resource "google_project_iam_member" "audit_sink_writer" {
  project = var.project_id
  role    = "roles/logging.bucketWriter"
  member  = google_logging_project_sink.audit.writer_identity

  condition {
    title       = "contract-audit-log-bucket"
    description = "Allow audit sink to write only to the contract audit log bucket"
    expression  = "resource.name.endsWith('/locations/${var.region}/buckets/${google_logging_project_bucket_config.audit.bucket_id}')"
  }
}
