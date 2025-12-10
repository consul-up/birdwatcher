variable "project_id" {
  description = "The GCP project ID"
  type        = string
}

variable "region" {
  description = "The GCP region to deploy resources"
  type        = string
  default     = "us-central1"
}

variable "backend_version" {
  description = "Backend version (v1 or v2)"
  type        = string
  default     = "v1"

  validation {
    condition     = contains(["v1", "v2"], var.backend_version)
    error_message = "Backend version must be either v1 or v2."
  }
}
