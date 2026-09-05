# Implementation Plan: Chat da aula ao vivo (tempo real)

**Branch**: `004-live-class-chat` | **Date**: 2026-09-05 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/004-live-class-chat/spec.md` + Clarifications (Session 2026-09-05). Baseline obrigatória: `001-aws-mvp-terraform` + `002-vod-library` + `003-live-streaming-ivs` (**não reabrir**).

## Summary

Acrescentar **chat por aula** sobre 001–003: WebSocket no **mesmo** serviço de API Go (ECS Fargate / Compose), sala disponível **sempre que a aula existe** (independente da live), mensagens **efêmeras à conexão** (sem histórico ao reabrir), fan-out **em memória** por `aula_id` (`desired_count = 1`), auth via **JWT na query** (`?token=...`), identidade na UI pelo **nome** do usuário. Sem Redis/ElastiCache, sem API Gateway WebSocket + Lambda, sem moderação P2 nesta rodada. UI: painel embutido na `AulasPage` (distinto de live/VOD).

## Technical Context

**Language/Version**: Go 1.25 (API); TypeScript / Node 22 (frontend Vite); Terraform ≥ 1.5

**Primary Dependencies**: Fiber v2 + `github.com/gofiber/contrib/websocket` (fasthttp); GORM; JWT+bcrypt; React + Vite. **Sem** Redis client, **sem** API GW WS, **sem** Lambda, **sem** IVS Chat

**Storage**: **Sem tabela persistente de mensagens em P1.** Sala = mapa em memória por `aula_id`; mensagem = payload em voo. Metadados de aula/usuário reutilizam Postgres existente (validar aula + carregar `nome` no handshake)

**Testing**: `go test` em handlers/hub/RBAC (SQLite `:memory:` ou harness HTTP/WS); `npm run build`; CI existente **sem** AWS e **sem** multi-cliente obrigatório. Validação multi-browser: [quickstart.md](./quickstart.md) no Compose

**Target Platform**: AWS `us-east-1` (demo efêmera, WS atrás de CloudFront → ALB → ECS); Docker Compose local (WS real, mesmo protocolo)

**Project Type**: Monorepo web (backend + frontend) + IaC flat em `infra/`

**Performance Goals**: SC-001 (primeira mensagem ≤ 2 min após login); SC-002 (fan-out ≤ 5 s entre dois conectados); warm-up da API antes da T−0 (sem cold start do chat)

**Constraints**: Sala ≠ exige `ao_vivo`; efêmero à conexão; JWT query + `aula_id` no handshake; texto ≤ 500 chars; erros PT; sem NAT/Cognito/domínio custom/Redis AWS; destroy remove API/task; chat **não** amplia RBAC live/VOD; não logar query com token; moderação P2 **fora**; baseline 001–003 intacta

**Scale/Scope**: Demo acadêmica com dezenas de avaliadores; 1 task ECS; hub in-process. Escala multi-task / pub-sub **fora** do P1

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Princípio / constraint | Status | Como o plan atende |
|------------------------|--------|-------------------|
| I. Custo-consciente | PASS | Sem serviço extra (sem API GW WS, Lambda, ElastiCache); custo = mesma task ECS da sessão; `destroy` zera; budget `infra/budget/` intocado |
| II. Pronto para o pico | PASS | Chat sobe com a API da sessão; runbook exige API saudável (T−15) **antes** da abertura; ping/keepalive contra idle ALB/CloudFront; cold start na T−0 inaceitável |
| III. Escopo mínimo / YAGNI | PASS | Só P1 (enviar/receber + WS + CI/docs); sem histórico, moderação, anexos, Redis, API GW+Lambda |
| IV. Segurança e RBAC | PASS | JWT na query; rejeitar sem token/aula inválida/sem leitura; papéis existentes; erros PT; nome na UI sem expor token; não logar query completa; chat não amplia live/VOD |
| V. Qualidade testável | PASS | Testes Go em handlers/contratos/RBAC; CI build+test sem AWS |
| VI. Observabilidade suficiente | PASS | CloudWatch da task 001; logs sem token na query; sem APM novo |
| VII. Documentação / Speckit | PASS | Artefatos nesta pasta; README/AGENTS na fase de docs (implement futura) |
| VIII. IaC | PASS | Ajuste enxuto ALB `idle_timeout` (e nota CloudFront) na stack flat existente; sem módulos novos |
| Stack fechada AWS | PASS | Go/Fiber/React/Postgres; JWT própria; `us-east-1`; sem Cognito/NAT/domínio |
| Self-hosted como demo | PASS | Compose = dev; demo = AWS gerenciada |
| Baseline 001–003 | PASS | Sem reabrir IVS/VOD/fundação; `desired_count = 1`; sem Redis AWS |

**Post-design re-check:** PASS — data-model sem tabela de chat; contracts WS + env mínimo; quickstart Compose + nota ALB/CloudFront; idle timeout ALB justificado (custo zero além da task); sticky **não** necessário com `desired_count = 1`; CloudFront já encaminha upgrade WS com policy existente.

## Project Structure

### Documentation (this feature)

```text
specs/004-live-class-chat/
├── plan.md                 # Este arquivo
├── research.md             # Phase 0
├── data-model.md           # Phase 1
├── quickstart.md           # Phase 1
├── contracts/
│   ├── chat-ws.md          # Protocolo WebSocket
│   └── chat-env.md         # Env / infra (mínimo)
├── checklists/
│   └── requirements.md
└── tasks.md                # NÃO criado aqui → /speckit-tasks
```

### Source Code (repository root — extensões previstas)

```text
estuda-ja/
├── backend/
│   ├── cmd/api/main.go                      # rota WS /aulas/:id/chat/ws
│   └── internal/
│       ├── handler/chat_handler.go          # NOVO — upgrade, auth query, validação
│       ├── handler/chat_handler_test.go     # NOVO
│       ├── chat/                            # NOVO — hub em memória por aula_id
│       │   ├── hub.go
│       │   └── hub_test.go
│       ├── auth/jwt.go                      # reutilizar Parse; nome via lookup User
│       └── middleware/                      # autenticação WS dedicada (query), sem Bearer obrigatório no upgrade
├── frontend/
│   └── src/
│       ├── pages/AulasPage.tsx              # painel Chat (distinto de Live/VOD)
│       ├── components/AulaChatPanel.tsx     # NOVO — connect/send/list/errors
│       ├── api/client.ts                    # helper URL WS (base API → ws/wss)
│       └── auth/auth.ts                     # reutilizar token + papéis (sem novos poderes)
├── infra/
│   └── alb.tf                               # idle_timeout ↑ para conexões longas (ex. 3600)
├── docker-compose.yml                       # sem Redis; API já expõe 8080
├── README.md                                # runbook: chat no ciclo apply→demo→destroy
└── AGENTS.md                                # chat 004 deixa de ser “fora do escopo” após implement
```

**Structure Decision:** Monorepo atual; domínio **chat** em `handler/chat_*` + pacote `internal/chat` (hub), rotas aninhadas sob aula; UI só na ficha; Terraform flat só com ajuste de idle do ALB. Sem segundo serviço, sem ElastiCache, sem tabela `chat_messages`.

## Complexity Tracking

> Nenhuma violação de constitution a justificar. Tabela omitida.

## Implementation Phases (chats futuros — sem `tasks.md` aqui)

Cada fase cabe em **um chat de implement**. Critério **Done** verificável. Ordem 1→3. Moderação P2 = **fora**.

| Fase | Escopo | Done quando |
|------|--------|-------------|
| **1 — Hub WS + handlers/auth + testes Go (Compose)** | Pacote hub por `aula_id`; rota `GET /api/v1/aulas/:id/chat/ws?token=...`; validar JWT + aula + leitura; send/broadcast/error; texto ≤ 500; rejeitar vazio; erros PT; testes Go (auth/RBAC/validação/isolamento de sala); **sem** UI | `go test` verde; dois clientes (teste ou ferramenta) no Compose trocam mensagem na mesma aula; aula B isolada; sem token / token inválido / aula inexistente rejeitados; reinício da API limpa salas |
| **2 — UI painel chat na AulasPage** | Componente embutido na ficha; connect com `aula_id` + JWT query; lista só mensagens da conexão atual; autor = **nome**; erros/estado degradado PT; painel distinto de Live/VOD; aluno **não** ganha controles live/VOD | Dois browsers (aluno + professor) na mesma aula veem fan-out; reabrir ficha = painel vazio; mensagem >500 ou vazia rejeitada com feedback; UI sem moderação |
| **3 — Infra ALB/docs runbook** | `idle_timeout` ALB adequado; nota CloudFront WS + keepalive; README/AGENTS/quickstart no ciclo apply→publish→warm-up→demo chat→destroy; **sem** Lambda/API GW/Redis | `terraform plan` só muda idle (e docs); demo AWS: WS via `wss` CloudFront; destroy remove API/chat; budget intacto; operador ≤ 15 min incremental (SC-005) |

**Não fazer neste plan:** implementar código, gerar `tasks.md`, reabrir 001–003, moderação P2, histórico, Redis AWS.

## Next Actions

1. **`/speckit-tasks`** — gerar `tasks.md` com **fases = chats** (1–3), dependências explícitas; moderação **excluída**.
2. **Não** iniciar `/speckit-implement` até as tasks existirem (salvo pedido explícito).
