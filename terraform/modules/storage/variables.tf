variable "project_id" {
  description = "GCP Project ID"
  type        = string
}

variable "environment" {
  description = "Environment name (staging, production)"
  type        = string
}

variable "bucket_location" {
  description = "Location for storage buckets"
  type        = string
  default     = "US"
}

variable "public_bucket_lifecycle_days" {
  description = "Number of days after which objects in public bucket are deleted"
  type        = number
  default     = 365
}

variable "private_bucket_lifecycle_days" {
  description = "Number of days after which objects in private bucket are deleted"
  type        = number
  default     = 180
}

variable "enable_versioning" {
  description = "Enable object versioning"
  type        = bool
  default     = true
}

variable "cors_origins" {
  description = "CORS origins for public bucket"
  type        = list(string)
  default     = ["*"]
}

variable "kms_key_name" {
  description = "KMS key name for encryption (optional)"
  type        = string
  default     = null
}

variable "api_service_account_email" {
  description = "Email of the API service account"
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

variable "enable_bucket_notifications" {
  description = "Enable bucket notifications to Pub/Sub"
  type        = bool
  default     = false
}

variable "notification_topic" {
  description = "Pub/Sub topic for bucket notifications"
  type        = string
  default     = ""
}