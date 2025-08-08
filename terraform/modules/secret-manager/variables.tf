variable "project_id" {
  description = "GCP Project ID"
  type        = string
}

variable "environment" {
  description = "Environment name (staging, production)"
  type        = string
}

variable "api_service_account_email" {
  description = "Email of the API service account"
  type        = string
}

variable "client_service_account_email" {
  description = "Email of the client service account"
  type        = string
}

variable "worker_service_account_email" {
  description = "Email of the worker service account"
  type        = string
}

variable "cicd_service_account_email" {
  description = "Email of the CI/CD service account"
  type        = string
}