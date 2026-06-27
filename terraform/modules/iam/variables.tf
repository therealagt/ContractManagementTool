variable "project_id" {
  type = string
}

variable "environment" {
  type = string
}

variable "labels" {
  type    = map(string)
  default = {}
}

locals {
  accounts = {
    api       = "sa-contract-api"
    ingestion = "sa-contract-ingestion"
    archive   = "sa-contract-archive"
    report    = "sa-contract-report"
  }
}
