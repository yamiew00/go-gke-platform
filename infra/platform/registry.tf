resource "google_artifact_registry_repository" "apps" {
  location      = var.region
  repository_id = "apps"
  format        = "DOCKER"
  description   = "Images built by GitHub Actions, tagged with the commit SHA."

  # A tag always points at the same image, so deploying or rolling back a SHA is exact.
  docker_config {
    immutable_tags = true
  }

  # Keep the recent images and let old ones age out instead of growing storage forever.
  cleanup_policy_dry_run = false
  cleanup_policies {
    id     = "keep-recent"
    action = "KEEP"
    most_recent_versions {
      keep_count = 10
    }
  }
  cleanup_policies {
    id     = "delete-older-than-30d"
    action = "DELETE"
    condition {
      older_than = "2592000s"
    }
  }

  depends_on = [google_project_service.apis]
}
