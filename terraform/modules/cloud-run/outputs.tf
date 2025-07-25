output "api_service_name" {
  description = "Name of the API Cloud Run service"
  value       = google_cloud_run_v2_service.api.name
}

output "api_service_url" {
  description = "URL of the API Cloud Run service"
  value       = google_cloud_run_v2_service.api.uri
}

output "api_service_id" {
  description = "ID of the API Cloud Run service"
  value       = google_cloud_run_v2_service.api.id
}

output "client_service_name" {
  description = "Name of the client Cloud Run service"
  value       = google_cloud_run_v2_service.client.name
}

output "client_service_url" {
  description = "URL of the client Cloud Run service"
  value       = google_cloud_run_v2_service.client.uri
}

output "client_service_id" {
  description = "ID of the client Cloud Run service"
  value       = google_cloud_run_v2_service.client.id
}

output "worker_service_name" {
  description = "Name of the worker Cloud Run service"
  value       = google_cloud_run_v2_service.worker.name
}

output "worker_service_url" {
  description = "URL of the worker Cloud Run service"
  value       = google_cloud_run_v2_service.worker.uri
}

output "worker_service_id" {
  description = "ID of the worker Cloud Run service"
  value       = google_cloud_run_v2_service.worker.id
}

output "service_urls" {
  description = "Map of service names to URLs"
  value = {
    api    = google_cloud_run_v2_service.api.uri
    client = google_cloud_run_v2_service.client.uri
    worker = google_cloud_run_v2_service.worker.uri
  }
}

output "service_names" {
  description = "Map of service types to names"
  value = {
    api    = google_cloud_run_v2_service.api.name
    client = google_cloud_run_v2_service.client.name
    worker = google_cloud_run_v2_service.worker.name
  }
}

output "public_urls" {
  description = "Public-facing service URLs"
  value = {
    client = google_cloud_run_v2_service.client.uri
  }
}

output "internal_urls" {
  description = "Internal-only service URLs"
  value = {
    api    = google_cloud_run_v2_service.api.uri
    worker = google_cloud_run_v2_service.worker.uri
  }
}