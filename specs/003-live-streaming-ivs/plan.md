# Implementation Plan: Streaming ao vivo da aula (AWS IVS)

**Branch**: `003-live-streaming-ivs` | **Date**: 2026-09-05 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/003-live-streaming-ivs/spec.md` + Clarifications (Session 2026-09-05). Baseline obrigatória: `001-aws-mvp-terraform` + `002-vod-library` (**não reabrir**).

## Summary

Acrescentar **transmissão ao vivo ligada a uma aula** sobre 001 (stack AWS efêmera) e 002 (VOD na ficha). **Um canal IVS compartilhado** por sessão Terraform; **no máximo uma live ativa**. P1: ação única **iniciar live nesta aula** (associa + marca ao vivo + mostra ingest OBS); encerrar; reinício na mesma sessão com a **mesma stream key**. Playback **somente via API JWT** (sem URL pública permanente; tokenização IVS fora do P1). Local/CI: **stub** (estados/RBAC/erros PT, sem vídeo). UI distingue live vs VOD; **sem** pipeline live→VOD. Destroy da stack 001 remove o canal.

## Technical Context

**Language/Version**: Go 1.25 (API); TypeScript / Node 22 (frontend Vite); Terraform ≥ 1.5; AWS CLI; OBS (encoder externo, não no repo)

**Primary Dependencies**: Fiber, GORM, gormigrate, JWT+bcrypt; React + Vite; **Amazon IVS Player** (Web SDK) na ficha; AWS provider (`aws_ivs_channel` + stream key; **sem** ElastiCache/NAT/Cognito/domínio/IVS Chat/recording)

**Storage**: PostgreSQL (metadados `AulaLive` ↔ aula; invariante ≤1 `ao_vivo`); canal IVS **não** é tabela — config da sessão via env/SSM. Stream key: **SSM SecureString** (não no git, não no DB)

**Testing**: `go test` em handlers/repositório live (stub; SQLite `:memory:`); `npm run build`; CI existente; E2E AWS + OBS via [quickstart.md](./quickstart.md)

**Target Platform**: AWS `us-east-1` (demo efêmera, vídeo real); Docker Compose / API local (dev/CI = stub)

**Project Type**: Monorepo web (backend + frontend) + IaC flat em `infra/`

**Performance Goals**: Aluno inicia reprodução in-app ≤ 2 min após login (SC-001/002); vídeo/áudio ≤ 15 s após “assistir” com sinal já no ar (SC-007); demo com dezenas de avaliadores (não load test de milhares)

**Constraints**: 1 canal / 1 live ativa; ingest só admin/professor; playback só JWT; stub local; destroy zera IVS; sem NAT/Redis AWS/OAuth/domínio custom; sem live→VOD; sem chat; `Aula.Status` CRUD **não** é o estado da live; datas BR; erros PT; budget 001 intocado

**Scale/Scope**: 1 canal BASIC LOW-latency; extensão enxuta `infra/ivs.tf` + domínio `live` no backend + bloco na `AulasPage` (sem rota “Ao vivo”)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Princípio / constraint | Status | Como o plan atende |
|------------------------|--------|-------------------|
| I. Custo-consciente | PASS | Canal **BASIC**; idle ≈ US$ 0 de input/output IVS; cobrança só com ingest/viewers; `destroy` remove canal/key; sem NAT/Redis; budget separado |
| II. Pronto para o pico | PASS | Checklist T−15 do 001 **estendido** (canal provisionado + smoke live **antes** da janela); cold start na abertura inaceitável; automação EventBridge 001 permanece gap; warm-up só streaming = P2/P3 opcional |
| III. Escopo mínimo / YAGNI | PASS | 1 canal; ação única P1; stub local; sem lives simultâneas, chat IVS, recording, playback authorization, player nativo, STANDARD/ABR |
| IV. Segurança e RBAC | PASS | JWT; gestão = escrita de aulas; ingest nunca ao aluno; playback só API autenticada; stream key no SSM; erros PT |
| V. Qualidade testável | PASS | Testes Go nos handlers live (stub); CI build+test |
| VI. Observabilidade suficiente | PASS | CloudWatch da task 001; sem APM; health `/health` inalterado |
| VII. Documentação / Speckit | PASS | Artefatos nesta pasta; README/AGENTS na fase de docs (implement futura) |
| VIII. IaC | PASS | Terraform flat `infra/ivs.tf` + SSM/ECS/IAM mínimos; sem módulos novos |
| Stack fechada AWS | PASS | IVS gerenciado `us-east-1`; Go/React/Postgres; auth JWT |
| Self-hosted como demo | PASS | Compose/stub só para dev/CI; vídeo real só AWS |
| Baseline 001/002 | PASS | Sem NAT, sem Redis AWS, destroy entre sessões, VOD um-por-aula intacto, sem página Biblioteca |

**Post-design re-check:** PASS — contracts/data-model não reabrem 001/002; canal `authorized=false` (tokenização fora do P1) com gate JWT na API; residual de HLS “descoberta se vazar URL” documentado em research; sem CloudFront de mídia extra; sem recording IVS→S3.

## Project Structure

### Documentation (this feature)

```text
specs/003-live-streaming-ivs/
├── plan.md                 # Este arquivo
├── research.md             # Phase 0
├── data-model.md           # Phase 1
├── quickstart.md           # Phase 1
├── contracts/
│   ├── live-api.md
│   ├── live-env.md
│   └── terraform-ivs.md
├── checklists/
│   └── requirements.md
└── tasks.md                # NÃO criado aqui → /speckit-tasks
```

### Source Code (repository root — extensões previstas)

```text
estuda-ja/
├── backend/
│   ├── cmd/api/main.go                 # rotas /aulas/:id/live*
│   └── internal/
│       ├── config/                     # LIVE_BACKEND, IVS_*
│       ├── database/migrations/        # AutoMigrate AulaLive (sem seed de sinal)
│       ├── handler/live_handler.go     # NOVO (domínio)
│       ├── handler/live_handler_test.go
│       ├── repository/live_repository.go
│       └── models/aula_live.go         # NOVO
├── frontend/
│   └── src/
│       ├── pages/AulasPage.tsx         # bloco Live (distinto do VOD)
│       ├── components/LivePlayer.tsx   # NOVO — IVS Player (AWS); stub = mensagem
│       ├── api/client.ts               # tipos/chamadas live
│       └── auth/auth.ts                # reutilizar canManageAulas
├── infra/
│   ├── ivs.tf                          # NOVO — canal + stream key
│   ├── ssm.tf                          # + IVS_STREAM_KEY (SecureString) e params de ingest/playback
│   ├── iam.tf                          # execution role: GetParameter dos novos SSM
│   ├── ecs.tf                          # env LIVE_BACKEND=ivs + secrets/env IVS_*
│   └── outputs.tf                      # ivs_channel_arn (sem stream key / sem playback_url)
├── docker-compose.yml                  # LIVE_BACKEND=stub (sem IVS)
├── README.md                           # runbook: OBS + warm-up streaming + destroy
└── AGENTS.md                           # streaming deixa de ser “fora do escopo” para esta feature
```

**Structure Decision:** Monorepo atual; domínio **live** em arquivos próprios (`live_*` / `aula_live.go`), rotas aninhadas sob aula (espelho VOD); player só na ficha existente; Terraform flat só com **um** canal IVS + SSM + env da task. Sem segundo CloudFront, sem IVS Chat, sem recording configuration.

## Complexity Tracking

> Nenhuma violação de constitution a justificar. Tabela omitida.

## Implementation Phases (chats futuros — sem `tasks.md` aqui)

Cada fase cabe em **um chat de implement**. Critério **Done** verificável. Ordem 1→4; fase 5 opcional.

| Fase | Escopo | Done quando |
|------|--------|-------------|
| **1 — Modelo / API / stub + testes** | `AulaLive`, handlers start/stop/status/playback/ingest, invariante 1 live ativa, RBAC, horário opcional na API, `LIVE_BACKEND=stub`, erros PT, testes Go | `go test` verde; professor inicia/encerra/reinicia no stub; aluno 403 em escrita/ingest; 409 se segunda aula ao vivo; playback stub **sem** URL de vídeo; aula sem `agendada_em` inicia; `Aula.Status` CRUD inalterado |
| **2 — UI ficha aula** | Em `AulasPage`: bloco **Ao vivo** distinto de **Gravação**; player (IVS na AWS / mensagem no stub); controles iniciar/encerrar só `canManageAulas`; ingest (endpoint+key) só gestor com live ativa; UI recomenda horário sem bloquear; sem rota “Ao vivo” | Aluno vê live vs ausência vs VOD; aluno não vê ingest nem botões; professor inicia e vê credenciais OBS; player vazio enganoso **não** aparece sem live |
| **3 — Terraform IVS + wire AWS** | `ivs.tf` canal BASIC LOW; stream key → SSM; env ECS; outputs só ARN; `LIVE_BACKEND=ivs`; destroy remove canal | `terraform apply` cria 1 canal; publish API; iniciar live na aula demo mostra ingest real; aluno assiste com OBS; `destroy` apaga IVS; budget intacto |
| **4 — Runbook / docs + warm-up streaming** | README + AGENTS + quickstart 001/003: apply→publish→**warm-up streaming**→OBS→demo→destroy; checklist T−15 estendido | Operador percorre runbook ≤ 15 min incremental (SC-005); T−0 sem cold start da stack nem do canal (SC-006); AGENTS deixa de listar streaming como proibido sem spec |
| **5 — P2 estados / P2–P3 automação (opcional)** | UI/API: `agendada` / `ao_vivo` / `encerrada` (agendada só com horário); MAY apontar VOD 002 se existir; automação **só** health/sinal de streaming (sem EventBridge apply/destroy do 001) | Aluno distingue os 3 estados; docs deixam claro o que é manual vs automatizado; 001 P2 EventBridge permanece gap |

**Não fazer neste plan:** implementar código, gerar `tasks.md`, reabrir decisões 001/002.

## Next Actions

1. **`/speckit-tasks`** — gerar `tasks.md` com **fases = chats** (1–4 obrigatórias, 5 opcional), dependências explícitas.
2. **Não** iniciar `/speckit-implement` até as tasks existirem (salvo pedido explícito).
