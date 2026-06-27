output "dataset_id" {
  value = google_bigquery_dataset.audit.dataset_id
}

output "audit_events_table" {
  value = google_bigquery_table.audit_events.table_id
}

output "access_events_table" {
  value = google_bigquery_table.access_events.table_id
}
