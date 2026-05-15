resource "aws_service_discovery_private_dns_namespace" "dev" {
  name = "${local.resource_prefix}.local"
  vpc  = aws_vpc.dev.id
}

resource "aws_service_discovery_service" "backend_api" {
  name = "backend-api"

  dns_config {
    namespace_id = aws_service_discovery_private_dns_namespace.dev.id

    dns_records {
      ttl  = 10
      type = "A"
    }

    routing_policy = "MULTIVALUE"
  }
}
