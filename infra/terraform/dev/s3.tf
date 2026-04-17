resource "aws_s3_bucket" "raw" {
  bucket = local.raw_bucket_name
}

resource "aws_s3_bucket_ownership_controls" "raw" {
  bucket = aws_s3_bucket.raw.id

  rule {
    object_ownership = "BucketOwnerEnforced"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "raw" {
  bucket = aws_s3_bucket.raw.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

resource "aws_s3_bucket_public_access_block" "raw" {
  bucket = aws_s3_bucket.raw.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

data "aws_iam_policy_document" "raw_access" {
  statement {
    sid    = "DenyInsecureTransport"
    effect = "Deny"

    principals {
      type        = "*"
      identifiers = ["*"]
    }

    actions = ["s3:*"]
    resources = [
      aws_s3_bucket.raw.arn,
      "${aws_s3_bucket.raw.arn}/*",
    ]

    condition {
      test     = "Bool"
      variable = "aws:SecureTransport"
      values   = ["false"]
    }
  }
}

resource "aws_s3_bucket_policy" "raw" {
  bucket = aws_s3_bucket.raw.id
  policy = data.aws_iam_policy_document.raw_access.json

  depends_on = [
    aws_s3_bucket_ownership_controls.raw,
    aws_s3_bucket_public_access_block.raw,
  ]
}

resource "aws_s3_bucket_cors_configuration" "raw" {
  bucket = aws_s3_bucket.raw.id

  cors_rule {
    allowed_headers = ["*"]
    allowed_methods = ["PUT"]
    allowed_origins = var.allowed_app_origins
    expose_headers  = ["ETag"]
    max_age_seconds = 300
  }
}

resource "aws_s3_bucket_lifecycle_configuration" "raw" {
  bucket = aws_s3_bucket.raw.id

  rule {
    id     = "expire-raw-after-14-days"
    status = "Enabled"

    filter {}

    expiration {
      days = 14
    }

    abort_incomplete_multipart_upload {
      days_after_initiation = 1
    }
  }
}

resource "aws_s3_bucket" "creator_avatar_upload" {
  bucket = local.creator_avatar_upload_bucket_name
}

resource "aws_s3_bucket_ownership_controls" "creator_avatar_upload" {
  bucket = aws_s3_bucket.creator_avatar_upload.id

  rule {
    object_ownership = "BucketOwnerEnforced"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "creator_avatar_upload" {
  bucket = aws_s3_bucket.creator_avatar_upload.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

resource "aws_s3_bucket_public_access_block" "creator_avatar_upload" {
  bucket = aws_s3_bucket.creator_avatar_upload.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

data "aws_iam_policy_document" "creator_avatar_upload_access" {
  statement {
    sid    = "DenyInsecureTransport"
    effect = "Deny"

    principals {
      type        = "*"
      identifiers = ["*"]
    }

    actions = ["s3:*"]
    resources = [
      aws_s3_bucket.creator_avatar_upload.arn,
      "${aws_s3_bucket.creator_avatar_upload.arn}/*",
    ]

    condition {
      test     = "Bool"
      variable = "aws:SecureTransport"
      values   = ["false"]
    }
  }
}

resource "aws_s3_bucket_policy" "creator_avatar_upload" {
  bucket = aws_s3_bucket.creator_avatar_upload.id
  policy = data.aws_iam_policy_document.creator_avatar_upload_access.json

  depends_on = [
    aws_s3_bucket_ownership_controls.creator_avatar_upload,
    aws_s3_bucket_public_access_block.creator_avatar_upload,
  ]
}

resource "aws_s3_bucket_cors_configuration" "creator_avatar_upload" {
  bucket = aws_s3_bucket.creator_avatar_upload.id

  cors_rule {
    allowed_headers = ["*"]
    allowed_methods = ["HEAD", "PUT"]
    allowed_origins = var.allowed_app_origins
    expose_headers  = ["ETag"]
    max_age_seconds = 300
  }
}

resource "aws_s3_bucket_lifecycle_configuration" "creator_avatar_upload" {
  bucket = aws_s3_bucket.creator_avatar_upload.id

  rule {
    id     = "expire-avatar-upload-after-1-day"
    status = "Enabled"

    filter {}

    expiration {
      days = 1
    }

    abort_incomplete_multipart_upload {
      days_after_initiation = 1
    }
  }
}

resource "aws_s3_bucket" "creator_avatar_delivery" {
  bucket = local.creator_avatar_delivery_bucket_name
}

resource "aws_s3_bucket_ownership_controls" "creator_avatar_delivery" {
  bucket = aws_s3_bucket.creator_avatar_delivery.id

  rule {
    object_ownership = "BucketOwnerEnforced"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "creator_avatar_delivery" {
  bucket = aws_s3_bucket.creator_avatar_delivery.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

resource "aws_s3_bucket_public_access_block" "creator_avatar_delivery" {
  bucket = aws_s3_bucket.creator_avatar_delivery.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

data "aws_iam_policy_document" "creator_avatar_delivery_access" {
  statement {
    sid    = "AllowCloudFrontReadOfCreatorAvatarObjects"
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["cloudfront.amazonaws.com"]
    }

    actions = ["s3:GetObject"]
    resources = [
      "${aws_s3_bucket.creator_avatar_delivery.arn}/*",
    ]

    condition {
      test     = "StringEquals"
      variable = "AWS:SourceArn"
      values   = [aws_cloudfront_distribution.creator_avatar.arn]
    }
  }

  statement {
    sid    = "DenyInsecureTransport"
    effect = "Deny"

    principals {
      type        = "*"
      identifiers = ["*"]
    }

    actions = ["s3:*"]
    resources = [
      aws_s3_bucket.creator_avatar_delivery.arn,
      "${aws_s3_bucket.creator_avatar_delivery.arn}/*",
    ]

    condition {
      test     = "Bool"
      variable = "aws:SecureTransport"
      values   = ["false"]
    }
  }
}

resource "aws_s3_bucket_policy" "creator_avatar_delivery" {
  bucket = aws_s3_bucket.creator_avatar_delivery.id
  policy = data.aws_iam_policy_document.creator_avatar_delivery_access.json

  depends_on = [
    aws_s3_bucket_ownership_controls.creator_avatar_delivery,
    aws_s3_bucket_public_access_block.creator_avatar_delivery,
  ]
}

resource "aws_s3_bucket_lifecycle_configuration" "creator_avatar_delivery" {
  bucket = aws_s3_bucket.creator_avatar_delivery.id

  rule {
    id     = "abort-incomplete-multipart-uploads"
    status = "Enabled"

    filter {}

    abort_incomplete_multipart_upload {
      days_after_initiation = 1
    }
  }
}

resource "aws_s3_bucket" "creator_review_evidence" {
  bucket = local.creator_review_evidence_bucket_name
}

resource "aws_s3_bucket_ownership_controls" "creator_review_evidence" {
  bucket = aws_s3_bucket.creator_review_evidence.id

  rule {
    object_ownership = "BucketOwnerEnforced"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "creator_review_evidence" {
  bucket = aws_s3_bucket.creator_review_evidence.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

resource "aws_s3_bucket_public_access_block" "creator_review_evidence" {
  bucket = aws_s3_bucket.creator_review_evidence.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

data "aws_iam_policy_document" "creator_review_evidence_access" {
  statement {
    sid    = "DenyInsecureTransport"
    effect = "Deny"

    principals {
      type        = "*"
      identifiers = ["*"]
    }

    actions = ["s3:*"]
    resources = [
      aws_s3_bucket.creator_review_evidence.arn,
      "${aws_s3_bucket.creator_review_evidence.arn}/*",
    ]

    condition {
      test     = "Bool"
      variable = "aws:SecureTransport"
      values   = ["false"]
    }
  }
}

resource "aws_s3_bucket_policy" "creator_review_evidence" {
  bucket = aws_s3_bucket.creator_review_evidence.id
  policy = data.aws_iam_policy_document.creator_review_evidence_access.json

  depends_on = [
    aws_s3_bucket_ownership_controls.creator_review_evidence,
    aws_s3_bucket_public_access_block.creator_review_evidence,
  ]
}

resource "aws_s3_bucket_cors_configuration" "creator_review_evidence" {
  bucket = aws_s3_bucket.creator_review_evidence.id

  cors_rule {
    allowed_headers = ["*"]
    allowed_methods = ["HEAD", "PUT"]
    allowed_origins = var.allowed_app_origins
    expose_headers  = ["ETag"]
    max_age_seconds = 300
  }
}

resource "aws_s3_bucket_lifecycle_configuration" "creator_review_evidence" {
  bucket = aws_s3_bucket.creator_review_evidence.id

  rule {
    id     = "expire-pending-review-evidence-after-1-day"
    status = "Enabled"

    filter {
      prefix = "creator-registration-evidence/"
    }

    expiration {
      days = 1
    }

    abort_incomplete_multipart_upload {
      days_after_initiation = 1
    }
  }
}

resource "aws_s3_bucket" "short_public" {
  bucket = local.short_public_bucket_name
}

resource "aws_s3_bucket_ownership_controls" "short_public" {
  bucket = aws_s3_bucket.short_public.id

  rule {
    object_ownership = "BucketOwnerEnforced"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "short_public" {
  bucket = aws_s3_bucket.short_public.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

resource "aws_s3_bucket_public_access_block" "short_public" {
  bucket = aws_s3_bucket.short_public.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

data "aws_iam_policy_document" "short_public_access" {
  statement {
    sid    = "AllowCloudFrontReadOfShortObjects"
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["cloudfront.amazonaws.com"]
    }

    actions = ["s3:GetObject"]
    resources = [
      "${aws_s3_bucket.short_public.arn}/*",
    ]

    condition {
      test     = "StringEquals"
      variable = "AWS:SourceArn"
      values   = [aws_cloudfront_distribution.short_public.arn]
    }
  }

  statement {
    sid    = "DenyInsecureTransport"
    effect = "Deny"

    principals {
      type        = "*"
      identifiers = ["*"]
    }

    actions = ["s3:*"]
    resources = [
      aws_s3_bucket.short_public.arn,
      "${aws_s3_bucket.short_public.arn}/*",
    ]

    condition {
      test     = "Bool"
      variable = "aws:SecureTransport"
      values   = ["false"]
    }
  }
}

resource "aws_s3_bucket_policy" "short_public" {
  bucket = aws_s3_bucket.short_public.id
  policy = data.aws_iam_policy_document.short_public_access.json
}

resource "aws_s3_bucket_lifecycle_configuration" "short_public" {
  bucket = aws_s3_bucket.short_public.id

  rule {
    id     = "abort-incomplete-multipart-uploads"
    status = "Enabled"

    filter {}

    abort_incomplete_multipart_upload {
      days_after_initiation = 1
    }
  }
}

resource "aws_s3_bucket" "main_private" {
  bucket = local.main_private_bucket_name
}

resource "aws_s3_bucket_ownership_controls" "main_private" {
  bucket = aws_s3_bucket.main_private.id

  rule {
    object_ownership = "BucketOwnerEnforced"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "main_private" {
  bucket = aws_s3_bucket.main_private.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

resource "aws_s3_bucket_public_access_block" "main_private" {
  bucket = aws_s3_bucket.main_private.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

data "aws_iam_policy_document" "main_private_access" {
  statement {
    sid    = "DenyInsecureTransport"
    effect = "Deny"

    principals {
      type        = "*"
      identifiers = ["*"]
    }

    actions = ["s3:*"]
    resources = [
      aws_s3_bucket.main_private.arn,
      "${aws_s3_bucket.main_private.arn}/*",
    ]

    condition {
      test     = "Bool"
      variable = "aws:SecureTransport"
      values   = ["false"]
    }
  }
}

resource "aws_s3_bucket_policy" "main_private" {
  bucket = aws_s3_bucket.main_private.id
  policy = data.aws_iam_policy_document.main_private_access.json

  depends_on = [
    aws_s3_bucket_ownership_controls.main_private,
    aws_s3_bucket_public_access_block.main_private,
  ]
}

resource "aws_s3_bucket_cors_configuration" "main_private" {
  bucket = aws_s3_bucket.main_private.id

  cors_rule {
    allowed_headers = ["*"]
    allowed_methods = ["GET", "HEAD"]
    allowed_origins = var.allowed_app_origins
    expose_headers  = ["ETag"]
    max_age_seconds = 300
  }
}

resource "aws_s3_bucket_lifecycle_configuration" "main_private" {
  bucket = aws_s3_bucket.main_private.id

  rule {
    id     = "abort-incomplete-multipart-uploads"
    status = "Enabled"

    filter {}

    abort_incomplete_multipart_upload {
      days_after_initiation = 1
    }
  }
}
