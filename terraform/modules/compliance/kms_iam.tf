data "google_project" "kms" {
  project_id = var.project_id
}

resource "google_kms_crypto_key_iam_member" "gcs_cmek" {
  crypto_key_id = google_kms_crypto_key.main.id
  role          = "roles/cloudkms.cryptoKeyEncrypterDecrypter"
  member        = "serviceAccount:service-${data.google_project.kms.number}@gs-project-accounts.iam.gserviceaccount.com"
}

resource "google_kms_crypto_key_iam_member" "bq_cmek" {
  crypto_key_id = google_kms_crypto_key.main.id
  role          = "roles/cloudkms.cryptoKeyEncrypterDecrypter"
  member        = "serviceAccount:bq-${data.google_project.kms.number}@gcp-sa-bigquery-iam.iam.gserviceaccount.com"
}

resource "google_kms_crypto_key_iam_member" "sql_cmek" {
  crypto_key_id = google_kms_crypto_key.main.id
  role          = "roles/cloudkms.cryptoKeyEncrypterDecrypter"
  member        = "serviceAccount:service-${data.google_project.kms.number}@gcp-sa-cloud-sql.iam.gserviceaccount.com"
}

resource "google_kms_crypto_key_iam_member" "secretmanager_cmek" {
  crypto_key_id = google_kms_crypto_key.main.id
  role          = "roles/cloudkms.cryptoKeyEncrypterDecrypter"
  member        = "serviceAccount:service-${data.google_project.kms.number}@gcp-sa-secretmanager.iam.gserviceaccount.com"
}

resource "google_kms_crypto_key_iam_member" "api_cmek" {
  count = var.api_service_account_email != "" ? 1 : 0

  crypto_key_id = google_kms_crypto_key.main.id
  role          = "roles/cloudkms.cryptoKeyEncrypterDecrypter"
  member        = "serviceAccount:${var.api_service_account_email}"
}
