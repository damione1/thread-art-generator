# Required Variables
variable "project_id" {
  description = "GCP Project ID (will be configured as Firebase project)"
  type        = string
}

variable "region" {
  description = "GCP Region for resources"
  type        = string
  default     = "us-central1"
}

variable "environment" {
  description = "Environment name (staging, production)"
  type        = string
}

variable "application_name" {
  description = "Application name for resource naming"
  type        = string
  default     = "thread-art"
}

# Firebase Authentication Configuration
variable "authorized_domains" {
  description = "List of authorized domains for Firebase Auth"
  type        = list(string)
  default     = ["localhost"]
}

variable "enable_oauth_providers" {
  description = "Enable OAuth identity providers"
  type        = bool
  default     = true
}

variable "enable_google_provider" {
  description = "Enable Google OAuth provider"
  type        = bool
  default     = false
}

variable "google_oauth_client_id" {
  description = "Google OAuth client ID"
  type        = string
  default     = ""
  sensitive   = true
}

variable "google_oauth_client_secret" {
  description = "Google OAuth client secret"
  type        = string
  default     = ""
  sensitive   = true
}

variable "enable_github_provider" {
  description = "Enable GitHub OAuth provider"
  type        = bool
  default     = false
}

variable "github_oauth_client_id" {
  description = "GitHub OAuth client ID"
  type        = string
  default     = ""
  sensitive   = true
}

variable "github_oauth_client_secret" {
  description = "GitHub OAuth client secret"
  type        = string
  default     = ""
  sensitive   = true
}

# Cloud Functions Configuration
variable "functions_source_dir" {
  description = "Directory containing Firebase Functions source code"
  type        = string
  default     = "../../../functions"
}

variable "functions_max_instances" {
  description = "Maximum number of function instances"
  type        = number
  default     = 10
}

variable "functions_memory" {
  description = "Memory allocation for Cloud Functions"
  type        = string
  default     = "512Mi"
}

variable "functions_timeout" {
  description = "Timeout for Cloud Functions in seconds"
  type        = number
  default     = 60
}

variable "functions_source_retention_days" {
  description = "Number of days to retain function source code in storage"
  type        = number
  default     = 7
}

# Backend Integration
variable "backend_api_url" {
  description = "Backend API URL for Cloud Functions to call"
  type        = string
}

variable "internal_api_key_secret_name" {
  description = "Name of the Secret Manager secret containing the internal API key"
  type        = string
}

# Service Account Integration
variable "api_service_account_email" {
  description = "Email of the API service account for IAM bindings"
  type        = string
  default     = ""
}

variable "worker_service_account_email" {
  description = "Email of the worker service account for IAM bindings"
  type        = string
  default     = ""
}

variable "cicd_service_account_email" {
  description = "Email of the CI/CD service account for IAM bindings"
  type        = string
  default     = ""
}

# Firebase Storage Configuration
variable "storage_lifecycle_days" {
  description = "Number of days after which to delete objects from storage buckets"
  type        = number
  default     = 365
}

variable "enable_versioning" {
  description = "Enable object versioning for storage buckets"
  type        = bool
  default     = true
}

variable "cors_origins" {
  description = "List of allowed CORS origins for storage buckets"
  type        = list(string)
  default     = ["*"]
}

variable "kms_key_name" {
  description = "KMS key name for bucket encryption (optional)"
  type        = string
  default     = null
}

variable "enable_bucket_notifications" {
  description = "Enable bucket notifications for storage events"
  type        = bool
  default     = false
}

variable "notification_topic" {
  description = "Pub/Sub topic name for bucket notifications"
  type        = string
  default     = ""
}

# Firebase Storage Security Rules Configuration
variable "storage_rules_file_path" {
  description = "Path to the Firebase Storage security rules file for production"
  type        = string
  default     = "../../../storage.production.rules"
}

variable "storage_rules_content_hash" {
  description = "Hash of the storage rules content to trigger deployment on changes"
  type        = string
  default     = ""
}