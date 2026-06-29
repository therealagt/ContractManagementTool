output "ops_notification_channel" {
  value = length(google_monitoring_notification_channel.ops_email) > 0 ? google_monitoring_notification_channel.ops_email[0].id : ""
}
