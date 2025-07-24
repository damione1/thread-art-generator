output "project_id" {
  description = "GCP Project ID"
  value       = var.project_id
}

output "region" {
  description = "GCP Region"
  value       = var.region
}

output "environment" {
  description = "Environment name"
  value       = var.environment
}

# Service URLs
output "client_url" {
  description = "Public URL for the client application"
  value       = module.cloud_run.client_service_url
}

output "api_url" {
  description = "Internal URL for the API service"
  value       = module.cloud_run.api_service_url
  sensitive   = true
}

output "worker_url" {
  description = "Internal URL for the worker service"
  value       = module.cloud_run.worker_service_url
  sensitive   = true
}

# Database Information
output "database_connection_name" {
  description = "Cloud SQL connection name"
  value       = module.database.instance_connection_name
}

output "database_private_ip" {
  description = "Database private IP address"
  value       = module.database.instance_private_ip
  sensitive   = true
}

# Storage Information
output "public_bucket_name" {
  description = "Name of the public storage bucket"
  value       = module.storage.public_bucket_name
}

output "public_bucket_url" {
  description = "URL of the public storage bucket"
  value       = module.storage.public_bucket_url
}

output "private_bucket_name" {
  description = "Name of the private storage bucket"
  value       = module.storage.private_bucket_name
}

# Redis Information (conditional)
output "redis_host" {
  description = "Redis instance host"
  value       = var.enable_redis ? module.redis[0].host : "disabled"
  sensitive   = true
}

output "redis_port" {
  description = "Redis instance port"
  value       = var.enable_redis ? module.redis[0].port : 6379
}

# Artifact Registry
output "docker_repository" {
  description = "Docker repository URL"
  value       = module.artifact_registry.repository_url
}

# Pub/Sub Topics
output "composition_topic" {
  description = "Composition processing topic name"
  value       = module.pubsub.composition_processing_topic
}

output "image_processing_topic" {
  description = "Image processing topic name"
  value       = module.pubsub.image_processing_topic
}

# Network Information
output "vpc_network_name" {
  description = "VPC network name"
  value       = module.networking.vpc_network_name
}

output "vpc_connector_name" {
  description = "VPC connector name for Cloud Run"
  value       = module.networking.vpc_connector_name
}

# Service Accounts
output "service_accounts" {
  description = "Service account emails"
  value = {
    api      = module.iam.api_service_account_email
    client   = module.iam.client_service_account_email
    worker   = module.iam.worker_service_account_email
    cicd     = module.iam.cicd_service_account_email
    migrator = module.iam.migrator_service_account_email
  }
  sensitive = true
}

# Workload Identity for GitHub Actions
output "workload_identity_provider" {
  description = "Workload Identity Provider for GitHub Actions"
  value       = module.iam.workload_identity_provider
}

output "cicd_service_account_email" {
  description = "CI/CD service account email for GitHub Actions"
  value       = module.iam.cicd_service_account_email
}

# Secret Manager
output "secret_names" {
  description = "Map of secret names in Secret Manager"
  value       = module.secrets.secret_names
  sensitive   = true
}

# Budget Information
# Temporarily disabled due to billing module issues
/*
output "budget_name" {
  description = "Billing budget name"
  value       = module.billing.budget_name
}
*/

output "budget_amount" {
  description = "Monthly budget amount"
  value       = var.monthly_budget_amount
}

# Infrastructure Status
output "infrastructure_summary" {
  description = "Summary of deployed infrastructure"
  value = {
    environment          = var.environment
    project_id          = var.project_id
    region              = var.region
    client_url          = module.cloud_run.client_service_url
    database_tier       = var.database_tier
    redis_memory_gb     = var.redis_memory_gb
    monthly_budget      = var.monthly_budget_amount
    auto_shutdown       = true
    services_deployed   = {
      api    = module.cloud_run.api_service_name
      client = module.cloud_run.client_service_name
      worker = module.cloud_run.worker_service_name
    }
    storage_buckets = {
      public  = module.storage.public_bucket_name
      private = module.storage.private_bucket_name
    }
  }
}