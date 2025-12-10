terraform {
  required_version = ">= 1.0"

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
  }
}

provider "google" {
  project = var.project_id
  region  = var.region
}

# Enable required APIs
resource "google_project_service" "cloud_run" {
  service = "run.googleapis.com"
  disable_on_destroy = false
}

resource "google_project_service" "cloud_build" {
  service = "cloudbuild.googleapis.com"
  disable_on_destroy = false
}

resource "google_project_service" "container_registry" {
  service = "containerregistry.googleapis.com"
  disable_on_destroy = false
}

# Backend Cloud Run Service
resource "google_cloud_run_service" "backend" {
  name     = "birdwatcher-backend"
  location = var.region

  template {
    spec {
      containers {
        image = "gcr.io/${var.project_id}/birdwatcher-backend:latest"

        ports {
          container_port = 7000
        }

        env {
          name  = "VERSION"
          value = var.backend_version
        }

        env {
          name  = "GIN_MODE"
          value = "release"
        }

        resources {
          limits = {
            cpu    = "1000m"
            memory = "512Mi"
          }
        }
      }
    }

    metadata {
      annotations = {
        "autoscaling.knative.dev/maxScale" = "10"
        "autoscaling.knative.dev/minScale" = "0"
      }
    }
  }

  traffic {
    percent         = 100
    latest_revision = true
  }

  depends_on = [google_project_service.cloud_run]
}

# Frontend Cloud Run Service
resource "google_cloud_run_service" "frontend" {
  name     = "birdwatcher-frontend"
  location = var.region

  template {
    spec {
      containers {
        image = "gcr.io/${var.project_id}/birdwatcher-frontend:latest"

        ports {
          container_port = 6060
        }

        env {
          name  = "BACKEND_URL"
          value = google_cloud_run_service.backend.status[0].url
        }

        env {
          name  = "GIN_MODE"
          value = "release"
        }

        resources {
          limits = {
            cpu    = "1000m"
            memory = "512Mi"
          }
        }
      }
    }

    metadata {
      annotations = {
        "autoscaling.knative.dev/maxScale" = "10"
        "autoscaling.knative.dev/minScale" = "0"
      }
    }
  }

  traffic {
    percent         = 100
    latest_revision = true
  }

  depends_on = [google_project_service.cloud_run]
}

# Allow unauthenticated access to backend
resource "google_cloud_run_service_iam_member" "backend_noauth" {
  service  = google_cloud_run_service.backend.name
  location = google_cloud_run_service.backend.location
  role     = "roles/run.invoker"
  member   = "allUsers"
}

# Allow unauthenticated access to frontend
resource "google_cloud_run_service_iam_member" "frontend_noauth" {
  service  = google_cloud_run_service.frontend.name
  location = google_cloud_run_service.frontend.location
  role     = "roles/run.invoker"
  member   = "allUsers"
}
