resource "aws_sqs_queue" "media_jobs_dlq" {
  name                      = local.media_jobs_dlq_queue_name
  message_retention_seconds = 1209600
  sqs_managed_sse_enabled   = true
}

resource "aws_sqs_queue" "media_jobs" {
  name                       = local.media_jobs_queue_name
  delay_seconds              = 0
  message_retention_seconds  = 345600
  receive_wait_time_seconds  = 10
  visibility_timeout_seconds = 300
  sqs_managed_sse_enabled    = true

  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.media_jobs_dlq.arn
    maxReceiveCount     = 3
  })
}
