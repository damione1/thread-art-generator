variable "project_id" {
  description = "GCP Project ID"
  type        = string
}

variable "environment" {
  description = "Environment name (staging, production)"
  type        = string
}

variable "region" {
  description = "GCP Region"
  type        = string
  default     = "us-central1"
}

variable "billing_account_id" {
  description = "Billing Account ID"
  type        = string
}

variable "monthly_budget_amount" {
  description = "Monthly budget amount in USD"
  type        = number
  default     = 50
}

variable "notification_channels" {
  description = "List of notification channels for budget alerts"
  type        = list(string)
  default     = []
}

variable "cicd_service_account_email" {
  description = "Email of the CI/CD service account"
  type        = string
}