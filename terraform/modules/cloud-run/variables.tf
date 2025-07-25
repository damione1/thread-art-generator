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


# VPC Configuration
variable "vpc_connector_name" {
  description = "Name of the VPC connector for private networking"
  type        = string
}

# Service Account Emails
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

# Container Image URLs
variable "api_image_url" {
  description = "Container image URL for the API service"
  type        = string
}

variable "client_image_url" {
  description = "Container image URL for the client service"
  type        = string
}

variable "worker_image_url" {
  description = "Container image URL for the worker service"
  type        = string
}

# Database Configuration
variable "database_host" {
  description = "Database host (Cloud SQL private IP)"
  type        = string
}

variable "database_name" {
  description = "Database name"
  type        = string
}

variable "database_connection_name" {
  description = "Cloud SQL connection name"
  type        = string
}

# Storage Configuration
variable "public_bucket_name" {
  description = "Name of the public storage bucket"
  type        = string
}

variable "private_bucket_name" {
  description = "Name of the private storage bucket"
  type        = string
}

# Queue Configuration
variable "composition_topic_name" {
  description = "Name of the composition processing Pub/Sub topic"
  type        = string
}

variable "composition_subscription_name" {
  description = "Name of the composition processing subscription"
  type        = string
}

# Redis Configuration (optional)
variable "redis_host" {
  description = "Redis instance host (empty string if Redis disabled)"
  type        = string
  default     = ""
}

variable "redis_port" {
  description = "Redis instance port"
  type        = number
  default     = 6379
}

# Firebase Configuration
variable "firebase_project_id" {
  description = "Firebase project ID"
  type        = string
}

# Database IAM Users
variable "api_database_user" {
  description = "Database username for API service (IAM authenticated)"
  type        = string
}

variable "worker_database_user" {
  description = "Database username for worker service (IAM authenticated)"
  type        = string
}

# Secret Manager Configuration (reduced to only essential secrets)
variable "secret_names" {
  description = "Map of secret names from Secret Manager"
  type = object({
    token_symmetric_key   = string
    internal_api_key      = string
    cookie_hash_key       = string
    cookie_block_key      = string
  })
}

# API Service Configuration
variable "api_min_instances" {
  description = "Minimum number of API instances"
  type        = number
  default     = 0
}

variable "api_max_instances" {
  description = "Maximum number of API instances"
  type        = number
  default     = 10
}

variable "api_cpu_limit" {
  description = "CPU limit for API service"
  type        = string
  default     = "1000m"
}

variable "api_memory_limit" {
  description = "Memory limit for API service"
  type        = string
  default     = "512Mi"
}

variable "api_cpu_idle" {
  description = "Whether to allocate CPU only during requests for API service"
  type        = bool
  default     = true
}

# Client Service Configuration
variable "client_min_instances" {
  description = "Minimum number of client instances"
  type        = number
  default     = 0
}

variable "client_max_instances" {
  description = "Maximum number of client instances"
  type        = number
  default     = 10
}

variable "client_cpu_limit" {
  description = "CPU limit for client service"
  type        = string
  default     = "1000m"
}

variable "client_memory_limit" {
  description = "Memory limit for client service"
  type        = string
  default     = "512Mi"
}

variable "client_cpu_idle" {
  description = "Whether to allocate CPU only during requests for client service"
  type        = bool
  default     = true
}

# Worker Service Configuration
variable "worker_min_instances" {
  description = "Minimum number of worker instances"
  type        = number
  default     = 0
}

variable "worker_max_instances" {
  description = "Maximum number of worker instances"
  type        = number
  default     = 5
}

variable "worker_cpu_limit" {
  description = "CPU limit for worker service"
  type        = string
  default     = "2000m"
}

variable "worker_memory_limit" {
  description = "Memory limit for worker service"
  type        = string
  default     = "1Gi"
}

variable "worker_cpu_idle" {
  description = "Whether to allocate CPU only during requests for worker service"
  type        = bool
  default     = false
}