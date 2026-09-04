---
description: "Task list — MVP AWS com custo controlado (fases = chats isolados)"
---

# Tasks: MVP AWS com custo controlado

**Input**: Design documents from `/specs/001-aws-mvp-terraform/`

**Prerequisites**: [plan.md](./plan.md), [spec.md](./spec.md), [research.md](./research.md), [data-model.md](./data-model.md), [contracts/](./contracts/), [quickstart.md](./quickstart.md), constitution

**Tests**: Obrigatórios só quando handlers/API mudarem (constitution V). E2E da demo = runbook (quickstart), não suite nova em P1.

**Organization**: Fases **A–G** = **um chat `/speckit-implement` cada**. Não misturar IaC + seed + publish + docs no mesmo chat.

---

## Formato e protocolo de chats

### Formato da task

`- [ ] [ID] [P?] [Story?] Descrição com path — depends: … — aceite: …`

- **[P]**: paralelizável (arquivos distintos, sem depender de task incompleta)
- **[USn]**: user story do [spec.md](./spec.md) (US1 demo hospedada · US2 IaC · US3 baseline local/CI · US4 ciclo warm-up · US5 CD)
- Setup/foundational embutidos nas fases A–E; fases F–G são **opcionais** (P2/P3)

### Regra de ouro (NÃO negociável)

1. Abrir **um** chat `/speckit-implement` **somente** para a fase atual.
2. Completar **todas** as tasks da fase + checklist **Done** verificável.
3. Só então abrir o próximo chat na fase seguinte.
4. **Não** começar Fase B+ sem Done da A; **não** misturar publish (D) com Terraform (B/C) nem docs (E) com IaC.

### Ordem dos chats

| Chat | Fase | Escopo | Prioridade |
|------|------|--------|------------|
| 1 | **A** | App/CI local (sem AWS) | P1 / US3 |
| 2 | **B** | Terraform base (rede + SSM + budget + outputs mínimos) | P1 / US2 |
| 3 | **C** | Terraform compute/data (RDS, ECR, ECS, ALB, CF, S3) | P1 / US2 |
| 4 | **D** | Publish da sessão (imagem + front) | P1 / US1 |
| 5 | **E** | Runbook + docs + custo/destroy | P1 polish |
| 6 | **F** | Warm-up automation ou gap documentado | **P2 / US4 — opcional** |
| 7 | **G** | CD no GitHub Actions | **P3 / US5 — opcional** |

**Começar implementação pela Fase A apenas.**

---

## Phase A — Endurecimento app/CI local (sem AWS) — CHAT 1

**Story**: US3 — Endurecimento da entrega local/CI como baseline  
**Goal**: App configurável por env, seed demo idempotente, CORS/front por sessão documentáveis localmente; CI verde.  
**Independent Test**: `docker compose` sobe stack; login admin/professor/aluno; listagem de ≥1 curso; `go test` + `npm run build` / CI passam.  
**FORA DESTE CHAT**: qualquer arquivo em `infra/`, push ECR, S3, Terraform, publish AWS, docs AWS longas (só `.env.example` / notas mínimas se necessário).

### Implementation

- [ ] T001 [US3] Auditar e alinhar `backend/internal/config/config.go` ao contrato [contracts/api-env.md](./contracts/api-env.md) (`PORT`, `DATABASE_URL`, `CORS_ORIGIN`, `JWT_SECRET`, `JWT_EXPIRATION`, `ADMIN_EMAIL`, `ADMIN_PASSWORD`) — depends: nenhuma — aceite: nomes e defaults locais documentáveis; sem hardcode de segredos AWS
- [ ] T002 [P] [US3] Garantir que `backend/cmd/api/main.go` usa `cfg.CORSOrigin` (origem única configurável) sem lista hardcoded de produção — depends: T001 — aceite: alterar `CORS_ORIGIN` muda o AllowOrigins sem recompilar lógica extra
- [ ] T003 [P] [US3] Atualizar `frontend/.env.example` e, se existir, `backend`/raiz `.env.example` com variáveis do contrato (sem valores secretos reais) — depends: T001 — aceite: operador local sabe quais envs setar; `VITE_API_URL` documentado
- [ ] T004 [US3] Criar migration de seed demo idempotente em `backend/internal/database/migrations/002_seed_demo.go` (professor, aluno, 1 curso, 1 aula) e registrar em `backend/internal/database/migrations/migrations.go` — depends: T001 — aceite: banco vazio após migrate tem admin (001) + demo (002); re-run não duplica
- [ ] T005 [US3] Credenciais demo do seed (emails/senhas de professor/aluno) só via env documentado ou defaults de **dev** explícitos — nunca secrets de conta AWS no git — depends: T004 — aceite: SC-010 localmente (login admin + ≥1 curso); senhas demo citáveis no guia local/README curto se necessário
- [ ] T006 [P] [US3] Confirmar `frontend/src/api/client.ts` e `frontend/Dockerfile` aceitam `VITE_API_URL` por build-arg/env (rebuild por “sessão” local = mudar env e rebuild) — depends: T003 — aceite: build com `VITE_API_URL` diferente embute a URL no bundle
- [ ] T007 [US3] Ajustar `docker-compose.yml` (e `Makefile` se necessário) para passar `DATABASE_URL`, `JWT_*`, `CORS_ORIGIN`, `ADMIN_*` ao serviço da API — depends: T001, T004 — aceite: `docker compose up` sobe API + Postgres; migrate/seed rodam na subida
- [ ] T008 [US3] Se T004/T005 alterarem comportamento testável da API/seed, adicionar/atualizar testes Go relevantes (ex. `backend/internal/database/...` ou handler smoke já existente); senão documentar N/A no PR/resumo do chat — depends: T004 — aceite: constitution V respeitada; `go test ./...` verde
- [ ] T009 [US3] Garantir CI em `.github/workflows/ci.yml` cobre o branch da feature (push/PR) com jobs backend test+build, frontend build, docker build — depends: T008 — aceite: SC-005 verificável no branch `001-aws-mvp-terraform`
- [ ] T010 [US3] Verificação manual local: compose up → login 3 perfis → listar cursos; rodar `cd backend && go test ./...` e `cd frontend && npm ci && npm run build` — depends: T007, T009 — aceite: checklist Done da Fase A abaixo

### Done — Fase A (obrigatório antes do Chat 2)

- [ ] `docker compose` sobe API + DB sem erro fatal de migrate/seed
- [ ] Login admin, professor e aluno (seed) funciona
- [ ] Listagem mostra ≥1 curso de exemplo (SC-010 local)
- [ ] `CORS_ORIGIN` / `VITE_API_URL` configuráveis por env (sem hardcode de hostname AWS)
- [ ] `go test ./...` e build frontend passam; CI do branch verde ou equivalente local documentado
- [ ] **Nenhum** recurso AWS criado neste chat

**Checkpoint**: Baseline US3 pronta. Parar. Abrir novo chat só para Fase B.

---

## Phase B — Terraform base (rede + SSM + budget + outputs) — CHAT 2

**Story**: US2 (fundação) — Provisionamento reproduzível (parte 1)  
**Goal**: Root Terraform enxuto em `infra/` com VPC pública 2 AZ, SG base, SSM params, budget/alerta de conta, outputs mínimos — **sem** ECS/RDS/S3/CloudFront/ALB ainda.  
**Independent Test**: `terraform init` + `terraform validate` (+ `plan` documentado); state local gitignored.  
**FORA DESTE CHAT**: RDS, ECR, ECS, ALB, S3 site, CloudFront, publish Docker/front, seed app, README AWS completo.

### Implementation

- [ ] T011 [US2] Criar skeleton `infra/` flat: `providers.tf` (AWS `us-east-1`), `versions.tf` (Terraform ≥ 1.5), `variables.tf`, `outputs.tf`, `main.tf` (ou split por arquivo lógico sem modules/) — depends: Fase A Done — aceite: `terraform init` funciona; região fixa `us-east-1` (FR-015)
- [ ] T012 [P] [US2] Adicionar `infra/.gitignore` (ou entradas na raiz `.gitignore`) para `*.tfstate*`, `.terraform/`, `*.tfvars` sensíveis — depends: T011 — aceite: state local não versionado (research §8)
- [ ] T013 [US2] Declarar VPC custom + Internet Gateway + **2 subnets públicas** (2 AZs) em `infra/vpc.tf` (ou `main.tf`) — **sem NAT Gateway** — depends: T011 — aceite: plan mostra VPC/subnets/IGW; zero NAT
- [ ] T014 [US2] Security groups base em `infra/sg.tf`: SG placeholder para futuros ALB/tasks/RDS (regras mínimas; RDS ainda não criado) — depends: T013 — aceite: SGs referenciam a VPC; sem abrir 0.0.0.0/0 no Postgres prematuramente sem necessidade
- [ ] T015 [P] [US2] Variáveis sensíveis e de projeto em `infra/variables.tf` (`project_name`, `admin_email`, senhas/JWT via TF_VAR / tfvars gitignored) — depends: T011 — aceite: defaults seguros só para não-secrets; secrets sem default commitado
- [ ] T016 [US2] Parâmetros SSM Parameter Store (SecureString/String) em `infra/ssm.tf` para `JWT_SECRET`, `ADMIN_PASSWORD`, placeholders de `DATABASE_URL`/senha DB, `ADMIN_EMAIL`, `CORS_ORIGIN` conforme [contracts/api-env.md](./contracts/api-env.md) — depends: T015 — aceite: params com prefixo do projeto; valores sensíveis SecureString
- [ ] T017 [US2] AWS Budget (ou billing alert) com limiar baixo (ex. US$ 5) + notificação — preferir recurso **fora do destroy da demo** (`infra/budget.tf` com lifecycle / stack separada documentada) — depends: T011 — aceite: FR-022 / SC-012 caminho claro; budget **não** some no destroy da stack de demo
- [ ] T018 [US2] Outputs mínimos em `infra/outputs.tf`: `aws_region`, IDs de VPC/subnets/SGs/SSM names úteis; stubs ou TODOs comentados para outputs futuros do contrato — depends: T013, T016 — aceite: `terraform output` lista região + rede; ainda **sem** exigir `frontend_url`/`api_url` reais
- [ ] T019 [US2] Rodar `terraform validate` e `terraform plan` em `infra/`; anotar resultado no resumo do chat (apply **opcional** e só se seguro/credenciais OK) — depends: T017, T018 — aceite: validate OK; plan sem erros; **app ainda não no ar é OK**

### Done — Fase B (obrigatório antes do Chat 3)

- [ ] `infra/` existe, flat, sem modules elaborados
- [ ] VPC 2 AZ públicas + IGW; **sem** NAT no plan
- [ ] SSM params do projeto declarados
- [ ] Budget/alerta configurável e separado do ciclo destroy da demo
- [ ] `terraform validate` OK; `plan` documentado no chat
- [ ] Sem ECS/RDS/ECR/S3/CloudFront/ALB neste chat (ou apenas refs/SG placeholders)
- [ ] Sem publish de imagem/front

**Checkpoint**: Fundação US2 (rede/segredos/budget). Parar. Abrir novo chat só para Fase C.

---

## Phase C — Terraform compute/data (RDS + ECR + ECS + ALB + CloudFront + S3) — CHAT 3

**Story**: US2 (completo) — Provisionamento reproduzível (parte 2)  
**Goal**: Completar recursos do plan/research; outputs do [contracts/terraform-outputs.md](./contracts/terraform-outputs.md).  
**Independent Test**: `terraform apply` sobe infra; outputs HTTPS/URLs presentes; health da API **pode** falhar até Fase D (sem imagem).  
**FORA DESTE CHAT**: `docker push`, `npm run build` de sessão, sync S3 de app, scripts de publish, README longo (só comentários TF se preciso).

### Implementation

- [ ] T020 [US2] RDS PostgreSQL `db.t4g.micro` single-AZ em subnet pública, SG só a partir da SG das tasks; `skip_final_snapshot = true`; senha → SSM/`DATABASE_URL` — depends: Fase B Done — aceite: FR-007; sem ElastiCache (FR-008)
- [ ] T021 [P] [US2] Repositório ECR (`force_delete = true`) em `infra/ecr.tf` — depends: Fase B Done — aceite: output futuro `ecr_repository_url`; destroy limpa imagens
- [ ] T022 [US2] ALB (público) + target group porta 8080 + listener HTTP :80 em `infra/alb.tf` — depends: T013/T014 (Fase B) — aceite: ALB em ≥2 AZs; DNS em output `alb_dns_name`
- [ ] T023 [US2] IAM roles de execução/tarefa ECS com `ssm:GetParameters` + decrypt KMS + awslogs + pull ECR em `infra/iam.tf` — depends: T016, T021 — aceite: task pode ler secrets SSM
- [ ] T024 [US2] Cluster ECS Fargate + task definition (256 CPU / 512 MiB, **linux/ARM64**) + service `desired_count = 1`, `assign_public_ip = true`, awslogs — depends: T020, T021, T022, T023 — aceite: research §3; sem NAT; imagem placeholder ou ECR empty OK até D
- [ ] T025 [P] [US2] S3 bucket site (private) + OAC em `infra/s3.tf` — depends: Fase B Done — aceite: bucket versionável no destroy; sem website público direto sem CF
- [ ] T026 [US2] CloudFront **frontend** (origin S3/OAC, cert default `*.cloudfront.net`) em `infra/cloudfront_frontend.tf` — depends: T025 — aceite: output `frontend_url` HTTPS
- [ ] T027 [US2] CloudFront **API** (origin ALB HTTP) para TLS sem domínio custom — depends: T022 — aceite: output `api_url` HTTPS; evita mixed content (research §5)
- [ ] T028 [US2] Wire env/secrets na task definition: `PORT`, `DATABASE_URL`, `JWT_*`, `ADMIN_*`, `CORS_ORIGIN` (atualizável pós-front) — depends: T024, T027 — aceite: contrato api-env; fail loud se secret ausente
- [ ] T029 [US2] Completar `infra/outputs.tf` com **todos** os outputs obrigatórios do contrato terraform-outputs — depends: T020–T027 — aceite: `frontend_url`, `api_url`, `ecr_repository_url`, `ecs_*`, `s3_bucket_name`, `cloudfront_frontend_distribution_id`, `rds_endpoint`, `aws_region`, `alb_dns_name`
- [ ] T030 [US2] `terraform validate` + `apply` (conta demo); confirmar recursos na console; **não** exigir `/health` 200 ainda — depends: T029 — aceite: SC-001 parcial (infra up); SC-002 re-plan idempotente sem clique manual

### Done — Fase C (obrigatório antes do Chat 4)

- [ ] `terraform apply` concluiu sem erro
- [ ] Outputs do contrato presentes e HTTPS onde aplicável
- [ ] Sem NAT Gateway; Fargate ARM 256/512; RDS micro; SSM; CF×2 + S3 + ALB + ECS + ECR
- [ ] Budget da conta ainda ativo (não destruído)
- [ ] **Nenhum** push de imagem de app nem sync do `frontend/dist` neste chat (OK se task unhealthy por falta de imagem)

**Checkpoint**: IaC US2 completa. Parar. Abrir novo chat só para Fase D.

---

## Phase D — Publish da sessão (imagem + front) — CHAT 4

**Story**: US1 — Demo acadêmica na plataforma hospedada (+ fechamento US2 cenários de app)  
**Goal**: Imagem ARM64 na ECR, force deploy ECS, CORS alinhado, rebuild front com `VITE_API_URL` da sessão, sync S3 + invalidate CF.  
**Independent Test**: HTTPS front + API; login admin seed; SC-010/SC-011; amostragem RBAC.  
**FORA DESTE CHAT**: reescrever Terraform estrutural; novos recursos AWS grandes; CD (G); docs longas (E) — só scripts mínimos de publish se couberem em `infra/` ou `scripts/` sem virar runbook completo.

### Implementation

- [ ] T031 [US1] Script ou passos documentados curtos para login ECR + `docker build --platform linux/arm64` a partir de `backend/` + tag/push `:latest` — depends: Fase C Done — aceite: imagem no `ecr_repository_url` (research §4)
- [ ] T032 [US1] Ajustar `backend/Dockerfile` se necessário para build ARM64/runtime compatível com Fargate 512 MiB — depends: T031 — aceite: container sobe na task sem OOM imediato no boot
- [ ] T033 [US1] Setar/atualizar `CORS_ORIGIN` (SSM/env) = origem de `frontend_url` e `aws ecs update-service --force-new-deployment` — depends: T031, outputs Fase C — aceite: task nova puxa imagem; eventualmente healthy
- [ ] T034 [US1] Esperar estabilização: `GET {api_url}/health` → 200; migrate/seed nos logs CloudWatch sem fatal — depends: T033 — aceite: API HTTPS da sessão responde
- [ ] T035 [US1] Build frontend: `VITE_API_URL={api_url}` → `npm run build` em `frontend/` conforme [contracts/frontend-publish.md](./contracts/frontend-publish.md) — depends: T034 — aceite: bundle aponta só para API desta sessão
- [ ] T036 [US1] `aws s3 sync frontend/dist/` → `s3_bucket_name` + invalidação CloudFront (`cloudfront_frontend_distribution_id`, paths `/*`) — depends: T035 — aceite: `frontend_url` serve a SPA nova
- [ ] T037 [US1] Verificação SC-010: login admin seed + listar ≥1 curso via API/UI — depends: T034, T036 — aceite: sem cadastro manual
- [ ] T038 [US1] Verificação SC-011 + RBAC amostral (professor/aluno) na UI hospedada — depends: T037 — aceite: SC-003/SC-004 amostral; front↔API mesma sessão

### Done — Fase D (obrigatório antes do Chat 5)

- [ ] `{api_url}/health` = 200 (HTTPS)
- [ ] `{frontend_url}` carrega SPA e chama `api_url` (sem mixed content / CORS quebrado)
- [ ] Login admin seed + curso exemplo (SC-010)
- [ ] Login pela UI hospedada OK (SC-011)
- [ ] Amostra RBAC (ação permitida + bloqueada em PT) OK
- [ ] Sem automação CD nem destroy documentado completo neste chat

**Checkpoint**: Demo US1 utilizável. Parar. Abrir novo chat só para Fase E.

---

## Phase E — Runbook + docs + custo/destroy — CHAT 5

**Story**: Cross-cutting P1 (FR-017, SC-006/007/012) — sem misturar com apply/publish de código  
**Goal**: Operador segue apply → publish → warm-up → destroy sem inventar passos; custo e billing alert documentados.  
**Independent Test**: Outra pessoa (ou checklist frio) executa só com README/AGENTS/quickstart.  
**FORA DESTE CHAT**: mudar IaC de recursos; novo seed; CD; automação EventBridge (F).

### Implementation

- [ ] T039 [P] Incorporar runbook de [quickstart.md](./quickstart.md) em `README.md` (seção AWS demo) com ordem: billing → apply → push API → CORS → health → rebuild front → warm-up → demo → destroy — depends: Fase D Done — aceite: FR-017; links aos contracts
- [ ] T040 [P] Atualizar `AGENTS.md` com caminho AWS (ECS/S3/CF, destroy entre sessões, sem NAT, sem Redis AWS, CD=P3) alinhado à constitution — depends: Fase D Done — aceite: agentes não sugerem self-hosted como demo
- [ ] T041 Estimativa de custo por sessão (~US$ 1–3 / 4 h; ~US$ 0 após destroy) em `README.md` ou `specs/001-aws-mvp-terraform/quickstart.md` — depends: T039 — aceite: SC-006; NAT = US$ 0 explícito
- [ ] T042 Checklist pós-destroy (console) + resíduos (ECR force_delete, snapshots, log groups, **budget permanece**) no runbook — depends: T039 — aceite: SC-007 ≤ 15 min de procedimento
- [ ] T043 Documentar limiar do budget/alerta e onde verificar na console (Billing → Budgets) — depends: T017 (já em B), T039 — aceite: SC-012
- [ ] T044 Documentar gap P2 (automação warm-up) e P3 (CD) como fora do caminho mínimo — depends: T039 — aceite: US4/US5 não confundidos com P1
- [ ] T045 [P] Revisar ausência de segredos em arquivos rastreados (`.tfvars`, `.env`, state) — depends: T012 — aceite: SC-008
- [ ] T046 Validar quickstart ponta a ponta contra a última sessão (ou dry-run checklist se destroy já feito) — depends: T039–T044 — aceite: operador não inventa passos

### Done — Fase E (fecha P1)

- [ ] README + AGENTS atualizados
- [ ] Custo por sessão + destroy + billing alert documentados
- [ ] Gap P2/P3 explícito
- [ ] SC-006, SC-007, SC-008, SC-012 endereçados na doc
- [ ] **P1 completo** — Fases F/G só se desejado

**Checkpoint**: MVP P1 documentado. Não iniciar F/G no mesmo chat.

---

## Phase F — Warm-up automation ou gap (P2) — CHAT 6 — OPCIONAL / DESEJÁVEL

**Story**: US4 — Ciclo apply → warm-up → destroy  
**Goal**: Automação mínima de agendamento/warm-up **ou** gap explicitamente documentado com procedimento manual reforçado.  
**Independent Test**: Runbook apply → health → (opcional scale) → destroy; console sem recursos cobráveis ociosos.  
**Pré-requisito**: P1 (A–E) completo.  
**FORA DESTE CHAT**: CD GitHub Actions (G); redesign de rede.

### Implementation

- [ ] T047 [US4] Decidir e registrar: (a) automação (ex. EventBridge + docs) **ou** (b) gap P2 apenas documentado — depends: Fase E Done — aceite: escolha única no README/quickstart
- [ ] T048 [US4] Se (a): implementar artefato mínimo de agendamento/warm-up alinhado a desired_count=1 e health check; se (b): expandir checklist T−15 min no quickstart — depends: T047 — aceite: US4 AC3 (doc clara P2 vs entregue)
- [ ] T049 [US4] Validar ciclo destroy + verificação pós-destroy uma vez com o runbook atualizado — depends: T048 — aceite: SC-007; custo contínuo ~0

### Done — Fase F

- [ ] Automação entregue **ou** gap P2 inequívoco na doc
- [ ] Ciclo warm-up operacional verificável
- [ ] Sem CD neste chat

---

## Phase G — CD no GitHub Actions (P3) — CHAT 7 — OPCIONAL

**Story**: US5 — Pipeline de deploy contínuo  
**Goal**: Após CI verde, publicar imagem API e/ou sync front a partir do Actions (credenciais em secrets).  
**Independent Test**: Commit na branch de deploy atualiza ambiente; falha de test bloqueia promote.  
**Pré-requisito**: P1 completo; idealmente stack aplicável.  
**FORA DESTE CHAT**: streaming/chat/OAuth; substituir caminho manual P1 (deve continuar documentado).

### Implementation

- [ ] T050 [US5] Desenhar job(s) CD em `.github/workflows/` (separado ou extendido de `ci.yml`) com secrets AWS — depends: Fase E Done — aceite: CI build+test continua obrigatório e bloqueia deploy
- [ ] T051 [US5] Job push imagem ARM64 → ECR + force deploy ECS — depends: T050 — aceite: US5 AC1 parcial API
- [ ] T052 [P] [US5] Job build front com `VITE_API_URL` (secret/var da sessão ou output) + S3 sync + invalidate CF — depends: T050 — aceite: US5 AC1 parcial front (notar URLs efêmeras pós-destroy)
- [ ] T053 [US5] Garantir que falha de test/build não promove — depends: T051, T052 — aceite: US5 AC2
- [ ] T054 [US5] Documentar no README: CD opcional vs publish manual P1 — depends: T053 — aceite: caminho manual permanece válido

### Done — Fase G

- [ ] CD documentado e, se habilitado, bloqueado por CI vermelho
- [ ] Publish manual P1 ainda descrito
- [ ] Sem escopo fora de US5

---

## Dependencies & Execution Order

### Phase / chat dependencies

```text
A (US3) ──► B (US2 base) ──► C (US2 compute) ──► D (US1 publish) ──► E (docs P1)
                                                                      │
                                                         ┌────────────┴────────────┐
                                                         ▼                         ▼
                                              F (US4 P2 opcional)       G (US5 P3 opcional)
```

- **A → B → C → D → E** estritamente sequenciais (chats separados).
- **F** e **G** só depois de **E**; não bloqueiam P1; podem ser chats distintos e em qualquer ordem entre si.
- **Nunca** paralelizar chats de fases diferentes no mesmo agente/contexto.

### User story mapping

| Story | Prioridade | Fases |
|-------|------------|-------|
| US3 Baseline local/CI | P1 | A |
| US2 IaC reproduzível | P1 | B + C |
| US1 Demo hospedada | P1 | D (verificação; depende B+C) |
| Docs/custo/billing | P1 | E |
| US4 Warm-up ciclo | P2 | F (opcional) |
| US5 CD | P3 | G (opcional) |

### Parallel opportunities (dentro do mesmo chat/fase)

- **A**: T003 ∥ T006 após T001; T002 ∥ T003
- **B**: T012 ∥ T015 após T011; budget T017 paralelo a VPC após skeleton
- **C**: T021 ∥ T025 após B; T026 após T025; T027 após T022
- **E**: T039 ∥ T040 ∥ T045
- **G**: T051 e T052 após T050 (com cuidado de secrets compartilhados)

### Within each phase

- Respeitar `depends:` nas tasks
- Completar checklist **Done** antes de fechar o chat
- Não puxar tasks da fase seguinte “porque sobrou contexto”

---

## Parallel Example: Phase C (mesmo chat)

```text
# Após VPC/SG da Fase B:
Task: T021 ECR
Task: T025 S3
# Depois:
Task: T022 ALB → T027 CloudFront API
Task: T025 → T026 CloudFront frontend
Task: T020 RDS → T024 ECS (com T021–T023)
```

---

## Implementation Strategy

### MVP First (só P1 — chats 1–5)

1. **Chat 1 — Fase A** apenas ← **começar aqui**
2. Chat 2 — Fase B
3. Chat 3 — Fase C
4. Chat 4 — Fase D
5. Chat 5 — Fase E
6. **STOP** — P1 demonstrável + documentado

### Depois (opcional)

7. Chat 6 — Fase F (P2)
8. Chat 7 — Fase G (P3)

### Sugestão ao operador

> Execute `/speckit-implement` com escopo explícito: **“Somente Fase A (T001–T010). Parar no Done da Fase A.”**  
> Não mencione Terraform/AWS neste primeiro chat.

---

## Notes

- Terraform flat em `infra/`; sem modules no MVP (constitution VIII)
- Dados efêmeros + seed a cada apply; sem backup entre sessões
- Sem NAT, sem Redis AWS, sem domínio customizado, sem streaming/chat/OAuth
- Budget permanece após destroy da demo
- Formato Speckit: checkbox + ID + [P]? + [Story]? + path + depends + aceite
- **Total tasks**: T001–T054 (54); P2 = T047–T049; P3 = T050–T054
