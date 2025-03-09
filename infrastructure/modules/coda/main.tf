resource "google_cloud_run_v2_service" "backend" {
  name     = var.service_name
  location = var.region
  ingress  = "INGRESS_TRAFFIC_ALL"

  template {
    containers {
      image = var.image

      # Configure environment variables dynamically
      dynamic "env" {
        for_each = local.container_config.env
        content {
          name  = env.value.name
          value = lookup(env.value, "value", null)
          dynamic "value_source" {
            for_each = lookup(env.value, "secret", null) != null ? [env.value.secret] : []
            content {
              secret_key_ref {
                secret  = value_source.value.secret_id
                version = value_source.value.version
              }
            }
          }
        }
      }
    }

    service_account = var.service_account
    
    scaling {
      min_instance_count = 0
      max_instance_count = 1
    }

    timeout     = "300s"
  }

  deletion_protection = false
}

# Make the service invocable by all users (no authentication)
# Note: this is only for demonstration purposes and should not be used in production
data "google_iam_policy" "noauth" {
  binding {
    role    = "roles/run.invoker"
    members = ["allUsers"]
  }
}

# Make the service invocable by all users (no authentication)
# Note: this is only for demonstration purposes and should not be used in production
resource "google_cloud_run_v2_service_iam_policy" "policy" {
  location    = var.region
  name        = google_cloud_run_v2_service.backend.name
  policy_data = data.google_iam_policy.noauth.policy_data
}
