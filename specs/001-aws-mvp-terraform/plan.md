# Implementation Plan: MVP AWS com custo controlado

**Branch**: `001-aws-mvp-terraform` | **Date**: 2026-09-04 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/001-aws-mvp-terraform/spec.md`

## Summary

Endurecer a entrega já existente (Go/Fiber + Vite/React + JWT/RBAC + Postgres) e provisionar um **ambiente de demo efêmero na AWS** com Terraform enxuto: frontend estático (S3 + CloudFront), API (ECS Fargate + ALB + CloudFront para HTTPS), RDS PostgreSQL `db.t4g.micro`, segredos no SSM Parameter Store, região `us-east-1`, **sem NAT Gateway** e **sem Redis AWS**. Operação padrão: `apply` → publish API → rebuild/publish front com `VITE_API_URL` da sessão → warm-up → demo → `destroy` completo. CI permanece obrigatório; CD é P3. Automação de agendamento de warm-up é P2.

## Technical Context

**Language/Version**: Go 1.25 (API); TypeScript / Node 22 (frontend Vite); Terraform ≥ 1.5; AWS CLI

**Primary Dependencies**: Fiber, GORM, gormigrate, JWT+bcrypt; React + Vite; AWS provider (S3, CloudFront, ECS, ECR, ALB, RDS, SSM, VPC, Budgets/SNS)

**Storage**: RDS PostgreSQL clássico (`db.t4g.micro`, single-AZ); S3 para assets do frontend; sem ElastiCache nesta feature

**Testing**: `go test ./...` (handlers); `npm run build` no frontend; CI GitHub Actions (build + test + docker build); validação E2E da demo via runbook (quickstart), não suite nova obrigatória em P1

**Target Platform**: AWS `us-east-1` (demo); Docker Compose local (dev)

**Project Type**: Monorepo web (backend API + frontend SPA) + IaC

**Performance Goals**: Demo estável com 1 task Fargate 0.25 vCPU / 512 MiB; login + listagem de cursos ≤ 2 min/perfil (SC-003); stack saudável antes da janela (sem cold start)

**Constraints**: Destroy entre sessões; sem domínio customizado; sem NAT; Parameter Store preferido; custos contínuos justificados; RBAC e datas BR inalterados; YAGNI (sem streaming/chat/OAuth); HTTPS via hostnames AWS (CloudFront)

**Scale/Scope**: Conta acadêmica Free/créditos; 1 ambiente de demo por sessão; CRUD existente; novos artefatos principais = `infra/` + seed demo + docs operacionais + ajustes de config/CORS/deploy

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Princípio / constraint | Status | Como o plan atende |
|------------------------|--------|-------------------|
| I. Custo-consciente | PASS | Destroy entre sessões; sem NAT; Fargate mínimo ARM; RDS micro; SSM; budget alert; estimativa por sessão |
| II. Pronto para o pico | PASS | Warm-up operacional documentado (P1/P2); desired_count=1 antes da janela; CDN CloudFront; gap de automação explícito em P2 |
| III. Escopo mínimo / YAGNI | PASS | Sem streaming, chat, OAuth, Redis AWS, módulos TF elaborados, CD (P3) |
| IV. Segurança e RBAC | PASS | JWT+bcrypt e papéis inalterados; segredos no SSM; erros PT; CORS por sessão |
| V. Qualidade testável | PASS | CI build+test obrigatório; testes Go ao alterar handlers; CD não isenta CI |
| VI. Observabilidade suficiente | PASS | CloudWatch Logs + métricas default; sem APM |
| VII. Documentação / Speckit | PASS | Artefatos em `specs/001-aws-mvp-terraform/`; README/AGENTS a atualizar na implement |
| VIII. IaC | PASS | Terraform enxuto em `infra/`; sem provisionamento manual permanente |
| Stack fechada AWS | PASS | S3+CloudFront, ECS Fargate+ALB (+CF TLS), RDS PG, Go/React |
| Self-hosted como demo | PASS | Apenas Compose local para dev |

**Post-design re-check:** PASS — contracts e data-model não introduzem serviços contínuos injustificados nem redesenho de domínio.

## Project Structure

### Documentation (this feature)

```text
specs/001-aws-mvp-terraform/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── api-env.md
│   ├── terraform-outputs.md
│   └── frontend-publish.md
├── checklists/
│   └── requirements.md
└── tasks.md                 # NÃO criado neste comando (/speckit-tasks)
```

### Source Code (repository root)

```text
estuda-ja/
├── backend/
│   ├── cmd/api/main.go
│   ├── Dockerfile
│   └── internal/
│       ├── config/          # env: DATABASE_URL, JWT_*, CORS_ORIGIN, ADMIN_*
│       ├── database/migrations/  # + seed demo (nova migration)
│       ├── handler/
│       ├── repository/
│       ├── auth/
│       └── models/          # domínio existente (sem redesign)
├── frontend/
│   ├── Dockerfile           # build arg VITE_API_URL (publish AWS usa dist→S3)
│   └── src/
│       ├── api/client.ts    # VITE_API_URL
│       └── auth/            # RBAC UI
├── infra/                   # NOVO — Terraform root (us-east-1)
│   ├── main.tf
│   ├── variables.tf
│   ├── outputs.tf
│   ├── providers.tf
│   └── ...                  # vpc, ecs, alb, rds, s3, cloudfront, ssm, ecr
├── docs/
├── .github/workflows/ci.yml # CI build+test
├── .github/workflows/cd.yml # CD opcional P3 (gate + ECR/ECS + S3/CF)
├── docker-compose.yml
├── Makefile
├── README.md
├── AGENTS.md
├── DESCRICAO.md
└── REQUISITOS.md
```

**Structure Decision:** Monorepo real atual (`backend/`, `frontend/`, `docs/`) + nova pasta `infra/` para Terraform flat. Sem opções alternativas de layout; domínio de app permanece um arquivo por domínio em handler/repository.

## Complexity Tracking

> Nenhuma violação de constitution a justificar. Tabela omitida.

## Phases (design → implementação futura)

| Fase | Entrega | Prioridade |
|------|---------|------------|
| Design (este comando) | plan, research, data-model, contracts, quickstart | — |
| P1 | IaC mínima + app configurável + seed demo + publish manual + CI verde + budget + docs | Obrigatório |
| P2 | Automação/agendamento de warm-up (ou gap documentado) | Desejável |
| P3 | CD (push imagem + sync S3 a partir do CI) | Opcional — `cd.yml` entregue; manual P1 permanece |

**Próximo comando Speckit:** `/speckit-tasks` (não executar implementação aqui).
