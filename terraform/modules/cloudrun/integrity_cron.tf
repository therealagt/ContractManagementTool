resource "google_cloud_run_v2_service" "integrity" {
  name     = "contract-integrity-${var.environment}"
  project  = var.project_id
  location = var.region
  ingress  = "INGRESS_TRAFFIC_INTERNAL_ONLY"

  template {
    service_account = var.integrity_service_account_email

    scaling {
      min_instance_count = 0
      max_instance_count = 2
    }

    vpc_access {
      connector = var.vpc_connector_id
      egress    = "PRIVATE_RANGES_ONLY"
    }

    containers {
      image = var.integrity_cron_image

      ports {
        container_port = 8080
      }

      env {
        name  = "ENVIRONMENT"
        value = var.environment
      }

      env {
        name  = "GCP_PROJECT_ID"
        value = var.project_id
      }

      env {
        name  = "GCP_REGION"
        value = var.region
      }

      env {
        name  = "CLOUD_SQL_CONNECTION_NAME"
        value = var.cloud_sql_connection_name
      }

      env {
        name  = "DB_NAME"
        value = var.db_name
      }

      env {
        name  = "DB_USER"
        value = var.db_user
      }

      env {
        name = "DB_PASSWORD"
        value_source {
          secret_key_ref {
            secret  = var.db_password_secret_id
            version = "latest"
          }
        }
      }

      env {
        name  = "GCS_ARCHIVE_BUCKET"
        value = var.archive_bucket_name
      }

      env {
        name  = "BIGQUERY_DATASET"
        value = var.bigquery_dataset_id
      }

      env {
        name  = "REVIEW_SLA_DAYS"
        value = tostring(var.review_sla_days)
      }

      resources {
        limits = {
          cpu    = "1"
          memory = "512Mi"
        }
      }

      startup_probe {
        http_get {
          path = "/health"
          port = 8080
        }
        initial_delay_seconds = 5
        timeout_seconds       = 3
        period_seconds        = 10
        failure_threshold     = 3
      }
    }

    annotations = {
      "run.googleapis.com/cloudsql-instances" = var.cloud_sql_connection_name
    }
  }

  lifecycle {
    ignore_changes = [
      template[0].containers[0].image,
    ]
  }
}

resource "google_cloud_run_v2_service_iam_member" "integrity_scheduler_invoker" {
  project  = var.project_id
  location = var.region
  name     = google_cloud_run_v2_service.integrity.name
  role     = "roles/run.invoker"
  member   = "serviceAccount:${var.integrity_service_account_email}"
}

resource "google_cloud_scheduler_job" "integrity_nightly" {
  name        = "contract-integrity-${var.environment}"
  project     = var.project_id
  region      = var.region
  schedule    = "0 2 * * *"
  time_zone   = "Europe/Berlin"
  description = "Nightly archive integrity recheck"

  http_target {
    http_method = "POST"
    uri         = "${google_cloud_run_v2_service.integrity.uri}/run"

    oidc_token {
      service_account_email = var.integrity_service_account_email
      audience              = google_cloud_run_v2_service.integrity.uri
    }
  }
}
