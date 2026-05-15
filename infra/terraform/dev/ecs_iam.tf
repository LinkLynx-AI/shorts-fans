data "aws_iam_policy_document" "ecs_tasks_assume_role" {
  statement {
    effect = "Allow"
    actions = [
      "sts:AssumeRole",
    ]

    principals {
      type        = "Service"
      identifiers = ["ecs-tasks.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "ecs_task_execution" {
  name               = "${local.resource_prefix}-ecs-task-execution"
  assume_role_policy = data.aws_iam_policy_document.ecs_tasks_assume_role.json
}

resource "aws_iam_role_policy_attachment" "ecs_task_execution_managed" {
  role       = aws_iam_role.ecs_task_execution.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

data "aws_iam_policy_document" "ecs_task_execution_secrets" {
  statement {
    sid    = "ReadRuntimeSecrets"
    effect = "Allow"
    actions = [
      "secretsmanager:GetSecretValue",
    ]
    resources = [
      aws_secretsmanager_secret.postgres_dsn.arn,
      aws_secretsmanager_secret.admin_api_token.arn,
    ]
  }
}

resource "aws_iam_role_policy" "ecs_task_execution_secrets" {
  name   = "${local.resource_prefix}-ecs-task-execution-secrets"
  role   = aws_iam_role.ecs_task_execution.id
  policy = data.aws_iam_policy_document.ecs_task_execution_secrets.json
}

resource "aws_iam_role" "app_task" {
  name               = "${local.resource_prefix}-app-task"
  assume_role_policy = data.aws_iam_policy_document.ecs_tasks_assume_role.json
}

resource "aws_iam_role_policy_attachment" "app_task_media" {
  role       = aws_iam_role.app_task.name
  policy_arn = aws_iam_policy.media_app_access.arn
}

resource "aws_iam_role_policy_attachment" "app_task_creator_avatar" {
  role       = aws_iam_role.app_task.name
  policy_arn = aws_iam_policy.creator_avatar_app_access.arn
}

resource "aws_iam_role_policy_attachment" "app_task_creator_review_evidence" {
  role       = aws_iam_role.app_task.name
  policy_arn = aws_iam_policy.creator_review_evidence_app_access.arn
}

data "aws_iam_policy_document" "app_task_cognito" {
  statement {
    sid    = "UseFanAuthPublicFlows"
    effect = "Allow"
    actions = [
      "cognito-idp:ConfirmForgotPassword",
      "cognito-idp:ConfirmSignUp",
      "cognito-idp:ForgotPassword",
      "cognito-idp:InitiateAuth",
      "cognito-idp:ResendConfirmationCode",
      "cognito-idp:SignUp",
    ]
    resources = [
      aws_cognito_user_pool.fan_auth.arn,
    ]
  }
}

resource "aws_iam_role_policy" "app_task_cognito" {
  name   = "${local.resource_prefix}-app-task-cognito"
  role   = aws_iam_role.app_task.id
  policy = data.aws_iam_policy_document.app_task_cognito.json
}
