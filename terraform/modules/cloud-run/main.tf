# API Service
resource "google_cloud_run_v2_service" "api" {
  name     = "thread-art-api-${var.environment}"
  location = var.region
  project  = var.project_id
  
  # Internal-only ingress for security
  ingress = "INGRESS_TRAFFIC_INTERNAL_ONLY"

  template {
    # Use the API service account
    service_account = var.api_service_account_email

    # Scaling configuration
    scaling {
      min_instance_count = var.api_min_instances
      max_instance_count = var.api_max_instances
    }

    # VPC configuration for private networking
    vpc_access {
      connector = var.vpc_connector_name
      egress    = "PRIVATE_RANGES_ONLY"
    }

    # Timeout and resource limits
    timeout = "300s"  # 5 minutes
    
    # Cloud SQL volume mount
    volumes {
      name = "cloudsql"
      cloud_sql_instance {
        instances = [var.database_connection_name]
      }
    }
    
    containers {
      image = var.api_image_url
      
      # Resource allocation
      resources {
        limits = {
          cpu    = var.api_cpu_limit
          memory = var.api_memory_limit
        }
        cpu_idle = var.api_cpu_idle
        startup_cpu_boost = true
      }

      # Port configuration
      ports {
        container_port = 9090
        name          = "http1"
      }

      # Environment variables
      env {
        name  = "ENVIRONMENT"
        value = var.environment
      }

      env {
        name  = "GCP_PROJECT_ID"
        value = var.project_id
      }

      env {
        name  = "HTTP_SERVER_PORT"
        value = "9090"
      }

      env {
        name  = "POSTGRES_HOST"
        value = var.database_host
      }

      env {
        name  = "POSTGRES_DB"
        value = var.database_name
      }

      env {
        name  = "STORAGE_PROVIDER"
        value = "gcs"
      }

      env {
        name  = "STORAGE_PUBLIC_BUCKET"
        value = var.public_bucket_name
      }

      env {
        name  = "STORAGE_PRIVATE_BUCKET"
        value = var.private_bucket_name
      }

      env {
        name  = "QUEUE_COMPOSITION_PROCESSING"
        value = var.composition_topic_name
      }

      env {
        name  = "REDIS_ADDR"
        value = "${var.redis_host}:${var.redis_port}"
      }

      # Database IAM authentication (no password needed)
      env {
        name  = "POSTGRES_USER"
        value = var.api_database_user
      }

      env {
        name  = "POSTGRES_IAM_AUTH"
        value = "true"
      }

      env {
        name = "TOKEN_SYMMETRIC_KEY"
        value_source {
          secret_key_ref {
            secret  = var.secret_names.token_symmetric_key
            version = "latest"
          }
        }
      }

      env {
        name = "INTERNAL_API_KEY"
        value_source {
          secret_key_ref {
            secret  = var.secret_names.internal_api_key
            version = "latest"
          }
        }
      }

      # Cloud SQL volume mount
      volume_mounts {
        name       = "cloudsql"
        mount_path = "/cloudsql"
      }

      # Health check
      startup_probe {
        initial_delay_seconds = 10
        timeout_seconds       = 5
        period_seconds        = 10
        failure_threshold     = 3
        
        http_get {
          path = "/health"
          port = 9090
        }
      }

      liveness_probe {
        initial_delay_seconds = 30
        timeout_seconds       = 5
        period_seconds        = 30
        failure_threshold     = 3
        
        http_get {
          path = "/health"
          port = 9090
        }
      }
    }

    # Annotations for specific configurations
    annotations = {
      "autoscaling.knative.dev/minScale" = tostring(var.api_min_instances)
      "autoscaling.knative.dev/maxScale" = tostring(var.api_max_instances)
      "run.googleapis.com/cloudsql-instances" = var.database_connection_name
      "run.googleapis.com/client-name" = "terraform"
    }
  }

  # Traffic allocation
  traffic {
    type    = "TRAFFIC_TARGET_ALLOCATION_TYPE_LATEST"
    percent = 100
  }

  labels = {
    environment = var.environment
    service     = "api"
  }
}

# Client/Web Service
resource "google_cloud_run_v2_service" "client" {
  name     = "thread-art-client-${var.environment}"
  location = var.region
  project  = var.project_id
  
  # Public ingress for web traffic
  ingress = "INGRESS_TRAFFIC_ALL"

  template {
    # Use the client service account
    service_account = var.client_service_account_email

    # Scaling configuration
    scaling {
      min_instance_count = var.client_min_instances
      max_instance_count = var.client_max_instances
    }

    # VPC configuration
    vpc_access {
      connector = var.vpc_connector_name
      egress    = "PRIVATE_RANGES_ONLY"
    }

    timeout = "60s"  # 1 minute for web requests
    
    containers {
      image = var.client_image_url
      
      # Resource allocation
      resources {
        limits = {
          cpu    = var.client_cpu_limit
          memory = var.client_memory_limit
        }
        cpu_idle = var.client_cpu_idle
        startup_cpu_boost = true
      }

      # Port configuration
      ports {
        container_port = 8080
        name          = "http1"
      }

      # Environment variables
      env {
        name  = "ENVIRONMENT"
        value = var.environment
      }

      env {
        name  = "FRONTEND_PORT"
        value = "8080"
      }

      env {
        name  = "API_URL"
        value = google_cloud_run_v2_service.api.uri
      }

      env {
        name  = "FIREBASE_PROJECT_ID"
        value = var.firebase_project_id
      }

      env {
        name  = "REDIS_ADDR"
        value = "${var.redis_host}:${var.redis_port}"
      }

      # Secrets from Secret Manager
      env {
        name = "COOKIE_HASH_KEY"
        value_source {
          secret_key_ref {
            secret  = var.secret_names.cookie_hash_key
            version = "latest"
          }
        }
      }

      env {
        name = "COOKIE_BLOCK_KEY"
        value_source {
          secret_key_ref {
            secret  = var.secret_names.cookie_block_key
            version = "latest"
          }
        }
      }

      env {
        name = "INTERNAL_API_KEY"
        value_source {
          secret_key_ref {
            secret  = var.secret_names.internal_api_key
            version = "latest"
          }
        }
      }

      # Health check
      startup_probe {
        initial_delay_seconds = 10
        timeout_seconds       = 5
        period_seconds        = 10
        failure_threshold     = 3
        
        http_get {
          path = "/health"
          port = 8080
        }
      }

      liveness_probe {
        initial_delay_seconds = 30
        timeout_seconds       = 5
        period_seconds        = 30
        failure_threshold     = 3
        
        http_get {
          path = "/health"
          port = 8080
        }
      }
    }

    # Annotations
    annotations = {
      "autoscaling.knative.dev/minScale" = tostring(var.client_min_instances)
      "autoscaling.knative.dev/maxScale" = tostring(var.client_max_instances)
      "run.googleapis.com/client-name" = "terraform"
    }
  }

  # Traffic allocation
  traffic {
    type    = "TRAFFIC_TARGET_ALLOCATION_TYPE_LATEST"
    percent = 100
  }

  labels = {
    environment = var.environment
    service     = "client"
  }
}

# Worker Service (for processing background tasks)
resource "google_cloud_run_v2_service" "worker" {
  name     = "thread-art-worker-${var.environment}"
  location = var.region
  project  = var.project_id
  
  # Internal-only ingress
  ingress = "INGRESS_TRAFFIC_INTERNAL_ONLY"

  template {
    # Use the worker service account
    service_account = var.worker_service_account_email

    # Scaling configuration
    scaling {
      min_instance_count = var.worker_min_instances
      max_instance_count = var.worker_max_instances
    }

    # VPC configuration
    vpc_access {
      connector = var.vpc_connector_name
      egress    = "PRIVATE_RANGES_ONLY"
    }

    timeout = "3600s"  # 1 hour for long-running tasks
    
    # Cloud SQL volume mount
    volumes {
      name = "cloudsql"
      cloud_sql_instance {
        instances = [var.database_connection_name]
      }
    }
    
    containers {
      image = var.worker_image_url
      
      # Resource allocation for intensive processing
      resources {
        limits = {
          cpu    = var.worker_cpu_limit
          memory = var.worker_memory_limit
        }
        cpu_idle = var.worker_cpu_idle
        startup_cpu_boost = true
      }

      # Port configuration
      ports {
        container_port = 8081
        name          = "http1"
      }

      # Environment variables
      env {
        name  = "ENVIRONMENT"
        value = var.environment
      }

      env {
        name  = "GCP_PROJECT_ID"
        value = var.project_id
      }

      env {
        name  = "POSTGRES_HOST"
        value = var.database_host
      }

      env {
        name  = "POSTGRES_DB"
        value = var.database_name
      }

      env {
        name  = "STORAGE_PROVIDER"
        value = "gcs"
      }

      env {
        name  = "STORAGE_PUBLIC_BUCKET"
        value = var.public_bucket_name
      }

      env {
        name  = "STORAGE_PRIVATE_BUCKET"
        value = var.private_bucket_name
      }

      env {
        name  = "QUEUE_COMPOSITION_PROCESSING"
        value = var.composition_subscription_name
      }

      # Database IAM authentication (no password needed)
      env {
        name  = "POSTGRES_USER"
        value = var.worker_database_user
      }

      env {
        name  = "POSTGRES_IAM_AUTH"
        value = "true"
      }

      # Cloud SQL volume mount
      volume_mounts {
        name       = "cloudsql"
        mount_path = "/cloudsql"
      }

      # Health check
      startup_probe {
        initial_delay_seconds = 15
        timeout_seconds       = 10
        period_seconds        = 15
        failure_threshold     = 5
        
        http_get {
          path = "/health"
          port = 8081
        }
      }

      liveness_probe {
        initial_delay_seconds = 60
        timeout_seconds       = 10
        period_seconds        = 60
        failure_threshold     = 3
        
        http_get {
          path = "/health"
          port = 8081
        }
      }
    }

    # Annotations
    annotations = {
      "autoscaling.knative.dev/minScale" = tostring(var.worker_min_instances)
      "autoscaling.knative.dev/maxScale" = tostring(var.worker_max_instances)
      "run.googleapis.com/cloudsql-instances" = var.database_connection_name
      "run.googleapis.com/client-name" = "terraform"
    }
  }

  # Traffic allocation
  traffic {
    type    = "TRAFFIC_TARGET_ALLOCATION_TYPE_LATEST"
    percent = 100
  }

  labels = {
    environment = var.environment
    service     = "worker"
  }
}

# IAM policy to allow client to invoke API service
resource "google_cloud_run_v2_service_iam_member" "client_invoke_api" {
  location = google_cloud_run_v2_service.api.location
  name     = google_cloud_run_v2_service.api.name
  role     = "roles/run.invoker"
  member   = "serviceAccount:${var.client_service_account_email}"
  project  = var.project_id
}

# IAM policy to allow unauthenticated access to client service (public web)
resource "google_cloud_run_v2_service_iam_member" "client_public_access" {
  location = google_cloud_run_v2_service.client.location
  name     = google_cloud_run_v2_service.client.name
  role     = "roles/run.invoker"
  member   = "allUsers"
  project  = var.project_id
}