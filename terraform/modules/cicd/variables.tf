variable "project_id" {
  type = string
}

variable "region" {
  type = string
}

variable "environment" {
  type = string
}

variable "github_org" {
  type = string
}

variable "github_repo" {
  type    = string
  default = "ContractManagementTool"
}

variable "artifact_registry_repo" {
  type    = string
  default = ""
}
