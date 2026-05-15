resource "aws_security_group" "alb" {
  name        = "${local.resource_prefix}-alb"
  description = "Public ALB access for the dev app"
  vpc_id      = aws_vpc.dev.id

  ingress {
    description = "HTTP from allowed dev clients"
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = var.dev_allowed_cidr_blocks
  }

  egress {
    description = "ALB to ECS tasks"
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_security_group" "backend_ecs_tasks" {
  name        = "${local.resource_prefix}-backend-ecs-tasks"
  description = "Backend API, worker, and maintenance ECS task access"
  vpc_id      = aws_vpc.dev.id

  ingress {
    description     = "Backend API from ALB"
    from_port       = 8080
    to_port         = 8080
    protocol        = "tcp"
    security_groups = [aws_security_group.alb.id]
  }

  ingress {
    description     = "Backend API from frontend server-side rendering"
    from_port       = 8080
    to_port         = 8080
    protocol        = "tcp"
    security_groups = [aws_security_group.frontend_ecs_tasks.id]
  }

  egress {
    description = "Outbound to AWS APIs and dependencies"
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_security_group" "frontend_ecs_tasks" {
  name        = "${local.resource_prefix}-frontend-ecs-tasks"
  description = "Frontend ECS task access"
  vpc_id      = aws_vpc.dev.id

  ingress {
    description     = "Frontend from ALB"
    from_port       = 3000
    to_port         = 3000
    protocol        = "tcp"
    security_groups = [aws_security_group.alb.id]
  }

  egress {
    description = "Outbound to backend API and AWS services"
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_security_group" "rds" {
  name        = "${local.resource_prefix}-rds"
  description = "PostgreSQL access from backend dev ECS tasks"
  vpc_id      = aws_vpc.dev.id

  ingress {
    description     = "PostgreSQL from backend ECS tasks"
    from_port       = local.db_port
    to_port         = local.db_port
    protocol        = "tcp"
    security_groups = [aws_security_group.backend_ecs_tasks.id]
  }
}

resource "aws_security_group" "redis" {
  name        = "${local.resource_prefix}-redis"
  description = "Redis access from backend dev ECS tasks"
  vpc_id      = aws_vpc.dev.id

  ingress {
    description     = "Redis from backend ECS tasks"
    from_port       = local.redis_port
    to_port         = local.redis_port
    protocol        = "tcp"
    security_groups = [aws_security_group.backend_ecs_tasks.id]
  }
}
