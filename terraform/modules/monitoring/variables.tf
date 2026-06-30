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

variable "alert_email_admin" {
  type        = string
  description = "Email for P1 escalation after alert_escalation_hours"
  default     = ""
}

variable "alert_escalation_hours" {
  type        = number
  description = "Hours before P1 alerts escalate to admin group"
  default     = 4
}
