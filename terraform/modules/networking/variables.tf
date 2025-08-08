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

variable "private_subnet_cidr" {
  description = "CIDR range for private subnet"
  type        = string
  default     = "10.0.0.0/24"
}

variable "pods_cidr_range" {
  description = "CIDR range for pods (if using GKE in future)"
  type        = string
  default     = "10.1.0.0/16"
}

variable "services_cidr_range" {
  description = "CIDR range for services (if using GKE in future)"
  type        = string
  default     = "10.2.0.0/16"
}

variable "connector_cidr_range" {
  description = "CIDR range for VPC connector"
  type        = string
  default     = "10.3.0.0/28"
}