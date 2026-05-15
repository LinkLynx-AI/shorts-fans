resource "aws_ecs_cluster" "dev" {
  name = local.ecs_cluster_name
}

resource "aws_cloudwatch_log_group" "backend" {
  name              = local.backend_log_group_name
  retention_in_days = 14
}

resource "aws_cloudwatch_log_group" "frontend" {
  name              = local.frontend_log_group_name
  retention_in_days = 14
}

resource "aws_cloudwatch_log_group" "worker" {
  name              = local.worker_log_group_name
  retention_in_days = 14
}

resource "aws_cloudwatch_log_group" "migration" {
  name              = local.migration_log_group_name
  retention_in_days = 14
}

locals {
  backend_image                 = "${aws_ecr_repository.backend.repository_url}:${var.backend_image_tag}"
  frontend_image                = "${aws_ecr_repository.frontend.repository_url}:${var.frontend_image_tag}"
  backend_api_internal_base_url = "http://${aws_service_discovery_service.backend_api.name}.${aws_service_discovery_private_dns_namespace.dev.name}:8080"

  backend_environment = [
    { name = "APP_ENV", value = "development" },
    { name = "API_ADDR", value = ":8080" },
    { name = "REDIS_ADDR", value = "${aws_elasticache_cluster.dev.cache_nodes[0].address}:${local.redis_port}" },
    { name = "AWS_REGION", value = var.aws_region },
    { name = "COGNITO_USER_POOL_ID", value = aws_cognito_user_pool.fan_auth.id },
    { name = "COGNITO_USER_POOL_CLIENT_ID", value = aws_cognito_user_pool_client.fan_auth.id },
    { name = "MEDIA_JOBS_QUEUE_URL", value = aws_sqs_queue.media_jobs.url },
    { name = "MEDIA_RAW_BUCKET_NAME", value = aws_s3_bucket.raw.bucket },
    { name = "MEDIA_SHORT_PUBLIC_BUCKET_NAME", value = aws_s3_bucket.short_public.bucket },
    { name = "MEDIA_SHORT_PUBLIC_BASE_URL", value = "https://${aws_cloudfront_distribution.short_public.domain_name}" },
    { name = "MEDIA_MAIN_PRIVATE_BUCKET_NAME", value = aws_s3_bucket.main_private.bucket },
    { name = "MEDIACONVERT_SERVICE_ROLE_ARN", value = aws_iam_role.mediaconvert_service.arn },
    { name = "CREATOR_AVATAR_UPLOAD_BUCKET_NAME", value = aws_s3_bucket.creator_avatar_upload.bucket },
    { name = "CREATOR_AVATAR_DELIVERY_BUCKET_NAME", value = aws_s3_bucket.creator_avatar_delivery.bucket },
    { name = "CREATOR_AVATAR_BASE_URL", value = "https://${aws_cloudfront_distribution.creator_avatar.domain_name}" },
    { name = "CREATOR_REVIEW_EVIDENCE_BUCKET_NAME", value = aws_s3_bucket.creator_review_evidence.bucket },
  ]

  backend_secrets = [
    { name = "POSTGRES_DSN", valueFrom = aws_secretsmanager_secret.postgres_dsn.arn },
    { name = "ADMIN_API_TOKEN", valueFrom = aws_secretsmanager_secret.admin_api_token.arn },
  ]
}

resource "aws_ecs_task_definition" "backend_api" {
  family                   = "${local.resource_prefix}-backend-api"
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"
  cpu                      = var.backend_ecs_cpu
  memory                   = var.backend_ecs_memory
  execution_role_arn       = aws_iam_role.ecs_task_execution.arn
  task_role_arn            = aws_iam_role.app_task.arn

  runtime_platform {
    cpu_architecture        = "X86_64"
    operating_system_family = "LINUX"
  }

  container_definitions = jsonencode([
    {
      name      = "backend-api"
      image     = local.backend_image
      essential = true
      command   = ["/app/api"]
      portMappings = [
        {
          containerPort = 8080
          hostPort      = 8080
          protocol      = "tcp"
        },
      ]
      environment = local.backend_environment
      secrets     = local.backend_secrets
      healthCheck = {
        command     = ["CMD-SHELL", "curl -fsS http://127.0.0.1:8080/readyz || exit 1"]
        interval    = 30
        retries     = 3
        startPeriod = 60
        timeout     = 5
      }
      logConfiguration = {
        logDriver = "awslogs"
        options = {
          awslogs-group         = aws_cloudwatch_log_group.backend.name
          awslogs-region        = var.aws_region
          awslogs-stream-prefix = "api"
        }
      }
    },
  ])
}

resource "aws_ecs_task_definition" "backend_worker" {
  family                   = "${local.resource_prefix}-backend-worker"
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"
  cpu                      = var.backend_ecs_cpu
  memory                   = var.backend_ecs_memory
  execution_role_arn       = aws_iam_role.ecs_task_execution.arn
  task_role_arn            = aws_iam_role.app_task.arn

  runtime_platform {
    cpu_architecture        = "X86_64"
    operating_system_family = "LINUX"
  }

  container_definitions = jsonencode([
    {
      name        = "backend-worker"
      image       = local.backend_image
      essential   = true
      command     = ["/app/worker"]
      environment = local.backend_environment
      secrets     = local.backend_secrets
      logConfiguration = {
        logDriver = "awslogs"
        options = {
          awslogs-group         = aws_cloudwatch_log_group.worker.name
          awslogs-region        = var.aws_region
          awslogs-stream-prefix = "worker"
        }
      }
    },
  ])
}

resource "aws_ecs_task_definition" "backend_maintenance" {
  family                   = "${local.resource_prefix}-backend-maintenance"
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"
  cpu                      = var.backend_ecs_cpu
  memory                   = var.backend_ecs_memory
  execution_role_arn       = aws_iam_role.ecs_task_execution.arn
  task_role_arn            = aws_iam_role.app_task.arn

  runtime_platform {
    cpu_architecture        = "X86_64"
    operating_system_family = "LINUX"
  }

  container_definitions = jsonencode([
    {
      name        = "backend-maintenance"
      image       = local.backend_image
      essential   = true
      command     = ["/app/migrate", "version"]
      environment = local.backend_environment
      secrets     = local.backend_secrets
      logConfiguration = {
        logDriver = "awslogs"
        options = {
          awslogs-group         = aws_cloudwatch_log_group.migration.name
          awslogs-region        = var.aws_region
          awslogs-stream-prefix = "maintenance"
        }
      }
    },
  ])
}

resource "aws_ecs_task_definition" "frontend" {
  family                   = "${local.resource_prefix}-frontend"
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"
  cpu                      = var.frontend_ecs_cpu
  memory                   = var.frontend_ecs_memory
  execution_role_arn       = aws_iam_role.ecs_task_execution.arn

  runtime_platform {
    cpu_architecture        = "X86_64"
    operating_system_family = "LINUX"
  }

  container_definitions = jsonencode([
    {
      name      = "frontend"
      image     = local.frontend_image
      essential = true
      healthCheck = {
        command     = ["CMD-SHELL", "curl -fsS http://127.0.0.1:3000/icon.svg || exit 1"]
        interval    = 30
        retries     = 3
        startPeriod = 30
        timeout     = 5
      }
      portMappings = [
        {
          containerPort = 3000
          hostPort      = 3000
          protocol      = "tcp"
        },
      ]
      environment = [
        { name = "NEXT_PUBLIC_API_BASE_URL", value = "__same_origin__" },
        { name = "API_BASE_URL_INTERNAL", value = local.backend_api_internal_base_url },
      ]
      logConfiguration = {
        logDriver = "awslogs"
        options = {
          awslogs-group         = aws_cloudwatch_log_group.frontend.name
          awslogs-region        = var.aws_region
          awslogs-stream-prefix = "frontend"
        }
      }
    },
  ])
}

resource "aws_ecs_service" "backend_api" {
  name            = "${local.resource_prefix}-backend-api"
  cluster         = aws_ecs_cluster.dev.id
  task_definition = aws_ecs_task_definition.backend_api.arn
  desired_count   = var.backend_desired_count
  launch_type     = "FARGATE"

  network_configuration {
    assign_public_ip = true
    security_groups  = [aws_security_group.backend_ecs_tasks.id]
    subnets          = local.dev_public_subnet_ids
  }

  load_balancer {
    target_group_arn = aws_lb_target_group.backend.arn
    container_name   = "backend-api"
    container_port   = 8080
  }

  service_registries {
    registry_arn = aws_service_discovery_service.backend_api.arn
  }

  depends_on = [aws_lb_listener.http]
}

resource "aws_ecs_service" "backend_worker" {
  name            = "${local.resource_prefix}-backend-worker"
  cluster         = aws_ecs_cluster.dev.id
  task_definition = aws_ecs_task_definition.backend_worker.arn
  desired_count   = var.worker_desired_count
  launch_type     = "FARGATE"

  network_configuration {
    assign_public_ip = true
    security_groups  = [aws_security_group.backend_ecs_tasks.id]
    subnets          = local.dev_public_subnet_ids
  }
}

resource "aws_ecs_service" "frontend" {
  name            = "${local.resource_prefix}-frontend"
  cluster         = aws_ecs_cluster.dev.id
  task_definition = aws_ecs_task_definition.frontend.arn
  desired_count   = var.frontend_desired_count
  launch_type     = "FARGATE"

  network_configuration {
    assign_public_ip = true
    security_groups  = [aws_security_group.frontend_ecs_tasks.id]
    subnets          = local.dev_public_subnet_ids
  }

  load_balancer {
    target_group_arn = aws_lb_target_group.frontend.arn
    container_name   = "frontend"
    container_port   = 3000
  }

  depends_on = [aws_lb_listener.http]
}
