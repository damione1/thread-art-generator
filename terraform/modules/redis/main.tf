# Random suffix for Redis instance name
resource "random_id" "redis_name_suffix" {
  byte_length = 4
}

# Redis instance for session storage and caching
resource "google_redis_instance" "cache" {
  name               = "thread-art-redis-${var.environment}-${random_id.redis_name_suffix.hex}"
  memory_size_gb     = var.memory_size_gb
  tier               = var.tier
  region             = var.region
  location_id        = var.location_id
  project            = var.project_id

  # Use the VPC network for private access
  authorized_network = var.vpc_network_id
  connect_mode       = "PRIVATE_SERVICE_ACCESS"

  # Redis version
  redis_version = var.redis_version

  # Display name for easier identification
  display_name = "Thread Art Redis Cache (${var.environment})"

  # Redis configuration
  redis_configs = var.redis_configs

  # Enable AUTH for security
  auth_enabled = var.auth_enabled

  # Transit encryption mode
  transit_encryption_mode = var.transit_encryption_mode

  # Alternative location for high availability (only for STANDARD_HA tier)
  alternative_location_id = var.tier == "STANDARD_HA" ? var.alternative_location_id : null

  # Reserved IP range for the instance
  reserved_ip_range = var.reserved_ip_range

  # Maintenance policy
  dynamic "maintenance_policy" {
    for_each = var.maintenance_policy != null ? [var.maintenance_policy] : []
    content {
      create_time = maintenance_policy.value.create_time
      update_time = maintenance_policy.value.update_time
      
      dynamic "weekly_maintenance_window" {
        for_each = maintenance_policy.value.weekly_maintenance_window != null ? [maintenance_policy.value.weekly_maintenance_window] : []
        content {
          day = weekly_maintenance_window.value.day
          
          start_time {
            hours   = weekly_maintenance_window.value.start_time.hours
            minutes = weekly_maintenance_window.value.start_time.minutes
            seconds = weekly_maintenance_window.value.start_time.seconds
            nanos   = weekly_maintenance_window.value.start_time.nanos
          }
        }
      }
    }
  }

  # Persistence configuration for data durability
  dynamic "persistence_config" {
    for_each = var.enable_persistence ? [1] : []
    content {
      persistence_mode    = "RDB"
      rdb_snapshot_period = var.rdb_snapshot_period
    }
  }

  # Labels for resource management
  labels = {
    environment = var.environment
    purpose     = "session-cache"
    cost-center = "thread-art"
  }

  # Lifecycle management for production environments
  # prevent_destroy cannot use variables, so set directly in staging/production configs

  depends_on = [var.private_vpc_connection]
}