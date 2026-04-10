data "aws_iam_policy_document" "mediaconvert_assume_role" {
  statement {
    effect = "Allow"
    actions = [
      "sts:AssumeRole",
    ]

    principals {
      type        = "Service"
      identifiers = ["mediaconvert.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "mediaconvert_service" {
  name               = "${local.resource_prefix}-mediaconvert-service"
  assume_role_policy = data.aws_iam_policy_document.mediaconvert_assume_role.json
}

data "aws_iam_policy_document" "mediaconvert_service" {
  statement {
    sid    = "ListBucketsForMediaJobs"
    effect = "Allow"
    actions = [
      "s3:GetBucketLocation",
      "s3:ListBucket",
    ]
    resources = [
      aws_s3_bucket.raw.arn,
      aws_s3_bucket.short_public.arn,
      aws_s3_bucket.main_private.arn,
    ]
  }

  statement {
    sid    = "ReadRawInputs"
    effect = "Allow"
    actions = [
      "s3:GetObject",
      "s3:GetObjectVersion",
    ]
    resources = [
      "${aws_s3_bucket.raw.arn}/*",
    ]
  }

  statement {
    sid    = "WriteDeliveryOutputs"
    effect = "Allow"
    actions = [
      "s3:AbortMultipartUpload",
      "s3:PutObject",
      "s3:PutObjectTagging",
    ]
    resources = [
      "${aws_s3_bucket.short_public.arn}/*",
      "${aws_s3_bucket.main_private.arn}/*",
    ]
  }
}

resource "aws_iam_role_policy" "mediaconvert_service" {
  name   = "${local.resource_prefix}-mediaconvert-service"
  role   = aws_iam_role.mediaconvert_service.id
  policy = data.aws_iam_policy_document.mediaconvert_service.json
}

data "aws_iam_policy_document" "media_app_access" {
  statement {
    sid    = "ListMediaBuckets"
    effect = "Allow"
    actions = [
      "s3:GetBucketLocation",
      "s3:ListBucket",
    ]
    resources = [
      aws_s3_bucket.raw.arn,
      aws_s3_bucket.short_public.arn,
      aws_s3_bucket.main_private.arn,
    ]
  }

  statement {
    sid    = "ManageMediaObjects"
    effect = "Allow"
    actions = [
      "s3:AbortMultipartUpload",
      "s3:DeleteObject",
      "s3:GetObject",
      "s3:PutObject",
    ]
    resources = [
      "${aws_s3_bucket.raw.arn}/*",
      "${aws_s3_bucket.short_public.arn}/*",
      "${aws_s3_bucket.main_private.arn}/*",
    ]
  }

  statement {
    sid    = "UseMediaQueue"
    effect = "Allow"
    actions = [
      "sqs:ChangeMessageVisibility",
      "sqs:DeleteMessage",
      "sqs:GetQueueAttributes",
      "sqs:GetQueueUrl",
      "sqs:ReceiveMessage",
      "sqs:SendMessage",
    ]
    resources = [
      aws_sqs_queue.media_jobs.arn,
      aws_sqs_queue.media_jobs_dlq.arn,
    ]
  }

  statement {
    sid    = "ManageMediaConvertJobs"
    effect = "Allow"
    actions = [
      "mediaconvert:CancelJob",
      "mediaconvert:CreateJob",
      "mediaconvert:DescribeEndpoints",
      "mediaconvert:GetJob",
      "mediaconvert:GetPreset",
      "mediaconvert:GetQueue",
      "mediaconvert:ListJobs",
      "mediaconvert:ListPresets",
      "mediaconvert:ListQueues",
    ]
    resources = ["*"]
  }

  statement {
    sid    = "PassMediaConvertServiceRole"
    effect = "Allow"
    actions = [
      "iam:PassRole",
    ]
    resources = [
      aws_iam_role.mediaconvert_service.arn,
    ]

    condition {
      test     = "StringEquals"
      variable = "iam:PassedToService"
      values   = ["mediaconvert.amazonaws.com"]
    }
  }
}

resource "aws_iam_policy" "media_app_access" {
  name   = "${local.resource_prefix}-media-app-access"
  policy = data.aws_iam_policy_document.media_app_access.json
}

data "aws_iam_policy_document" "creator_avatar_app_access" {
  statement {
    sid    = "ListCreatorAvatarBuckets"
    effect = "Allow"
    actions = [
      "s3:GetBucketLocation",
      "s3:ListBucket",
    ]
    resources = [
      aws_s3_bucket.creator_avatar_upload.arn,
      aws_s3_bucket.creator_avatar_delivery.arn,
    ]
  }

  statement {
    sid    = "ManageCreatorAvatarObjects"
    effect = "Allow"
    actions = [
      "s3:AbortMultipartUpload",
      "s3:DeleteObject",
      "s3:GetObject",
      "s3:PutObject",
    ]
    resources = [
      "${aws_s3_bucket.creator_avatar_upload.arn}/*",
      "${aws_s3_bucket.creator_avatar_delivery.arn}/*",
    ]
  }
}

resource "aws_iam_policy" "creator_avatar_app_access" {
  name   = "${local.resource_prefix}-creator-avatar-app-access"
  policy = data.aws_iam_policy_document.creator_avatar_app_access.json
}
