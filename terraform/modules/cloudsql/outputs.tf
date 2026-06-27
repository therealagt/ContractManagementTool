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

output "private_ip_address" {
  value = google_sql_database_instance.main.private_ip_address
}
