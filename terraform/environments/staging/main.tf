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
  project               = var.project_id
  region                = var.region
  user_project_override = true
  billing_project       = var.project_id
}

provider "google-beta" {
  project               = var.project_id
  region                = var.region
  user_project_override = true
  billing_project       = var.project_id
}

# Data sources
data "google_project" "current" {
  project_id = var.project_id
}


# Billing Budget Module (includes auto-shutdown)
# Temporarily disabled due to budget API and Cloud Function issues
# Re-enable after resolving:
# 1. Billing budget "invalid argument" error 
# 2. Cloud Function container startup issues
/*
module "billing" {
  source = "../../modules/billing"

  project_id                   = var.project_id
  environment                 = var.environment
  monthly_budget_amount       = var.monthly_budget_amount
  notification_channels       = []
  billing_account_id          = var.billing_account_id
  cicd_service_account_email  = module.iam.cicd_service_account_email
}
*/


# IAM Module
module "iam" {
  source = "../../modules/iam"

  project_id        = var.project_id
  environment       = var.environment
  github_repository = var.github_repository
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
  api_service_account_email    = module.iam.api_service_account_email
  client_service_account_email = module.iam.client_service_account_email
  worker_service_account_email = module.iam.worker_service_account_email
  cicd_service_account_email   = module.iam.cicd_service_account_email
}

# Cloud SQL Module
module "database" {
  source = "../../modules/cloud-sql"

  project_id             = var.project_id
  environment            = var.environment
  region                 = var.region
  application_name       = var.application_name
  vpc_network_self_link  = module.networking.vpc_network_self_link
  private_vpc_connection = module.networking.private_vpc_connection

  # Free tier configuration for staging
  database_tier                 = "db-f1-micro"
  enable_point_in_time_recovery = false
  deletion_protection           = false

  # Database configuration with IAM authentication
  database_name                  = var.database_name
  api_service_account_email      = module.iam.api_service_account_email
  worker_service_account_email   = module.iam.worker_service_account_email
  migrator_service_account_email = module.iam.migrator_service_account_email
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

# Redis Module (Cloud Memorystore) - Conditional
module "redis" {
  count  = var.enable_redis ? 1 : 0
  source = "../../modules/redis"

  project_id  = var.project_id
  environment = var.environment
  region      = var.region

  # Free tier configuration
  memory_size_gb          = 1
  tier                    = "BASIC"
  auth_enabled            = true
  transit_encryption_mode = "DISABLED" # Disabled for cost savings in staging
  enable_persistence      = false      # Disabled for cost savings
  prevent_destroy         = false      # Allow easy teardown in staging

  # Network configuration
  vpc_network_id         = module.networking.vpc_network_id
  private_vpc_connection = module.networking.private_vpc_connection
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

# Firebase Module
module "firebase" {
  source = "../../modules/firebase"

  project_id       = var.project_id
  environment      = var.environment
  region           = var.region
  application_name = var.application_name

  # Firebase Authentication configuration
  authorized_domains = [
    "localhost",
    "${var.project_id}.web.app",
    "${var.project_id}.firebaseapp.com"
  ]

  # OAuth providers configuration
  enable_oauth_providers   = var.enable_google_provider || var.enable_github_provider
  enable_google_provider   = var.enable_google_provider
  google_oauth_client_id   = var.google_oauth_client_id
  google_oauth_client_secret = var.google_oauth_client_secret
  enable_github_provider   = var.enable_github_provider
  github_oauth_client_id   = var.github_oauth_client_id
  github_oauth_client_secret = var.github_oauth_client_secret

  # Cloud Functions configuration (cost-optimized for staging)
  functions_source_dir    = "../../../functions"  # Use existing functions directory
  functions_max_instances = 5
  functions_memory        = "256Mi"
  functions_timeout       = 30

  # Backend integration (will be configured after deployment)
  backend_api_url              = "https://thread-art-api-staging-uc.a.run.app"
  internal_api_key_secret_name = module.secrets.secret_names.internal_api_key

  # Service account integration
  api_service_account_email    = module.iam.api_service_account_email
  worker_service_account_email = module.iam.worker_service_account_email
  cicd_service_account_email   = module.iam.cicd_service_account_email

  # Storage security configuration
  storage_rules_file_path    = "../../../storage.production.rules"
  storage_rules_content_hash = filemd5("../../../storage.production.rules")

  depends_on = [module.secrets]
}

# Cloud Run Module
module "cloud_run" {
  source = "../../modules/cloud-run"

  project_id  = var.project_id
  environment = var.environment
  region      = var.region


  # VPC configuration
  vpc_connector_name = module.networking.vpc_connector_name

  # Service account emails
  api_service_account_email    = module.iam.api_service_account_email
  client_service_account_email = module.iam.client_service_account_email
  worker_service_account_email = module.iam.worker_service_account_email

  # Container images (using hello world until real images are built)
  api_image_url    = "us-docker.pkg.dev/cloudrun/container/hello"
  client_image_url = "us-docker.pkg.dev/cloudrun/container/hello"
  worker_image_url = "us-docker.pkg.dev/cloudrun/container/hello"

  # Database configuration with IAM authentication
  database_host            = module.database.instance_private_ip
  database_name            = var.database_name
  database_connection_name = module.database.instance_connection_name
  api_database_user        = module.database.api_iam_user
  worker_database_user     = module.database.worker_iam_user

  # Firebase Storage configuration
  public_bucket_name  = module.firebase.public_bucket_name
  private_bucket_name = module.firebase.private_bucket_name

  # Queue configuration
  composition_topic_name        = module.pubsub.composition_processing_topic

  # Redis configuration (conditional)
  redis_host = var.enable_redis ? module.redis[0].host : ""
  redis_port = var.enable_redis ? module.redis[0].port : 6379

  # Firebase configuration
  firebase_project_id = module.firebase.firebase_project_id

  # Secret names
  secret_names = module.secrets.secret_names

  # Staging-specific resource limits (cost-optimized)
  api_min_instances = 0
  api_max_instances = 3
  api_cpu_limit     = "1000m"
  api_memory_limit  = "512Mi"
  api_cpu_idle      = true

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
