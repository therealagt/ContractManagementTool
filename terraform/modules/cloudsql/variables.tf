variable "project_id" {
  type = string
}

variable "region" {
  type = string
}

variable "environment" {
  type = string
}

variable "network_id" {
  type = string
}

variable "kms_key_id" {
  type = string
}

variable "api_service_account_email" {
  type = string
}

variable "db_tier" {
  type    = string
  default = "db-f1-micro"
}

variable "db_name" {
  type    = string
  default = "contracts"
}

variable "db_user" {
  type    = string
  default = "contract_api"
}

variable "deletion_protection" {
  type    = bool
  default = false
}
