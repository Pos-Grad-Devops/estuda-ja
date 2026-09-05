# Implementation Plan: Biblioteca VOD (vídeos gravados sob demanda)

**Branch**: `002-vod-library` | **Date**: 2026-09-04 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/002-vod-library/spec.md` + Clarifications (Session 2026-09-04). Baseline obrigatória: `001-aws-mvp-terraform` (não reabrir).

## Summary

Acrescentar **um VOD vigente por aula** (MP4 H.264/AAC, ≤ 50 MB) sobre a fundação 001: admin/professor publicam/substituem/removem (apagando o objeto antigo na hora); aluno/professor/admin autenticados assistem **só na ficha de aula existente** via reprodutor in-app (sem download, sem página Biblioteca). Persistência de metadados no Postgres; bytes no **S3 efêmero** da mesma stack Terraform (local: disco/volume Compose). Playback via **URL temporária ~15 min** após JWT. Seed recria 1 MP4 na aula de demo a cada apply. P2 (rascunho/metadados) não bloqueia P1.

## Technical Context

**Language/Version**: Go 1.25 (API); TypeScript / Node 22 (frontend Vite); Terraform ≥ 1.5; AWS CLI

**Primary Dependencies**: Fiber, GORM, gormigrate, JWT+bcrypt, AWS SDK v2 (S3); React + Vite; AWS provider (extensão S3 + IAM task; **sem** ElastiCache/NAT/domínio)

**Storage**: PostgreSQL (metadados VOD ↔ aula); objetos: S3 bucket **novo** efêmero na stack 001 (`force_destroy`); local: diretório/volume (`VOD_LOCAL_DIR`)

**Testing**: `go test` em handlers/repositório VOD (+ integração local storage); `npm run build`; CI existente; E2E demo via [quickstart.md](./quickstart.md)

**Target Platform**: AWS `us-east-1` (demo efêmera); Docker Compose / API local (dev)

**Project Type**: Monorepo web (backend + frontend) + IaC flat em `infra/`

**Performance Goals**: Início de playback ≤ 10 s (SC-006); jornadas ≤ 2 min (SC-001/002); demo com dezenas de avaliadores (não pico magna)

**Constraints**: Teto **50 MB**; só **MP4 H.264/AAC**; playback **~15 min**; sem download; sem transcodificação; destroy apaga objetos VOD; sem NAT/Redis AWS/OAuth/página Biblioteca; RBAC = escrita aulas; datas BR; erros PT

**Scale/Scope**: 1 objeto vigente/aula; seed 1 clipe; extensão enxuta de `infra/` + domínio `vod` no backend + UI em `AulasPage` (e contexto curso se já exibir aula)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Princípio / constraint | Status | Como o plan atende |
|------------------------|--------|-------------------|
| I. Custo-consciente | PASS | Bucket VOD na stack efêmera + `force_destroy`; sem CF de mídia permanente; clipes ≤ 50 MB; destroy zera objetos; budget 001 intocado |
| II. Pronto para o pico | PASS | VOD **não** substitui ao vivo; sem warm-up extra além do runbook 001 |
| III. Escopo mínimo / YAGNI | PASS | Sem Biblioteca dedicada, transcode, multi-qualidade, download, P2 obrigatório |
| IV. Segurança e RBAC | PASS | JWT; escrita admin+professor; playback só autenticado; URL ~15 min; sem URL permanente pública |
| V. Qualidade testável | PASS | Testes Go em handlers VOD; CI build+test |
| VI. Observabilidade suficiente | PASS | Logs API existentes + CloudWatch task; sem APM |
| VII. Documentação / Speckit | PASS | Artefatos nesta pasta; README/AGENTS na fase de docs |
| VIII. IaC | PASS | Terraform flat estendendo `infra/` (sem módulos novos) |
| Stack fechada AWS | PASS | S3 objetos + ECS já existente; Go/React/Postgres |
| Self-hosted como demo | PASS | Compose/disco só para dev |

**Post-design re-check:** PASS — contracts/data-model não reabrem 001; playback via S3 presigned (sem distribuição CloudFront extra de mídia); upload multipart pela API unifica local/AWS.

## Project Structure

### Documentation (this feature)

```text
specs/002-vod-library/
├── plan.md                 # Este arquivo
├── research.md             # Phase 0
├── data-model.md           # Phase 1
├── quickstart.md           # Phase 1
├── contracts/
│   ├── vod-api.md
│   ├── vod-env.md
│   └── terraform-vod.md
├── checklists/
│   └── requirements.md
└── tasks.md                # NÃO criado aqui → /speckit-tasks
```

### Source Code (repository root — extensões previstas)

```text
estuda-ja/
├── backend/
│   ├── assets/vod/                    # NOVO — MP4 seed (pequeno, ≤ teto)
│   ├── cmd/api/main.go                # rotas /aulas/:id/vod*
│   ├── Dockerfile                     # volume/path assets se necessário
│   └── internal/
│       ├── config/                    # VOD_* env
│       ├── database/migrations/       # 003_seed_vod_demo (ou similar)
│       ├── handler/vod_handler.go     # NOVO (domínio)
│       ├── repository/vod_repository.go
│       ├── models/                    # AulaVod (ou equivalente)
│       └── vodstorage/                # NOVO — Local + S3 (Put/Delete/Presign)
├── frontend/
│   └── src/
│       ├── pages/AulasPage.tsx        # player + upload gestor (sem rota nova)
│       ├── api/client.ts              # tipos/chamadas VOD
│       └── auth/auth.ts               # reutilizar canManageAulas
├── infra/
│   ├── s3_vod.tf                      # NOVO — bucket efêmero
│   ├── iam.tf                         # policy s3 na task role
│   ├── ecs.tf                         # env VOD_*
│   ├── outputs.tf                     # vod_bucket_name
│   └── variables.tf                   # se necessário (TTL etc.)
├── docker-compose.yml                 # volume ./data/vod → container
├── README.md                          # runbook VOD
└── AGENTS.md                          # menção VOD se stack/auth mudar
```

**Structure Decision:** Monorepo atual; domínio **VOD** em arquivos próprios (`vod_*`), rotas aninhadas sob aula; storage atrás de interface local/S3; Terraform flat só com bucket + IAM + env (sem segundo CloudFront de mídia — ver research).

## Complexity Tracking

> Nenhuma violação de constitution a justificar. Tabela omitida.

## Implementation Phases (chats futuros — sem `tasks.md` aqui)

Cada fase é independente o bastante para **um chat de implement**. Critério **done** verificável abaixo. Ordem sugerida 1→4; fase 5 opcional.

| Fase | Escopo | Done quando |
|------|--------|-------------|
| **1 — Modelo / API / storage local + testes** | `AulaVod`, `vodstorage` local, handlers upload/replace/delete/metadata/playback, validação MP4+50MB, delete objeto antigo, RBAC, testes Go | `go test` verde; curl/local: professor sobe MP4, aluno obtém playback URL, aluno não escreve; arquivo >50MB / não-MP4 rejeitado |
| **2 — UI ficha aula** | Em `AulasPage` (e contexto curso se couber): indicar VOD, `<video>` com URL de playback, formulário upload/remover só se `canManageAulas`; sem rota Biblioteca; sem download | Login aluno vê/toca seed ou upload; professor publica pela UI; aluno não vê botões de escrita |
| **3 — Terraform S3 + wire AWS** | `s3_vod.tf`, IAM task, env ECS, outputs; API em modo S3; destroy limpa objetos | `terraform apply` cria bucket; publish API; upload/playback na sessão; `destroy` remove bucket/objetos |
| **4 — Seed asset + runbook/docs** | Asset MP4 no repo + migration seed; README/quickstart/AGENTS; passos apply→publish→demo→destroy | Pós-apply aluno assiste sem upload manual (SC-005); runbook ≤ 10 min incremental (SC-008) |
| **5 — P2 (opcional)** | Estados rascunho/publicado; título/duração visíveis | Aluno não reproduz rascunho; publish torna assistível |

**Não fazer neste plan:** implementar código, gerar `tasks.md`, reabrir decisões 001.

## Next Actions

1. **`/speckit-tasks`** — gerar `tasks.md` com fases = chats (1–4 obrigatórias, 5 opcional), dependências explícitas.
2. **Não** iniciar `/speckit-implement` até as tasks existirem (salvo pedido explícito).
