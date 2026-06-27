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

data "google_project" "current" {
  project_id = var.project_id
}

locals {
  labels = {
    environment = var.environment
    managed_by  = "terraform"
    application = "contract-management"
  }

  required_apis = [
    "compute.googleapis.com",
    "run.googleapis.com",
    "sqladmin.googleapis.com",
    "servicenetworking.googleapis.com",
    "vpcaccess.googleapis.com",
    "secretmanager.googleapis.com",
    "artifactregistry.googleapis.com",
    "bigquery.googleapis.com",
    "cloudkms.googleapis.com",
    "iap.googleapis.com",
    "logging.googleapis.com",
  ]
}

resource "google_project_service" "apis" {
  for_each = toset(local.required_apis)

  project            = var.project_id
  service            = each.value
  disable_on_destroy = false
}

module "compliance" {
  source = "../../modules/compliance"

  project_id                  = var.project_id
  region                      = var.region
  environment                 = var.environment
  enable_org_policies         = var.enable_org_policies
  enable_vpc_service_controls = var.enable_vpc_service_controls
  vpc_sc_access_policy_id     = var.vpc_sc_access_policy_id
  labels                      = local.labels
  vpc_sc_ingress_identities = [
    "serviceAccount:${module.iam.service_account_emails["api"]}",
    "serviceAccount:${module.iam.service_account_emails["ingestion"]}",
    "serviceAccount:${module.iam.service_account_emails["archive"]}",
    "serviceAccount:${module.iam.service_account_emails["report"]}",
  ]

  depends_on = [google_project_service.apis]
}

module "iam" {
  source = "../../modules/iam"

  project_id  = var.project_id
  environment = var.environment
  labels      = local.labels

  depends_on = [google_project_service.apis]
}

module "networking" {
  source = "../../modules/networking"

  project_id  = var.project_id
  region      = var.region
  environment = var.environment

  depends_on = [google_project_service.apis]
}

module "storage" {
  source = "../../modules/storage"

  project_id         = var.project_id
  region             = var.region
  environment        = var.environment
  kms_key_id         = module.compliance.kms_key_id
  staging_ttl_days   = var.staging_ttl_days
  retention_years    = var.retention_years
  enable_bucket_lock = var.enable_bucket_lock
  labels             = local.labels

  depends_on = [module.compliance]
}

module "cloudsql" {
  source = "../../modules/cloudsql"

  project_id                = var.project_id
  region                    = var.region
  environment               = var.environment
  network_id                = module.networking.network_id
  kms_key_id                = module.compliance.kms_key_id
  api_service_account_email = module.iam.service_account_emails["api"]
  db_tier                   = var.db_tier
  deletion_protection       = var.environment == "prod"

  depends_on = [module.networking, module.compliance, module.iam]
}

module "bigquery" {
  source = "../../modules/bigquery"

  project_id  = var.project_id
  region      = var.region
  environment = var.environment
  kms_key_id  = module.compliance.kms_key_id

  depends_on = [module.compliance]
}

module "artifactregistry" {
  source = "../../modules/artifactregistry"

  project_id  = var.project_id
  region      = var.region
  environment = var.environment

  depends_on = [google_project_service.apis]
}

module "cloudrun" {
  source = "../../modules/cloudrun"

  project_id                  = var.project_id
  region                      = var.region
  environment                 = var.environment
  api_service_account_email   = module.iam.service_account_emails["api"]
  vpc_connector_id            = module.networking.vpc_connector_id
  api_image                   = var.api_image
  cloud_sql_connection_name   = module.cloudsql.instance_connection_name
  db_name                     = module.cloudsql.database_name
  db_user                     = module.cloudsql.database_user
  db_password_secret_id       = module.cloudsql.db_password_secret_id
  staging_bucket_name         = module.storage.staging_bucket_name
  archive_bucket_name         = module.storage.archive_bucket_name
  iap_audience                = var.enable_iap ? var.iap_oauth_client_id : ""
  iap_jwt_validation_disabled = !var.enable_iap
  allowed_email_domains       = var.allowed_email_domains
  auth_uploader_emails        = var.auth_uploader_emails
  auth_reviewer_emails        = var.auth_reviewer_emails
  auth_auditor_emails         = var.auth_auditor_emails
  auth_admin_emails           = var.auth_admin_emails

  depends_on = [module.cloudsql, module.storage, module.networking, module.iam]
}

module "iap" {
  count  = var.enable_iap ? 1 : 0
  source = "../../modules/iap"

  project_id              = var.project_id
  region                  = var.region
  environment             = var.environment
  domain                  = var.domain
  api_service_name        = module.cloudrun.api_service_name
  iap_oauth_client_id     = var.iap_oauth_client_id
  iap_oauth_client_secret = var.iap_oauth_client_secret
  iap_access_groups       = var.iap_access_groups
  cloud_armor_allowed_ips = var.cloud_armor_allowed_ips

  depends_on = [module.cloudrun]
}

# KMS: service agents and workload SAs
resource "google_kms_crypto_key_iam_member" "gcs_cmek" {
  crypto_key_id = module.compliance.kms_key_id
  role          = "roles/cloudkms.cryptoKeyEncrypterDecrypter"
  member        = "serviceAccount:service-${data.google_project.current.number}@gs-project-accounts.iam.gserviceaccount.com"
}

resource "google_kms_crypto_key_iam_member" "bq_cmek" {
  crypto_key_id = module.compliance.kms_key_id
  role          = "roles/cloudkms.cryptoKeyEncrypterDecrypter"
  member        = "serviceAccount:bq-${data.google_project.current.number}@gcp-sa-bigquery-iam.iam.gserviceaccount.com"
}

resource "google_kms_crypto_key_iam_member" "sql_cmek" {
  crypto_key_id = module.compliance.kms_key_id
  role          = "roles/cloudkms.cryptoKeyEncrypterDecrypter"
  member        = "serviceAccount:service-${data.google_project.current.number}@gcp-sa-cloud-sql.iam.gserviceaccount.com"
}

resource "google_kms_crypto_key_iam_member" "api_cmek" {
  crypto_key_id = module.compliance.kms_key_id
  role          = "roles/cloudkms.cryptoKeyEncrypterDecrypter"
  member        = "serviceAccount:${module.iam.service_account_emails["api"]}"
}

# Storage IAM
resource "google_storage_bucket_iam_member" "api_staging_create" {
  bucket = module.storage.staging_bucket_name
  role   = "roles/storage.objectCreator"
  member = "serviceAccount:${module.iam.service_account_emails["api"]}"
}

resource "google_storage_bucket_iam_member" "api_staging_read" {
  bucket = module.storage.staging_bucket_name
  role   = "roles/storage.objectViewer"
  member = "serviceAccount:${module.iam.service_account_emails["api"]}"
}

resource "google_storage_bucket_iam_member" "api_archive_read" {
  bucket = module.storage.archive_bucket_name
  role   = "roles/storage.objectViewer"
  member = "serviceAccount:${module.iam.service_account_emails["api"]}"
}

resource "google_storage_bucket_iam_member" "ingestion_staging_read" {
  bucket = module.storage.staging_bucket_name
  role   = "roles/storage.objectViewer"
  member = "serviceAccount:${module.iam.service_account_emails["ingestion"]}"
}

resource "google_storage_bucket_iam_member" "archive_worker" {
  bucket = module.storage.archive_bucket_name
  role   = "roles/storage.objectAdmin"
  member = "serviceAccount:${module.iam.service_account_emails["archive"]}"
}

resource "google_storage_bucket_iam_member" "archive_staging_admin" {
  bucket = module.storage.staging_bucket_name
  role   = "roles/storage.objectAdmin"
  member = "serviceAccount:${module.iam.service_account_emails["archive"]}"
}

resource "google_storage_bucket_iam_member" "report_archive_read" {
  bucket = module.storage.archive_bucket_name
  role   = "roles/storage.objectViewer"
  member = "serviceAccount:${module.iam.service_account_emails["report"]}"
}

# BigQuery IAM
resource "google_bigquery_dataset_iam_member" "api_audit_writer" {
  dataset_id = module.bigquery.dataset_id
  project    = var.project_id
  role       = "roles/bigquery.dataEditor"
  member     = "serviceAccount:${module.iam.service_account_emails["api"]}"
}

check "prod_iap_required" {
  assert {
    condition     = var.enable_iap
    error_message = "IAP must be enabled in prod."
  }
}

check "prod_iap_groups" {
  assert {
    condition     = length(var.iap_access_groups) > 0
    error_message = "Configure at least one IAP access group in prod."
  }
}

check "prod_email_domains" {
  assert {
    condition     = length(var.allowed_email_domains) > 0
    error_message = "Configure allowed_email_domains in prod."
  }
}

check "vpc_sc_ingress_identities" {
  assert {
    condition = !var.enable_vpc_service_controls || length([
      module.iam.service_account_emails["api"],
      module.iam.service_account_emails["ingestion"],
      module.iam.service_account_emails["archive"],
      module.iam.service_account_emails["report"],
    ]) == 4
    error_message = "VPC Service Controls requires workload service accounts."
  }
}
