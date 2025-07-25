variable "project_id" {
  description = "GCP Project ID"
  type        = string
}

variable "environment" {
  description = "Environment name (staging, production)"
  type        = string
}

variable "application_name" {
  description = "Application name used for resource naming"
  type        = string
  default     = "thread-art"
}

variable "region" {
  description = "GCP Region"
  type        = string
  default     = "us-central1"
}

variable "database_tier" {
  description = "Database tier"
  type        = string
  default     = "db-f1-micro"  # Free tier
}

variable "disk_size_gb" {
  description = "Initial disk size in GB"
  type        = number
  default     = 10
}

variable "max_disk_size_gb" {
  description = "Maximum disk size for autoresize in GB"
  type        = number
  default     = 20
}

variable "availability_type" {
  description = "Availability type (ZONAL or REGIONAL)"
  type        = string
  default     = "ZONAL"  # Single zone for staging to save costs
}

variable "backup_retention_days" {
  description = "Number of days to retain backups"
  type        = number
  default     = 7
}

variable "enable_point_in_time_recovery" {
  description = "Enable point-in-time recovery"
  type        = bool
  default     = false  # Disable for staging to save costs
}

variable "deletion_protection" {
  description = "Enable deletion protection"
  type        = bool
  default     = false  # False for staging for easier teardown
}

variable "database_name" {
  description = "Name of the main database"
  type        = string
  default     = "threadmachine"
}

# Service account emails for IAM authentication
variable "api_service_account_email" {
  description = "Email of the API service account for IAM authentication"
  type        = string
}

variable "worker_service_account_email" {
  description = "Email of the worker service account for IAM authentication"
  type        = string
}

variable "migrator_service_account_email" {
  description = "Email of the migrator service account for IAM authentication"
  type        = string
}

variable "vpc_network_self_link" {
  description = "Self link of the VPC network"
  type        = string
}

variable "private_vpc_connection" {
  description = "Private VPC connection dependency"
  type        = any
}