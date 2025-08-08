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

variable "application_name" {
  description = "Application name used for resource naming (e.g., thread-art)"
  type        = string
  default     = "thread-art"
}

variable "github_repository" {
  description = "GitHub repository for Workload Identity (format: owner/repo-name)"
  type        = string
}