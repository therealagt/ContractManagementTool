output "load_balancer_ip" {
  value       = var.enable_iap ? module.iap[0].load_balancer_ip : null
  description = "Point domain A record to this IP when IAP is enabled"
}

output "api_service_uri" {
  value       = module.cloudrun.api_service_uri
  sensitive   = true
  description = "Internal Cloud Run URI; not for direct public access"
}

output "staging_bucket" {
  value = module.storage.staging_bucket_name
}

output "archive_bucket" {
  value = module.storage.archive_bucket_name
}

output "artifact_registry_url" {
  value = module.artifactregistry.repository_url
}

output "cloud_sql_connection_name" {
  value = module.cloudsql.instance_connection_name
}

output "bigquery_dataset" {
  value = module.bigquery.dataset_id
}

output "api_service_account" {
  value = module.iam.service_account_emails["api"]
}
