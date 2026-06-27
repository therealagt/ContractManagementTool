output "staging_bucket_name" {
  value = google_storage_bucket.staging.name
}

output "archive_bucket_name" {
  value = google_storage_bucket.archive.name
}
