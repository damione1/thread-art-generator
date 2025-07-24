output "public_bucket_name" {
  description = "Name of the public storage bucket"
  value       = google_storage_bucket.public_bucket.name
}

output "public_bucket_url" {
  description = "URL of the public storage bucket"
  value       = google_storage_bucket.public_bucket.url
}

output "public_bucket_self_link" {
  description = "Self link of the public storage bucket"
  value       = google_storage_bucket.public_bucket.self_link
}

output "private_bucket_name" {
  description = "Name of the private storage bucket"
  value       = google_storage_bucket.private_bucket.name
}

output "private_bucket_url" {
  description = "URL of the private storage bucket"
  value       = google_storage_bucket.private_bucket.url
}

output "private_bucket_self_link" {
  description = "Self link of the private storage bucket"
  value       = google_storage_bucket.private_bucket.self_link
}

output "bucket_names" {
  description = "Map of bucket purposes to names"
  value = {
    public  = google_storage_bucket.public_bucket.name
    private = google_storage_bucket.private_bucket.name
  }
}

output "bucket_urls" {
  description = "Map of bucket purposes to URLs"
  value = {
    public  = google_storage_bucket.public_bucket.url
    private = google_storage_bucket.private_bucket.url
  }
}