# Keyless CI/CD: GitHub Actions trades its OIDC token for short-lived Google credentials,
# so no service-account key exists anywhere to leak or rotate.

resource "random_id" "pool_suffix" {
  # A deleted pool keeps its ID reserved for 30 days; the suffix lets destroy + apply start clean.
  byte_length = 2
}

resource "google_iam_workload_identity_pool" "github" {
  workload_identity_pool_id = "github-${random_id.pool_suffix.hex}"
  display_name              = "GitHub Actions"

  depends_on = [google_project_service.apis]
}

resource "google_iam_workload_identity_pool_provider" "github" {
  workload_identity_pool_id          = google_iam_workload_identity_pool.github.workload_identity_pool_id
  workload_identity_pool_provider_id = "github-oidc"
  display_name                       = "GitHub OIDC"

  attribute_mapping = {
    "google.subject"       = "assertion.sub"
    "attribute.repository" = "assertion.repository"
    "attribute.ref"        = "assertion.ref"
  }

  # Only this repository's main branch can authenticate; any other token is rejected
  # before IAM is consulted, so a fork or a feature branch cannot deploy.
  attribute_condition = "assertion.repository == '${var.github_repository}' && assertion.ref == 'refs/heads/main'"

  oidc {
    issuer_uri = "https://token.actions.githubusercontent.com"
  }
}

resource "google_service_account" "deployer" {
  account_id   = "gha-deployer"
  display_name = "GitHub Actions deployer"
  description  = "Pushes images and deploys Helm releases. Has no access to secrets."
}

resource "google_service_account_iam_member" "github_impersonates_deployer" {
  service_account_id = google_service_account.deployer.name
  role               = "roles/iam.workloadIdentityUser"
  member             = "principalSet://iam.googleapis.com/${google_iam_workload_identity_pool.github.name}/attribute.repository/${var.github_repository}"
}

resource "google_artifact_registry_repository_iam_member" "deployer_pushes_images" {
  location   = google_artifact_registry_repository.apps.location
  repository = google_artifact_registry_repository.apps.name
  role       = "roles/artifactregistry.writer"
  member     = google_service_account.deployer.member
}

resource "google_project_iam_member" "deployer_deploys_to_gke" {
  project = var.project_id
  role    = "roles/container.developer"
  member  = google_service_account.deployer.member
}
