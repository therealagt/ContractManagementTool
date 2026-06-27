output "kms_key_id" {
  value = google_kms_crypto_key.main.id
}

output "kms_key_name" {
  value = google_kms_crypto_key.main.name
}

output "log_bucket_id" {
  value = google_logging_project_bucket_config.audit.bucket_id
}
