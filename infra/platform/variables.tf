variable "project_id" {
  type        = string
  description = "GCP project that hosts the platform."
}

variable "region" {
  type        = string
  description = "Region for the cluster, registry, and network."
  default     = "asia-east1"
}

variable "github_repository" {
  type        = string
  description = "owner/name of the only GitHub repository allowed to deploy through OIDC."
}

variable "cluster_name" {
  type    = string
  default = "gke-platform"
}

variable "app_name" {
  type        = string
  description = "Image name, Helm release, and Kubernetes service account of the app."
  default     = "go-gke-platform"
}

variable "app_namespace" {
  type    = string
  default = "app"
}
