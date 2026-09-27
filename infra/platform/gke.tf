# Autopilot: Google runs the nodes and billing follows the pods' resource requests,
# which suits a small, bursty workload better than paying for idle nodes.
resource "google_container_cluster" "autopilot" {
  name     = var.cluster_name
  location = var.region

  enable_autopilot    = true
  deletion_protection = false # lab cluster: `terraform destroy` must be able to remove it

  network    = google_compute_network.vpc.id
  subnetwork = google_compute_subnetwork.gke.id

  ip_allocation_policy {
    cluster_secondary_range_name  = "pods"
    services_secondary_range_name = "services"
  }

  release_channel {
    channel = "REGULAR"
  }

  depends_on = [google_project_service.apis]
}
