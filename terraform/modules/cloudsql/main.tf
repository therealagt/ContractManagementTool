resource "random_password" "db_password" {
  length  = 32
  special = false
}

resource "google_sql_database_instance" "main" {
  name             = "contract-db-${var.environment}"
  project          = var.project_id
  region           = var.region
  database_version = "POSTGRES_15"

  deletion_protection = var.deletion_protection

  settings {
    tier              = var.db_tier
    availability_type = "ZONAL"
    disk_autoresize   = true
    disk_size         = 10

    ip_configuration {
      ipv4_enabled    = false
      private_network = var.network_id
    }

    database_flags {
      name  = "cloudsql.iam_authentication"
      value = "on"
    }

    backup_configuration {
      enabled                        = true
      point_in_time_recovery_enabled = true
      transaction_log_retention_days = 7
    }
  }

  encryption_key_name = var.kms_key_id
}

resource "google_sql_database" "main" {
  name     = var.db_name
  instance = google_sql_database_instance.main.name
  project  = var.project_id
}

resource "google_sql_user" "main" {
  name     = var.db_user
  instance = google_sql_database_instance.main.name
  project  = var.project_id
  password = random_password.db_password.result
}

resource "google_secret_manager_secret" "db_password" {
  project   = var.project_id
  secret_id = "contract-db-password-${var.environment}"

  replication {
    auto {
      customer_managed_encryption {
        kms_key_name = var.kms_key_id
      }
    }
  }
}

resource "google_secret_manager_secret_version" "db_password" {
  secret      = google_secret_manager_secret.db_password.id
  secret_data = random_password.db_password.result
}

resource "google_secret_manager_secret_iam_member" "api_db_password" {
  project   = var.project_id
  secret_id = google_secret_manager_secret.db_password.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${var.api_service_account_email}"
}

resource "google_secret_manager_secret_iam_member" "ingestion_db_password" {
  project   = var.project_id
  secret_id = google_secret_manager_secret.db_password.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${var.ingestion_service_account_email}"
}
