data "google_project" "this" {}

locals {
  services = toset([
    "artifactregistry.googleapis.com",
    "compute.googleapis.com",
    "container.googleapis.com",
    "iam.googleapis.com",
    "iamcredentials.googleapis.com",
    "secretmanager.googleapis.com",
    "sts.googleapis.com",
  ])

  # Workload Identity Federation for GKE lets IAM name a Kubernetes service account directly,
  # so the pod needs neither a Google service account nor a key file.
  app_principal = join("/", [
    "principal://iam.googleapis.com/projects/${data.google_project.this.number}/locations/global",
    "workloadIdentityPools/${var.project_id}.svc.id.goog/subject/ns/${var.app_namespace}/sa/${var.app_name}",
  ])
}

resource "google_project_service" "apis" {
  for_each           = local.services
  service            = each.value
  disable_on_destroy = false
}
