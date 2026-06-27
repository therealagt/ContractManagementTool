variable "project_id" {
  type = string
}

variable "region" {
  type = string
}

variable "environment" {
  type = string
}

variable "kms_key_id" {
  type = string
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

variable "labels" {
  type    = map(string)
  default = {}
}

locals {
  staging_name = "${var.project_id}-contracts-staging-${var.environment}"
  archive_name = "${var.project_id}-contracts-archive-${var.environment}"
}
