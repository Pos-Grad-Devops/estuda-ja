output "aws_region" {
  description = "Região da sessão (sempre us-east-1 nesta feature)."
  value       = var.aws_region
}

output "vpc_id" {
  description = "ID da VPC custom."
  value       = aws_vpc.main.id
}

output "public_subnet_ids" {
  description = "Subnets públicas (2 AZs) para ALB/ECS/RDS."
  value       = aws_subnet.public[*].id
}

output "alb_security_group_id" {
  description = "SG do ALB."
  value       = aws_security_group.alb.id
}

output "ecs_tasks_security_group_id" {
  description = "SG das tasks ECS."
  value       = aws_security_group.ecs_tasks.id
}

output "rds_security_group_id" {
  description = "SG do RDS (Postgres só a partir das tasks)."
  value       = aws_security_group.rds.id
}

output "ssm_prefix" {
  description = "Prefixo dos parâmetros SSM do projeto."
  value       = local.ssm_prefix
}

output "ssm_parameter_names" {
  description = "Nomes (não valores) dos parâmetros SSM."
  value = {
    jwt_secret     = aws_ssm_parameter.jwt_secret.name
    admin_password = aws_ssm_parameter.admin_password.name
    admin_email    = aws_ssm_parameter.admin_email.name
    cors_origin    = aws_ssm_parameter.cors_origin.name
    db_password    = aws_ssm_parameter.db_password.name
    database_url   = aws_ssm_parameter.database_url.name
  }
}

# --- Contrato contracts/terraform-outputs.md ---

output "frontend_url" {
  description = "URL HTTPS do CloudFront do frontend (SPA)."
  value       = "https://${aws_cloudfront_distribution.frontend.domain_name}"
}

output "api_url" {
  description = "URL HTTPS do CloudFront da API (origin ALB). Valor de VITE_API_URL."
  value       = "https://${aws_cloudfront_distribution.api.domain_name}"
}

output "alb_dns_name" {
  description = "DNS do ALB (diagnóstico; não usar no browser — mixed content)."
  value       = aws_lb.api.dns_name
}

output "ecr_repository_url" {
  description = "URI do repositório ECR para docker push da API."
  value       = aws_ecr_repository.api.repository_url
}

output "ecs_cluster_name" {
  description = "Nome do cluster ECS (update-service)."
  value       = aws_ecs_cluster.main.name
}

output "ecs_service_name" {
  description = "Nome do service ECS (force new deployment)."
  value       = aws_ecs_service.api.name
}

output "s3_bucket_name" {
  description = "Bucket S3 para aws s3 sync do frontend/dist."
  value       = aws_s3_bucket.frontend.id
}

output "cloudfront_frontend_distribution_id" {
  description = "ID da distribution CloudFront do frontend (invalidação)."
  value       = aws_cloudfront_distribution.frontend.id
}

output "rds_endpoint" {
  description = "Hostname do RDS (conexão via DATABASE_URL na task)."
  value       = aws_db_instance.main.address
}
