output "extraction_topic" {
  value = google_pubsub_topic.extraction.name
}

output "extraction_topic_id" {
  value = google_pubsub_topic.extraction.id
}

output "archive_topic" {
  value = google_pubsub_topic.archive.name
}

output "archive_topic_id" {
  value = google_pubsub_topic.archive.id
}

output "push_service_account_email" {
  value = google_service_account.push.email
}
