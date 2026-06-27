output "service_account_emails" {
  value = { for k, sa in google_service_account.accounts : k => sa.email }
}

output "service_account_ids" {
  value = { for k, sa in google_service_account.accounts : k => sa.id }
}
