# The demo secret never lands in Terraform state: an ephemeral random value feeds a
# write-only argument, so state only records that version 1 was written.
ephemeral "random_password" "demo" {
  length  = 32
  special = false
}

resource "google_secret_manager_secret" "demo" {
  secret_id = "demo-api-key"

  replication {
    auto {}
  }

  depends_on = [google_project_service.apis]
}

resource "google_secret_manager_secret_version" "demo" {
  secret                 = google_secret_manager_secret.demo.id
  secret_data_wo         = ephemeral.random_password.demo.result
  secret_data_wo_version = 1 # bump to rotate: Terraform cannot diff a value it never stores
}

# Only the app's Kubernetes service account may read the secret. The CI deployer cannot.
resource "google_secret_manager_secret_iam_member" "app_reads_demo" {
  secret_id = google_secret_manager_secret.demo.id
  role      = "roles/secretmanager.secretAccessor"
  member    = local.app_principal
}
