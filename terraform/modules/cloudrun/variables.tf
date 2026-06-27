variable "project_id" {
  type = string
}

variable "region" {
  type = string
}

variable "environment" {
  type = string
}

variable "api_service_account_email" {
  type = string
}

variable "vpc_connector_id" {
  type = string
}

variable "api_image" {
  type = string
}

variable "cloud_sql_connection_name" {
  type = string
}

variable "db_name" {
  type = string
}

variable "db_user" {
  type = string
}

variable "db_password_secret_id" {
  type = string
}

variable "staging_bucket_name" {
  type = string
}

variable "archive_bucket_name" {
  type = string
}

variable "iap_audience" {
  type        = string
  description = "OAuth client ID used as IAP JWT audience"
  default     = ""
}

variable "iap_jwt_validation_disabled" {
  type    = bool
  default = false
}

variable "allowed_email_domains" {
  type    = list(string)
  default = []
}

variable "auth_uploader_emails" {
  type    = list(string)
  default = []
}

variable "auth_reviewer_emails" {
  type    = list(string)
  default = []
}

variable "auth_auditor_emails" {
  type    = list(string)
  default = []
}

variable "auth_admin_emails" {
  type    = list(string)
  default = []
}

variable "extraction_worker_image" {
  type        = string
  description = "Container image for extraction worker"
  default     = "us-docker.pkg.dev/cloudrun/container/hello"
}

variable "ingestion_service_account_email" {
  type    = string
  default = ""
}

variable "push_invoker_member" {
  type    = string
  default = ""
}

variable "gemini_model" {
  type    = string
  default = "gemini-2.0-flash"
}
