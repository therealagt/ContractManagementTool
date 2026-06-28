resource "google_pubsub_topic" "archive" {
  name    = "contract-archive-${var.environment}"
  project = var.project_id
}

resource "google_pubsub_topic" "archive_dlq" {
  name    = "contract-archive-dlq-${var.environment}"
  project = var.project_id
}

resource "google_pubsub_subscription" "archive" {
  name    = "contract-archive-${var.environment}"
  project = var.project_id
  topic   = google_pubsub_topic.archive.id

  ack_deadline_seconds = 120

  push_config {
    push_endpoint = "${var.archive_worker_uri}/pubsub/archive"

    oidc_token {
      service_account_email = google_service_account.push.email
      audience              = var.archive_worker_uri
    }
  }

  dead_letter_policy {
    dead_letter_topic     = google_pubsub_topic.archive_dlq.id
    max_delivery_attempts = 5
  }

  retry_policy {
    minimum_backoff = "10s"
    maximum_backoff = "600s"
  }
}

resource "google_pubsub_subscription" "archive_dlq" {
  name    = "contract-archive-dlq-${var.environment}"
  project = var.project_id
  topic   = google_pubsub_topic.archive_dlq.id

  ack_deadline_seconds = 600
}

resource "google_pubsub_topic_iam_member" "archive_dlq_publisher" {
  project = var.project_id
  topic   = google_pubsub_topic.archive_dlq.name
  role    = "roles/pubsub.publisher"
  member  = "serviceAccount:service-${data.google_project.current.number}@gcp-sa-pubsub.iam.gserviceaccount.com"
}
