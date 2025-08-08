# Get the billing account
data "google_billing_account" "account" {
  billing_account = var.billing_account_id
}

# Get project information
data "google_project" "current" {
  project_id = var.project_id
}

# Create a budget with hard limits
resource "google_billing_budget" "monthly_budget" {
  billing_account = data.google_billing_account.account.id
  display_name    = "Thread Art ${var.environment} Monthly Budget"

  budget_filter {
    projects = ["projects/${data.google_project.current.number}"]
  }

  amount {
    specified_amount {
      currency_code = "USD"
      units         = tostring(var.monthly_budget_amount)
    }
  }

  threshold_rules {
    threshold_percent = 0.5  # 50%
    spend_basis       = "CURRENT_SPEND"
  }

  threshold_rules {
    threshold_percent = 0.8  # 80%
    spend_basis       = "CURRENT_SPEND"
  }

  threshold_rules {
    threshold_percent = 0.9  # 90%
    spend_basis       = "CURRENT_SPEND"
  }

  threshold_rules {
    threshold_percent = 1.0  # 100%
    spend_basis       = "CURRENT_SPEND"
  }

  # Send alerts to pub/sub topic for automated shutdown
  all_updates_rule {
    pubsub_topic                = google_pubsub_topic.budget_alerts.id
    schema_version              = "1.0"
    monitoring_notification_channels = var.notification_channels
  }
}

# Pub/Sub topic for budget alerts
resource "google_pubsub_topic" "budget_alerts" {
  name    = "budget-alerts-${var.environment}"
  project = var.project_id

  labels = {
    environment = var.environment
    purpose     = "budget-monitoring"
  }
}

# Pub/Sub subscription for budget alerts
resource "google_pubsub_subscription" "budget_alerts" {
  name    = "budget-alerts-subscription-${var.environment}"
  topic   = google_pubsub_topic.budget_alerts.name
  project = var.project_id

  # Message retention for 7 days
  message_retention_duration = "604800s"

  # Acknowledge deadline of 10 seconds
  ack_deadline_seconds = 10

  labels = {
    environment = var.environment
    purpose     = "budget-monitoring"
  }
}

# Cloud Function to handle budget alerts and shutdown services
resource "google_cloudfunctions2_function" "budget_enforcer" {
  name        = "budget-enforcer-${var.environment}"
  location    = var.region
  description = "Function to enforce budget limits by shutting down services"
  project     = var.project_id

  build_config {
    runtime     = "python311"
    entry_point = "budget_enforcer"
    source {
      storage_source {
        bucket = google_storage_bucket.function_source.name
        object = google_storage_bucket_object.function_source.name
      }
    }
  }

  service_config {
    max_instance_count = 1
    available_memory   = "256M"
    timeout_seconds    = 60
    
    environment_variables = {
      PROJECT_ID  = var.project_id
      ENVIRONMENT = var.environment
      BUDGET_THRESHOLD = var.monthly_budget_amount
    }

    service_account_email = var.cicd_service_account_email
  }

  event_trigger {
    trigger_region = var.region
    event_type     = "google.cloud.pubsub.topic.v1.messagePublished"
    pubsub_topic   = google_pubsub_topic.budget_alerts.id
  }

  depends_on = [
    google_storage_bucket_object.function_source
  ]
}

# Storage bucket for Cloud Function source
resource "google_storage_bucket" "function_source" {
  name     = "thread-art-budget-enforcer-${var.environment}-${random_id.bucket_suffix.hex}"
  location = var.region
  project  = var.project_id

  uniform_bucket_level_access = true
  
  lifecycle_rule {
    condition {
      age = 30  # Delete after 30 days
    }
    action {
      type = "Delete"
    }
  }
}

resource "random_id" "bucket_suffix" {
  byte_length = 4
}

# Upload the budget enforcer function source
resource "google_storage_bucket_object" "function_source" {
  name   = "budget-enforcer-${random_id.function_version.hex}.zip"
  bucket = google_storage_bucket.function_source.name
  source = data.archive_file.function_source.output_path
}

resource "random_id" "function_version" {
  byte_length = 4
}

# Create the function source code archive
data "archive_file" "function_source" {
  type        = "zip"
  output_path = "/tmp/budget-enforcer-${var.environment}.zip"
  
  source {
    content = templatefile("${path.module}/budget_enforcer.py", {
      project_id = var.project_id
      environment = var.environment
    })
    filename = "main.py"
  }

  source {
    content = file("${path.module}/requirements.txt")
    filename = "requirements.txt"
  }
}