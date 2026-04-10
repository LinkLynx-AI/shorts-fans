locals {
  creator_avatar_origin_id = "${local.resource_prefix}-creator-avatar-origin"
  short_public_origin_id   = "${local.resource_prefix}-short-public-origin"
}

resource "aws_cloudfront_origin_access_control" "creator_avatar" {
  name                              = "${local.resource_prefix}-creator-avatar"
  description                       = "Origin access control for dev creator avatar delivery"
  origin_access_control_origin_type = "s3"
  signing_behavior                  = "always"
  signing_protocol                  = "sigv4"
}

resource "aws_cloudfront_distribution" "creator_avatar" {
  enabled         = true
  is_ipv6_enabled = true
  comment         = "Dev creator avatar delivery distribution"
  price_class     = "PriceClass_100"

  origin {
    domain_name              = aws_s3_bucket.creator_avatar_delivery.bucket_regional_domain_name
    origin_access_control_id = aws_cloudfront_origin_access_control.creator_avatar.id
    origin_id                = local.creator_avatar_origin_id
  }

  default_cache_behavior {
    allowed_methods        = ["GET", "HEAD"]
    cached_methods         = ["GET", "HEAD"]
    compress               = true
    target_origin_id       = local.creator_avatar_origin_id
    viewer_protocol_policy = "redirect-to-https"

    forwarded_values {
      query_string = false

      cookies {
        forward = "none"
      }
    }

    min_ttl     = 0
    default_ttl = 300
    max_ttl     = 3600
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }
}

resource "aws_cloudfront_origin_access_control" "short_public" {
  name                              = "${local.resource_prefix}-short-public"
  description                       = "Origin access control for dev short public delivery"
  origin_access_control_origin_type = "s3"
  signing_behavior                  = "always"
  signing_protocol                  = "sigv4"
}

resource "aws_cloudfront_distribution" "short_public" {
  enabled         = true
  is_ipv6_enabled = true
  comment         = "Dev public short delivery distribution"
  price_class     = "PriceClass_100"

  origin {
    domain_name              = aws_s3_bucket.short_public.bucket_regional_domain_name
    origin_access_control_id = aws_cloudfront_origin_access_control.short_public.id
    origin_id                = local.short_public_origin_id
  }

  default_cache_behavior {
    allowed_methods        = ["GET", "HEAD"]
    cached_methods         = ["GET", "HEAD"]
    compress               = true
    target_origin_id       = local.short_public_origin_id
    viewer_protocol_policy = "redirect-to-https"

    forwarded_values {
      query_string = false

      cookies {
        forward = "none"
      }
    }

    min_ttl     = 0
    default_ttl = 300
    max_ttl     = 3600
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }
}
