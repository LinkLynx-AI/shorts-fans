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
