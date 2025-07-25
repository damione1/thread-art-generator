terraform {
  backend "gcs" {
    bucket = "thread-art-terraform-state-thread-art-staging-466319"
    prefix = "staging"
  }
}
