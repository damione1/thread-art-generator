# Random suffix for bucket names to ensure uniqueness
resource "random_id" "bucket_suffix" {
  byte_length = 4
}

# Public bucket for publicly accessible images (with CDN caching)
resource "google_storage_bucket" "public_bucket" {
  name     = "thread-art-public-${var.environment}-${random_id.bucket_suffix.hex}"
  location = var.bucket_location
  project  = var.project_id

  # Enable uniform bucket-level access for better security
  uniform_bucket_level_access = true

  # Lifecycle management to control costs
  lifecycle_rule {
    condition {
      age = var.public_bucket_lifecycle_days
    }
    action {
      type = "Delete"
    }
  }

  # Lifecycle rule to transition to cheaper storage class
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

  # Versioning for better data protection
  versioning {
    enabled = var.enable_versioning
  }

  # Website configuration for static hosting if needed
  website {
    main_page_suffix = "index.html"
    not_found_page   = "404.html"
  }

  labels = {
    environment = var.environment
    purpose     = "public-images"
    cost-center = "thread-art"
  }
}

# Private bucket for private images and user data
resource "google_storage_bucket" "private_bucket" {
  name     = "thread-art-private-${var.environment}-${random_id.bucket_suffix.hex}"
  location = var.bucket_location
  project  = var.project_id

  # Enable uniform bucket-level access
  uniform_bucket_level_access = true

  # Lifecycle management
  lifecycle_rule {
    condition {
      age = var.private_bucket_lifecycle_days
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
}

# Make public bucket publicly readable
resource "google_storage_bucket_iam_member" "public_bucket_public_read" {
  bucket = google_storage_bucket.public_bucket.name
  role   = "roles/storage.objectViewer"
  member = "allUsers"
}

# IAM bindings for service accounts

# API service account - full access to both buckets
resource "google_storage_bucket_iam_member" "api_public_bucket_admin" {
  bucket = google_storage_bucket.public_bucket.name
  role   = "roles/storage.objectAdmin"
  member = "serviceAccount:${var.api_service_account_email}"
}

resource "google_storage_bucket_iam_member" "api_private_bucket_admin" {
  bucket = google_storage_bucket.private_bucket.name
  role   = "roles/storage.objectAdmin"
  member = "serviceAccount:${var.api_service_account_email}"
}

# Worker service account - full access to both buckets for processing
resource "google_storage_bucket_iam_member" "worker_public_bucket_admin" {
  bucket = google_storage_bucket.public_bucket.name
  role   = "roles/storage.objectAdmin"
  member = "serviceAccount:${var.worker_service_account_email}"
}

resource "google_storage_bucket_iam_member" "worker_private_bucket_admin" {
  bucket = google_storage_bucket.private_bucket.name
  role   = "roles/storage.objectAdmin"
  member = "serviceAccount:${var.worker_service_account_email}"
}

# CI/CD service account - admin access for deployment and cleanup
resource "google_storage_bucket_iam_member" "cicd_public_bucket_admin" {
  bucket = google_storage_bucket.public_bucket.name
  role   = "roles/storage.admin"
  member = "serviceAccount:${var.cicd_service_account_email}"
}

resource "google_storage_bucket_iam_member" "cicd_private_bucket_admin" {
  bucket = google_storage_bucket.private_bucket.name
  role   = "roles/storage.admin"
  member = "serviceAccount:${var.cicd_service_account_email}"
}

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