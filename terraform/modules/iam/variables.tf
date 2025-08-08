variable "project_id" {
  description = "GCP Project ID"
  type        = string
}

variable "environment" {
  description = "Environment name (staging, production)"
  type        = string
}

variable "github_repository" {
  description = "GitHub repository for Workload Identity"
  type        = string
}