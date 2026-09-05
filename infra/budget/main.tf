# Stack SEPARADA da demo (FR-022 / research §13).
#
# Aplicar UMA VEZ por conta, ANTES da primeira demo:
#   cd infra/budget
#   terraform init
#   terraform apply
#
# NÃO rodar terraform destroy aqui no ciclo da sessão.
# terraform destroy em infra/ (VPC/SSM/…) não inclui estes recursos.
#
# Conferir: Console AWS → Billing → Budgets (SC-012).

resource "aws_budgets_budget" "monthly_cost" {
  name              = "${var.project_name}-monthly-cost"
  budget_type       = "COST"
  limit_amount      = var.budget_limit_usd
  limit_unit        = "USD"
  time_unit         = "MONTHLY"
  time_period_start = "2026-01-01_00:00"

  notification {
    comparison_operator        = "GREATER_THAN"
    threshold                  = var.threshold_percent
    threshold_type             = "PERCENTAGE"
    notification_type          = "ACTUAL"
    subscriber_email_addresses = [var.notification_email]
  }

  notification {
    comparison_operator        = "GREATER_THAN"
    threshold                  = 100
    threshold_type             = "PERCENTAGE"
    notification_type          = "FORECASTED"
    subscriber_email_addresses = [var.notification_email]
  }

  lifecycle {
    prevent_destroy = true
  }
}
