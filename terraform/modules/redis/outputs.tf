output "instance_name" {
  description = "Name of the Redis instance"
  value       = google_redis_instance.cache.name
}

output "instance_id" {
  description = "ID of the Redis instance"
  value       = google_redis_instance.cache.id
}

output "host" {
  description = "Redis instance hostname or IP address"
  value       = google_redis_instance.cache.host
}

output "port" {
  description = "Redis instance port"
  value       = google_redis_instance.cache.port
}

output "current_location_id" {
  description = "Current zone where the Redis instance resides"
  value       = google_redis_instance.cache.current_location_id
}

output "create_time" {
  description = "Creation time of the Redis instance"
  value       = google_redis_instance.cache.create_time
}

output "auth_string" {
  description = "Redis AUTH string (sensitive)"
  value       = google_redis_instance.cache.auth_string
  sensitive   = true
}

output "server_ca_certs" {
  description = "Server CA certificates for TLS connections"
  value       = google_redis_instance.cache.server_ca_certs
  sensitive   = true
}

output "connection_details" {
  description = "Connection details for applications"
  value = {
    host        = google_redis_instance.cache.host
    port        = google_redis_instance.cache.port
    auth_string = google_redis_instance.cache.auth_string
  }
  sensitive = true
}

output "redis_url" {
  description = "Redis connection URL"
  value       = var.auth_enabled ? "redis://:${google_redis_instance.cache.auth_string}@${google_redis_instance.cache.host}:${google_redis_instance.cache.port}" : "redis://${google_redis_instance.cache.host}:${google_redis_instance.cache.port}"
  sensitive   = true
}