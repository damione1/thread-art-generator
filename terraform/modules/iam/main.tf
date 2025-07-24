# Get current project information
data "google_project" "current" {
  project_id = var.project_id
}

# Service Account for API service
resource "google_service_account" "api_sa" {
  account_id   = "api-sa-${var.environment}"
  display_name = "Thread Art API Service Account (${var.environment})"
  description  = "Service account for API service in ${var.environment} environment"
  project      = var.project_id
}

# Service Account for Client/Web service  
resource "google_service_account" "client_sa" {
  account_id   = "client-sa-${var.environment}"
  display_name = "Thread Art Client Service Account (${var.environment})"
  description  = "Service account for client service in ${var.environment} environment"
  project      = var.project_id
}

# Service Account for Worker service
resource "google_service_account" "worker_sa" {
  account_id   = "worker-sa-${var.environment}"
  display_name = "Thread Art Worker Service Account (${var.environment})"
  description  = "Service account for worker service in ${var.environment} environment"
  project      = var.project_id
}

# Service Account for CI/CD
resource "google_service_account" "cicd_sa" {
  account_id   = "cicd-sa-${var.environment}"
  display_name = "Thread Art CI/CD Service Account (${var.environment})"
  description  = "Service account for CI/CD pipeline in ${var.environment} environment"
  project      = var.project_id
}

# Service Account for Database Migrations
resource "google_service_account" "migrator_sa" {
  account_id   = "migrator-sa-${var.environment}"
  display_name = "Thread Art DB Migrator Service Account (${var.environment})"
  description  = "Service account for database migrations in ${var.environment} environment"
  project      = var.project_id
}

# Workload Identity Pool
resource "google_iam_workload_identity_pool" "github" {
  workload_identity_pool_id = "github-${var.environment}"
  display_name              = "GitHub Actions (${var.environment})"
  description               = "Workload Identity Pool for GitHub Actions in ${var.environment}"
  project                   = var.project_id
}

# Workload Identity Provider for GitHub
resource "google_iam_workload_identity_pool_provider" "github" {
  workload_identity_pool_id          = google_iam_workload_identity_pool.github.workload_identity_pool_id
  workload_identity_pool_provider_id = "github-provider"
  display_name                       = "GitHub Provider"
  project                            = var.project_id

  attribute_mapping = {
    "google.subject"       = "assertion.sub"
    "attribute.actor"      = "assertion.actor"
    "attribute.repository" = "assertion.repository"
    "attribute.ref"        = "assertion.ref"
  }

  oidc {
    issuer_uri = "https://token.actions.githubusercontent.com"
  }
}

# Allow GitHub Actions to impersonate CI/CD service account
resource "google_service_account_iam_member" "github_cicd_impersonation" {
  service_account_id = google_service_account.cicd_sa.name
  role               = "roles/iam.workloadIdentityUser"
  member             = "principalSet://iam.googleapis.com/${google_iam_workload_identity_pool.github.name}/attribute.repository/${var.github_repository}"
}

# IAM Roles for API Service Account
resource "google_project_iam_member" "api_cloudsql_client" {
  project = var.project_id
  role    = "roles/cloudsql.client"
  member  = "serviceAccount:${google_service_account.api_sa.email}"
}

resource "google_project_iam_member" "api_storage_admin" {
  project = var.project_id
  role    = "roles/storage.objectAdmin"
  member  = "serviceAccount:${google_service_account.api_sa.email}"
}

resource "google_project_iam_member" "api_pubsub_editor" {
  project = var.project_id
  role    = "roles/pubsub.editor"
  member  = "serviceAccount:${google_service_account.api_sa.email}"
}

# IAM Roles for Client Service Account
resource "google_project_iam_member" "client_run_invoker" {
  project = var.project_id
  role    = "roles/run.invoker"
  member  = "serviceAccount:${google_service_account.client_sa.email}"
}

# IAM Roles for Worker Service Account
resource "google_project_iam_member" "worker_storage_admin" {
  project = var.project_id
  role    = "roles/storage.objectAdmin"
  member  = "serviceAccount:${google_service_account.worker_sa.email}"
}

resource "google_project_iam_member" "worker_pubsub_subscriber" {
  project = var.project_id
  role    = "roles/pubsub.subscriber"
  member  = "serviceAccount:${google_service_account.worker_sa.email}"
}

resource "google_project_iam_member" "worker_cloudsql_client" {
  project = var.project_id
  role    = "roles/cloudsql.client"
  member  = "serviceAccount:${google_service_account.worker_sa.email}"
}

# IAM Roles for CI/CD Service Account
resource "google_project_iam_member" "cicd_run_admin" {
  project = var.project_id
  role    = "roles/run.admin"
  member  = "serviceAccount:${google_service_account.cicd_sa.email}"
}

resource "google_project_iam_member" "cicd_artifact_registry_writer" {
  project = var.project_id
  role    = "roles/artifactregistry.writer"
  member  = "serviceAccount:${google_service_account.cicd_sa.email}"
}

resource "google_project_iam_member" "cicd_cloudfunctions_admin" {
  project = var.project_id
  role    = "roles/cloudfunctions.admin"
  member  = "serviceAccount:${google_service_account.cicd_sa.email}"
}

resource "google_project_iam_member" "cicd_service_account_user" {
  project = var.project_id
  role    = "roles/iam.serviceAccountUser"
  member  = "serviceAccount:${google_service_account.cicd_sa.email}"
}

# Allow CI/CD to impersonate migrator service account
resource "google_service_account_iam_member" "cicd_migrator_impersonation" {
  service_account_id = google_service_account.migrator_sa.name
  role               = "roles/iam.serviceAccountTokenCreator"
  member             = "serviceAccount:${google_service_account.cicd_sa.email}"
}

# IAM Roles for Migrator Service Account
resource "google_project_iam_member" "migrator_cloudsql_admin" {
  project = var.project_id
  role    = "roles/cloudsql.admin"
  member  = "serviceAccount:${google_service_account.migrator_sa.email}"
}