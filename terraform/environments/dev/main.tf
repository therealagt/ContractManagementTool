terraform {
  required_version = ">= 1.5.0"

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.45"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.6"
    }
  }
}

provider "google" {
  project = var.project_id
  region  = var.region
}

module "stack" {
  source = "../../modules/stack"

  project_id  = var.project_id
  region      = var.region
  environment = var.environment
  domain      = var.domain

  staging_ttl_days   = var.staging_ttl_days
  retention_years    = var.retention_years
  enable_bucket_lock = var.enable_bucket_lock
  db_tier            = var.db_tier

  enable_iap              = var.enable_iap
  iap_oauth_client_id     = var.iap_oauth_client_id
  iap_oauth_client_secret = var.iap_oauth_client_secret
  iap_access_groups       = var.iap_access_groups
  cloud_armor_allowed_ips = var.cloud_armor_allowed_ips

  enable_org_policies         = var.enable_org_policies
  enable_vpc_service_controls = var.enable_vpc_service_controls
  vpc_sc_access_policy_id     = var.vpc_sc_access_policy_id

  api_image             = var.api_image
  allowed_email_domains = var.allowed_email_domains
  auth_uploader_emails  = var.auth_uploader_emails
  auth_reviewer_emails  = var.auth_reviewer_emails
  auth_auditor_emails   = var.auth_auditor_emails
  auth_admin_emails     = var.auth_admin_emails
}
