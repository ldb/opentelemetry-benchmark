terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "4.5.0"
    }
  }
}

provider "google" {
  project = "opentelemetry-benchmark"
  region  = "europe-west1"
}