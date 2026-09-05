# Root Terraform da demo (flat, sem modules/).
# Fase B: VPC pública 2 AZ + IGW (sem NAT) + SG base + SSM.
# Fase C: RDS, ECR, ECS, ALB, S3, CloudFront (arquivos dedicados).
#
# Budget/alerta da conta: stack SEPARADA em infra/budget/
#   cd infra/budget && terraform init && terraform apply
# terraform destroy NESTE diretório não toca o budget (FR-022).

data "aws_availability_zones" "available" {
  state = "available"
}

data "aws_caller_identity" "current" {}

locals {
  azs        = slice(data.aws_availability_zones.available.names, 0, 2)
  ssm_prefix = "/${var.project_name}"

  jwt_secret     = var.jwt_secret != null ? var.jwt_secret : random_password.jwt_secret.result
  admin_password = var.admin_password != null ? var.admin_password : random_password.admin_password.result
  db_password    = var.db_password != null ? var.db_password : random_password.db_password.result

  # sslmode=require alinhado ao contrato api-env. Host = RDS após T020.
  database_url = var.database_url != null ? var.database_url : format(
    "postgres://%s:%s@%s:5432/%s?sslmode=require",
    var.db_username,
    urlencode(local.db_password),
    aws_db_instance.main.address,
    var.db_name,
  )

  # Default do variable é placeholder até o CloudFront do front existir.
  cors_origin = var.cors_origin != "https://pending-frontend.invalid" ? var.cors_origin : "https://${aws_cloudfront_distribution.frontend.domain_name}"
}

resource "random_password" "jwt_secret" {
  length  = 48
  special = false
}

resource "random_password" "admin_password" {
  length  = 24
  special = false
}

resource "random_password" "db_password" {
  length  = 24
  special = false
}
