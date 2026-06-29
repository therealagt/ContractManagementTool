output "api_service_name" {
  value = google_cloud_run_v2_service.api.name
}

output "api_service_uri" {
  value = google_cloud_run_v2_service.api.uri
}

output "api_service_id" {
  value = google_cloud_run_v2_service.api.id
}

output "extraction_service_name" {
  value = google_cloud_run_v2_service.extraction.name
}

output "extraction_service_uri" {
  value = google_cloud_run_v2_service.extraction.uri
}

output "archive_service_name" {
  value = google_cloud_run_v2_service.archive.name
}

output "archive_service_uri" {
  value = google_cloud_run_v2_service.archive.uri
}

output "integrity_service_uri" {
  value = google_cloud_run_v2_service.integrity.uri
}
