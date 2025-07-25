# Generate random values for secrets that need to be auto-generated
resource "random_password" "token_symmetric_key" {
  length  = 32
  special = true
}

resource "random_password" "internal_api_key" {
  length  = 32
  special = false
}

resource "random_password" "cookie_hash_key" {
  length  = 32
  special = false
}

resource "random_password" "cookie_block_key" {
  length  = 32
  special = false
}

# Note: Database secrets removed - using IAM authentication instead

# Application secrets
resource "google_secret_manager_secret" "token_symmetric_key" {
  secret_id = "token-symmetric-key-${var.environment}"
  project   = var.project_id

  labels = {
    environment = var.environment
    component   = "application"
  }

  replication {
    auto {}
  }

  lifecycle {
    ignore_changes = [secret_id]
  }
}

resource "google_secret_manager_secret_version" "token_symmetric_key" {
  secret      = google_secret_manager_secret.token_symmetric_key.id
  secret_data = random_password.token_symmetric_key.result
}

resource "google_secret_manager_secret" "internal_api_key" {
  secret_id = "internal-api-key-${var.environment}"
  project   = var.project_id

  labels = {
    environment = var.environment
    component   = "application"
  }

  replication {
    auto {}
  }

  lifecycle {
    ignore_changes = [secret_id]
  }
}

resource "google_secret_manager_secret_version" "internal_api_key" {
  secret      = google_secret_manager_secret.internal_api_key.id
  secret_data = random_password.internal_api_key.result
}

resource "google_secret_manager_secret" "cookie_hash_key" {
  secret_id = "cookie-hash-key-${var.environment}"
  project   = var.project_id

  labels = {
    environment = var.environment
    component   = "application"
  }

  replication {
    auto {}
  }

  lifecycle {
    ignore_changes = [secret_id]
  }
}

resource "google_secret_manager_secret_version" "cookie_hash_key" {
  secret      = google_secret_manager_secret.cookie_hash_key.id
  secret_data = random_password.cookie_hash_key.result
}

resource "google_secret_manager_secret" "cookie_block_key" {
  secret_id = "cookie-block-key-${var.environment}"
  project   = var.project_id

  labels = {
    environment = var.environment
    component   = "application"
  }

  replication {
    auto {}
  }

  lifecycle {
    ignore_changes = [secret_id]
  }
}

resource "google_secret_manager_secret_version" "cookie_block_key" {
  secret      = google_secret_manager_secret.cookie_block_key.id
  secret_data = random_password.cookie_block_key.result
}

# Firebase configuration secrets (placeholders - need to be updated manually)
resource "google_secret_manager_secret" "firebase_web_config" {
  secret_id = "firebase-web-config-${var.environment}"
  project   = var.project_id

  labels = {
    environment = var.environment
    component   = "firebase"
  }

  replication {
    auto {}
  }
}

resource "google_secret_manager_secret_version" "firebase_web_config" {
  secret      = google_secret_manager_secret.firebase_web_config.id
  secret_data = jsonencode({
    apiKey            = "placeholder-api-key"
    authDomain        = "thread-art-${var.environment}.firebaseapp.com"
    projectId         = "thread-art-${var.environment}"
    storageBucket     = "thread-art-${var.environment}.appspot.com"
    messagingSenderId = "placeholder-sender-id"
    appId             = "placeholder-app-id"
  })
}

resource "google_secret_manager_secret" "firebase_service_account" {
  secret_id = "firebase-service-account-${var.environment}"
  project   = var.project_id

  labels = {
    environment = var.environment
    component   = "firebase"
  }

  replication {
    auto {}
  }
}

resource "google_secret_manager_secret_version" "firebase_service_account" {
  secret      = google_secret_manager_secret.firebase_service_account.id
  secret_data = jsonencode({
    type     = "service_account"
    project_id = "thread-art-${var.environment}"
    # This needs to be updated with actual Firebase service account key
    placeholder = "Update this with actual Firebase service account JSON"
  })
}

# External service secrets (placeholders - need to be updated manually)
resource "google_secret_manager_secret" "sendinblue_api_key" {
  secret_id = "sendinblue-api-key-${var.environment}"
  project   = var.project_id

  labels = {
    environment = var.environment
    component   = "external-service"
  }

  replication {
    auto {}
  }
}

resource "google_secret_manager_secret_version" "sendinblue_api_key" {
  secret      = google_secret_manager_secret.sendinblue_api_key.id
  secret_data = "placeholder-sendinblue-api-key"
}

# Storage secrets (for GCS service account keys if needed)
resource "google_secret_manager_secret" "storage_access_key" {
  secret_id = "storage-access-key-${var.environment}"
  project   = var.project_id

  labels = {
    environment = var.environment
    component   = "storage"
  }

  replication {
    auto {}
  }
}

resource "google_secret_manager_secret_version" "storage_access_key" {
  secret      = google_secret_manager_secret.storage_access_key.id
  secret_data = "placeholder-storage-access-key"
}

resource "google_secret_manager_secret" "storage_secret_key" {
  secret_id = "storage-secret-key-${var.environment}"
  project   = var.project_id

  labels = {
    environment = var.environment
    component   = "storage"
  }

  replication {
    auto {}
  }
}

resource "google_secret_manager_secret_version" "storage_secret_key" {
  secret      = google_secret_manager_secret.storage_secret_key.id
  secret_data = "placeholder-storage-secret-key"
}

# IAM bindings for service accounts to access secrets
resource "google_secret_manager_secret_iam_member" "api_secrets_access" {
  for_each = {
    "token_symmetric_key" = google_secret_manager_secret.token_symmetric_key.id,
    "internal_api_key" = google_secret_manager_secret.internal_api_key.id,
    "storage_access_key" = google_secret_manager_secret.storage_access_key.id,
    "storage_secret_key" = google_secret_manager_secret.storage_secret_key.id,
  }

  secret_id = each.value
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${var.api_service_account_email}"

  depends_on = [
    google_secret_manager_secret.token_symmetric_key,
    google_secret_manager_secret.internal_api_key,
    google_secret_manager_secret.storage_access_key,
    google_secret_manager_secret.storage_secret_key
  ]
}

resource "google_secret_manager_secret_iam_member" "client_secrets_access" {
  for_each = {
    "cookie_hash_key" = google_secret_manager_secret.cookie_hash_key.id,
    "cookie_block_key" = google_secret_manager_secret.cookie_block_key.id,
    "firebase_web_config" = google_secret_manager_secret.firebase_web_config.id,
    "internal_api_key" = google_secret_manager_secret.internal_api_key.id,
  }

  secret_id = each.value
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${var.client_service_account_email}"

  depends_on = [
    google_secret_manager_secret.cookie_hash_key,
    google_secret_manager_secret.cookie_block_key,
    google_secret_manager_secret.firebase_web_config,
    google_secret_manager_secret.internal_api_key
  ]
}

resource "google_secret_manager_secret_iam_member" "worker_secrets_access" {
  for_each = {
    "storage_access_key" = google_secret_manager_secret.storage_access_key.id,
    "storage_secret_key" = google_secret_manager_secret.storage_secret_key.id,
  }

  secret_id = each.value
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${var.worker_service_account_email}"

  depends_on = [
    google_secret_manager_secret.storage_access_key,
    google_secret_manager_secret.storage_secret_key
  ]
}

resource "google_secret_manager_secret_iam_member" "cicd_secrets_access" {
  for_each = {
    "firebase_service_account" = google_secret_manager_secret.firebase_service_account.id,
  }

  secret_id = each.value
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${var.cicd_service_account_email}"

  depends_on = [
    google_secret_manager_secret.firebase_service_account
  ]
}