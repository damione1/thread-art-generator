# Staging Environment Configuration
terraform {
  required_version = ">= 1.5"

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
    google-beta = {
      source  = "hashicorp/google-beta"
      version = "~> 5.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.5"
    }
  }

}

# Configure the Google Cloud Provider
provider "google" {
  project = var.project_id
  region  = var.region
}

provider "google-beta" {
  project = var.project_id
  region  = var.region
}

# Data sources
data "google_project" "current" {
  project_id = var.project_id
}

# Billing Budget Module (includes auto-shutdown)
module "billing" {
  source = "../../modules/billing"

  project_id                   = var.project_id
  environment                 = var.environment
  monthly_budget_amount       = var.monthly_budget_amount
  notification_channels       = []
  billing_account_id          = var.billing_account_id
  cicd_service_account_email  = module.iam.cicd_service_account_email
}

# IAM Module
module "iam" {
  source = "../../modules/iam"

  project_id            = var.project_id
  environment           = var.environment
  github_repository     = var.github_repository
}

# Networking Module
module "networking" {
  source = "../../modules/networking"

  project_id  = var.project_id
  environment = var.environment
  region      = var.region
}

# Secret Manager Module
module "secrets" {
  source = "../../modules/secret-manager"

  project_id  = var.project_id
  environment = var.environment

  # Service account emails for IAM bindings
  api_service_account_email      = module.iam.api_service_account_email
  client_service_account_email   = module.iam.client_service_account_email
  worker_service_account_email   = module.iam.worker_service_account_email
  cicd_service_account_email     = module.iam.cicd_service_account_email
}

# Cloud SQL Module
module "database" {
  source = "../../modules/cloud-sql"

  project_id               = var.project_id
  environment              = var.environment
  region                   = var.region
  vpc_network_self_link    = module.networking.vpc_network_self_link
  private_vpc_connection   = module.networking.private_vpc_connection

  # Free tier configuration for staging
  database_tier                   = "db-f1-micro"
  enable_point_in_time_recovery   = false
  deletion_protection             = false

  # Database configuration with IAM authentication
  database_name                   = var.database_name
  api_service_account_email       = module.iam.api_service_account_email
  worker_service_account_email    = module.iam.worker_service_account_email
  migrator_service_account_email  = module.iam.migrator_service_account_email
}

# Cloud Storage Module
module "storage" {
  source = "../../modules/storage"

  project_id  = var.project_id
  environment = var.environment

  # Service account emails for IAM
  api_service_account_email    = module.iam.api_service_account_email
  worker_service_account_email = module.iam.worker_service_account_email
  cicd_service_account_email   = module.iam.cicd_service_account_email
}

# Artifact Registry Module
module "artifact_registry" {
  source = "../../modules/artifact-registry"

  project_id  = var.project_id
  environment = var.environment
  region      = var.region

  # Service account emails
  api_service_account_email    = module.iam.api_service_account_email
  client_service_account_email = module.iam.client_service_account_email
  worker_service_account_email = module.iam.worker_service_account_email
  cicd_service_account_email   = module.iam.cicd_service_account_email
}

# Redis Module (Cloud Memorystore)
module "redis" {
  source = "../../modules/redis"

  project_id  = var.project_id
  environment = var.environment
  region      = var.region

  # Free tier configuration
  memory_size_gb           = 1
  tier                     = "BASIC"
  auth_enabled             = true
  transit_encryption_mode  = "DISABLED"  # Disabled for cost savings in staging
  enable_persistence       = false       # Disabled for cost savings
  prevent_destroy          = false       # Allow easy teardown in staging

  # Network configuration
  vpc_network_id           = module.networking.vpc_network_id
  private_vpc_connection   = module.networking.private_vpc_connection
}

# Pub/Sub Module (replacing RabbitMQ)
module "pubsub" {
  source = "../../modules/pubsub"

  project_id  = var.project_id
  environment = var.environment

  # Service account emails for IAM
  api_service_account_email    = module.iam.api_service_account_email
  worker_service_account_email = module.iam.worker_service_account_email
}

# Cloud Run Module
module "cloud_run" {
  source = "../../modules/cloud-run"

  project_id    = var.project_id
  environment   = var.environment
  region        = var.region


  # VPC configuration
  vpc_connector_name = module.networking.vpc_connector_name

  # Service account emails
  api_service_account_email    = module.iam.api_service_account_email
  client_service_account_email = module.iam.client_service_account_email
  worker_service_account_email = module.iam.worker_service_account_email

  # Container images (will be updated via CI/CD)
  api_image_url    = "${var.region}-docker.pkg.dev/${var.project_id}/${module.artifact_registry.repository_id}/thread-art-api:latest"
  client_image_url = "${var.region}-docker.pkg.dev/${var.project_id}/${module.artifact_registry.repository_id}/thread-art-client:latest"
  worker_image_url = "${var.region}-docker.pkg.dev/${var.project_id}/${module.artifact_registry.repository_id}/thread-art-worker:latest"

  # Database configuration with IAM authentication
  database_host            = module.database.instance_private_ip
  database_name            = var.database_name
  database_connection_name = module.database.instance_connection_name
  api_database_user        = module.database.api_iam_user
  worker_database_user     = module.database.worker_iam_user

  # Storage configuration
  public_bucket_name  = module.storage.public_bucket_name
  private_bucket_name = module.storage.private_bucket_name

  # Queue configuration
  composition_topic_name        = module.pubsub.composition_processing_topic
  composition_subscription_name = module.pubsub.composition_processing_subscription

  # Redis configuration
  redis_host = module.redis.host
  redis_port = module.redis.port

  # Firebase configuration
  firebase_project_id = var.firebase_project_id

  # Secret names
  secret_names = module.secrets.secret_names

  # Staging-specific resource limits (cost-optimized)
  api_min_instances    = 0
  api_max_instances    = 3
  api_cpu_limit        = "1000m"
  api_memory_limit     = "512Mi"
  api_cpu_idle         = true

  client_min_instances = 0
  client_max_instances = 3
  client_cpu_limit     = "1000m"
  client_memory_limit  = "512Mi"
  client_cpu_idle      = true

  worker_min_instances = 0
  worker_max_instances = 2
  worker_cpu_limit     = "1000m"
  worker_memory_limit  = "1Gi"
  worker_cpu_idle      = false
}
