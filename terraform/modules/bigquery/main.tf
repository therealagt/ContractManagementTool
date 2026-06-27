resource "google_bigquery_dataset" "audit" {
  dataset_id                 = "contract_audit_${var.environment}"
  project                    = var.project_id
  location                   = var.region
  delete_contents_on_destroy = var.environment != "prod"

  default_encryption_configuration {
    kms_key_name = var.kms_key_id
  }
}

resource "google_bigquery_table" "audit_events" {
  dataset_id = google_bigquery_dataset.audit.dataset_id
  table_id   = "audit_events"
  project    = var.project_id

  schema = jsonencode([
    { name = "id", type = "STRING", mode = "REQUIRED" },
    { name = "contract_id", type = "STRING", mode = "NULLABLE" },
    { name = "actor", type = "STRING", mode = "REQUIRED" },
    { name = "action", type = "STRING", mode = "REQUIRED" },
    { name = "payload_json", type = "JSON", mode = "NULLABLE" },
    { name = "prev_event_hash", type = "STRING", mode = "NULLABLE" },
    { name = "event_hash", type = "STRING", mode = "NULLABLE" },
    { name = "created_at", type = "TIMESTAMP", mode = "REQUIRED" },
  ])
}

resource "google_bigquery_table" "access_events" {
  dataset_id = google_bigquery_dataset.audit.dataset_id
  table_id   = "access_events"
  project    = var.project_id

  schema = jsonencode([
    { name = "id", type = "STRING", mode = "REQUIRED" },
    { name = "actor", type = "STRING", mode = "REQUIRED" },
    { name = "resource_type", type = "STRING", mode = "REQUIRED" },
    { name = "resource_id", type = "STRING", mode = "NULLABLE" },
    { name = "action", type = "STRING", mode = "REQUIRED" },
    { name = "ip", type = "STRING", mode = "NULLABLE" },
    { name = "created_at", type = "TIMESTAMP", mode = "REQUIRED" },
  ])
}

resource "google_bigquery_table" "alert_events" {
  dataset_id = google_bigquery_dataset.audit.dataset_id
  table_id   = "alert_events"
  project    = var.project_id

  schema = jsonencode([
    { name = "id", type = "STRING", mode = "REQUIRED" },
    { name = "severity", type = "STRING", mode = "REQUIRED" },
    { name = "source", type = "STRING", mode = "REQUIRED" },
    { name = "payload_json", type = "JSON", mode = "NULLABLE" },
    { name = "incident_id", type = "STRING", mode = "NULLABLE" },
    { name = "created_at", type = "TIMESTAMP", mode = "REQUIRED" },
  ])
}
