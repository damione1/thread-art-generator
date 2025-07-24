# Artifact Registry repository for Docker images
resource "google_artifact_registry_repository" "docker_repo" {
  location      = var.region
  repository_id = "thread-art-${var.environment}"
  description   = "Docker repository for Thread Art Generator ${var.environment}"
  format        = "DOCKER"
  project       = var.project_id

  # Cleanup policies to manage storage costs
  cleanup_policy_dry_run = false
  
  cleanup_policies {
    id     = "delete-untagged"
    action = "DELETE"
    condition {
      tag_state = "UNTAGGED"
    }
  }

  cleanup_policies {
    id     = "keep-recent-untagged"
    action = "KEEP"
    condition {
      tag_state  = "UNTAGGED"
      newer_than = "604800s"
    }
  }

  cleanup_policies {
    id     = "delete-old-images"
    action = "DELETE"
    condition {
      tag_state  = "TAGGED"
      older_than = "2592000s"
    }
  }

  cleanup_policies {
    id     = "keep-recent-releases"
    action = "KEEP"
    condition {
      tag_state             = "TAGGED"
      tag_prefixes          = ["v", "release"]
      package_name_prefixes = ["thread-art-api", "thread-art-client"]
    }
  }

  cleanup_policies {
    id     = "keep-minimum-versions"
    action = "KEEP"
    most_recent_versions {
      package_name_prefixes = ["thread-art-api", "thread-art-client", "thread-art-worker"]
      keep_count            = 5
    }
  }

  labels = {
    environment = var.environment
    purpose     = "container-images"
  }
}

# IAM bindings for service accounts

# CI/CD service account - writer access for pushing images
resource "google_artifact_registry_repository_iam_member" "cicd_writer" {
  location   = google_artifact_registry_repository.docker_repo.location
  repository = google_artifact_registry_repository.docker_repo.name
  role       = "roles/artifactregistry.writer"
  member     = "serviceAccount:${var.cicd_service_account_email}"
  project    = var.project_id
}

# Cloud Run services need reader access to pull images
resource "google_artifact_registry_repository_iam_member" "api_reader" {
  location   = google_artifact_registry_repository.docker_repo.location
  repository = google_artifact_registry_repository.docker_repo.name
  role       = "roles/artifactregistry.reader"
  member     = "serviceAccount:${var.api_service_account_email}"
  project    = var.project_id
}

resource "google_artifact_registry_repository_iam_member" "client_reader" {
  location   = google_artifact_registry_repository.docker_repo.location
  repository = google_artifact_registry_repository.docker_repo.name
  role       = "roles/artifactregistry.reader"
  member     = "serviceAccount:${var.client_service_account_email}"
  project    = var.project_id
}

resource "google_artifact_registry_repository_iam_member" "worker_reader" {
  location   = google_artifact_registry_repository.docker_repo.location
  repository = google_artifact_registry_repository.docker_repo.name
  role       = "roles/artifactregistry.reader"
  member     = "serviceAccount:${var.worker_service_account_email}"
  project    = var.project_id
}