variable "project_id" {
  type = string
}

variable "region" {
  type    = string
  default = "europe-west3"
}

variable "environment" {
  type    = string
  default = "dev"
}

variable "domain" {
  type        = string
  description = "Domain for HTTPS LB and managed SSL certificate"
}

variable "staging_ttl_days" {
  type    = number
  default = 30
}

variable "retention_years" {
  type    = number
  default = 10
}

variable "enable_bucket_lock" {
  type    = bool
  default = false
}

variable "db_tier" {
  type    = string
  default = "db-f1-micro"
}

variable "enable_iap" {
  type        = bool
  default     = false
  description = "Enable IAP + HTTPS LB (requires OAuth client credentials)"
}

variable "iap_oauth_client_id" {
  type      = string
  default   = ""
  sensitive = true
}

variable "iap_oauth_client_secret" {
  type      = string
  default   = ""
  sensitive = true
}

variable "iap_access_groups" {
  type    = list(string)
  default = []
}

variable "cloud_armor_allowed_ips" {
  type    = list(string)
  default = []
}

variable "enable_org_policies" {
  type    = bool
  default = false
}

variable "enable_vpc_service_controls" {
  type    = bool
  default = false
}

variable "vpc_sc_access_policy_id" {
  type    = string
  default = ""
}

variable "api_image" {
  type        = string
  description = "Container image for contract-api"
  default     = "us-docker.pkg.dev/cloudrun/container/hello"
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

variable "integrity_cron_image" {
  type        = string
  description = "Container image for integrity cron"
  default     = "us-docker.pkg.dev/cloudrun/container/hello"
}

variable "review_sla_days" {
  type    = number
  default = 7
}

variable "alert_email_ops" {
  type        = string
  description = "Email for P1/P2 operational alerts"
  default     = ""
}

variable "gemini_model" {
  type    = string
  default = "gemini-2.0-flash"
}

variable "allowed_email_domains" {
  type        = list(string)
  description = "Allowed Google account email domains for IAP-authenticated API access"
  default     = []
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
