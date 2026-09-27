output "cluster_name" {
  value = google_container_cluster.autopilot.name
}

output "cluster_location" {
  value = google_container_cluster.autopilot.location
}

output "image_repository" {
  value = "${var.region}-docker.pkg.dev/${var.project_id}/${google_artifact_registry_repository.apps.repository_id}/${var.app_name}"
}

output "workload_identity_provider" {
  value = google_iam_workload_identity_pool_provider.github.name
}

output "deployer_service_account" {
  value = google_service_account.deployer.email
}

# Everything the deploy workflow reads from repository variables, in one place:
#   terraform output -json github_actions_variables
output "github_actions_variables" {
  value = {
    GCP_PROJECT_ID   = var.project_id
    GCP_REGION       = var.region
    GKE_CLUSTER      = google_container_cluster.autopilot.name
    IMAGE_REPOSITORY = "${var.region}-docker.pkg.dev/${var.project_id}/${google_artifact_registry_repository.apps.repository_id}/${var.app_name}"
    WIF_PROVIDER     = google_iam_workload_identity_pool_provider.github.name
    DEPLOYER_SA      = google_service_account.deployer.email
  }
}
