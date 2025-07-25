output "instance_name" {
  description = "Name of the Cloud SQL instance"
  value       = google_sql_database_instance.main.name
}

output "instance_connection_name" {
  description = "Connection name of the Cloud SQL instance"
  value       = google_sql_database_instance.main.connection_name
}

output "instance_private_ip" {
  description = "Private IP address of the Cloud SQL instance"
  value       = google_sql_database_instance.main.private_ip_address
}

output "instance_self_link" {
  description = "Self link of the Cloud SQL instance"
  value       = google_sql_database_instance.main.self_link
}

output "database_name" {
  description = "Name of the main database"
  value       = google_sql_database.main_database.name
}

output "api_iam_user" {
  description = "Database username for API service (IAM)"
  value       = google_sql_user.api_iam_user.name
}

output "worker_iam_user" {
  description = "Database username for worker service (IAM)"
  value       = google_sql_user.worker_iam_user.name
}

output "migrator_iam_user" {
  description = "Database username for migrator service (IAM)"
  value       = google_sql_user.migrator_iam_user.name
}

output "ssl_cert" {
  description = "SSL certificate for secure connections"
  value = {
    cert        = google_sql_ssl_cert.client_cert.cert
    private_key = google_sql_ssl_cert.client_cert.private_key
    server_ca_cert = google_sql_ssl_cert.client_cert.server_ca_cert
  }
  sensitive = true
}

# Connection details for applications using IAM authentication
output "connection_details" {
  description = "Database connection details for IAM authentication"
  value = {
    host               = google_sql_database_instance.main.private_ip_address
    port               = 5432
    database           = google_sql_database.main_database.name
    connection_name    = google_sql_database_instance.main.connection_name
    ssl_mode          = "require"
    iam_authentication = true
  }
}