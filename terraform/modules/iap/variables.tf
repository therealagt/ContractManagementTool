variable "project_id" {
  type = string
}

variable "region" {
  type = string
}

variable "environment" {
  type = string
}

variable "domain" {
  type = string
}

variable "api_service_name" {
  type = string
}

variable "iap_oauth_client_id" {
  type        = string
  description = "OAuth client ID for IAP (create manually or via brand)"
}

variable "iap_oauth_client_secret" {
  type        = string
  sensitive   = true
  description = "OAuth client secret for IAP backend"
}

variable "iap_access_groups" {
  type        = list(string)
  description = "Google Groups emails granted IAP access"
  default     = []
}

variable "cloud_armor_allowed_ips" {
  type    = list(string)
  default = []
}
