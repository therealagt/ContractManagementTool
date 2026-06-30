output "load_balancer_ip" {
  value       = module.iap.load_balancer_ip
  description = "Point domain A record to this IP"
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

output "extraction_topic" {
  value = module.pubsub.extraction_topic
}

output "cicd_workload_identity_provider" {
  value = var.enable_cicd ? module.cicd[0].workload_identity_provider : ""
}

output "cicd_deploy_service_account" {
  value = var.enable_cicd ? module.cicd[0].deploy_service_account_email : ""
}
