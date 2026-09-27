# Dot-source to point the current shell at the lab cluster:
#     . ./scripts/use-cluster.ps1
# Credentials go to this repo's .kube/config (git-ignored) and gcloud uses the "personal"
# configuration for this shell only, so ~/.kube/config and the default gcloud account stay untouched.

$env:CLOUDSDK_ACTIVE_CONFIG_NAME = if ($env:GCLOUD_CONFIG) { $env:GCLOUD_CONFIG } else { 'personal' }
$env:KUBECONFIG = Join-Path (Split-Path $PSScriptRoot -Parent) '.kube/config'

gcloud container clusters get-credentials gke-platform --region asia-east1 --project go-gke-platform
if ($LASTEXITCODE -eq 0) {
    Write-Host "kubectl now targets gke-platform (KUBECONFIG=$env:KUBECONFIG)" -ForegroundColor Green
}
