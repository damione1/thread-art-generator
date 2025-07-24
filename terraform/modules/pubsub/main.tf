# Main topic for composition processing (replacing RabbitMQ)
resource "google_pubsub_topic" "composition_processing" {
  name    = "composition-processing-${var.environment}"
  project = var.project_id

  # Message retention for 7 days
  message_retention_duration = "604800s"

  labels = {
    environment = var.environment
    purpose     = "composition-processing"
  }
}

# Subscription for the worker service
resource "google_pubsub_subscription" "composition_processing_worker" {
  name    = "composition-processing-worker-${var.environment}"
  topic   = google_pubsub_topic.composition_processing.name
  project = var.project_id

  # Acknowledge deadline of 600 seconds (10 minutes) for long-running tasks
  ack_deadline_seconds = 600

  # Message retention for 7 days
  message_retention_duration = "604800s"

  # Retry policy for failed messages
  retry_policy {
    minimum_backoff = "10s"
    maximum_backoff = "300s"
  }

  # Dead letter policy for permanently failed messages
  dead_letter_policy {
    dead_letter_topic     = google_pubsub_topic.composition_processing_dead_letter.id
    max_delivery_attempts = 5
  }

  # Enable message ordering if needed
  enable_message_ordering = false

  # Filter messages if needed (commented out for now)
  # filter = "attributes.processing_type=\"composition\""

  labels = {
    environment = var.environment
    purpose     = "worker-subscription"
  }
}

# Dead letter topic for failed messages
resource "google_pubsub_topic" "composition_processing_dead_letter" {
  name    = "composition-processing-dead-letter-${var.environment}"
  project = var.project_id

  # Longer retention for dead letter messages
  message_retention_duration = "2592000s"  # 30 days

  labels = {
    environment = var.environment
    purpose     = "dead-letter"
  }
}

# Subscription for monitoring dead letter messages
resource "google_pubsub_subscription" "composition_processing_dead_letter_monitor" {
  name    = "composition-processing-dead-letter-monitor-${var.environment}"
  topic   = google_pubsub_topic.composition_processing_dead_letter.name
  project = var.project_id

  # Short ack deadline for monitoring
  ack_deadline_seconds = 60

  # Long retention for analysis
  message_retention_duration = "2592000s"  # 30 days

  labels = {
    environment = var.environment
    purpose     = "monitoring"
  }
}

# Topic for image processing notifications
resource "google_pubsub_topic" "image_processing" {
  name    = "image-processing-${var.environment}"
  project = var.project_id

  message_retention_duration = "604800s"

  labels = {
    environment = var.environment
    purpose     = "image-processing"
  }
}

# Subscription for image processing
resource "google_pubsub_subscription" "image_processing_worker" {
  name    = "image-processing-worker-${var.environment}"
  topic   = google_pubsub_topic.image_processing.name
  project = var.project_id

  ack_deadline_seconds       = 300  # 5 minutes for image processing
  message_retention_duration = "604800s"

  retry_policy {
    minimum_backoff = "5s"
    maximum_backoff = "120s"
  }

  dead_letter_policy {
    dead_letter_topic     = google_pubsub_topic.image_processing_dead_letter.id
    max_delivery_attempts = 3
  }

  labels = {
    environment = var.environment
    purpose     = "image-worker"
  }
}

# Dead letter topic for image processing
resource "google_pubsub_topic" "image_processing_dead_letter" {
  name    = "image-processing-dead-letter-${var.environment}"
  project = var.project_id

  message_retention_duration = "2592000s"

  labels = {
    environment = var.environment
    purpose     = "dead-letter"
  }
}

# IAM bindings for service accounts

# API service account - publisher access
resource "google_pubsub_topic_iam_member" "api_composition_publisher" {
  topic   = google_pubsub_topic.composition_processing.name
  role    = "roles/pubsub.publisher"
  member  = "serviceAccount:${var.api_service_account_email}"
  project = var.project_id
}

resource "google_pubsub_topic_iam_member" "api_image_publisher" {
  topic   = google_pubsub_topic.image_processing.name
  role    = "roles/pubsub.publisher"
  member  = "serviceAccount:${var.api_service_account_email}"
  project = var.project_id
}

# Worker service account - subscriber access
resource "google_pubsub_subscription_iam_member" "worker_composition_subscriber" {
  subscription = google_pubsub_subscription.composition_processing_worker.name
  role         = "roles/pubsub.subscriber"
  member       = "serviceAccount:${var.worker_service_account_email}"
  project      = var.project_id
}

resource "google_pubsub_subscription_iam_member" "worker_image_subscriber" {
  subscription = google_pubsub_subscription.image_processing_worker.name
  role         = "roles/pubsub.subscriber"
  member       = "serviceAccount:${var.worker_service_account_email}"
  project      = var.project_id
}

# Additional permissions for worker to acknowledge messages
resource "google_pubsub_subscription_iam_member" "worker_composition_viewer" {
  subscription = google_pubsub_subscription.composition_processing_worker.name
  role         = "roles/pubsub.viewer"
  member       = "serviceAccount:${var.worker_service_account_email}"
  project      = var.project_id
}

resource "google_pubsub_subscription_iam_member" "worker_image_viewer" {
  subscription = google_pubsub_subscription.image_processing_worker.name
  role         = "roles/pubsub.viewer"
  member       = "serviceAccount:${var.worker_service_account_email}"
  project      = var.project_id
}