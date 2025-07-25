variable "project_id" {
  description = "GCP Project ID"
  type        = string
}

variable "environment" {
  description = "Environment name (staging, production)"
  type        = string
}

variable "region" {
  description = "GCP Region"
  type        = string
  default     = "us-central1"
}

variable "location_id" {
  description = "Zone where the instance will be provisioned"
  type        = string
  default     = "us-central1-a"
}

variable "alternative_location_id" {
  description = "Alternative zone for STANDARD_HA tier"
  type        = string
  default     = "us-central1-b"
}

variable "memory_size_gb" {
  description = "Redis memory size in GiB"
  type        = number
  default     = 1
}

variable "tier" {
  description = "Service tier of the instance (BASIC or STANDARD_HA)"
  type        = string
  default     = "BASIC"
  
  validation {
    condition     = contains(["BASIC", "STANDARD_HA"], var.tier)
    error_message = "Tier must be either BASIC or STANDARD_HA."
  }
}

variable "redis_version" {
  description = "Redis version"
  type        = string
  default     = "REDIS_7_0"
}

variable "auth_enabled" {
  description = "Enable Redis AUTH"
  type        = bool
  default     = true
}

variable "transit_encryption_mode" {
  description = "TLS mode of the Redis instance"
  type        = string
  default     = "DISABLED"
  
  validation {
    condition     = contains(["DISABLED", "SERVER_AUTHENTICATION"], var.transit_encryption_mode)
    error_message = "Transit encryption mode must be either DISABLED or SERVER_AUTHENTICATION."
  }
}

variable "vpc_network_id" {
  description = "ID of the VPC network"
  type        = string
}

variable "reserved_ip_range" {
  description = "CIDR range of internal addresses reserved for this instance"
  type        = string
  default     = null
}

variable "redis_configs" {
  description = "Redis configuration parameters"
  type        = map(string)
  default = {
    maxmemory-policy = "allkeys-lru"
    notify-keyspace-events = "Ex"
  }
}

variable "maintenance_policy" {
  description = "Maintenance policy configuration"
  type = object({
    create_time = optional(string)
    update_time = optional(string)
    weekly_maintenance_window = optional(object({
      day = string
      start_time = object({
        hours   = number
        minutes = number
        seconds = number
        nanos   = number
      })
    }))
  })
  default = {
    weekly_maintenance_window = {
      day = "SUNDAY"
      start_time = {
        hours   = 2
        minutes = 0
        seconds = 0
        nanos   = 0
      }
    }
  }
}

variable "enable_persistence" {
  description = "Enable RDB persistence"
  type        = bool
  default     = false  # Disabled for staging to save costs
}

variable "rdb_snapshot_period" {
  description = "RDB snapshot period"
  type        = string
  default     = "TWENTY_FOUR_HOURS"
  
  validation {
    condition     = contains(["ONE_HOUR", "SIX_HOURS", "TWELVE_HOURS", "TWENTY_FOUR_HOURS"], var.rdb_snapshot_period)
    error_message = "RDB snapshot period must be one of: ONE_HOUR, SIX_HOURS, TWELVE_HOURS, TWENTY_FOUR_HOURS."
  }
}

variable "prevent_destroy" {
  description = "Prevent Terraform from destroying the instance"
  type        = bool
  default     = false  # False for staging for easier teardown
}

variable "private_vpc_connection" {
  description = "Private VPC connection dependency"
  type        = any
}