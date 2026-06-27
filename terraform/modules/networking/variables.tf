variable "project_id" {
  type = string
}

variable "region" {
  type = string
}

variable "environment" {
  type = string
}

locals {
  network_name = "contract-vpc-${var.environment}"
}
