variable "project_id" {
  type = string
}

variable "region" {
  type = string
}

variable "environment" {
  type = string
}

variable "kms_keyring_name" {
  type    = string
  default = "contract-mgmt-keyring"
}

variable "kms_key_name" {
  type    = string
  default = "contract-mgmt-key"
}

variable "log_bucket_retention_days" {
  type    = number
  default = 2555
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

variable "labels" {
  type    = map(string)
  default = {}
}

variable "vpc_sc_ingress_identities" {
  type        = list(string)
  default     = []
  description = "Service account identities allowed to ingress the VPC SC perimeter"
}
