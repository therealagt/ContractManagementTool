resource "google_cloud_run_v2_service" "weekly_report" {
  name     = "contract-weekly-report-${var.environment}"
  project  = var.project_id
  location = var.region
  ingress  = "INGRESS_TRAFFIC_INTERNAL_ONLY"

  template {
    service_account = var.report_service_account_email

    scaling {
      min_instance_count = 0
      max_instance_count = 1
    }

    vpc_access {
      connector = var.vpc_connector_id
      egress    = "PRIVATE_RANGES_ONLY"
    }

    containers {
      image = var.weekly_report_image

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
        name  = "CLOUD_SQL_CONNECTION_NAME"
        value = var.cloud_sql_connection_name
      }

      env {
        name  = "DB_NAME"
        value = var.db_name
      }

      env {
        name  = "DB_USER"
        value = var.report_db_user
      }

      env {
        name = "DB_PASSWORD"
        value_source {
          secret_key_ref {
            secret  = var.report_db_password_secret_id
            version = "latest"
          }
        }
      }

      env {
        name  = "REVIEW_SLA_DAYS"
        value = tostring(var.review_sla_days)
      }

      env {
        name  = "REPORT_EMAIL_OPS"
        value = var.alert_email_ops
      }

      env {
        name  = "REPORT_EMAIL_AUDIT"
        value = var.alert_email_audit
      }

      env {
        name  = "EMAIL_FROM"
        value = var.email_from
      }

      env {
        name  = "SMTP_HOST"
        value = var.smtp_host
      }

      env {
        name  = "SMTP_PORT"
        value = tostring(var.smtp_port)
      }

      env {
        name  = "SMTP_USER"
        value = var.smtp_user
      }

      dynamic "env" {
        for_each = var.smtp_password_secret_id != "" ? [1] : []
        content {
          name = "SMTP_PASSWORD"
          value_source {
            secret_key_ref {
              secret  = var.smtp_password_secret_id
              version = "latest"
            }
          }
        }
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

resource "google_cloud_run_v2_service_iam_member" "weekly_report_scheduler_invoker" {
  project  = var.project_id
  location = var.region
  name     = google_cloud_run_v2_service.weekly_report.name
  role     = "roles/run.invoker"
  member   = "serviceAccount:${var.report_service_account_email}"
}

resource "google_cloud_scheduler_job" "weekly_report" {
  count = var.alert_email_ops != "" && var.alert_email_audit != "" ? 1 : 0

  name        = "contract-weekly-report-${var.environment}"
  project     = var.project_id
  region      = var.region
  schedule    = var.weekly_report_schedule
  time_zone   = "Europe/Berlin"
  description = "Weekly operational status report email"

  http_target {
    http_method = "POST"
    uri         = "${google_cloud_run_v2_service.weekly_report.uri}/run"

    oidc_token {
      service_account_email = var.report_service_account_email
      audience              = google_cloud_run_v2_service.weekly_report.uri
    }
  }
}
