resource "aws_sesv2_email_identity" "cognito_sender" {
  count = trimspace(var.cognito_email_from_address) == "" ? 0 : 1

  email_identity = trimspace(var.cognito_email_from_address)
}

data "aws_sesv2_email_identity" "cognito_sender_verified" {
  count = var.cognito_use_ses_developer_email ? 1 : 0

  email_identity = trimspace(var.cognito_email_from_address)
}
