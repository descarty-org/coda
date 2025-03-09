data "google_secret_manager_secret_version" "allowed_origins" {
  secret  = "ALLOWED_ORIGINS"
  version = "latest"
}

data "google_secret_manager_secret_version" "langfuse_private_key" {
  secret  = "LANGFUSE_PRIVATE_KEY"
  version = "latest"
}

data "google_secret_manager_secret_version" "langfuse_public_key" {
  secret  = "LANGFUSE_PUBLIC_KEY"
  version = "latest"
}

data "google_secret_manager_secret_version" "ollama_base_url" {
  secret  = "OLLAMA_BASE_URL"
  version = "latest"
}

data "google_secret_manager_secret_version" "openai_api_key" {
  secret  = "OPENAI_API_KEY"
  version = "latest"
}

locals {
  container_config = {
    env = [
      {
        name  = "CODA_ENV",
        value = var.env
      },
      {
        name = "ALLOWED_ORIGINS",
        secret = {
          secret_id = "ALLOWED_ORIGINS",
          version   = data.google_secret_manager_secret_version.allowed_origins.version
        }
      },
      {
        name = "LANGFUSE_PRIVATE_KEY",
        secret = {
          secret_id = "LANGFUSE_PRIVATE_KEY",
          version   = data.google_secret_manager_secret_version.langfuse_private_key.version
        }
      },
      {
        name = "LANGFUSE_PUBLIC_KEY",
        secret = {
          secret_id = "LANGFUSE_PUBLIC_KEY",
          version   = data.google_secret_manager_secret_version.langfuse_public_key.version
        }
      },
      {
        name = "OLLAMA_BASE_URL",
        secret = {
          secret_id = "OLLAMA_BASE_URL",
          version   = data.google_secret_manager_secret_version.ollama_base_url.version
        }
      },
      {
        name = "OPENAI_API_KEY",
        secret = {
          secret_id = "OPENAI_API_KEY",
          version   = data.google_secret_manager_secret_version.openai_api_key.version
        }
      },
    ]
  }
}
