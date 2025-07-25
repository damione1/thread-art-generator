output "composition_processing_topic" {
  description = "Name of the composition processing topic"
  value       = google_pubsub_topic.composition_processing.name
}

output "composition_processing_subscription" {
  description = "Name of the composition processing subscription"
  value       = google_pubsub_subscription.composition_processing_worker.name
}

output "image_processing_topic" {
  description = "Name of the image processing topic"
  value       = google_pubsub_topic.image_processing.name
}

output "image_processing_subscription" {
  description = "Name of the image processing subscription"
  value       = google_pubsub_subscription.image_processing_worker.name
}

output "topic_names" {
  description = "Map of topic purposes to names"
  value = {
    composition_processing = google_pubsub_topic.composition_processing.name
    image_processing      = google_pubsub_topic.image_processing.name
    composition_dead_letter = google_pubsub_topic.composition_processing_dead_letter.name
    image_dead_letter     = google_pubsub_topic.image_processing_dead_letter.name
  }
}

output "subscription_names" {
  description = "Map of subscription purposes to names"
  value = {
    composition_worker = google_pubsub_subscription.composition_processing_worker.name
    image_worker      = google_pubsub_subscription.image_processing_worker.name
    dead_letter_monitor = google_pubsub_subscription.composition_processing_dead_letter_monitor.name
  }
}