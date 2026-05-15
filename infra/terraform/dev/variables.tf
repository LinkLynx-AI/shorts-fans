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

variable "dev_allowed_cidr_blocks" {
  description = "CIDR blocks allowed to reach the public dev ALB."
  type        = list(string)

  validation {
    condition = (
      length(var.dev_allowed_cidr_blocks) > 0 &&
      alltrue([for cidr in var.dev_allowed_cidr_blocks : can(cidrhost(cidr, 0))]) &&
      !contains(var.dev_allowed_cidr_blocks, "0.0.0.0/0")
    )
    error_message = "dev_allowed_cidr_blocks must contain valid CIDR blocks and must not include 0.0.0.0/0."
  }
}

variable "backend_image_tag" {
  description = "Backend ECR image tag used by API, worker, migration, and seed ECS tasks."
  type        = string
  default     = "dev"

  validation {
    condition     = trimspace(var.backend_image_tag) != ""
    error_message = "backend_image_tag must not be empty."
  }
}

variable "frontend_image_tag" {
  description = "Frontend ECR image tag used by the Next.js ECS task."
  type        = string
  default     = "dev"

  validation {
    condition     = trimspace(var.frontend_image_tag) != ""
    error_message = "frontend_image_tag must not be empty."
  }
}

variable "db_instance_class" {
  description = "RDS PostgreSQL instance class for dev."
  type        = string
  default     = "db.t4g.micro"
}

variable "redis_node_type" {
  description = "ElastiCache Redis node type for dev."
  type        = string
  default     = "cache.t4g.micro"
}

variable "backend_ecs_cpu" {
  description = "CPU units for backend API, worker, migration, and seed tasks."
  type        = number
  default     = 512
}

variable "backend_ecs_memory" {
  description = "Memory MiB for backend API, worker, migration, and seed tasks."
  type        = number
  default     = 1024
}

variable "frontend_ecs_cpu" {
  description = "CPU units for the frontend task."
  type        = number
  default     = 512
}

variable "frontend_ecs_memory" {
  description = "Memory MiB for the frontend task."
  type        = number
  default     = 1024
}

variable "backend_desired_count" {
  description = "Desired count for the backend API ECS service."
  type        = number
  default     = 1
}

variable "frontend_desired_count" {
  description = "Desired count for the frontend ECS service."
  type        = number
  default     = 1
}

variable "worker_desired_count" {
  description = "Desired count for the backend worker ECS service."
  type        = number
  default     = 1
}

variable "admin_api_token" {
  description = "Dev admin API token stored in Secrets Manager and passed to backend API tasks."
  type        = string
  sensitive   = true
  default     = "dev-admin-token"

  validation {
    condition     = trimspace(var.admin_api_token) != ""
    error_message = "admin_api_token must not be empty."
  }
}
