# Ollama Module

This Terraform module deploys Ollama as a Cloud Run service with GPU support in Google Cloud Platform.

## Purpose

The Ollama module provisions a Cloud Run service for Ollama, which is an open-source LLM server. It configures the service with GPU support, appropriate scaling settings, and IAM policies to enable high-performance inference capabilities.

## Features

- Deploys Ollama as a Cloud Run service
- Configures GPU support (NVIDIA T4 GPU)
- Sets up appropriate scaling parameters
- Configures service account and IAM policies
- Optimizes for LLM inference with concurrency settings

## Usage

```hcl
module "ollama" {
  source = "../../modules/ollama"

  project_id      = "your-project-id"
  service_name    = "ollama"
  region          = "us-central1"
  image           = "gcr.io/your-project/ollama:latest"
  service_account = "your-service-account@your-project.iam.gserviceaccount.com"
}
```

## Required Inputs

| Name | Description | Type | Required |
|------|-------------|------|:--------:|
| project_id | The project to deploy resources | `string` | Yes |
| service_name | The name of the Cloud Run service | `string` | Yes |
| region | The region to deploy resources | `string` | Yes |
| image | The Docker image to deploy | `string` | Yes |
| service_account | The service account to use | `string` | Yes |

## Resources Created

- Google Cloud Run service for Ollama with GPU support
- IAM policy binding for the Cloud Run service

## GPU Configuration

This module configures the Cloud Run service with the following GPU resources:

- 1 NVIDIA GPU
- 8 vCPUs
- 32 GB memory
- CPU throttling disabled for optimal performance

## Security Considerations

**Note:** The current configuration allows public access to the Cloud Run service (`allUsers` as invokers). This is intended for demonstration purposes only and should be restricted in production environments.

## Related Modules

- [coda](../coda) - Deploys the Coda application that integrates with this Ollama service
- [artifact-registry](../artifact-registry) - Creates the Artifact Registry for storing Docker images

## Important Notes

- The Cloud Run service is configured in BETA launch stage, which is required for GPU support
- The service is configured with a 300-second timeout to accommodate longer inference requests
- Instance concurrency is set to 4 to optimize for LLM inference workloads
- The service scales from 0 to 1 instances to manage costs while providing availability