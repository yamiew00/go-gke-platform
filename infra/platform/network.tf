# A dedicated VPC instead of the auto-created "default" network: explicit IP ranges,
# and no race with the default network being created when the Compute API turns on.
resource "google_compute_network" "vpc" {
  name                    = "gke-platform-vpc"
  auto_create_subnetworks = false

  depends_on = [google_project_service.apis]
}

resource "google_compute_subnetwork" "gke" {
  name                     = "gke-${var.region}"
  region                   = var.region
  network                  = google_compute_network.vpc.id
  ip_cidr_range            = "10.10.0.0/20" # nodes
  private_ip_google_access = true

  # VPC-native cluster: pods and services get their own secondary ranges.
  secondary_ip_range {
    range_name    = "pods"
    ip_cidr_range = "10.20.0.0/16"
  }
  secondary_ip_range {
    range_name    = "services"
    ip_cidr_range = "10.30.0.0/20"
  }
}
