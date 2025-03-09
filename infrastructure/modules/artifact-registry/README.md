# Artifact Registry Module

This Terraform module creates a Docker Artifact Registry in Google Cloud Platform.

## Purpose

The Artifact Registry module provisions a Docker container registry that can be used to store and manage Docker images. This is a key component for CI/CD pipelines and container-based deployments.

## Features

- Creates a Docker-format Artifact Registry repository
- Configurable region and project settings
- Uses the official GoogleCloudPlatform Artifact Registry module

## Usage

```hcl
module "artifact_registry" {
  source = "../../modules/artifact-registry"

  project_id             = "your-project-id"
  region                 = "us-central1"
  artifact_repository_id = "my-docker-repo"
}
```

## Required Inputs

| Name | Description | Type | Required |
|------|-------------|------|:--------:|
| project_id | The Google Cloud project ID | `string` | Yes |
| region | The region to deploy resources | `string` | Yes |
| artifact_repository_id | The ID of the artifact repository | `string` | Yes |

## Resources Created

- Google Cloud Artifact Registry repository configured for Docker images

## Notes

- This module uses the official GoogleCloudPlatform Artifact Registry module
- The repository is configured specifically for Docker images
- Make sure the Artifact Registry API is enabled in your project before using this module