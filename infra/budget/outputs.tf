output "budget_name" {
  description = "Nome do AWS Budget da conta (permanece após destroy da demo)."
  value       = aws_budgets_budget.monthly_cost.name
}

output "budget_limit_usd" {
  description = "Limiar mensal em USD."
  value       = var.budget_limit_usd
}

output "notification_email" {
  description = "E-mail inscrito nas notificações."
  value       = var.notification_email
}
