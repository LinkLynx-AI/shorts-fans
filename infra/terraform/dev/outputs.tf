output "aws_region" {
  description = "AWS region configured for the dev media sandbox."
  value       = var.aws_region
}

output "cognito_user_pool_id" {
  description = "Cognito User Pool ID for dev fan auth."
  value       = aws_cognito_user_pool.fan_auth.id
}

output "cognito_user_pool_arn" {
  description = "Cognito User Pool ARN for dev fan auth."
  value       = aws_cognito_user_pool.fan_auth.arn
}

output "cognito_user_pool_client_id" {
  description = "Secretless Cognito App Client ID for backend fan auth."
  value       = aws_cognito_user_pool_client.fan_auth.id
}

output "cognito_user_pool_issuer_url" {
  description = "HTTPS issuer URL for the dev Cognito User Pool."
  value       = format("https://%s", aws_cognito_user_pool.fan_auth.endpoint)
}

output "cognito_email_sending_account" {
  description = "Configured Cognito email delivery mode for dev fan auth."
  value       = var.cognito_use_ses_developer_email ? "DEVELOPER" : "COGNITO_DEFAULT"
}

output "cognito_email_from_address" {
  description = "Branded FROM address prepared for Cognito SES delivery."
  value       = local.cognito_email_from_formatted
}

output "cognito_ses_email_identity_arn" {
  description = "SES email identity ARN created for Cognito sender branding."
  value       = trimspace(var.cognito_email_from_address) == "" ? null : aws_sesv2_email_identity.cognito_sender[0].arn
}

output "cognito_ses_email_identity_verification_status" {
  description = "Last refreshed SES sender identity verification status recorded in Terraform state. Manual SES verification changes require terraform refresh/plan/apply before this output updates."
  value       = trimspace(var.cognito_email_from_address) == "" ? null : aws_sesv2_email_identity.cognito_sender[0].verification_status
}

output "raw_bucket_name" {
  description = "Private raw upload bucket name."
  value       = aws_s3_bucket.raw.bucket
}

output "raw_bucket_arn" {
  description = "Private raw upload bucket ARN."
  value       = aws_s3_bucket.raw.arn
}

output "creator_avatar_upload_bucket_name" {
  description = "Private upload bucket name for creator avatar direct uploads."
  value       = aws_s3_bucket.creator_avatar_upload.bucket
}

output "creator_avatar_upload_bucket_arn" {
  description = "Private upload bucket ARN for creator avatar direct uploads."
  value       = aws_s3_bucket.creator_avatar_upload.arn
}

output "creator_avatar_delivery_bucket_name" {
  description = "Private S3 origin bucket name for creator avatar delivery."
  value       = aws_s3_bucket.creator_avatar_delivery.bucket
}

output "creator_avatar_delivery_bucket_arn" {
  description = "Private S3 origin bucket ARN for creator avatar delivery."
  value       = aws_s3_bucket.creator_avatar_delivery.arn
}

output "creator_avatar_base_url" {
  description = "Base URL for creator avatar objects via CloudFront."
  value       = "https://${aws_cloudfront_distribution.creator_avatar.domain_name}"
}

output "creator_avatar_cloudfront_distribution_id" {
  description = "CloudFront distribution ID for creator avatar delivery."
  value       = aws_cloudfront_distribution.creator_avatar.id
}

output "creator_avatar_cloudfront_distribution_arn" {
  description = "CloudFront distribution ARN for creator avatar delivery."
  value       = aws_cloudfront_distribution.creator_avatar.arn
}

output "creator_avatar_cloudfront_domain_name" {
  description = "CloudFront domain name for creator avatar delivery."
  value       = aws_cloudfront_distribution.creator_avatar.domain_name
}

output "creator_review_evidence_bucket_name" {
  description = "Private bucket name for creator registration evidence uploads and finalized review evidence."
  value       = aws_s3_bucket.creator_review_evidence.bucket
}

output "creator_review_evidence_bucket_arn" {
  description = "Private bucket ARN for creator registration evidence uploads and finalized review evidence."
  value       = aws_s3_bucket.creator_review_evidence.arn
}

output "short_public_bucket_name" {
  description = "Private S3 origin bucket name for short delivery."
  value       = aws_s3_bucket.short_public.bucket
}

output "short_public_bucket_arn" {
  description = "Private S3 origin bucket ARN for short delivery."
  value       = aws_s3_bucket.short_public.arn
}

output "short_public_base_url" {
  description = "Base URL for public short objects via CloudFront."
  value       = "https://${aws_cloudfront_distribution.short_public.domain_name}"
}

output "short_public_cloudfront_distribution_id" {
  description = "CloudFront distribution ID for public short delivery."
  value       = aws_cloudfront_distribution.short_public.id
}

output "short_public_cloudfront_distribution_arn" {
  description = "CloudFront distribution ARN for public short delivery."
  value       = aws_cloudfront_distribution.short_public.arn
}

output "short_public_cloudfront_domain_name" {
  description = "CloudFront domain name for public short delivery."
  value       = aws_cloudfront_distribution.short_public.domain_name
}

output "main_private_bucket_name" {
  description = "Private main delivery bucket name."
  value       = aws_s3_bucket.main_private.bucket
}

output "main_private_bucket_arn" {
  description = "Private main delivery bucket ARN."
  value       = aws_s3_bucket.main_private.arn
}

output "media_jobs_queue_url" {
  description = "Primary media jobs queue URL."
  value       = aws_sqs_queue.media_jobs.url
}

output "media_jobs_queue_arn" {
  description = "Primary media jobs queue ARN."
  value       = aws_sqs_queue.media_jobs.arn
}

output "media_jobs_dlq_url" {
  description = "Dead-letter queue URL for media jobs."
  value       = aws_sqs_queue.media_jobs_dlq.url
}

output "media_jobs_dlq_arn" {
  description = "Dead-letter queue ARN for media jobs."
  value       = aws_sqs_queue.media_jobs_dlq.arn
}

output "mediaconvert_service_role_arn" {
  description = "IAM role ARN that MediaConvert jobs must assume."
  value       = aws_iam_role.mediaconvert_service.arn
}

output "media_app_access_policy_arn" {
  description = "Managed policy ARN to attach manually to the dev app principal."
  value       = aws_iam_policy.media_app_access.arn
}

output "creator_avatar_app_access_policy_arn" {
  description = "Managed policy ARN to attach manually to the app principal for creator avatar upload/delivery."
  value       = aws_iam_policy.creator_avatar_app_access.arn
}

output "creator_review_evidence_app_access_policy_arn" {
  description = "Managed policy ARN to attach manually to the app principal for creator registration evidence uploads."
  value       = aws_iam_policy.creator_review_evidence_app_access.arn
}

output "dev_app_url" {
  description = "Public HTTP URL for the AWS dev app ALB."
  value       = "http://${aws_lb.dev.dns_name}"
}

output "backend_ecr_repository_url" {
  description = "Backend ECR repository URL."
  value       = aws_ecr_repository.backend.repository_url
}

output "frontend_ecr_repository_url" {
  description = "Frontend ECR repository URL."
  value       = aws_ecr_repository.frontend.repository_url
}

output "rds_endpoint" {
  description = "RDS PostgreSQL endpoint for dev."
  value       = aws_db_instance.dev.endpoint
}

output "redis_endpoint" {
  description = "ElastiCache Redis endpoint for dev."
  value       = "${aws_elasticache_cluster.dev.cache_nodes[0].address}:${local.redis_port}"
}

output "ecs_cluster_name" {
  description = "ECS cluster name for the dev app."
  value       = aws_ecs_cluster.dev.name
}

output "backend_service_name" {
  description = "Backend API ECS service name."
  value       = aws_ecs_service.backend_api.name
}

output "frontend_service_name" {
  description = "Frontend ECS service name."
  value       = aws_ecs_service.frontend.name
}

output "worker_service_name" {
  description = "Backend worker ECS service name."
  value       = aws_ecs_service.backend_worker.name
}

output "backend_maintenance_task_definition_arn" {
  description = "Backend maintenance task definition ARN for migration and seed one-off tasks."
  value       = aws_ecs_task_definition.backend_maintenance.arn
}

output "ecs_task_security_group_id" {
  description = "Security group ID used by backend maintenance ECS tasks."
  value       = aws_security_group.backend_ecs_tasks.id
}

output "backend_ecs_task_security_group_id" {
  description = "Security group ID used by backend API, worker, and maintenance ECS tasks."
  value       = aws_security_group.backend_ecs_tasks.id
}

output "frontend_ecs_task_security_group_id" {
  description = "Security group ID used by frontend ECS tasks."
  value       = aws_security_group.frontend_ecs_tasks.id
}

output "ecs_public_subnet_ids" {
  description = "Public subnet IDs used by dev ECS tasks."
  value       = local.dev_public_subnet_ids
}
