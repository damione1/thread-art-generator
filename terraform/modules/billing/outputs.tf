output "budget_name" {
  description = "Name of the billing budget"
  value       = google_billing_budget.monthly_budget.display_name
}

output "budget_alerts_topic" {
  description = "Pub/Sub topic for budget alerts"
  value       = google_pubsub_topic.budget_alerts.name
}

output "budget_enforcer_function" {
  description = "Name of the budget enforcer Cloud Function"
  value       = google_cloudfunctions2_function.budget_enforcer.name
}