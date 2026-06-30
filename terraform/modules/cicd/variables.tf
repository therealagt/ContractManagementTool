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
  type        = string
  description = "GitHub organization or user that owns the repository"
}

variable "github_repo" {
  type        = string
  description = "GitHub repository name"
  default     = "ContractManagementTool"
}

variable "artifact_registry_repo" {
  type    = string
  default = ""
}
