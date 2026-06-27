resource "google_kms_key_ring" "main" {
  name     = "${var.kms_keyring_name}-${var.environment}"
  location = var.region
  project  = var.project_id
}

resource "google_kms_crypto_key" "main" {
  name            = "${var.kms_key_name}-${var.environment}"
  key_ring        = google_kms_key_ring.main.id
  rotation_period = "7776000s"

  lifecycle {
    prevent_destroy = true
  }

  labels = var.labels
}
