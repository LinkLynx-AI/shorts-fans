locals {
  project_name    = "shorts-fans"
  environment     = "dev"
  resource_prefix = "${local.project_name}-${local.environment}"
  bucket_suffix   = "${data.aws_caller_identity.current.account_id}-${var.aws_region}"

  raw_bucket_name                     = "${local.resource_prefix}-raw-${local.bucket_suffix}"
  creator_avatar_upload_bucket_name   = "${local.resource_prefix}-avatar-upload-${local.bucket_suffix}"
  creator_avatar_delivery_bucket_name = "${local.resource_prefix}-avatar-delivery-${local.bucket_suffix}"
  creator_review_evidence_bucket_name = "${local.resource_prefix}-review-evidence-${local.bucket_suffix}"
  short_public_bucket_name            = "${local.resource_prefix}-short-public-${local.bucket_suffix}"
  main_private_bucket_name            = "${local.resource_prefix}-main-private-${local.bucket_suffix}"
  backend_ecr_repository_name         = "${local.resource_prefix}-backend"
  frontend_ecr_repository_name        = "${local.resource_prefix}-frontend"
  ecs_cluster_name                    = "${local.resource_prefix}-cluster"
  backend_log_group_name              = "/ecs/${local.resource_prefix}/backend"
  frontend_log_group_name             = "/ecs/${local.resource_prefix}/frontend"
  worker_log_group_name               = "/ecs/${local.resource_prefix}/worker"
  migration_log_group_name            = "/ecs/${local.resource_prefix}/migration"
  db_name                             = "shorts_fans"
  db_username                         = "shorts_fans"
  db_port                             = 5432
  redis_port                          = 6379
  dev_app_origin                      = "http://${aws_lb.dev.dns_name}"
  dev_allowed_app_origins             = distinct(concat(var.allowed_app_origins, [local.dev_app_origin]))
  cognito_user_pool_name              = "${local.resource_prefix}-fan-auth"
  cognito_user_pool_client_name       = "${local.resource_prefix}-fan-auth-app-client"
  cognito_email_from_display_name     = "shortsfans"
  cognito_email_from_formatted        = trimspace(var.cognito_email_from_address) == "" ? null : format("%s <%s>", local.cognito_email_from_display_name, trimspace(var.cognito_email_from_address))

  media_jobs_queue_name     = "${local.resource_prefix}-media-jobs"
  media_jobs_dlq_queue_name = "${local.resource_prefix}-media-jobs-dlq"

  tags = {
    Project     = local.project_name
    Environment = local.environment
    ManagedBy   = "terraform"
    Scope       = "dev-environment"
  }
}
