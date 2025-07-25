terraform {
  backend "gcs" {
    bucket = "thread-art-terraform-state"
    prefix = "environments"
  }
}