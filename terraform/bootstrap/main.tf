terraform {
  required_version = ">= 1.5.0"

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.45"
    }
  }
}

variable "project_id" {
  type = string
}

variable "region" {
  type    = string
  default = "europe-west3"
}

variable "state_bucket_name" {
  type = string
}

provider "google" {
  project = var.project_id
  region  = var.region
}

resource "google_storage_bucket" "terraform_state" {
  name                        = var.state_bucket_name
  location                    = var.region
  project                     = var.project_id
  uniform_bucket_level_access = true
  public_access_prevention    = "enforced"
  force_destroy               = false

  labels = {
    purpose    = "terraform-state"
    managed_by = "terraform"
  }

  versioning {
    enabled = true
  }

  lifecycle_rule {
    condition {
      num_newer_versions = 10
    }
    action {
      type = "Delete"
    }
  }
}

output "state_bucket" {
  value = google_storage_bucket.terraform_state.name
}
