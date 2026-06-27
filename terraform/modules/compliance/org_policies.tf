resource "google_org_policy_policy" "uniform_bucket_access" {
  count = var.enable_org_policies ? 1 : 0

  name   = "projects/${var.project_id}/policies/storage.uniformBucketLevelAccess"
  parent = "projects/${var.project_id}"

  spec {
    rules {
      enforce = true
    }
  }
}

resource "google_org_policy_policy" "disable_sa_keys" {
  count = var.enable_org_policies ? 1 : 0

  name   = "projects/${var.project_id}/policies/iam.disableServiceAccountKeyCreation"
  parent = "projects/${var.project_id}"

  spec {
    rules {
      enforce = true
    }
  }
}
