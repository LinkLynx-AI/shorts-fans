locals {
  postgres_dsn = format(
    "postgres://%s:%s@%s:%d/%s?sslmode=require",
    local.db_username,
    random_password.db.result,
    aws_db_instance.dev.address,
    local.db_port,
    local.db_name,
  )
}

resource "aws_secretsmanager_secret" "postgres_dsn" {
  name                    = "${local.resource_prefix}/backend/postgres-dsn"
  recovery_window_in_days = 0
}

resource "aws_secretsmanager_secret_version" "postgres_dsn" {
  secret_id     = aws_secretsmanager_secret.postgres_dsn.id
  secret_string = local.postgres_dsn
}

resource "aws_secretsmanager_secret" "admin_api_token" {
  name                    = "${local.resource_prefix}/backend/admin-api-token"
  recovery_window_in_days = 0
}

resource "aws_secretsmanager_secret_version" "admin_api_token" {
  secret_id     = aws_secretsmanager_secret.admin_api_token.id
  secret_string = var.admin_api_token
}

locals {
  ssm_runtime_parameters = {
    "/${local.project_name}/${local.environment}/app/url"                         = "http://${aws_lb.dev.dns_name}"
    "/${local.project_name}/${local.environment}/backend/aws-region"              = var.aws_region
    "/${local.project_name}/${local.environment}/backend/cognito-user-pool-id"    = aws_cognito_user_pool.fan_auth.id
    "/${local.project_name}/${local.environment}/backend/cognito-app-client-id"   = aws_cognito_user_pool_client.fan_auth.id
    "/${local.project_name}/${local.environment}/backend/media-jobs-queue-url"    = aws_sqs_queue.media_jobs.url
    "/${local.project_name}/${local.environment}/backend/raw-bucket"              = aws_s3_bucket.raw.bucket
    "/${local.project_name}/${local.environment}/backend/short-public-bucket"     = aws_s3_bucket.short_public.bucket
    "/${local.project_name}/${local.environment}/backend/short-public-base-url"   = "https://${aws_cloudfront_distribution.short_public.domain_name}"
    "/${local.project_name}/${local.environment}/backend/main-private-bucket"     = aws_s3_bucket.main_private.bucket
    "/${local.project_name}/${local.environment}/backend/creator-avatar-base-url" = "https://${aws_cloudfront_distribution.creator_avatar.domain_name}"
  }
}

resource "aws_ssm_parameter" "runtime" {
  for_each = local.ssm_runtime_parameters

  name  = each.key
  type  = "String"
  value = each.value
}
