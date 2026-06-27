resource "google_service_account" "accounts" {
  for_each = local.accounts

  project      = var.project_id
  account_id   = "${each.value}-${var.environment}"
  display_name = "Contract Management ${each.key} (${var.environment})"
}

resource "google_project_iam_member" "api_cloudsql_client" {
  project = var.project_id
  role    = "roles/cloudsql.client"
  member  = "serviceAccount:${google_service_account.accounts["api"].email}"
}

resource "google_project_iam_member" "ingestion_vertex" {
  project = var.project_id
  role    = "roles/aiplatform.user"
  member  = "serviceAccount:${google_service_account.accounts["ingestion"].email}"
}

resource "google_project_iam_member" "report_bigquery" {
  project = var.project_id
  role    = "roles/bigquery.dataViewer"
  member  = "serviceAccount:${google_service_account.accounts["report"].email}"
}

resource "google_project_iam_member" "report_bigquery_job" {
  project = var.project_id
  role    = "roles/bigquery.jobUser"
  member  = "serviceAccount:${google_service_account.accounts["report"].email}"
}
