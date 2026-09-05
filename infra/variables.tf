variable "aws_region" {
  type        = string
  description = "Região AWS. Fixa em us-east-1 nesta feature (FR-015)."
  default     = "us-east-1"

  validation {
    condition     = var.aws_region == "us-east-1"
    error_message = "Esta feature exige aws_region = us-east-1 (FR-015)."
  }
}

variable "project_name" {
  type        = string
  description = "Prefixo de nomes de recursos e parâmetros SSM."
  default     = "estudaja"
}

variable "vpc_cidr" {
  type        = string
  description = "CIDR da VPC custom (2 subnets públicas derivadas)."
  default     = "10.0.0.0/16"
}

variable "admin_email" {
  type        = string
  description = "E-mail do admin seed (não é segredo). Espelha ADMIN_EMAIL."
  default     = "admin@estudaja.com"
}

variable "cors_origin" {
  type        = string
  description = "CORS_ORIGIN da API. Placeholder até o CloudFront do front (Fase C/D)."
  default     = "https://pending-frontend.invalid"
}

variable "db_username" {
  type        = string
  description = "Usuário do RDS (Fase C). Usado no placeholder de DATABASE_URL."
  default     = "estudaja"
}

variable "db_name" {
  type        = string
  description = "Nome do banco RDS (Fase C)."
  default     = "estudaja"
}

# Segredos: sem default commitado. Se omitidos, random_password gera valores no apply
# (ficam no state local gitignored). Preferir TF_VAR_* ou terraform.tfvars gitignored.

variable "jwt_secret" {
  type        = string
  description = "JWT_SECRET (SSM SecureString). Null = gerar no apply."
  sensitive   = true
  default     = null
}

variable "admin_password" {
  type        = string
  description = "ADMIN_PASSWORD (SSM SecureString). Null = gerar no apply."
  sensitive   = true
  default     = null
}

variable "db_password" {
  type        = string
  description = "Senha RDS / DATABASE_URL (SSM SecureString). Null = gerar no apply."
  sensitive   = true
  default     = null
}

variable "database_url" {
  type        = string
  description = "DATABASE_URL completo. Null = montar a partir do RDS (address + senha)."
  sensitive   = true
  default     = null
}

variable "db_instance_class" {
  type        = string
  description = "Classe da instância RDS (FR-007)."
  default     = "db.t4g.micro"
}

variable "fargate_cpu" {
  type        = string
  description = "CPU da task Fargate (unidades AWS). Default 256 = 0.25 vCPU."
  default     = "256"
}

variable "fargate_memory" {
  type        = string
  description = "Memória da task Fargate em MiB. Default 512."
  default     = "512"
}

variable "ecs_desired_count" {
  type        = number
  description = "Réplicas Fargate na janela de demo (research §3)."
  default     = 1
}
