# Uncomment and use this for accessing secrets
# data "google_secret_manager_secret_version" "db_user" {
#   secret  = "db_user"
#   version = "latest"
# }

locals {
  container_config = {
    env = [
      # Add environment variables here. Example:
      # {
      #   name = "DB_USER",
      #   secret = {
      #     secret_id = "db_user",
      #     version   = "latest"
      #   }
      # },
      # {
      #   name  = "LOG_LEVEL",
      #   value = "info"
      # }
    ]
  }
}
