# Coda Module

This Terraform module deploys the Coda application as a Cloud Run service in Google Cloud Platform.

## Purpose

The Coda module provisions a Cloud Run service for the Coda application, which is designed to provide Local GenAI Models (Large Language Model) services. It configures the service with appropriate environment variables, scaling settings, and IAM policies.

## Features

- Deploys a Cloud Run service for the Coda application
- Configures environment variables from Secret Manager
- Sets up appropriate scaling parameters
- Configures service account and IAM policies

## Usage

```hcl
module "coda" {
  source = "../../modules/coda"

  project_id      = "your-project-id"
  service_name    = "coda"
  region          = "us-central1"
  image           = "gcr.io/your-project/coda:latest"
  service_account = "your-service-account@your-project.iam.gserviceaccount.com"
  env             = "dev"
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
| env | The environment to deploy to | `string` | Yes |

## Secret Manager Requirements

This module expects the following secrets to be available in Secret Manager:

- `ALLOWED_ORIGINS` - Allowed CORS origins
- `LANGFUSE_PRIVATE_KEY` - Langfuse private key for observability
- `LANGFUSE_PUBLIC_KEY` - Langfuse public key for observability
- `OLLAMA_BASE_URL` - Base URL for the Ollama service
- `OPENAI_API_KEY` - OpenAI API key

## Resources Created

- Google Cloud Run service for the Coda application
- IAM policy binding for the Cloud Run service

## Security Considerations

**Note:** The current configuration allows public access to the Cloud Run service (`allUsers` as invokers). This is intended for demonstration purposes only and should be restricted in production environments.

## Related Modules

- [ollama](../ollama) - Deploys the Ollama service that this module integrates with
- [artifact-registry](../artifact-registry) - Creates the Artifact Registry for storing Docker images