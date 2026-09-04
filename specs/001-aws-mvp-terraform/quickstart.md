# Quickstart: sessão de demo AWS

Runbook operacional (validação ponta a ponta). Sem código de implementação — apenas passos e resultados esperados.

Contratos: [api-env](./contracts/api-env.md) · [terraform-outputs](./contracts/terraform-outputs.md) · [frontend-publish](./contracts/frontend-publish.md)

## Pré-requisitos

- Conta AWS com permissões para VPC, ECS, ECR, ALB, RDS, S3, CloudFront, SSM, IAM, Budgets
- Região: **us-east-1**
- Ferramentas: Terraform, AWS CLI, Docker (build `linux/arm64`), Node 22, Go 1.25 (opcional local)
- Repositório clonado na branch `001-aws-mvp-terraform`
- **Antes da primeira demo:** alerta de billing/orçamento ativo (limiar baixo, ex. US$ 5) — ver FR-022

## 0. Alerta de billing (uma vez por conta)

1. Criar AWS Budget (custo) com limiar baixo em USD e notificação por e-mail/SNS.
2. Confirmar na console Billing → Budgets.
3. **Não** destruir este alerta junto com a stack de demo.

**Esperado:** SC-012 satisfeito.

## 1. Apply (infra)

```text
cd infra
terraform init
terraform apply
```

**Esperado (≤ 90 min na primeira vez com guia):** outputs `frontend_url`, `api_url`, `ecr_repository_url`, `s3_bucket_name`, `ecs_*`, `cloudfront_frontend_distribution_id`.

## 2. Publish API

1. Login ECR; `docker build --platform linux/arm64` a partir de `backend/`.
2. Tag + push para `ecr_repository_url:latest`.
3. Garantir secrets SSM / env (`DATABASE_URL`, `JWT_SECRET`, `ADMIN_*`, `CORS_ORIGIN` — este último pode ser ajustado após saber `frontend_url`).
4. `aws ecs update-service --force-new-deployment` (cluster/service dos outputs).
5. Aguardar task healthy; `GET {api_url}/health` → 200.

**Esperado:** API HTTPS responde; logs no CloudWatch sem fatal de migrate.

## 3. Rebuild / publish frontend

1. `VITE_API_URL={api_url}` → `npm run build` em `frontend/`.
2. `aws s3 sync dist/` → bucket; invalidar CloudFront do frontend.
3. Se necessário, alinhar `CORS_ORIGIN` e redeploy da API.

**Esperado:** `frontend_url` carrega a SPA; network do browser chama `api_url`.

## 4. Warm-up operacional (antes da janela)

Checklist (P1 manual; automação = P2):

- [ ] Apply concluído há tempo suficiente (RDS available + migrations ok)
- [ ] `desired_count = 1` e task running (não scale-to-zero)
- [ ] `GET {api_url}/health` estável
- [ ] Login admin seed + listagem de cursos (SC-010)
- [ ] Login pela UI hospedada (SC-011)

**Antecedência sugerida:** alguns minutos após healthy (RDS cold + first migrate); para “aula simulada”, aplicar com margem documentada no README (ex. T−15 min).

## 5. Verificar demo (RBAC)

Com admin, professor e aluno seedados:

1. Admin: login + gestão permitida.
2. Professor: criar/editar aula; escrita de curso negada (PT).
3. Aluno: leitura ok; escritas negadas.
4. Credencial inválida: acesso negado claro.

**Esperado:** SC-003, SC-004.

## 6. Destroy

```text
cd infra
terraform destroy
```

Checklist pós-destroy (≤ 15 min — SC-007):

- [ ] ECS/ALB/RDS/CloudFront/S3/VPC ausentes na console
- [ ] Sem NAT Gateway na conta deste projeto
- [ ] Budget ainda ativo
- [ ] Resíduos: ver [research.md](./research.md) §10 (ECR force_delete, snapshots, log groups)

**Esperado:** custo contínuo de compute/banco/ALB ~ US$ 0.

## 7. Checar billing

1. Cost Explorer / Bills — consumo só na janela da sessão.
2. Confirmar que o alerta dispararia se a stack ficasse ligada por engano.

## CI local / pipeline (baseline P1)

Sem deploy:

```text
cd backend && go test ./...
cd frontend && npm ci && npm run build
```

CI em `.github/workflows/ci.yml` deve permanecer verde (SC-005). CD = P3.

## Estimativa de custo (ordem de grandeza)

Sessão ~4 h ligada: tipicamente **~US$ 1–3** sem Free Tier; menos se RDS/ALB Free Tier elegíveis. Detalhes: [research.md](./research.md) §9. Fora da sessão após destroy: ~US$ 0 de stack (exceto budget/alertas e possíveis resíduos mal limpos).

## Fora de escopo neste runbook

Streaming, chat, OAuth, Redis AWS, domínio customizado, CD, automação EventBridge de apply/destroy (P2).
