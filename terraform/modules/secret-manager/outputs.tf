output "secret_ids" {
  description = "Map of secret names to their IDs"
  value = {
    token_symmetric_key        = google_secret_manager_secret.token_symmetric_key.secret_id
    internal_api_key          = google_secret_manager_secret.internal_api_key.secret_id
    cookie_hash_key           = google_secret_manager_secret.cookie_hash_key.secret_id
    cookie_block_key          = google_secret_manager_secret.cookie_block_key.secret_id
    firebase_web_config       = google_secret_manager_secret.firebase_web_config.secret_id
    firebase_service_account  = google_secret_manager_secret.firebase_service_account.secret_id
    sendinblue_api_key       = google_secret_manager_secret.sendinblue_api_key.secret_id
    storage_access_key       = google_secret_manager_secret.storage_access_key.secret_id
    storage_secret_key       = google_secret_manager_secret.storage_secret_key.secret_id
  }
}

output "secret_names" {
  description = "Map of secret names to their full resource names"
  value = {
    token_symmetric_key        = google_secret_manager_secret.token_symmetric_key.name
    internal_api_key          = google_secret_manager_secret.internal_api_key.name
    cookie_hash_key           = google_secret_manager_secret.cookie_hash_key.name
    cookie_block_key          = google_secret_manager_secret.cookie_block_key.name
    firebase_web_config       = google_secret_manager_secret.firebase_web_config.name
    firebase_service_account  = google_secret_manager_secret.firebase_service_account.name
    sendinblue_api_key       = google_secret_manager_secret.sendinblue_api_key.name
    storage_access_key       = google_secret_manager_secret.storage_access_key.name
    storage_secret_key       = google_secret_manager_secret.storage_secret_key.name
  }
}

# Note: PostgreSQL password output removed - using IAM authentication instead

output "generated_token_symmetric_key" {
  description = "Generated token symmetric key"
  value       = random_password.token_symmetric_key.result
  sensitive   = true
}

output "generated_internal_api_key" {
  description = "Generated internal API key"
  value       = random_password.internal_api_key.result
  sensitive   = true
}