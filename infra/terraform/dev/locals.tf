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
  cognito_user_pool_name              = "${local.resource_prefix}-fan-auth"
  cognito_user_pool_client_name       = "${local.resource_prefix}-fan-auth-app-client"

  media_jobs_queue_name     = "${local.resource_prefix}-media-jobs"
  media_jobs_dlq_queue_name = "${local.resource_prefix}-media-jobs-dlq"

  tags = {
    Project     = local.project_name
    Environment = local.environment
    ManagedBy   = "terraform"
    Scope       = "media-sandbox"
  }
}
