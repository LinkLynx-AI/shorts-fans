data "aws_availability_zones" "available" {
  state = "available"
}

locals {
  dev_azs                = slice(data.aws_availability_zones.available.names, 0, 2)
  dev_vpc_cidr           = "10.42.0.0/16"
  dev_public_subnets     = [for index, az in local.dev_azs : { az = az, cidr = cidrsubnet(local.dev_vpc_cidr, 8, index) }]
  dev_private_subnets    = [for index, az in local.dev_azs : { az = az, cidr = cidrsubnet(local.dev_vpc_cidr, 8, index + 10) }]
  dev_public_subnet_ids  = [for subnet in aws_subnet.public : subnet.id]
  dev_private_subnet_ids = [for subnet in aws_subnet.private : subnet.id]
}

resource "aws_vpc" "dev" {
  cidr_block           = local.dev_vpc_cidr
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name = "${local.resource_prefix}-vpc"
  }
}

resource "aws_internet_gateway" "dev" {
  vpc_id = aws_vpc.dev.id

  tags = {
    Name = "${local.resource_prefix}-igw"
  }
}

resource "aws_subnet" "public" {
  for_each = { for index, subnet in local.dev_public_subnets : tostring(index) => subnet }

  vpc_id                  = aws_vpc.dev.id
  availability_zone       = each.value.az
  cidr_block              = each.value.cidr
  map_public_ip_on_launch = true

  tags = {
    Name = "${local.resource_prefix}-public-${each.key}"
    Tier = "public"
  }
}

resource "aws_subnet" "private" {
  for_each = { for index, subnet in local.dev_private_subnets : tostring(index) => subnet }

  vpc_id            = aws_vpc.dev.id
  availability_zone = each.value.az
  cidr_block        = each.value.cidr

  tags = {
    Name = "${local.resource_prefix}-private-${each.key}"
    Tier = "private"
  }
}

resource "aws_route_table" "public" {
  vpc_id = aws_vpc.dev.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.dev.id
  }

  tags = {
    Name = "${local.resource_prefix}-public"
  }
}

resource "aws_route_table_association" "public" {
  for_each = aws_subnet.public

  subnet_id      = each.value.id
  route_table_id = aws_route_table.public.id
}

resource "aws_route_table" "private" {
  vpc_id = aws_vpc.dev.id

  tags = {
    Name = "${local.resource_prefix}-private"
  }
}

resource "aws_route_table_association" "private" {
  for_each = aws_subnet.private

  subnet_id      = each.value.id
  route_table_id = aws_route_table.private.id
}
