variable "project_id" {
  description = "GCP Project ID"
  type        = string
}

variable "region" {
  description = "GCP Region"
  type        = string
  default     = "us-central1"
}

variable "environment" {
  description = "Environment name (staging, production)"
  type        = string
}

variable "github_repository" {
  description = "GitHub repository for Workload Identity"
  type        = string
  default     = "damione1/thread-art-generator"
}