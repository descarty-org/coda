variable "region" {
  description = "The region to deploy resources"
  type        = string
}

variable "region_cloud_run" {
  description = "The region for Cloud Run services (GPU support)"
  type        = string
}

variable "project_id" {
  description = "The Google Cloud project ID"
  type        = string
}

variable "artifact_repository_id" {
  description = "The ID of the artifact repository"
  type        = string
}

variable "github_owner" {
  description = "The owner of the GitHub repository"
  type        = string
}

variable "github_name" {
  description = "The name of the GitHub repository"
  type        = string
}
