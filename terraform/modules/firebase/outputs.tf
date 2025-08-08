# Firebase Project Outputs
output "firebase_project_id" {
  description = "Firebase project ID"
  value       = google_firebase_project.default.project
}

output "firebase_project_number" {
  description = "Firebase project number"
  value       = google_firebase_project.default.project_number
}

# Firebase Web App Outputs
output "firebase_web_app_id" {
  description = "Firebase Web App ID"
  value       = google_firebase_web_app.default.app_id
}

output "firebase_config" {
  description = "Firebase configuration for web client"
  value = {
    apiKey            = data.google_firebase_web_app_config.default.api_key
    authDomain        = data.google_firebase_web_app_config.default.auth_domain
    projectId         = var.project_id
    storageBucket     = data.google_firebase_web_app_config.default.storage_bucket
    messagingSenderId = data.google_firebase_web_app_config.default.messaging_sender_id
    appId             = google_firebase_web_app.default.app_id
  }
  sensitive = true
}

# Cloud Functions Outputs
output "sync_user_function_name" {
  description = "Name of the user sync Cloud Function"
  value       = google_cloudfunctions2_function.sync_user_on_create.name
}

output "sync_user_function_url" {
  description = "URL of the user sync Cloud Function"
  value       = google_cloudfunctions2_function.sync_user_on_create.service_config[0].uri
}

output "health_check_function_name" {
  description = "Name of the health check Cloud Function"
  value       = google_cloudfunctions2_function.health_check.name
}

output "health_check_function_url" {
  description = "URL of the health check Cloud Function"
  value       = google_cloudfunctions2_function.health_check.service_config[0].uri
}

# Storage Outputs
output "functions_source_bucket" {
  description = "Cloud Storage bucket for Functions source code"
  value       = google_storage_bucket.functions_source.name
}

# Firebase Storage Outputs
output "firebase_default_bucket_name" {
  description = "Firebase default storage bucket name"
  value       = google_storage_bucket.firebase_default_bucket.name
}

output "public_bucket_name" {
  description = "Public storage bucket name"
  value       = google_storage_bucket.public_bucket.name
}

output "private_bucket_name" {
  description = "Private storage bucket name"
  value       = google_storage_bucket.private_bucket.name
}

# Authentication Configuration
output "firebase_auth_domain" {
  description = "Firebase Auth domain"
  value       = data.google_firebase_web_app_config.default.auth_domain
}

output "identity_platform_config_name" {
  description = "Identity Platform configuration name"
  value       = google_identity_platform_config.auth_config.name
}