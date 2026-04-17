variable "aws_region" {
  description = "AWS region for the dev media sandbox."
  type        = string

  validation {
    condition     = trimspace(var.aws_region) != ""
    error_message = "aws_region must not be empty."
  }
}

variable "cognito_email_from_address" {
  description = "SES verified sender email address for Cognito transactional emails in dev."
  type        = string
  default     = ""

  validation {
    condition = (
      trimspace(var.cognito_email_from_address) == "" ||
      can(regex("^[^\\s@]+@[^\\s@]+\\.[^\\s@]+$", trimspace(var.cognito_email_from_address)))
    )
    error_message = "cognito_email_from_address must be empty or a valid email address."
  }
}

variable "cognito_use_ses_developer_email" {
  description = "Enable Cognito DEVELOPER email sending with SES after the SES sender identity has been verified."
  type        = bool
  default     = false

  validation {
    condition = (
      !var.cognito_use_ses_developer_email ||
      trimspace(var.cognito_email_from_address) != ""
    )
    error_message = "cognito_email_from_address must be set when cognito_use_ses_developer_email is true."
  }
}

variable "allowed_app_origins" {
  description = "Origins allowed to access raw upload, creator avatar upload, creator review evidence upload, and main delivery buckets from local app environments."
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
