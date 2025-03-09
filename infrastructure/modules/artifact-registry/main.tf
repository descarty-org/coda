module "artifact_registry" {
  source  = "GoogleCloudPlatform/artifact-registry/google"
  version = "~> 0.3"

  project_id    = var.project_id
  location      = var.region
  format        = "DOCKER"
  repository_id = var.artifact_repository_id
}
