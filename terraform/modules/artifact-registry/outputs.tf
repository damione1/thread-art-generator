output "repository_name" {
  description = "Name of the Artifact Registry repository"
  value       = google_artifact_registry_repository.docker_repo.name
}

output "repository_id" {
  description = "ID of the Artifact Registry repository"
  value       = google_artifact_registry_repository.docker_repo.repository_id
}

output "repository_url" {
  description = "URL of the Artifact Registry repository"
  value       = "${var.region}-docker.pkg.dev/${var.project_id}/${google_artifact_registry_repository.docker_repo.repository_id}"
}

output "full_repository_name" {
  description = "Full name of the Artifact Registry repository"
  value       = "projects/${var.project_id}/locations/${var.region}/repositories/${google_artifact_registry_repository.docker_repo.repository_id}"
}