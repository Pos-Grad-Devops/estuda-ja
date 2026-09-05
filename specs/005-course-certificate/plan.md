# Implementation Plan: Certificado ao final do curso

**Branch**: `005-course-certificate` | **Date**: 2026-09-05 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/005-course-certificate/spec.md` + Clarifications (Session 2026-09-05). Baseline obrigatória: `001`–`004` (**não reabrir**).

## Summary

Acrescentar **certificado PDF de conclusão de curso** sobre 001–004: admin marca **elegibilidade** (par User `aluno`–Curso); a **primeira** solicitação do aluno elegível cria registro `valido` + PDF; downloads seguintes reutilizam o ativo; invalidar remove elegibilidade até reabilitação; depois nova solicitação = novo certificado. PDF **on-the-fly** (sem S3 de certificado). UI: painel mínimo em `CursosPage` (sem página “Certificados”). Seed: só elegibilidade no curso demo. P2 (template rico / lista) **fora**. Gestão **só admin**.

## Technical Context

**Language/Version**: Go 1.25 (API); TypeScript / Node 22 (frontend Vite); Terraform ≥ 1.5 (sem mudança obrigatória nesta feature)

**Primary Dependencies**: Fiber, GORM, gormigrate, JWT+bcrypt; **`github.com/go-pdf/fpdf`** + TTF embutida (acentos PT); React + Vite. **Sem** S3/certs, **sem** Redis AWS, **sem** Cognito, **sem** Chromium/wkhtmltopdf

**Storage**: PostgreSQL — tabelas `certificado_elegibilidades` + `certificados`. Artefato PDF **não** persistido (geração on-the-fly a partir do registro)

**Testing**: `go test` em handlers/RBAC/lazy emit/invalidação (SQLite `:memory:`); `npm run build`; CI existente **sem** AWS. Validação E2E: [quickstart.md](./quickstart.md)

**Target Platform**: AWS `us-east-1` (demo efêmera, mesma stack 001); Docker Compose / API local (dev)

**Project Type**: Monorepo web (backend + frontend) + IaC flat em `infra/` (certificado P1 **sem** recurso Terraform novo)

**Performance Goals**: SC-001 (PDF ≤ 2 min após login no caminho feliz); certificado **fora** do T−0 da live (sem cold start na abertura)

**Constraints**: Elegibilidade = flag admin (sem matrícula/aulas); emissão lazy; só admin gerencia; identidade = User aluno (não CRUD Alunos); PDF baixável; on-the-fly; erros PT; datas BR; destroy remove registros com RDS; sem NAT/Redis/Cognito; budget intocado; P2 fora; baseline 001–004 intacta

**Scale/Scope**: Demo acadêmica; 1 certificado ativo por par user–curso; seed 1 elegibilidade; UI só no contexto do curso

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Princípio / constraint | Status | Como o plan atende |
|------------------------|--------|-------------------|
| I. Custo-consciente | PASS | Sem bucket/S3/IAM de certificado; PDF on-the-fly; custo = mesma task ECS + RDS; `destroy` zera registros; budget intocado |
| II. Pronto para o pico | PASS | Fora do caminho crítico T−0; sem warm-up extra; não altera IVS/chat |
| III. Escopo mínimo / YAGNI | PASS | Só P1; flag admin; PDF simples; sem página dedicada; P2 excluído |
| IV. Segurança e RBAC | PASS | JWT existente; gestão só admin; aluno só próprio PDF; erros PT; sem Cognito |
| V. Qualidade testável | PASS | Testes Go em handlers/contratos/RBAC; CI build+test sem AWS |
| VI. Observabilidade suficiente | PASS | CloudWatch/task logs 001; sem APM novo |
| VII. Documentação / Speckit | PASS | Artefatos nesta pasta; README/AGENTS na fase 3 (implement futura) |
| VIII. IaC | PASS | Sem Terraform novo no P1 (justificado: sem artefato S3); stack flat existente intocada para cert |
| Stack fechada AWS | PASS | Go/Fiber/React/Postgres; JWT própria; `us-east-1` |
| Self-hosted como demo | PASS | Compose = dev; demo = AWS gerenciada |
| Baseline 001–004 | PASS | Sem reabrir VOD/live/chat/fundação AWS |

**Post-design re-check:** PASS — research resolve PDF lib + on-the-fly; data-model com elegibilidade + certificado; contracts API + env (sem `CERT_*`); quickstart admin→aluno→destroy; sem violação a justificar.

## Project Structure

### Documentation (this feature)

```text
specs/005-course-certificate/
├── plan.md                 # Este arquivo
├── research.md             # Phase 0
├── data-model.md           # Phase 1
├── quickstart.md           # Phase 1
├── contracts/
│   ├── certificado-api.md  # Endpoints / RBAC
│   └── certificado-env.md  # Env / destroy (sem S3)
├── checklists/
│   └── requirements.md
└── tasks.md                # NÃO criado aqui → /speckit-tasks
```

### Source Code (repository root — extensões previstas)

```text
estuda-ja/
├── backend/
│   ├── assets/certs/                         # NOVO — TTF (ex. DejaVuSans) para PDF PT
│   ├── cmd/api/main.go                       # rotas /cursos/:id/certificado*
│   ├── Dockerfile                            # COPY assets/certs se necessário
│   └── internal/
│       ├── models/                           # CertificadoElegibilidade, Certificado
│       ├── handler/certificado_handler.go    # NOVO
│       ├── handler/certificado_handler_test.go
│       ├── repository/certificado_repository.go  # NOVO
│       ├── certpdf/                          # NOVO — wrapper go-pdf/fpdf
│       └── database/migrations/              # 004_seed_certificado_demo
├── frontend/
│   └── src/
│       ├── pages/CursosPage.tsx              # painel certificado (aluno + admin)
│       ├── components/CursoCertificadoPanel.tsx  # NOVO (opcional se extrair)
│       ├── api/client.ts                     # chamadas certificado
│       └── auth/auth.ts                      # canManageCertificados (admin)
├── infra/                                    # SEM tf novo no P1 (on-the-fly)
├── README.md                                 # runbook: certificado + destroy
└── AGENTS.md                                 # certificado 005 deixa de ser “fora do escopo”
```

**Structure Decision:** Monorepo atual; domínio **certificado** em `handler/certificado_*` + `repository/certificado_*` + `certpdf/`; rotas aninhadas sob **curso**; UI só em `CursosPage`; **sem** recurso Terraform de storage. Identidade = `User`, não `Aluno` CRUD.

## Complexity Tracking

> Nenhuma violação de constitution a justificar. Tabela omitida.

## Implementation Phases (chats futuros — sem `tasks.md` aqui)

Cada fase cabe em **um chat de implement**. Critério **Done** verificável. Ordem 1→3. P2 = **fora**.

| Fase | Escopo | Done quando |
|------|--------|-------------|
| **1 — Model/API + PDF local + testes Go** | Models + migrations AutoMigrate; repositório; `certpdf` (fpdf+TTF); handlers elegibilidade / status / pdf lazy / list admin / invalidar; RBAC; erros PT; testes Go | `go test` verde; curl: admin marca elegibilidade (sem emitir); aluno 1ª `GET .../pdf` cria registro + PDF; 2ª reutiliza; invalidar bloqueia; reabilitar + nova solicitação = novo id; professor 403 em gestão |
| **2 — UI no contexto do curso** | Painel em `CursosPage` (ou componente embutido): aluno status + baixar; admin marcar/reabilitar, consultar, invalidar; `canManageCertificados`; sem rota “Certificados”; professor sem controles | Aluno seed (elegível) baixa PDF na UI do curso; admin invalida/reabilita pela UI; professor não vê gestão; live/VOD/chat intactos |
| **3 — Seed + wire AWS/docs** | Migration seed só elegibilidade no curso demo; README/AGENTS/quickstart no ciclo apply→publish→demo certificado→destroy; confirmar **zero** tf S3 de cert; nota destroy = RDS | Pós-seed: elegível sem certificado pré-emitido; demo AWS (se sessão up) mesmo fluxo; destroy documentado; budget intacto; operador ≤ 15 min incremental (SC-005) |

**Não fazer neste plan:** implementar código, gerar `tasks.md`, reabrir 001–004, P2 (template rico / lista), S3 de certificado, gestão por professor, página dedicada.

## Next Actions

1. **`/speckit-tasks`** — gerar `tasks.md` com **fases = chats** (1–3), dependências explícitas; P2 **excluído**.
2. **Não** iniciar `/speckit-implement` até as tasks existirem (salvo pedido explícito).
