# Quickstart: sessão de demo AWS

Runbook operacional (validação ponta a ponta). Sem código de implementação — apenas passos e resultados esperados.

Espelho resumido no [README.md](../../README.md) (seção **Demo AWS**). Guidelines: [AGENTS.md](../../AGENTS.md).

Contratos: [api-env](./contracts/api-env.md) · [terraform-outputs](./contracts/terraform-outputs.md) · [frontend-publish](./contracts/frontend-publish.md)

## Pré-requisitos

- Conta AWS com permissões para VPC, ECS, ECR, ALB, RDS, S3, CloudFront, SSM, IAM, Budgets
- Região: **us-east-1**
- Ferramentas: Terraform (≥ 1.5), AWS CLI, Docker (build `linux/arm64`), Node 22, Go 1.25 (opcional local)
- Repositório clonado na branch da feature
- **Antes da primeira demo:** alerta de billing/orçamento ativo — ver §0

## 0. Alerta de billing (uma vez por conta)

Stack **separada** em `infra/budget/` (não faz parte do destroy da demo):

```bash
cd infra/budget
cp terraform.tfvars.example terraform.tfvars   # defina notification_email
terraform init && terraform apply
```

| Campo | Default / limiar |
|-------|------------------|
| `budget_limit_usd` | **5** (US$ 5 / mês) |
| `threshold_percent` | **80** (alerta ACTUAL ao atingir 80% do limiar) |
| FORECASTED | 100% do limiar |
| Console | **Billing → Budgets** |

1. Confirmar budget na console (nome com prefixo do projeto + e-mail).
2. **Não** destruir este alerta junto com a stack de demo (`cd infra && terraform destroy` **não** toca `infra/budget/`).

**Esperado:** SC-012 satisfeito.

## 1. Apply (infra)

```bash
cd infra
cp terraform.tfvars.example terraform.tfvars   # opcional
terraform init
terraform apply
```

Segredos omitidos em tfvars são gerados no apply (`random_password`) e ficam no **state local** (gitignored) + SSM.

**Esperado (≤ 90 min na primeira vez com guia):** outputs `frontend_url`, `api_url`, `ecr_repository_url`, `s3_bucket_name`, `ecs_*`, `cloudfront_frontend_distribution_id`.

`CORS_ORIGIN` no SSM = origem HTTPS do CloudFront frontend (salvo override `cors_origin`).

## 2. Publish API (+ CORS + health)

Preferir script:

```powershell
cd infra
.\publish-api.ps1
```

Passos equivalentes:

1. Login ECR; `docker build --platform linux/arm64` a partir de `backend/`.
2. Tag + push para `ecr_repository_url:latest`.
3. Confirmar secrets SSM (`DATABASE_URL`, `JWT_SECRET`, `ADMIN_*`, `CORS_ORIGIN`).
4. `aws ecs update-service --force-new-deployment` (cluster/service dos outputs).
5. Aguardar task healthy; `GET {api_url}/health` → **200**.

Se precisar realinhar CORS depois: atualizar SSM `CORS_ORIGIN` = origem de `frontend_url` e force deploy.

**Esperado:** API HTTPS responde; logs CloudWatch sem fatal de migrate/seed.

## 3. Rebuild / publish frontend

```powershell
cd infra
.\publish-frontend.ps1
```

Ou manual ([frontend-publish](./contracts/frontend-publish.md)):

1. `VITE_API_URL={api_url}` → `npm run build` em `frontend/`.
2. `aws s3 sync dist/` → `s3_bucket_name`; invalidar CloudFront (`/*`).

**Esperado:** `frontend_url` carrega a SPA; network do browser chama `api_url` (sem mixed content).

## 4. Warm-up operacional (antes da janela)

### Decisão Fase F / P2 (escolha única)

**Escolha (b) — gap P2 documentado:** **não** há automação EventBridge/Lambda/CodeBuild de apply, warm-up ou destroy neste MVP. O warm-up é **100% operacional e manual**, com checklist T−15 min abaixo. Automação de agendamento na nuvem permanece **fora do caminho mínimo** (US4 AC3 / FR-012).

| Alternativa | Status |
|-------------|--------|
| (a) EventBridge + artefato mínimo de agendamento | **Não adotada** (custo/complexidade vs. conta acadêmica; constitution I) |
| (b) Gap P2 + procedimento manual reforçado | **Adotada** (este runbook) |

Warm-up **não** é scale-from-zero: com a stack já aplicada, `desired_count = 1` (research §3) e a task running, o risco de cold start na abertura da janela some.

### Checklist T−15 min (obrigatório antes da janela simulada)

Comece **no máximo 15 minutos antes** do horário combinado (RDS cold + first migrate + propagação CF + **canal IVS / OBS** se a demo incluir ao vivo). Marque na ordem:

**T−15 — infra e capacidade**

- [ ] `terraform apply` da sessão já concluído (outputs `api_url` / `frontend_url` conhecidos)
- [ ] RDS em estado **available** (console RDS ou describe-db-instances)
- [ ] Serviço ECS com `desired_count = 1` e ≥1 task **RUNNING** (não 0; não “scale-to-zero”)
- [ ] Target group / ALB com targets **healthy** (ou equivalente no console ECS)
- [ ] **Live (003):** `terraform output ivs_channel_arn` presente (1 canal); task com `LIVE_BACKEND=ivs` após publish — detalhe [003 quickstart §B3](../003-live-streaming-ivs/quickstart.md)

**T−10 — API saudável**

- [ ] `GET {api_url}/health` → **200** de forma estável (2–3 chamadas espaçadas; HTTPS CloudFront da API)
- [ ] Logs CloudWatch da task sem fatal de migrate/seed (opcional, se health oscilar)
- [ ] **Live (003):** status live da aula demo = inativa (ou estado conhecido); sem cold start de stack pendente

**T−5 — smoke da demo (SC-010 / SC-011)**

- [ ] Login **admin** seed + listagem de ≥1 curso (API ou UI)
- [ ] Login pela UI em `{frontend_url}` (SPA carrega; network chama `api_url` sem mixed content)
- [ ] Amostra rápida RBAC: professor consegue escrita de aula **ou** aluno só lê (mensagem de negação em PT)
- [ ] **Live (003):** professor **Inicia live** + OBS **Live**; player do gestor com sinal; aluno em outra sessão vê ao vivo (não confundir com VOD) — **antes** de T−0

**T−0 — janela aberta**

- [ ] Não alterar `desired_count` para 0 durante a demo
- [ ] Não iniciar `terraform destroy` até o encerramento combinado
- [ ] **Live (003):** encoder continua; **não** apply/destroy nem “subir OBS agora” no minuto da abertura

**Se pular o warm-up:** cold start / RDS ainda subindo / first migrate / **canal ou OBS frios** na abertura da janela — viola o critério de preparação para o pico (constitution II / FR-012; live SC-006). Não use a sessão como “demo de pico” sem este checklist.

## 5. Verificar demo (RBAC)

Com admin, professor e aluno seedados (defaults de dev — ver README):

1. Admin: login + gestão permitida.
2. Professor: criar/editar aula; escrita de curso negada (mensagem em PT).
3. Aluno: leitura ok; escritas negadas.
4. Credencial inválida: acesso negado claro.

**Esperado:** SC-003, SC-004 amostral.

## 6. Destroy

```bash
cd infra
terraform destroy
```

### Checklist pós-destroy (≤ 15 min — SC-007)

- [ ] ECS / ALB / RDS / CloudFront ×2 / S3 site / VPC / SG do projeto **ausentes**
- [ ] Sem **NAT Gateway** na conta deste projeto (nunca provisionado)
- [ ] ECR removido (`force_delete = true` no TF)
- [ ] Sem snapshot final RDS (`skip_final_snapshot = true`)
- [ ] Log groups órfãos: apagar se sobrarem fora do TF
- [ ] **Budget ainda ativo** (Billing → Budgets) — **não** destruído
- [ ] State local na máquina do operador (não versionado)

Resíduos detalhados: [research.md](./research.md) §10.

**Esperado:** custo contínuo de compute/banco/ALB ~ **US$ 0**.

## 7. Checar billing

1. Cost Explorer / Bills — consumo só na janela da sessão.
2. Confirmar que o alerta (limiar US$ 5 / 80%) dispararia se a stack ficasse ligada por engano.

## Estimativa de custo (ordem de grandeza — SC-006)

| Cenário | Faixa |
|---------|--------|
| Sessão ~4 h ligada, sem Free Tier | **~US$ 1–3** |
| Mesma janela com Free Tier RDS/ALB | **~US$ 0–1** |
| Após destroy da stack de demo | **~US$ 0** (exceto budget/alertas e resíduos mal limpos) |
| NAT Gateway | **US$ 0** (não provisionado) |

Componentes: [research.md](./research.md) §9. Não é cotação contratual — usar Pricing Calculator na data da demo.

## CI local / pipeline (baseline P1)

Sem deploy:

```text
cd backend && go test ./...
cd frontend && npm ci && npm run build
```

CI em `.github/workflows/ci.yml` deve permanecer verde (SC-005). **CD = P3 opcional** em `.github/workflows/cd.yml` (habilitar com `AWS_CD_ENABLED=true` + secrets da sessão). Publish manual (`infra/publish-*.ps1`) **continua o caminho P1** e deve ser usado quando a stack está destruída ou os secrets estão desatualizados.

## Gaps / opcionais além do caminho mínimo

| Item | Prioridade | Neste runbook |
|------|------------|---------------|
| Automação EventBridge apply / warm-up / destroy | **P2** | **Gap explícito (Fase F / escolha b)** — checklist T−15 min §4; sem IaC de agendamento |
| CD GitHub Actions (ECR + S3/CF) | **P3** | **Entregue (opcional)** — `.github/workflows/cd.yml`; secrets = outputs da sessão (efêmeros pós-destroy); manual P1 permanece |

### CD — secrets de sessão (efêmeros)

Após cada `terraform apply`, se for usar Actions: preencha secrets com os outputs (`ecr_repository_url`, `ecs_*`, `api_url` → `VITE_API_URL`, `s3_bucket_name`, `cloudfront_frontend_distribution_id`) e `AWS_CD_ENABLED=true`. Após `destroy`, desative o flag — **não** reutilize URLs de sessão anterior.

## Segredos (SC-008)

Não versionar: `*.tfvars` (exceto `*.tfvars.example`), `*.tfstate*`, `.terraform/`, `.env`. Revisar com `git status` antes de commit. Secrets do Actions também são por sessão (não commitados).

## Fora de escopo neste runbook

Chat, OAuth, Redis AWS, domínio customizado, automação EventBridge apply/destroy (P2 — escolha b / gap).  
**Streaming ao vivo** não é detalhado aqui — runbook e critérios: [003-live-streaming-ivs/quickstart.md](../003-live-streaming-ivs/quickstart.md) (encaixa no ciclo apply→publish→warm-up→demo→destroy **sem** reabrir NAT/Redis/budget).

## Validação deste quickstart (checklist frio)

Use quando a stack da última sessão já foi destruída — percorra sem inventar passos:

- [ ] §0 budget: limiar US$ 5, console Billing → Budgets, destroy da demo não remove
- [ ] Ordem: billing → apply → push API → CORS → health → front → warm-up → demo → destroy
- [ ] Scripts `publish-api.ps1` / `publish-frontend.ps1` ou passos manuais equivalentes citados
- [ ] §4: decisão Fase F = **(b) gap**; checklist T−15 min (T−15 / T−10 / T−5 / T−0) presente
- [ ] Pós-destroy: tabela de resíduos + budget permanece
- [ ] Custo ~US$ 1–3 / 4 h; NAT = US$ 0; ~US$ 0 após destroy
- [ ] P2 = gap (escolha b); P3 = CD opcional documentado (manual P1 ainda válido)
- [ ] Links aos três contracts + research §9–10 / §10b

**Última validação (Fase G / T050–T054):** CD em `cd.yml` com gate test/build; secrets de sessão efêmeros; publish manual P1 intacto.
