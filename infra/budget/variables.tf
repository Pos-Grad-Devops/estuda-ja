variable "aws_region" {
  type        = string
  description = "Região do provider (Budgets é serviço de conta; FR-015 usa us-east-1)."
  default     = "us-east-1"

  validation {
    condition     = var.aws_region == "us-east-1"
    error_message = "Esta feature exige aws_region = us-east-1 (FR-015)."
  }
}

variable "project_name" {
  type        = string
  description = "Prefixo do nome do budget."
  default     = "estudaja"
}

variable "budget_limit_usd" {
  type        = string
  description = "Limiar mensal em USD (FR-022 / SC-012). Baixo de propósito."
  default     = "5"
}

variable "notification_email" {
  type        = string
  description = "E-mail que recebe o alerta. Sobrescreva para o operador da conta."
  default     = "admin@estudaja.com"
}

variable "threshold_percent" {
  type        = number
  description = "Percentual do limiar que dispara notificação ACTUAL."
  default     = 80
}
