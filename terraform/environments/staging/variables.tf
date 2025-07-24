variable "project_id" {
  description = "GCP Project ID for staging environment"
  type        = string
}

variable "region" {
  description = "GCP Region"
  type        = string
  default     = "us-central1"
}

variable "environment" {
  description = "Environment name"
  type        = string
  default     = "staging"
}

# Billing and Cost Management
variable "billing_account_id" {
  description = "GCP Billing Account ID for budget enforcement"
  type        = string
}

variable "monthly_budget_amount" {
  description = "Monthly budget amount in USD"
  type        = number
  default     = 50
}

variable "alert_emails" {
  description = "List of email addresses to receive billing alerts"
  type        = list(string)
}

# GitHub Configuration for CI/CD
variable "github_repository" {
  description = "GitHub repository name (e.g., thread-art-generator)"
  type        = string
}

variable "github_repository_owner" {
  description = "GitHub repository owner/organization"
  type        = string
}

# Database Configuration
variable "database_name" {
  description = "PostgreSQL database name"
  type        = string
  default     = "threadartdb"
}

# Firebase Configuration
variable "firebase_project_id" {
  description = "Firebase project ID"
  type        = string
}

# Feature Flags
variable "enable_monitoring" {
  description = "Enable monitoring and alerting (additional cost)"
  type        = bool
  default     = false
}

variable "enable_backup" {
  description = "Enable database backups (additional cost)"
  type        = bool
  default     = false
}

variable "enable_ssl_certificates" {
  description = "Enable managed SSL certificates"
  type        = bool
  default     = true
}

# Network Configuration
variable "authorized_networks" {
  description = "List of CIDR blocks authorized to access resources"
  type        = list(string)
  default     = ["0.0.0.0/0"]  # Open for staging - restrict in production
}

# Resource Sizing (for cost optimization)
variable "database_tier" {
  description = "Cloud SQL instance tier"
  type        = string
  default     = "db-f1-micro"  # Free tier
}

variable "redis_memory_gb" {
  description = "Redis instance memory in GB"
  type        = number
  default     = 1  # Minimum for staging
}

# Service Configuration
variable "api_max_instances" {
  description = "Maximum number of API instances"
  type        = number
  default     = 3
}

variable "client_max_instances" {
  description = "Maximum number of client instances"
  type        = number
  default     = 3
}

variable "worker_max_instances" {
  description = "Maximum number of worker instances"
  type        = number
  default     = 2
}

# Storage Configuration
variable "public_bucket_force_destroy" {
  description = "Allow force destruction of public bucket (staging only)"
  type        = bool
  default     = true
}

variable "private_bucket_force_destroy" {
  description = "Allow force destruction of private bucket (staging only)"
  type        = bool
  default     = true
}