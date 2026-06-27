resource "google_cloud_run_v2_service" "api" {
  name     = "contract-api-${var.environment}"
  project  = var.project_id
  location = var.region
  ingress  = "INGRESS_TRAFFIC_INTERNAL_LOAD_BALANCER"

  template {
    service_account = var.api_service_account_email

    scaling {
      min_instance_count = 0
      max_instance_count = 5
    }

    vpc_access {
      connector = var.vpc_connector_id
      egress    = "PRIVATE_RANGES_ONLY"
    }

    containers {
      image = var.api_image

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
        name  = "GCS_STAGING_BUCKET"
        value = var.staging_bucket_name
      }

      env {
        name  = "GCS_ARCHIVE_BUCKET"
        value = var.archive_bucket_name
      }

      env {
        name  = "IAP_AUDIENCE"
        value = var.iap_audience
      }

      env {
        name  = "IAP_JWT_VALIDATION_DISABLED"
        value = tostring(var.iap_jwt_validation_disabled)
      }

      env {
        name  = "ALLOWED_EMAIL_DOMAINS"
        value = join(",", var.allowed_email_domains)
      }

      env {
        name  = "AUTH_UPLOADER_EMAILS"
        value = join(",", var.auth_uploader_emails)
      }

      env {
        name  = "AUTH_REVIEWER_EMAILS"
        value = join(",", var.auth_reviewer_emails)
      }

      env {
        name  = "AUTH_AUDITOR_EMAILS"
        value = join(",", var.auth_auditor_emails)
      }

      env {
        name  = "AUTH_ADMIN_EMAILS"
        value = join(",", var.auth_admin_emails)
      }

      env {
        name  = "PUBSUB_EXTRACTION_TOPIC"
        value = "contract-extraction-${var.environment}"
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

      liveness_probe {
        http_get {
          path = "/health"
          port = 8080
        }
        timeout_seconds   = 3
        period_seconds    = 30
        failure_threshold = 3
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

data "google_project" "current" {
  project_id = var.project_id
}

resource "google_cloud_run_v2_service_iam_member" "api_invoker_lb" {
  project  = var.project_id
  location = var.region
  name     = google_cloud_run_v2_service.api.name
  role     = "roles/run.invoker"
  member   = "serviceAccount:service-${data.google_project.current.number}@serverless-robot-prod.iam.gserviceaccount.com"
}
