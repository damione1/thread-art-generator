output "api_service_account_email" {
  description = "Email of the API service account"
  value       = google_service_account.api_sa.email
}

output "client_service_account_email" {
  description = "Email of the client service account"
  value       = google_service_account.client_sa.email
}

output "worker_service_account_email" {
  description = "Email of the worker service account"
  value       = google_service_account.worker_sa.email
}

output "cicd_service_account_email" {
  description = "Email of the CI/CD service account"
  value       = google_service_account.cicd_sa.email
}

output "migrator_service_account_email" {
  description = "Email of the migrator service account"
  value       = google_service_account.migrator_sa.email
}

output "workload_identity_provider" {
  description = "Workload Identity Provider for GitHub Actions"
  value       = google_iam_workload_identity_pool_provider.github.name
}