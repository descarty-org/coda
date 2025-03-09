terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 6.6.0"
    }
    google-beta = {
      source  = "hashicorp/google-beta"
      version = "~> 6.6.0"
    }
  }
}

############################################
# Provider
############################################

provider "google" {
  project = var.project_id
  region  = var.region
}

provider "google-beta" {
  project = var.project_id
  region  = var.region
}

############################################
# APIs
############################################

resource "google_project_service" "enabled_apis" {
  for_each = toset([
    "run.googleapis.com",
    "cloudbuild.googleapis.com",
    "artifactregistry.googleapis.com",
  ])

  service            = each.key
  disable_on_destroy = false
}

############################################
# Service Account
############################################
resource "google_project_iam_member" "cloudbuildsa_cloud_run_admin" {
  project = var.project_id
  role    = "roles/run.admin"
  member  = "serviceAccount:${google_service_account.cloudbuild_service_account.email}"
}

resource "google_service_account" "cloudbuild_service_account" {
  account_id   = "cloudbuild-sa"
  display_name = "Cloud Build Service Account"
}

############################################
# Artifact Registry
############################################

module "artifact_registry" {
  source = "../../modules/artifact-registry"

  region = var.region
  project_id = var.project_id
  artifact_repository_id = var.artifact_repository_id
}

resource "google_project_iam_member" "cloudbuild_artifacts" {
  project = var.project_id
  role    = "roles/artifactregistry.writer"
  member  = "serviceAccount:${google_service_account.cloudbuild_service_account.email}"
}

############################################
# Ollama Backend
############################################

resource "google_cloudbuild_trigger" "ollama_backend_push" {
  project  = var.project_id
  name     = "ollama-backend-push"
  filename = "cloudbuild.ollama.yaml"

  github {
    owner = "descarty-org"
    name  = "coda"
    push {
      branch = "^main$"
    }
  }

  included_files = [
    "docker/**",
  ]

  substitutions = {
    _PROJECT_ID : var.project_id
    _REPOSITORY : var.artifact_repository_id
    _REGION     : var.region
  }

  include_build_logs = "INCLUDE_BUILD_LOGS_WITH_STATUS"
  service_account    = google_service_account.cloudbuild_service_account.id

  depends_on = [
    google_project_service.enabled_apis,
    google_service_account.cloudbuild_service_account,
    module.artifact_registry,
  ]
}

module "ollama" {
  source = "../../modules/ollama"

  region = var.region_cloud_run
  project_id = var.project_id
  image = "${var.region}-docker.pkg.dev/${var.project_id}/${var.artifact_repository_id}/ollama-backend:latest"
  service_name = "ollama"
  service_account = google_service_account.cloudbuild_service_account.email
}