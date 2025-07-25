# Random suffix for database instance name
resource "random_id" "db_name_suffix" {
  byte_length = 4
}

# Cloud SQL PostgreSQL instance
resource "google_sql_database_instance" "main" {
  name             = "${var.application_name}-db-${var.environment}-${random_id.db_name_suffix.hex}"
  database_version = "POSTGRES_17"
  region           = var.region
  project          = var.project_id

  # Free tier settings for staging
  settings {
    tier              = var.database_tier
    disk_type         = "PD_SSD"
    disk_size         = var.disk_size_gb
    disk_autoresize   = true
    disk_autoresize_limit = var.max_disk_size_gb
    
    # High availability for production, single zone for staging to save costs
    availability_type = var.availability_type

    backup_configuration {
      enabled                        = true
      start_time                     = "02:00"  # 2 AM UTC
      point_in_time_recovery_enabled = var.enable_point_in_time_recovery
      location                       = var.region
      backup_retention_settings {
        retained_backups = var.backup_retention_days
      }
    }

    ip_configuration {
      ipv4_enabled                                  = false
      private_network                               = var.vpc_network_self_link
      enable_private_path_for_google_cloud_services = true
      ssl_mode                                      = "ENCRYPTED_ONLY"
    }

    database_flags {
      name  = "log_checkpoints"
      value = "on"
    }

    database_flags {
      name  = "log_connections"
      value = "on"
    }

    database_flags {
      name  = "log_disconnections"
      value = "on"
    }

    database_flags {
      name  = "log_lock_waits"
      value = "on"
    }

    database_flags {
      name  = "log_min_duration_statement"
      value = "1000"  # Log queries taking longer than 1 second
    }

    # Enable IAM authentication
    database_flags {
      name  = "cloudsql.iam_authentication"
      value = "on"
    }

    # Enable query insights
    insights_config {
      query_insights_enabled  = true
      record_application_tags = true
      record_client_address   = true
    }

    # Maintenance window
    maintenance_window {
      day  = 7  # Sunday
      hour = 3  # 3 AM UTC
    }

    # Deletion protection for production
    deletion_protection_enabled = var.deletion_protection
  }

  deletion_protection = var.deletion_protection

  depends_on = [var.private_vpc_connection]
}

# Create the main database
resource "google_sql_database" "main_database" {
  name     = var.database_name
  instance = google_sql_database_instance.main.name
  project  = var.project_id
}

# Create IAM database user for API service
resource "google_sql_user" "api_iam_user" {
  name     = trimsuffix(var.api_service_account_email, ".gserviceaccount.com")
  instance = google_sql_database_instance.main.name
  type     = "CLOUD_IAM_SERVICE_ACCOUNT"
  project  = var.project_id

  depends_on = [google_sql_database.main_database]
}

# Create IAM database user for worker service
resource "google_sql_user" "worker_iam_user" {
  name     = trimsuffix(var.worker_service_account_email, ".gserviceaccount.com")
  instance = google_sql_database_instance.main.name
  type     = "CLOUD_IAM_SERVICE_ACCOUNT"
  project  = var.project_id

  depends_on = [google_sql_database.main_database]
}

# Create IAM database user for migrator service (for migrations)
resource "google_sql_user" "migrator_iam_user" {
  name     = trimsuffix(var.migrator_service_account_email, ".gserviceaccount.com")
  instance = google_sql_database_instance.main.name
  type     = "CLOUD_IAM_SERVICE_ACCOUNT"
  project  = var.project_id

  depends_on = [google_sql_database.main_database]
}

# Grant necessary privileges to migration user
resource "google_sql_database" "migration_grants" {
  name     = "${var.database_name}_migration_grants"
  instance = google_sql_database_instance.main.name
  project  = var.project_id
}

# SSL Certificate for secure connections
resource "google_sql_ssl_cert" "client_cert" {
  common_name = "${var.application_name}-${var.environment}"
  instance    = google_sql_database_instance.main.name
  project     = var.project_id
}