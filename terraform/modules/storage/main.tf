resource "google_storage_bucket" "staging" {
  name                        = local.staging_name
  location                    = var.region
  project                     = var.project_id
  uniform_bucket_level_access = true
  public_access_prevention    = "enforced"
  force_destroy               = var.environment != "prod"

  encryption {
    default_kms_key_name = var.kms_key_id
  }

  lifecycle_rule {
    condition {
      age = var.staging_ttl_days
    }
    action {
      type = "Delete"
    }
  }

  labels = var.labels
}

resource "google_storage_bucket" "archive" {
  name                        = local.archive_name
  location                    = var.region
  project                     = var.project_id
  uniform_bucket_level_access = true
  public_access_prevention    = "enforced"
  force_destroy               = false
  versioning {
    enabled = true
  }

  encryption {
    default_kms_key_name = var.kms_key_id
  }

  retention_policy {
    retention_period = var.retention_years * 365 * 24 * 60 * 60
    is_locked        = var.enable_bucket_lock
  }

  labels = var.labels
}
