output "aws_region" {
  description = "AWS region configured for the dev media sandbox."
  value       = var.aws_region
}

output "raw_bucket_name" {
  description = "Private raw upload bucket name."
  value       = aws_s3_bucket.raw.bucket
}

output "raw_bucket_arn" {
  description = "Private raw upload bucket ARN."
  value       = aws_s3_bucket.raw.arn
}

output "short_public_bucket_name" {
  description = "Public short delivery bucket name."
  value       = aws_s3_bucket.short_public.bucket
}

output "short_public_bucket_arn" {
  description = "Public short delivery bucket ARN."
  value       = aws_s3_bucket.short_public.arn
}

output "short_public_base_url" {
  description = "Base URL for public short objects when enable_public_short_delivery is true."
  value       = var.enable_public_short_delivery ? "https://${aws_s3_bucket.short_public.bucket_regional_domain_name}" : null
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
