# Parameter Store — contrato contracts/api-env.md.
# SecureString: JWT_SECRET, ADMIN_PASSWORD, DATABASE_URL, senha DB.
# String: ADMIN_EMAIL, CORS_ORIGIN.
# PORT e JWT_EXPIRATION ficam como env da task (Fase C), não SSM.

resource "aws_ssm_parameter" "jwt_secret" {
  name        = "${local.ssm_prefix}/JWT_SECRET"
  description = "JWT_SECRET da API (nunca default de dev na AWS)"
  type        = "SecureString"
  value       = local.jwt_secret
}

resource "aws_ssm_parameter" "admin_password" {
  name        = "${local.ssm_prefix}/ADMIN_PASSWORD"
  description = "ADMIN_PASSWORD do seed"
  type        = "SecureString"
  value       = local.admin_password
}

resource "aws_ssm_parameter" "admin_email" {
  name        = "${local.ssm_prefix}/ADMIN_EMAIL"
  description = "ADMIN_EMAIL do seed"
  type        = "String"
  value       = var.admin_email
}

resource "aws_ssm_parameter" "cors_origin" {
  name        = "${local.ssm_prefix}/CORS_ORIGIN"
  description = "Origem HTTPS do CloudFront do frontend (ajustável na Fase D)"
  type        = "String"
  value       = local.cors_origin
}

resource "aws_ssm_parameter" "db_password" {
  name        = "${local.ssm_prefix}/DB_PASSWORD"
  description = "Senha RDS (também embutida em DATABASE_URL)"
  type        = "SecureString"
  value       = local.db_password
}

resource "aws_ssm_parameter" "database_url" {
  name        = "${local.ssm_prefix}/DATABASE_URL"
  description = "DATABASE_URL da API (sslmode=require). Fail loud se ausente na task."
  type        = "SecureString"
  value       = local.database_url
}

# Live IVS (contracts/live-env.md / terraform-ivs.md) — key nunca em output plaintext.
resource "aws_ssm_parameter" "ivs_stream_key" {
  name        = "${local.ssm_prefix}/IVS_STREAM_KEY"
  description = "Stream key IVS da sessão (SecureString; task secret)"
  type        = "SecureString"
  value       = data.aws_ivs_stream_key.live.value
}
