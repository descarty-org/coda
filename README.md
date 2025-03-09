# Coda: Local GenAI Models on Cloud Run - A Go & HTMX Integration Example

## Overview

Coda is an example application for building a secure, scalable, and observable Large Language Model (LLM) powered application. The web application is built with Go and HTMX, and integrates with OpenAI, Langfuse, and Local GenAI Models models. The project includes Terraform configurations for deploying to Google Cloud Run with GPU support.

### Architecture

![Coda Architecture](./assets/images/coda-architecture.jpg)

### Demo

![Coda Demo](./assets/images/coda-demo.gif)

### Core Features

1. **Local GenAI Integration**: Run GenAI models locally with Ollama and integrate with the web application.
2. **Terraform IaC**: Infrastructure as Code (IaC) for deploying the application to Google Cloud Run with GPU support.
3. **Lightweight Frontend**: The frontend is built with Go and HTMX for simplicity.
4. **Tracing**: Built-in Observability with Langfuse for LLM engineering.

### Codebase Structure

```
coda/
├── cmd/                  # Application entry points
├── config/               # Configuration management
├── docker/               # Docker configurations
├── gguf/                 # GGUF model management
├── infrastructure/       # Terraform IaC for Google Cloud
└── internal/             # Core application packages
    ├── config/           # Configuration loading
    ├── frontend/         # Web UI components
    ├── infrastructure/   # Server and middleware
    ├── llm/              # LLM integration layer
    │   ├── ollama/       # Ollama provider
    │   ├── openai/       # OpenAI provider
    │   └── langfuse/     # Observability
    ├── logger/           # Structured logging
    └── review/           # Code review features
```

## Prerequisites

- [Go](https://golang.org/dl/) (v1.24.0 or later)
- [Terraform](https://www.terraform.io/downloads.html) (latest stable version)
- [Docker](https://docs.docker.com/get-docker/) (for containerization)
- [Google Cloud SDK](https://cloud.google.com/sdk/docs/install) (for deployment)

## Setup Instructions

### Cloning the Repository
```bash
git clone git@github.com:descarty-org/coda.git
cd coda
```

### Environment Configuration

Copy the `.env.example` file to `.env` and fill in the required variables.

| Category | Variable | Description | Required |
|----------|----------|-------------| -------- |
| Global | `CODA_ENV` | Environment name (e.g., `local`, `dev`, `prod`) | - |
| Server | `PORT` | Port number for the server (default: 8080) | - |
| | `HOST` | Hostname for the server to listen on (default: 0.0.0.0) | - |
| | `ALLOWED_ORIGINS` | Comma-separated list of allowed origins | - |
| LLM | `OPENAI_API_KEY` | API key for OpenAI | Yes |
| | `OLLAMA_BASE_URL` | Base URL for the OLLAMA REST API | - |
| | `LANGFUSE_PUBLIC_KEY` | Public key for Langfuse observability | - |
| | `LANGFUSE_PRIVATE_KEY` | Private key for Langfuse observability | - |

### Local Development

#### Running Ollama Locally

```bash
ollama run "yottahmd/tiny-swallow-1.5b-instruct" # Download the model
ollama serve # Start the REST API server
```

#### Running the Server

```bash
make run
```

The web interface will be available at http://localhost:8080

To run the server with a custom Ollama base URL:
```bash
OLLAMA_BASE_URL=http://localhost:11434 make run
```

#### Testing

Run the test suite with coverage reporting:

```bash
make test
```

#### Linting and Code Quality

```bash
make lint
```

### Deployment

Coda can be deployed to Google Cloud Run with GPU support:

Note: Refer to the [infrastructure README](infrastructure/README.md) for detailed deployment instructions.

1. Set up your Google Cloud project:
   ```bash
   export PROJECT_ID=your-project-id
   gcloud config set project $PROJECT_ID
   ```

2. Copy `infrastructure/environments/exp/terraform.tfvars.example` to `terraform.tfvars` and fill in the required variables.

3. Initialize Terraform:
   ```bash
   cd infrastructure/environments/exp
   terraform init
   ```

4. Apply the Terraform configuration:
   ```bash
   terraform apply
   ```

## Contributing

Feel free to contribute to the project by opening an issue or submitting a pull request.

## License

This project is licensed under MIT License.

**Note:** The Tiny Swallow 1.5B model used in this project is not covered by this license. For the model's licensing information, please refer to the [Hugging Face page for TinySwallow-1.5B](https://huggingface.co/SakanaAI/TinySwallow-1.5B).
