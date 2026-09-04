# Contract: outputs Terraform

Root module em `infra/`. Outputs são a interface operador ↔ publish (front/API).

## Outputs obrigatórios

| Output | Tipo | Uso |
|--------|------|-----|
| `frontend_url` | string (HTTPS) | URL pública CloudFront do site estático |
| `api_url` | string (HTTPS) | URL pública CloudFront (origin ALB) — valor de `VITE_API_URL` |
| `alb_dns_name` | string | Diagnóstico / origin; não usar no browser se mixed content |
| `ecr_repository_url` | string | `docker push` da imagem API |
| `ecs_cluster_name` | string | `aws ecs update-service` |
| `ecs_service_name` | string | force new deployment |
| `s3_bucket_name` | string | `aws s3 sync` do `frontend/dist` |
| `cloudfront_frontend_distribution_id` | string | invalidação pós-publish |
| `rds_endpoint` | string | diagnóstico (conexão via `DATABASE_URL` na task) |
| `aws_region` | string | sempre `us-east-1` nesta feature |

## Inputs / variables relevantes

| Variable | Descrição |
|----------|-----------|
| `project_name` | prefixo de nomes (ex. `estudaja`) |
| `db_instance_class` | default `db.t4g.micro` |
| `fargate_cpu` / `fargate_memory` | default `256` / `512` |
| `admin_email` | seed |
| Sensitive vars | senhas/JWT geradas ou passadas via TF_VAR / arquivo `*.tfvars` gitignored |

## Estado e destroy

- Apply idempotente sem mudanças materiais (SC-002).
- Destroy remove recursos da stack de demo; **não** remove AWS Budget da conta.
- ECR com `force_delete` para permitir destroy com imagens.
- RDS: `skip_final_snapshot = true` (dados efêmeros).
