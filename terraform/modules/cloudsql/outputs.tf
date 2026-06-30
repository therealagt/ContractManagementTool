output "instance_connection_name" {
  value = google_sql_database_instance.main.connection_name
}

output "database_name" {
  value = google_sql_database.main.name
}

output "database_user" {
  value = google_sql_user.main.name
}

output "db_password_secret_id" {
  value = google_secret_manager_secret.db_password.secret_id
}

output "ingestion_database_user" {
  value = google_sql_user.ingestion.name
}

output "ingestion_db_password_secret_id" {
  value = google_secret_manager_secret.db_password_ingestion.secret_id
}

output "archive_database_user" {
  value = google_sql_user.archive.name
}

output "archive_db_password_secret_id" {
  value = google_secret_manager_secret.db_password_archive.secret_id
}

output "integrity_database_user" {
  value = google_sql_user.integrity.name
}

output "integrity_db_password_secret_id" {
  value = google_secret_manager_secret.db_password_integrity.secret_id
}

output "report_database_user" {
  value = google_sql_user.report.name
}

output "report_db_password_secret_id" {
  value = google_secret_manager_secret.db_password_report.secret_id
}

output "private_ip_address" {
  value = google_sql_database_instance.main.private_ip_address
}
