data "google_project" "current" {
  project_id = var.project_id
}

resource "google_compute_global_address" "lb_ip" {
  name    = "contract-lb-ip-${var.environment}"
  project = var.project_id
}

resource "google_compute_managed_ssl_certificate" "main" {
  name    = "contract-cert-${var.environment}"
  project = var.project_id

  managed {
    domains = [var.domain]
  }
}

resource "google_compute_region_network_endpoint_group" "api" {
  name                  = "contract-api-neg-${var.environment}"
  project               = var.project_id
  region                = var.region
  network_endpoint_type = "SERVERLESS"

  cloud_run {
    service = var.api_service_name
  }
}

resource "google_compute_backend_service" "api" {
  name                  = "contract-api-backend-${var.environment}"
  project               = var.project_id
  protocol              = "HTTPS"
  load_balancing_scheme = "EXTERNAL_MANAGED"
  security_policy       = length(var.cloud_armor_allowed_ips) > 0 ? google_compute_security_policy.main[0].id : null

  backend {
    group = google_compute_region_network_endpoint_group.api.id
  }

  iap {
    oauth2_client_id     = var.iap_oauth_client_id
    oauth2_client_secret = var.iap_oauth_client_secret
  }

  log_config {
    enable      = true
    sample_rate = 1.0
  }
}

resource "google_compute_security_policy" "main" {
  count = length(var.cloud_armor_allowed_ips) > 0 ? 1 : 0

  name    = "contract-armor-${var.environment}"
  project = var.project_id

  rule {
    action   = "allow"
    priority = 1000
    match {
      versioned_expr = "SRC_IPS_V1"
      config {
        src_ip_ranges = var.cloud_armor_allowed_ips
      }
    }
  }

  rule {
    action   = "deny(403)"
    priority = 2147483647
    match {
      versioned_expr = "SRC_IPS_V1"
      config {
        src_ip_ranges = ["*"]
      }
    }
  }
}

resource "google_compute_url_map" "main" {
  name            = "contract-url-map-${var.environment}"
  project         = var.project_id
  default_service = google_compute_backend_service.api.id
}

resource "google_compute_target_https_proxy" "main" {
  name    = "contract-https-proxy-${var.environment}"
  project = var.project_id
  url_map = google_compute_url_map.main.id

  ssl_certificates = [google_compute_managed_ssl_certificate.main.id]
}

resource "google_compute_global_forwarding_rule" "main" {
  name                  = "contract-https-fw-${var.environment}"
  project               = var.project_id
  ip_address            = google_compute_global_address.lb_ip.id
  ip_protocol           = "TCP"
  port_range            = "443"
  load_balancing_scheme = "EXTERNAL_MANAGED"
  target                = google_compute_target_https_proxy.main.id
}

resource "google_iap_web_backend_service_iam_member" "access" {
  for_each = toset(var.iap_access_groups)

  project             = var.project_id
  web_backend_service = google_compute_backend_service.api.name
  role                = "roles/iap.httpsResourceAccessor"
  member              = "group:${each.value}"
}
