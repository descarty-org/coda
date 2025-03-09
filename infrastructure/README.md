# Infrastructure as Code (IaC)

This repository contains the infrastructure-as-code for the project, managed with Terraform on Google Cloud Platform (Google Cloud).

This project does not include any sensitive information, such as credentials or private keys. All sensitive data should be stored securely and managed separately outside of version control.

Disclaimer: This repository is for experimental purposes only. Do not use this code in production without proper review and testing.

## Overview

The infrastructure is organized by environment (`dev`, `staging`, `prod`), with each environment having its own Terraform configuration.

## Prerequisites

- [Google Cloud SDK](https://cloud.google.com/sdk/docs/install) (latest stable)
- [Terraform](https://developer.hashicorp.com/terraform/downloads) (latest stable)
- Google Cloud account with appropriate permissions
- Repository mapping configured in [Cloud Build Triggers](https://console.cloud.google.com/cloud-build/triggers)

## Quick Start

```bash
# 1. Authenticate with Google Cloud
gcloud config set core/account your-email-account@gmail.com
gcloud auth login
gcloud auth application-default login

# 2. Navigate to your environment directory
cd environments/dev

# 3. Initialize Terraform
terraform init

# 4. Plan your changes
terraform plan

# 5. Apply changes
terraform apply
```

## Detailed Deployment Guide

### Creating a New Google Cloud Project

```bash
# Create a new project
gcloud projects create PROJECT_ID --name="Project Display Name"

# Set as default project
gcloud config set project PROJECT_ID

# Enable required APIs
gcloud services enable cloudbuild.googleapis.com \
    cloudresourcemanager.googleapis.com \
    compute.googleapis.com \
    iam.googleapis.com \
    storage-api.googleapis.com
```

### Setting Up GCS Bucket for Terraform State

```bash
# Set variables
PROJECT_ID=$(gcloud config get-value project)
BUCKET_NAME=${PROJECT_ID}-tfstate
REGION=asia

# Create bucket with uniform bucket-level access
gsutil mb -l ${REGION} -b on gs://${BUCKET_NAME}

# Enable versioning
gsutil versioning set on gs://${BUCKET_NAME}

# Set public access prevention to enforced
gsutil pap set enforced gs://${BUCKET_NAME}

# Display success message
echo "Terraform state bucket created: gs://${BUCKET_NAME}"
```

After creating the bucket, create the `backend.tf` file in the environment directory with the following content:

```hcl
terraform {
  backend "gcs" {
    bucket  = "BUCKET_NAME"
    prefix  = "terraform/state"
  }
}
```

### Setting Up Terraform variables

Copy the `environments/exp/terraform.tfvars.example` file to `environments/exp/terraform.tfvars` and fill in the required variables.

Note: change `exp` to the appropriate environment (e.g., `dev`, `staging`, `prod`).

### Working with Terraform

#### Environment Management

The repository is organized by environments:

```
infrastructure/
├── environments/
│   ├── dev/
│   ├── staging/
│   └── prod/
├── modules/
│   ├── compute/
│   ├── networking/
│   └── storage/
└── scripts/
```

Navigate to the appropriate environment directory before running Terraform commands:

```bash
cd environments/<env>  # Where <env> is dev, staging, or prod
```

#### Common Terraform Operations

```bash
# Initialize Terraform (required once per environment)
terraform init

# Preview changes
terraform plan -out=tfplan

# Apply changes
terraform apply tfplan  # Apply the saved plan
# OR
terraform apply        # Plan and apply in one step (requires confirmation)

# Destroy all resources (use with extreme caution)
terraform destroy
```

#### Troubleshooting

If authentication fails:
```bash
gcloud auth application-default login
```

If state is locked:
```bash
terraform force-unlock LOCK_ID
```

## CI/CD Pipeline

This repository uses Cloud Build for CI/CD. The pipeline:

1. Validates Terraform configuration
2. Plans changes
3. Applies changes (in approved environments)

Triggers are configured in the [Cloud Build Triggers Console](https://console.cloud.google.com/cloud-build/triggers).

## Best Practices

### Terraform Usage

- Use modules for reusable components
- Set explicit provider versions
- Use variables for environment-specific values
- Follow naming conventions for resources

### Security Considerations

- Use service accounts with minimal permissions
- Enable audit logging
- Implement least privilege access
- Encrypt sensitive data

## Terraform Naming Conventions

### Resource Names
- Use `snake_case` for all resource names
- Use a purpose-based identifier without duplicating the resource type information
- Example: `app_logs` for a Google Storage Bucket (not `bucket_app_logs`)
- Be descriptive and functional (avoid generic names like `logs1`)
- Use singular/plural consistently and appropriately

### Examples
```hcl
# Good - clear and non-redundant
resource "google_storage_bucket" "app_logs" {
  bucket = "${var.project}-logs" 
}

# Good - simplified for main resource in module
resource "google_storage_bucket" "this" {
  bucket = var.bucket_name
}

# Bad - too generic
resource "google_storage_bucket" "logs1" {
  bucket = "logs1"
}

# Bad - redundant (includes resource type)
resource "google_storage_bucket" "bucket_app_logs" {
  bucket = "app-logs"
}
```

### When to use "this"
- Use `this` as a resource name when working with single-resource modules
- Appropriate when the purpose of the resource is self-evident from the module context
- Useful in reusable modules where the resource purpose is defined by module variables

### Resource Naming vs Resource Attributes
- The resource identifier is for Terraform references only
- Actual cloud resource names are set via specific attributes (`bucket`, `name`, etc.)
- Many resources generate default names if not explicitly provided

## Resources

- [Terraform Google Cloud Provider Documentation](https://registry.terraform.io/providers/hashicorp/google/latest/docs)
- [Terraform Best Practices](https://www.terraform-best-practices.com/)
- [Google Cloud Architecture Center](https://cloud.google.com/architecture)
- [Terraform Style Guide](https://developer.hashicorp.com/terraform/language/style)

## Contributing

1. Create a feature branch from `main`
2. Make your changes
3. Run `terraform fmt` and `terraform validate`
4. Submit a pull request
