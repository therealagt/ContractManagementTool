output "load_balancer_ip" {
  value = google_compute_global_address.lb_ip.address
}

output "iap_audience" {
  value = var.iap_oauth_client_id
}

output "backend_service_name" {
  value = google_compute_backend_service.api.name
}
