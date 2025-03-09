resource "google_cloud_run_v2_service" "backend" {
  name     = var.service_name
  location = var.region
  ingress  = "INGRESS_TRAFFIC_ALL"

  # Required for GPU support
  launch_stage = "BETA"

  template {
    containers {
      image = var.image

      # Configure resources with GPU
      resources {
        limits = {
          "cpu" = "8"
          "memory" = "32Gi"
          "nvidia.com/gpu" = "1"
        }
        startup_cpu_boost = true
      }
    }

    service_account = var.service_account

    # Required for GPU support
    annotations = {
      "run.googleapis.com/cpu-throttling" = false
      "run.googleapis.com/launch-stage" = "BETA"
    }
    
    scaling {
      min_instance_count = 0
      max_instance_count = 1
    }

    timeout     = "300s"

    # Configure concurrency
    max_instance_request_concurrency = 4
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
