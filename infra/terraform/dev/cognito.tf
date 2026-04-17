resource "aws_cognito_user_pool" "fan_auth" {
  name                     = local.cognito_user_pool_name
  username_attributes      = ["email"]
  auto_verified_attributes = ["email"]
  deletion_protection      = "INACTIVE"
  mfa_configuration        = "OFF"

  admin_create_user_config {
    allow_admin_create_user_only = false
  }

  username_configuration {
    case_sensitive = false
  }

  account_recovery_setting {
    recovery_mechanism {
      name     = "verified_email"
      priority = 1
    }
  }

  email_configuration {
    email_sending_account = var.cognito_use_ses_developer_email ? "DEVELOPER" : "COGNITO_DEFAULT"
    from_email_address    = var.cognito_use_ses_developer_email ? local.cognito_email_from_formatted : null
    source_arn            = var.cognito_use_ses_developer_email ? data.aws_sesv2_email_identity.cognito_sender_verified[0].arn : null
  }

  verification_message_template {
    default_email_option = "CONFIRM_WITH_CODE"
  }

  lifecycle {
    precondition {
      condition = (
        !var.cognito_use_ses_developer_email ||
        (
          data.aws_sesv2_email_identity.cognito_sender_verified[0].verification_status == "SUCCESS" &&
          data.aws_sesv2_email_identity.cognito_sender_verified[0].verified_for_sending_status
        )
      )
      error_message = "cognito_use_ses_developer_email requires an existing SES sender identity with VerificationStatus=SUCCESS. First apply with cognito_use_ses_developer_email = false, complete SES verification for cognito_email_from_address, confirm the verified status, and then re-apply with cognito_use_ses_developer_email = true."
    }
  }
}

resource "aws_cognito_user_pool_client" "fan_auth" {
  name                          = local.cognito_user_pool_client_name
  user_pool_id                  = aws_cognito_user_pool.fan_auth.id
  explicit_auth_flows           = ["ALLOW_USER_PASSWORD_AUTH"]
  generate_secret               = false
  prevent_user_existence_errors = "ENABLED"
  supported_identity_providers  = ["COGNITO"]
}
