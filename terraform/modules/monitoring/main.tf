resource "google_monitoring_notification_channel" "ops_email" {
  count = var.alert_email_ops != "" ? 1 : 0

  project      = var.project_id
  display_name = "Contract Ops (${var.environment})"
  type         = "email"
  labels = {
    email_address = var.alert_email_ops
  }
}

resource "google_monitoring_notification_channel" "admin_email" {
  count = var.alert_email_admin != "" ? 1 : 0

  project      = var.project_id
  display_name = "Contract Admin escalation (${var.environment})"
  type         = "email"
  labels = {
    email_address = var.alert_email_admin
  }
}

locals {
  escalation_duration = "${var.alert_escalation_hours * 3600}s"
  ops_channels        = var.alert_email_ops != "" ? [google_monitoring_notification_channel.ops_email[0].id] : []
  admin_channels      = var.alert_email_admin != "" ? [google_monitoring_notification_channel.admin_email[0].id] : []
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

  notification_channels = local.ops_channels
  alert_strategy {
    auto_close = "604800s"
  }

  documentation {
    content   = "Archive SHA-256 mismatch detected. Do not auto-repair — follow docs/verfahrensdokumentation.md P1 runbook."
    mime_type = "text/markdown"
  }
}

resource "google_monitoring_alert_policy" "integrity_failed_escalation" {
  count = var.alert_email_ops != "" && var.alert_email_admin != "" ? 1 : 0

  project      = var.project_id
  display_name = "Contract archive integrity failure escalation (${var.environment})"
  combiner     = "OR"

  conditions {
    display_name = "integrity_check_failed > 0 for ${var.alert_escalation_hours}h"
    condition_threshold {
      filter          = "resource.type = \"global\" AND metric.type = \"custom.googleapis.com/contract/integrity_check_failed\""
      comparison      = "COMPARISON_GT"
      threshold_value = 0
      duration        = local.escalation_duration
      aggregations {
        alignment_period   = "300s"
        per_series_aligner = "ALIGN_MAX"
      }
    }
  }

  notification_channels = local.admin_channels

  documentation {
    content   = "P1 integrity incident open for ${var.alert_escalation_hours}+ hours. Escalate to admin — see docs/verfahrensdokumentation.md."
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

  notification_channels = local.ops_channels
}

resource "google_monitoring_alert_policy" "audit_chain_broken_escalation" {
  count = var.alert_email_ops != "" && var.alert_email_admin != "" ? 1 : 0

  project      = var.project_id
  display_name = "Contract audit chain broken escalation (${var.environment})"
  combiner     = "OR"

  conditions {
    display_name = "audit_chain_broken > 0 for ${var.alert_escalation_hours}h"
    condition_threshold {
      filter          = "resource.type = \"global\" AND metric.type = \"custom.googleapis.com/contract/audit_chain_broken\""
      comparison      = "COMPARISON_GT"
      threshold_value = 0
      duration        = local.escalation_duration
      aggregations {
        alignment_period   = "300s"
        per_series_aligner = "ALIGN_MAX"
      }
    }
  }

  notification_channels = local.admin_channels
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

  notification_channels = local.ops_channels
}
