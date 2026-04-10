variable "aws_region" {
  description = "AWS region for the dev media sandbox."
  type        = string

  validation {
    condition     = trimspace(var.aws_region) != ""
    error_message = "aws_region must not be empty."
  }
}

variable "allowed_app_origins" {
  description = "Origins allowed to access creator avatar upload and main delivery buckets from local app environments."
  type        = list(string)
  default = [
    "http://localhost:3000",
    "http://127.0.0.1:3000",
  ]

  validation {
    condition     = length(var.allowed_app_origins) > 0
    error_message = "allowed_app_origins must contain at least one origin."
  }
}
