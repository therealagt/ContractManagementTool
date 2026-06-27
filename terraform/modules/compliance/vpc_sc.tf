resource "google_access_context_manager_service_perimeter" "main" {
  count = var.enable_vpc_service_controls && var.vpc_sc_access_policy_id != "" ? 1 : 0

  name   = "accessPolicies/${var.vpc_sc_access_policy_id}/servicePerimeters/contract_mgmt_${var.environment}"
  parent = "accessPolicies/${var.vpc_sc_access_policy_id}"
  title  = "contract-mgmt-${var.environment}"

  status {
    restricted_services = [
      "storage.googleapis.com",
      "bigquery.googleapis.com",
      "aiplatform.googleapis.com",
      "sqladmin.googleapis.com",
    ]

    resources = ["projects/${data.google_project.current.number}"]

    dynamic "ingress_policies" {
      for_each = length(var.vpc_sc_ingress_identities) > 0 ? [1] : []
      content {
        ingress_from {
          identities = var.vpc_sc_ingress_identities
        }
        ingress_to {
          operations {
            service_name = "storage.googleapis.com"
          }
          operations {
            service_name = "bigquery.googleapis.com"
          }
          operations {
            service_name = "aiplatform.googleapis.com"
          }
          operations {
            service_name = "sqladmin.googleapis.com"
          }
        }
      }
    }
  }
}

data "google_project" "current" {
  project_id = var.project_id
}
