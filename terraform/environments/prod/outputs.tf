output "load_balancer_ip" {
  value       = module.stack.load_balancer_ip
  description = "Point domain A record to this IP when IAP is enabled"
}

output "api_service_uri" {
  value       = module.stack.api_service_uri
  sensitive   = true
  description = "Internal Cloud Run URI; not for direct public access"
}

output "staging_bucket" {
  value = module.stack.staging_bucket
}

output "archive_bucket" {
  value = module.stack.archive_bucket
}

output "artifact_registry_url" {
  value = module.stack.artifact_registry_url
}

output "cloud_sql_connection_name" {
  value = module.stack.cloud_sql_connection_name
}

output "bigquery_dataset" {
  value = module.stack.bigquery_dataset
}

output "api_service_account" {
  value = module.stack.api_service_account
}
