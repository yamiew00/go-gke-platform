# One-time stack: the GCS bucket that stores Terraform state for infra/platform.
# Its own state stays local because a bucket cannot hold state before it exists.

terraform {
  required_version = ">= 1.11"
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 8.4"
    }
  }
}

provider "google" {
  project = var.project_id
  region  = var.region
}

variable "project_id" {
  type        = string
  description = "GCP project that hosts the platform."
}

variable "region" {
  type        = string
  description = "Region for the state bucket."
  default     = "asia-east1"
}

# APIs the platform stack needs before it can even read the project or its state.
resource "google_project_service" "base" {
  for_each = toset([
    "cloudresourcemanager.googleapis.com",
    "serviceusage.googleapis.com",
    "storage.googleapis.com",
  ])
  service            = each.value
  disable_on_destroy = false
}

resource "google_storage_bucket" "tfstate" {
  name                        = "${var.project_id}-tfstate"
  location                    = var.region
  uniform_bucket_level_access = true
  public_access_prevention    = "enforced"

  # Every state write is kept, so a bad apply can be traced and rolled back.
  versioning {
    enabled = true
  }

  lifecycle_rule {
    condition {
      num_newer_versions = 10
      with_state         = "ARCHIVED"
    }
    action {
      type = "Delete"
    }
  }

  depends_on = [google_project_service.base]
}

output "state_bucket" {
  value = google_storage_bucket.tfstate.name
}
