variable "project_id" {
  type = string
}

variable "region" {
  type = string
}

variable "environment" {
  type = string
}

variable "extraction_worker_uri" {
  type        = string
  description = "Cloud Run URI for extraction worker push endpoint (without path)"
}

resource "google_pubsub_topic" "extraction" {
  name    = "contract-extraction-${var.environment}"
  project = var.project_id
}

resource "google_pubsub_topic" "extraction_dlq" {
  name    = "contract-extraction-dlq-${var.environment}"
  project = var.project_id
}

resource "google_pubsub_subscription" "extraction" {
  name    = "contract-extraction-${var.environment}"
  project = var.project_id
  topic   = google_pubsub_topic.extraction.id

  ack_deadline_seconds = 60

  push_config {
    push_endpoint = "${var.extraction_worker_uri}/pubsub/extraction"

    oidc_token {
      service_account_email = google_service_account.push.email
      audience              = var.extraction_worker_uri
    }
  }

  dead_letter_policy {
    dead_letter_topic     = google_pubsub_topic.extraction_dlq.id
    max_delivery_attempts = 5
  }

  retry_policy {
    minimum_backoff = "10s"
    maximum_backoff = "600s"
  }
}

resource "google_pubsub_subscription" "extraction_dlq" {
  name    = "contract-extraction-dlq-${var.environment}"
  project = var.project_id
  topic   = google_pubsub_topic.extraction_dlq.id

  ack_deadline_seconds = 600
}

resource "google_service_account" "push" {
  account_id   = "sa-pubsub-push-${var.environment}"
  display_name = "Pub/Sub push to extraction worker (${var.environment})"
  project      = var.project_id
}

resource "google_pubsub_subscription_iam_member" "push_subscriber" {
  project      = var.project_id
  subscription = google_pubsub_subscription.extraction.name
  role         = "roles/pubsub.subscriber"
  member       = "serviceAccount:${google_service_account.push.email}"
}

resource "google_pubsub_topic_iam_member" "dlq_publisher" {
  project = var.project_id
  topic   = google_pubsub_topic.extraction_dlq.name
  role    = "roles/pubsub.publisher"
  member  = "serviceAccount:service-${data.google_project.current.number}@gcp-sa-pubsub.iam.gserviceaccount.com"
}

data "google_project" "current" {
  project_id = var.project_id
}
