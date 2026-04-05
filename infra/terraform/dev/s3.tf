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
  block_public_policy     = false
  ignore_public_acls      = true
  restrict_public_buckets = false
}

data "aws_iam_policy_document" "short_public_read" {
  statement {
    sid    = "AllowAnonymousReadOfShortObjects"
    effect = "Allow"

    principals {
      type        = "*"
      identifiers = ["*"]
    }

    actions = ["s3:GetObject"]
    resources = [
      "${aws_s3_bucket.short_public.arn}/*",
    ]
  }
}

resource "aws_s3_bucket_policy" "short_public" {
  bucket = aws_s3_bucket.short_public.id
  policy = data.aws_iam_policy_document.short_public_read.json

  depends_on = [
    aws_s3_bucket_ownership_controls.short_public,
    aws_s3_bucket_public_access_block.short_public,
  ]
}

resource "aws_s3_bucket_cors_configuration" "short_public" {
  bucket = aws_s3_bucket.short_public.id

  cors_rule {
    allowed_headers = ["*"]
    allowed_methods = ["GET", "HEAD"]
    allowed_origins = var.allowed_app_origins
    expose_headers  = ["ETag"]
    max_age_seconds = 300
  }
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
