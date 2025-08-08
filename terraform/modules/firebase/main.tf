# Firebase Project and Services Configuration
terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
    google-beta = {
      source  = "hashicorp/google-beta"
      version = "~> 5.0"
    }
    time = {
      source  = "hashicorp/time"
      version = "~> 0.9"
    }
  }
}

# Data sources
data "google_project" "current" {
  project_id = var.project_id
}

# Create Firebase project (this creates a new GCP project configured for Firebase)
resource "google_firebase_project" "default" {
  provider = google-beta
  project  = var.project_id
}

# Enable required APIs for Firebase
resource "google_project_service" "firebase_apis" {
  for_each = toset([
    "firebase.googleapis.com",
    "identitytoolkit.googleapis.com",
    "cloudfunctions.googleapis.com",
    "cloudbuild.googleapis.com",
    "artifactregistry.googleapis.com",
    "run.googleapis.com",
    "eventarc.googleapis.com",
    "pubsub.googleapis.com"
  ])

  project = var.project_id
  service = each.value

  disable_dependent_services = false
  disable_on_destroy         = false

  depends_on = [google_firebase_project.default]
}

# Wait for APIs to be fully enabled
resource "time_sleep" "wait_for_apis" {
  create_duration = "60s"
  
  depends_on = [google_project_service.firebase_apis]
}

# Firebase Authentication Configuration
resource "google_identity_platform_config" "auth_config" {
  provider = google-beta
  project  = var.project_id

  sign_in {
    allow_duplicate_emails = false
    
    email {
      enabled           = true
      password_required = true
    }

    phone_number {
      enabled = false
    }

    anonymous {
      enabled = false
    }
  }

  # Configure OAuth providers
  dynamic "sign_in" {
    for_each = var.enable_oauth_providers ? [1] : []
    content {
      email {
        enabled           = true
        password_required = true
      }
    }
  }

  # Authorized domains for Firebase Auth
  authorized_domains = var.authorized_domains

  depends_on = [
    google_project_service.firebase_apis,
    google_firebase_project.default,
    time_sleep.wait_for_apis
  ]
}

# Configure OAuth Identity Providers (if enabled)
resource "google_identity_platform_default_supported_idp_config" "google_provider" {
  count    = var.enable_google_provider ? 1 : 0
  provider = google-beta
  project  = var.project_id

  idp_id        = "google.com"
  client_id     = var.google_oauth_client_id
  client_secret = var.google_oauth_client_secret
  enabled       = true

  depends_on = [google_identity_platform_config.auth_config]
}

resource "google_identity_platform_default_supported_idp_config" "github_provider" {
  count    = var.enable_github_provider ? 1 : 0
  provider = google-beta
  project  = var.project_id

  idp_id        = "github.com"
  client_id     = var.github_oauth_client_id
  client_secret = var.github_oauth_client_secret
  enabled       = true

  depends_on = [google_identity_platform_config.auth_config]
}

# Firebase Web App Configuration
resource "google_firebase_web_app" "default" {
  provider     = google-beta
  project      = var.project_id
  display_name = "${var.application_name}-${var.environment}"

  depends_on = [google_firebase_project.default]
}

# Generate Firebase config for the web app
data "google_firebase_web_app_config" "default" {
  provider   = google-beta
  project    = var.project_id
  web_app_id = google_firebase_web_app.default.app_id

  depends_on = [google_firebase_web_app.default]
}

# Cloud Functions Storage Configuration
resource "google_storage_bucket" "functions_source" {
  name                        = "${var.project_id}-firebase-functions-source"
  location                    = var.region
  force_destroy               = var.environment != "production"
  uniform_bucket_level_access = true

  lifecycle_rule {
    condition {
      age = var.functions_source_retention_days
    }
    action {
      type = "Delete"
    }
  }

  depends_on = [google_project_service.firebase_apis]
}

# IAM binding for Cloud Functions to access the source bucket
resource "google_storage_bucket_iam_member" "functions_source_reader" {
  bucket = google_storage_bucket.functions_source.name
  role   = "roles/storage.objectViewer"
  member = "serviceAccount:${var.project_id}@appspot.gserviceaccount.com"

  depends_on = [google_storage_bucket.functions_source]
}

# Create archive from source directory
data "archive_file" "functions_source" {
  type        = "zip"
  source_dir  = var.functions_source_dir
  output_path = "${path.module}/functions-source.zip"
}

# Upload the functions source to Cloud Storage
resource "google_storage_bucket_object" "functions_source" {
  name   = "functions-source-${data.archive_file.functions_source.output_md5}.zip"
  bucket = google_storage_bucket.functions_source.name
  source = data.archive_file.functions_source.output_path

  depends_on = [google_storage_bucket.functions_source]
}

# Deploy Cloud Function for user sync
resource "google_cloudfunctions2_function" "sync_user_on_create" {
  name        = "${var.application_name}-sync-user-${var.environment}"
  location    = var.region
  project     = var.project_id
  description = "Sync Firebase user to backend database on user creation"

  build_config {
    runtime     = "nodejs18"
    entry_point = "syncUserOnCreate"
    
    source {
      storage_source {
        bucket = google_storage_bucket.functions_source.name
        object = google_storage_bucket_object.functions_source.name
      }
    }
  }

  service_config {
    max_instance_count = var.functions_max_instances
    min_instance_count = 0
    available_memory   = var.functions_memory
    timeout_seconds    = var.functions_timeout
    
    environment_variables = {
      NODE_ENV         = var.environment
      BACKEND_URL      = var.backend_api_url
    }

    secret_environment_variables {
      key        = "INTERNAL_API_KEY"
      project_id = var.project_id
      secret     = "internal-api-key-${var.environment}"
      version    = "latest"
    }

    ingress_settings               = "ALLOW_INTERNAL_ONLY"
    all_traffic_on_latest_revision = true
  }

  # HTTP trigger - will be called manually by API when users are created

  depends_on = [
    google_project_service.firebase_apis,
    google_identity_platform_config.auth_config
  ]
}

# Deploy health check Cloud Function
resource "google_cloudfunctions2_function" "health_check" {
  name        = "${var.application_name}-functions-health-${var.environment}"
  location    = var.region
  project     = var.project_id
  description = "Health check endpoint for Firebase Functions monitoring"

  build_config {
    runtime     = "nodejs18"
    entry_point = "healthCheck"
    
    source {
      storage_source {
        bucket = google_storage_bucket.functions_source.name
        object = google_storage_bucket_object.functions_source.name
      }
    }
  }

  service_config {
    max_instance_count = 2
    min_instance_count = 0
    available_memory   = "256Mi"
    timeout_seconds    = 60
    
    environment_variables = {
      NODE_ENV = var.environment
    }

    ingress_settings               = "ALLOW_ALL"
    all_traffic_on_latest_revision = true
  }

  depends_on = [
    google_project_service.firebase_apis
  ]
}

# IAM binding for health check function (allow public access)
resource "google_cloud_run_service_iam_member" "health_check_invoker" {
  location = google_cloudfunctions2_function.health_check.location
  project  = google_cloudfunctions2_function.health_check.project
  service  = google_cloudfunctions2_function.health_check.name
  role     = "roles/run.invoker"
  member   = "allUsers"

  depends_on = [google_cloudfunctions2_function.health_check]
}

# Grant Firebase service account necessary permissions
resource "google_project_iam_member" "firebase_service_account" {
  for_each = toset([
    "roles/firebase.admin",
    "roles/serviceusage.serviceUsageConsumer",
  ])

  project = var.project_id
  role    = each.value
  member  = "serviceAccount:${var.project_id}@appspot.gserviceaccount.com"

  depends_on = [google_firebase_project.default]
}

# Grant API service account access to call Cloud Functions
resource "google_project_iam_member" "api_functions_invoker" {
  count   = var.api_service_account_email != "" ? 1 : 0
  project = var.project_id
  role    = "roles/cloudfunctions.invoker"
  member  = "serviceAccount:${var.api_service_account_email}"

  depends_on = [google_cloudfunctions2_function.sync_user_on_create]
}

# Grant Cloud Functions service account access to the internal API key secret
resource "google_secret_manager_secret_iam_member" "functions_internal_api_key_access" {
  secret_id = var.internal_api_key_secret_name
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${var.project_id}@appspot.gserviceaccount.com"
}

# Grant default compute service account access to the internal API key secret
resource "google_secret_manager_secret_iam_member" "compute_internal_api_key_access" {
  secret_id = var.internal_api_key_secret_name
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${data.google_project.current.number}-compute@developer.gserviceaccount.com"
}

# ================================================
# FIREBASE STORAGE CONFIGURATION
# ================================================

# Default Firebase Storage bucket (automatically created with Firebase project)
# This bucket is used for Firebase Storage operations and has the format {project-id}.appspot.com
resource "google_storage_bucket" "firebase_default_bucket" {
  name     = "${var.project_id}.appspot.com"
  location = var.region
  project  = var.project_id

  # Force destroy for non-production environments
  force_destroy = var.environment != "production"

  # Enable uniform bucket-level access for better security
  uniform_bucket_level_access = true

  # Lifecycle management to control costs
  lifecycle_rule {
    condition {
      age = var.storage_lifecycle_days
    }
    action {
      type = "Delete"
    }
  }

  # Transition to cheaper storage class after 30 days
  lifecycle_rule {
    condition {
      age = 30
    }
    action {
      type          = "SetStorageClass"
      storage_class = "NEARLINE"
    }
  }

  # CORS configuration for web access
  cors {
    origin          = var.cors_origins
    method          = ["GET", "HEAD", "PUT", "POST", "DELETE", "OPTIONS"]
    response_header = ["*"]
    max_age_seconds = 3600
  }

  # Versioning for better data protection
  versioning {
    enabled = var.enable_versioning
  }

  labels = {
    environment = var.environment
    purpose     = "firebase-storage-default"
    cost-center = "thread-art"
  }

  depends_on = [google_firebase_project.default]
}

# Public bucket for CDN-cacheable content
resource "google_storage_bucket" "public_bucket" {
  name     = "${var.project_id}-public"
  location = var.region
  project  = var.project_id

  # Force destroy for non-production environments
  force_destroy = var.environment != "production"

  # Enable uniform bucket-level access for better security
  uniform_bucket_level_access = true

  # Lifecycle management
  lifecycle_rule {
    condition {
      age = var.storage_lifecycle_days
    }
    action {
      type = "Delete"
    }
  }

  # Transition to cheaper storage class
  lifecycle_rule {
    condition {
      age = 30
    }
    action {
      type          = "SetStorageClass"
      storage_class = "NEARLINE"
    }
  }

  # CORS configuration for web access
  cors {
    origin          = var.cors_origins
    method          = ["GET", "HEAD", "OPTIONS"]
    response_header = ["*"]
    max_age_seconds = 3600
  }

  # Versioning
  versioning {
    enabled = var.enable_versioning
  }

  labels = {
    environment = var.environment
    purpose     = "public-images"
    cost-center = "thread-art"
  }

  depends_on = [google_firebase_project.default]
}

# Private bucket for signed URLs only
resource "google_storage_bucket" "private_bucket" {
  name     = "${var.project_id}-private"
  location = var.region
  project  = var.project_id

  # Force destroy for non-production environments
  force_destroy = var.environment != "production"

  # Enable uniform bucket-level access
  uniform_bucket_level_access = true

  # Lifecycle management
  lifecycle_rule {
    condition {
      age = var.storage_lifecycle_days
    }
    action {
      type = "Delete"
    }
  }

  # Transition to cheaper storage class after 60 days
  lifecycle_rule {
    condition {
      age = 60
    }
    action {
      type          = "SetStorageClass"
      storage_class = "COLDLINE"
    }
  }

  # Enable versioning for data protection
  versioning {
    enabled = var.enable_versioning
  }

  # Encryption with customer-managed keys if provided
  dynamic "encryption" {
    for_each = var.kms_key_name != null ? [1] : []
    content {
      default_kms_key_name = var.kms_key_name
    }
  }

  labels = {
    environment = var.environment
    purpose     = "private-user-data"
    cost-center = "thread-art"
  }

  depends_on = [google_firebase_project.default]
}

# ================================================
# FIREBASE STORAGE IAM PERMISSIONS
# ================================================

# Make public bucket publicly readable only in non-production environments
# In production, we rely on signed URLs and Firebase Storage rules for access control
resource "google_storage_bucket_iam_member" "public_bucket_public_read" {
  count  = var.environment != "production" ? 1 : 0
  bucket = google_storage_bucket.public_bucket.name
  role   = "roles/storage.objectViewer"
  member = "allUsers"
}

# Firebase service account permissions for all buckets
# Note: In production, Firebase Storage security rules control access, not IAM
# The Firebase service account needs admin access to manage objects via signed URLs
resource "google_storage_bucket_iam_member" "firebase_default_bucket_admin" {
  bucket = google_storage_bucket.firebase_default_bucket.name
  role   = "roles/storage.admin"
  member = "serviceAccount:${var.project_id}@appspot.gserviceaccount.com"

  depends_on = [google_firebase_project.default]
}

resource "google_storage_bucket_iam_member" "firebase_public_bucket_admin" {
  bucket = google_storage_bucket.public_bucket.name
  role   = "roles/storage.objectAdmin"
  member = "serviceAccount:${var.project_id}@appspot.gserviceaccount.com"

  depends_on = [google_firebase_project.default]
}

resource "google_storage_bucket_iam_member" "firebase_private_bucket_admin" {
  bucket = google_storage_bucket.private_bucket.name
  role   = "roles/storage.objectAdmin"
  member = "serviceAccount:${var.project_id}@appspot.gserviceaccount.com"

  depends_on = [google_firebase_project.default]
}

# API service account - full access to storage buckets
resource "google_storage_bucket_iam_member" "api_default_bucket_admin" {
  count  = var.api_service_account_email != "" ? 1 : 0
  bucket = google_storage_bucket.firebase_default_bucket.name
  role   = "roles/storage.objectAdmin"
  member = "serviceAccount:${var.api_service_account_email}"
}

resource "google_storage_bucket_iam_member" "api_public_bucket_admin" {
  count  = var.api_service_account_email != "" ? 1 : 0
  bucket = google_storage_bucket.public_bucket.name
  role   = "roles/storage.objectAdmin"
  member = "serviceAccount:${var.api_service_account_email}"
}

resource "google_storage_bucket_iam_member" "api_private_bucket_admin" {
  count  = var.api_service_account_email != "" ? 1 : 0
  bucket = google_storage_bucket.private_bucket.name
  role   = "roles/storage.objectAdmin"
  member = "serviceAccount:${var.api_service_account_email}"
}

# Worker service account - full access to storage buckets for processing
resource "google_storage_bucket_iam_member" "worker_default_bucket_admin" {
  count  = var.worker_service_account_email != "" ? 1 : 0
  bucket = google_storage_bucket.firebase_default_bucket.name
  role   = "roles/storage.objectAdmin"
  member = "serviceAccount:${var.worker_service_account_email}"
}

resource "google_storage_bucket_iam_member" "worker_public_bucket_admin" {
  count  = var.worker_service_account_email != "" ? 1 : 0
  bucket = google_storage_bucket.public_bucket.name
  role   = "roles/storage.objectAdmin"
  member = "serviceAccount:${var.worker_service_account_email}"
}

resource "google_storage_bucket_iam_member" "worker_private_bucket_admin" {
  count  = var.worker_service_account_email != "" ? 1 : 0
  bucket = google_storage_bucket.private_bucket.name
  role   = "roles/storage.objectAdmin"
  member = "serviceAccount:${var.worker_service_account_email}"
}

# CI/CD service account - admin access for deployment and cleanup
resource "google_storage_bucket_iam_member" "cicd_default_bucket_admin" {
  count  = var.cicd_service_account_email != "" ? 1 : 0
  bucket = google_storage_bucket.firebase_default_bucket.name
  role   = "roles/storage.admin"
  member = "serviceAccount:${var.cicd_service_account_email}"
}

resource "google_storage_bucket_iam_member" "cicd_public_bucket_admin" {
  count  = var.cicd_service_account_email != "" ? 1 : 0
  bucket = google_storage_bucket.public_bucket.name
  role   = "roles/storage.admin"
  member = "serviceAccount:${var.cicd_service_account_email}"
}

resource "google_storage_bucket_iam_member" "cicd_private_bucket_admin" {
  count  = var.cicd_service_account_email != "" ? 1 : 0
  bucket = google_storage_bucket.private_bucket.name
  role   = "roles/storage.admin"
  member = "serviceAccount:${var.cicd_service_account_email}"
}

# ================================================
# FIREBASE STORAGE BUCKET NOTIFICATIONS
# ================================================

# Notification configuration for bucket events (optional)
resource "google_storage_notification" "public_bucket_notification" {
  count  = var.enable_bucket_notifications ? 1 : 0
  bucket = google_storage_bucket.public_bucket.name
  
  topic         = var.notification_topic
  payload_format = "JSON_API_V1"
  
  event_types = [
    "OBJECT_FINALIZE",
    "OBJECT_DELETE"
  ]

  object_name_prefix = "images/"
}

resource "google_storage_notification" "private_bucket_notification" {
  count  = var.enable_bucket_notifications ? 1 : 0
  bucket = google_storage_bucket.private_bucket.name
  
  topic         = var.notification_topic
  payload_format = "JSON_API_V1"
  
  event_types = [
    "OBJECT_FINALIZE",
    "OBJECT_DELETE"
  ]
}

# ================================================
# FIREBASE STORAGE SECURITY RULES DEPLOYMENT
# ================================================

# Deploy Firebase Storage security rules for production
resource "google_storage_bucket_object" "storage_rules" {
  count  = var.environment == "production" ? 1 : 0
  name   = "storage.rules"
  bucket = google_storage_bucket.functions_source.name
  source = var.storage_rules_file_path

  depends_on = [google_storage_bucket.functions_source]
}

# Firebase Storage security rules configuration
# Note: Firebase Storage rules are deployed through Firebase CLI or console in production
# This is a placeholder for future automated deployment integration
resource "null_resource" "deploy_storage_rules" {
  count = var.environment == "production" ? 1 : 0

  # Trigger deployment when rules file changes
  triggers = {
    rules_content = var.storage_rules_content_hash
  }

  # In a real CI/CD pipeline, this would deploy rules via Firebase CLI
  # For now, this serves as a reminder that rules need manual deployment
  provisioner "local-exec" {
    command = "echo 'REMINDER: Deploy Firebase Storage rules manually in production using: firebase deploy --only storage'"
  }

  depends_on = [
    google_firebase_project.default,
    google_storage_bucket.firebase_default_bucket
  ]
}