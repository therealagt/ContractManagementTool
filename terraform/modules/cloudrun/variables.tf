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

variable "ingestion_db_user" {
  type = string
}

variable "ingestion_db_password_secret_id" {
  type = string
}

variable "archive_db_user" {
  type = string
}

variable "archive_db_password_secret_id" {
  type = string
}

variable "integrity_db_user" {
  type = string
}

variable "integrity_db_password_secret_id" {
  type = string
}

variable "report_db_user" {
  type = string
}

variable "report_db_password_secret_id" {
  type = string
}

variable "pades_allow_untrusted_roots" {
  type = bool
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

variable "archive_worker_image" {
  type        = string
  description = "Container image for archive worker"
  default     = "us-docker.pkg.dev/cloudrun/container/hello"
}

variable "ingestion_service_account_email" {
  type    = string
  default = ""
}

variable "archive_service_account_email" {
  type    = string
  default = ""
}

variable "integrity_service_account_email" {
  type    = string
  default = ""
}

variable "integrity_cron_image" {
  type        = string
  description = "Container image for integrity cron"
  default     = "us-docker.pkg.dev/cloudrun/container/hello"
}

variable "review_sla_days" {
  type    = number
  default = 7
}

variable "bigquery_dataset_id" {
  type    = string
  default = ""
}

variable "retention_years" {
  type    = number
  default = 10
}

variable "push_invoker_member" {
  type    = string
  default = ""
}

variable "gemini_model" {
  type    = string
  default = "gemini-2.0-flash"
}

variable "weekly_report_image" {
  type    = string
  default = "us-docker.pkg.dev/cloudrun/container/hello"
}

variable "report_service_account_email" {
  type    = string
  default = ""
}

variable "alert_email_ops" {
  type    = string
  default = ""
}

variable "alert_email_audit" {
  type    = string
  default = ""
}

variable "email_from" {
  type    = string
  default = ""
}

variable "smtp_host" {
  type    = string
  default = ""
}

variable "smtp_port" {
  type    = number
  default = 587
}

variable "smtp_user" {
  type    = string
  default = ""
}

variable "smtp_password_secret_id" {
  type    = string
  default = ""
}

variable "weekly_report_schedule" {
  type    = string
  default = "0 8 * * 1"
}
