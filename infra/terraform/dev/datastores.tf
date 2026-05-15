resource "random_password" "db" {
  length  = 32
  special = false
}

resource "aws_db_subnet_group" "dev" {
  name       = "${local.resource_prefix}-db"
  subnet_ids = local.dev_private_subnet_ids
}

resource "aws_db_instance" "dev" {
  identifier             = "${local.resource_prefix}-postgres"
  allocated_storage      = 20
  max_allocated_storage  = 100
  db_name                = local.db_name
  engine                 = "postgres"
  engine_version         = "16"
  instance_class         = var.db_instance_class
  username               = local.db_username
  password               = random_password.db.result
  db_subnet_group_name   = aws_db_subnet_group.dev.name
  vpc_security_group_ids = [aws_security_group.rds.id]
  port                   = local.db_port
  publicly_accessible    = false
  skip_final_snapshot    = true
  deletion_protection    = false
  storage_encrypted      = true
}

resource "aws_elasticache_subnet_group" "dev" {
  name       = "${local.resource_prefix}-redis"
  subnet_ids = local.dev_private_subnet_ids
}

resource "aws_elasticache_cluster" "dev" {
  cluster_id           = "${local.resource_prefix}-redis"
  engine               = "redis"
  node_type            = var.redis_node_type
  num_cache_nodes      = 1
  parameter_group_name = "default.redis7"
  port                 = local.redis_port
  subnet_group_name    = aws_elasticache_subnet_group.dev.name
  security_group_ids   = [aws_security_group.redis.id]
}
