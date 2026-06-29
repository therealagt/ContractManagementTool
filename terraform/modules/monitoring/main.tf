variable "project_id" {
  type = string
}

variable "environment" {
  type = string
}

variable "alert_email_ops" {
  type        = string
  description = "Email for P1/P2 operational alerts"
  default     = ""
}

resource "google_monitoring_notification_channel" "ops_email" {
  count = var.alert_email_ops != "" ? 1 : 0

  project      = var.project_id
  display_name = "Contract Ops (${var.environment})"
  type         = "email"
  labels = {
    email_address = var.alert_email_ops
  }
}

resource "google_monitoring_alert_policy" "integrity_failed" {
  count = var.alert_email_ops != "" ? 1 : 0

  project      = var.project_id
  display_name = "Contract archive integrity failure (${var.environment})"
  combiner     = "OR"

  conditions {
    display_name = "integrity_check_failed > 0"
    condition_threshold {
      filter          = "resource.type = \"global\" AND metric.type = \"custom.googleapis.com/contract/integrity_check_failed\""
      comparison      = "COMPARISON_GT"
      threshold_value = 0
      duration        = "0s"
      aggregations {
        alignment_period   = "60s"
        per_series_aligner = "ALIGN_MAX"
      }
    }
  }

  notification_channels = [google_monitoring_notification_channel.ops_email[0].id]
  alert_strategy {
    auto_close = "604800s"
  }

  documentation {
    content   = "Archive SHA-256 mismatch detected. Do not auto-repair — follow incident runbook."
    mime_type = "text/markdown"
  }
}

resource "google_monitoring_alert_policy" "audit_chain_broken" {
  count = var.alert_email_ops != "" ? 1 : 0

  project      = var.project_id
  display_name = "Contract audit chain broken (${var.environment})"
  combiner     = "OR"

  conditions {
    display_name = "audit_chain_broken > 0"
    condition_threshold {
      filter          = "resource.type = \"global\" AND metric.type = \"custom.googleapis.com/contract/audit_chain_broken\""
      comparison      = "COMPARISON_GT"
      threshold_value = 0
      duration        = "0s"
      aggregations {
        alignment_period   = "60s"
        per_series_aligner = "ALIGN_MAX"
      }
    }
  }

  notification_channels = [google_monitoring_notification_channel.ops_email[0].id]
}

resource "google_monitoring_alert_policy" "review_sla" {
  count = var.alert_email_ops != "" ? 1 : 0

  project      = var.project_id
  display_name = "Contract review SLA exceeded (${var.environment})"
  combiner     = "OR"

  conditions {
    display_name = "pending_review_sla_exceeded > 0"
    condition_threshold {
      filter          = "resource.type = \"global\" AND metric.type = \"custom.googleapis.com/contract/pending_review_sla_exceeded\""
      comparison      = "COMPARISON_GT"
      threshold_value = 0
      duration        = "300s"
      aggregations {
        alignment_period   = "300s"
        per_series_aligner = "ALIGN_MAX"
      }
    }
  }

  notification_channels = [google_monitoring_notification_channel.ops_email[0].id]
}
