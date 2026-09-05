---
description: "Task list — Chat da aula ao vivo (fases = chats isolados)"
---

# Tasks: Chat da aula ao vivo (tempo real)

**Input**: Design documents from `/specs/004-live-class-chat/`

**Prerequisites**: [plan.md](./plan.md), [spec.md](./spec.md), [research.md](./research.md), [data-model.md](./data-model.md), [contracts/](./contracts/), [quickstart.md](./quickstart.md), constitution

**Baseline**: `001-aws-mvp-terraform` + `002-vod-library` + `003-live-streaming-ivs` — **NÃO reabrir** (destroy, sem NAT, sem Redis AWS, VOD/live intactos, budget separado, `desired_count = 1`).

**Tests**: Obrigatórios na Fase 1 (handlers/hub/RBAC — constitution V / FR-012). Multi-browser Compose = [quickstart.md](./quickstart.md) §A (Fase 2). CI **sem** AWS e **sem** multi-cliente obrigatório.

**Organization**: Fases **1–3** = **um chat `/speckit-implement` cada**. Não misturar hub/API + UI + infra/docs no mesmo chat.

**FORA DESTA ENTREGA**: Moderação P2 (US5 / FR-016) — **sem tasks**; histórico ao reabrir; Redis/ElastiCache; API Gateway WebSocket + Lambda; tabela de mensagens; redesenho 001–003.

---

## Formato e protocolo de chats

### Formato da task

`- [ ] [ID] [P?] [Story?] Descrição com path — depends: … — aceite: …`

- **[P]**: paralelizável (arquivos distintos, sem depender de task incompleta)
- **[USn]**: user story do [spec.md](./spec.md) (US1 aluno chat · US2 professor/admin + RBAC · US3 ciclo efêmero · US4 local/CI · **US5 moderação = fora**)
- Setup/foundational embutidos na Fase 1 (sem fases Setup/Foundational separadas)

### Regra de ouro (NÃO negociável)

1. Abrir **um** chat `/speckit-implement` **somente** para a fase atual.
2. Completar **todas** as tasks da fase + checklist **Done** verificável.
3. Só então abrir o próximo chat na fase seguinte.
4. **Não** começar Fase 2+ sem Done da 1; **não** misturar UI (2) com API (1), nem infra/docs (3) com API/UI.

### Ordem dos chats

| Chat | Fase | Escopo | Prioridade |
|------|------|--------|------------|
| 1 | **1** | Hub WS + handlers/auth + testes Go (Compose) | P1 / US1+US2+US4 (API) |
| 2 | **2** | UI painel chat na `AulasPage` | P1 / US1+US2 (UI) |
| 3 | **3** | Infra ALB `idle_timeout` + docs/runbook | P1 / US3 (+ docs US4) |

**Começar implementação pela Fase 1 apenas.** Não implementar neste artefato.

---

## Phase 1 — Hub WS + handlers/auth + testes Go — CHAT 1

**Stories**: US1 (enviar/receber) + US2 (RBAC / nome) + US4 (local/CI) — **só backend**  
**Goal**: Pacote hub em memória por `aula_id`; rota `GET /api/v1/aulas/:id/chat/ws?token=...`; JWT + lookup `nome`; validação ≤500 runes; broadcast isolado; erros PT; testes Go; Compose sem vars Redis/chat extra. **Sem UI.**  
**Independent Test**: `go test ./...` verde; dois clientes (teste ou ferramenta) no Compose trocam mensagem na mesma aula; aula B isolada; sem token / token inválido / aula inexistente rejeitados; reinício da API limpa salas.  
**FORA DESTE CHAT**: UI (`AulaChatPanel` / `AulasPage`); Terraform/`alb.tf`; README/AGENTS longos; moderação; tabela de mensagens; Redis; API GW WS+Lambda.

### Implementation

- [ ] T001 [P] [US1] Criar pacote hub em `backend/internal/chat/hub.go`: `map[aula_id]*Room`, register/unregister/broadcast em goroutine única, GC de sala vazia, write mutex por conexão — conforme [data-model.md](./data-model.md) e [research.md](./research.md) §2 — depends: nenhuma — aceite: broadcast só para clientes da mesma `aula_id`; sala vazia não vaza indefinidamente; reinício do processo zera mapa
- [ ] T002 [P] [US1] Criar testes de hub em `backend/internal/chat/hub_test.go`: isolamento aula A vs B; register/unregister; broadcast inclui remetente — depends: T001 — aceite: `go test ./internal/chat/...` verde; mensagens de A nunca chegam a B
- [ ] T003 [US1] Implementar `backend/internal/handler/chat_handler.go`: upgrade Fiber `contrib/websocket`; path `GET /api/v1/aulas/:id/chat/ws`; handshake com `?token=` (JWT existente); rejeitar sem token/inválido (401/close); aula inexistente 404; papéis leitura (`admin`/`professor`/`aluno`); lookup `User.Nome` no DB; **não** logar query com token — depends: T001 — aceite: alinhado a [contracts/chat-ws.md](./contracts/chat-ws.md); erros PT; identidade do autor vem do servidor (não do payload cliente)
- [ ] T004 [US1] No mesmo `chat_handler.go`: frames `chat.send` → validar texto (trim; 1..500 runes Unicode); `chat.message` broadcast (UUID, `aula_id`, `autor.{id,nome}`, `texto`, `enviado_em` BR); `chat.error` PT; `chat.ping`/`chat.pong` keepalive — depends: T003 — aceite: vazio/>500 → erro PT sem broadcast; tipo inválido → erro PT; eco ao remetente; sem frames de moderação/histórico
- [ ] T005 [US4] Wire em `backend/cmd/api/main.go`: rota WS sob `/api/v1/aulas/:id/chat/ws` com middleware de upgrade; hub singleton no processo; **sem** dependência de `AulaLive`/`AulaVod` — depends: T003, T004 — aceite: Compose/local sobe WS real na porta 8080; sala disponível com aula existente independente do status live
- [ ] T006 [P] [US4] Confirmar `docker-compose.yml` (e env API): **sem** Redis para chat; **sem** vars `REDIS_URL`/`CHAT_AWS_*`/API GW; JWT/CORS/PORT existentes bastam — [contracts/chat-env.md](./contracts/chat-env.md) — depends: nenhuma — aceite: Compose sobe API com WS na 8080; chat P1 não exige serviço novo
- [ ] T007 [US1] Testes Go em `backend/internal/handler/chat_handler_test.go` (SQLite `:memory:` / harness HTTP-WS): auth ausente/inválido rejeitado; aula inexistente rejeitada; send válido → `chat.message`; vazio e >500 → `chat.error` PT; isolamento duas aulas; anônimo sem token rejeitado — depends: T005 — aceite: `go test ./...` verde; constitution V / FR-012; multi-cliente E2E **não** obrigatório no CI
- [ ] T008 [US4] Smoke manual Compose (websocat/curl-ws ou dois clientes de teste): login → WS mesma aula troca mensagem; segunda aula isolada; sem token rejeitado — depends: T007, T006 — aceite: checklist Done da Fase 1

### Done — Fase 1 (obrigatório antes do Chat 2)

- [ ] `go test ./...` verde (hub + chat handlers)
- [ ] Dois clientes na mesma aula trocam mensagem via WS (Compose)
- [ ] Aulas isoladas (A ≠ B); sem token / token inválido / aula inexistente rejeitados
- [ ] Reinício da API limpa salas (efêmero)
- [ ] **Nenhuma** mudança UI; **nenhuma** mudança Terraform ALB; **sem** tabela de mensagens; **sem** Redis; **sem** moderação

**Checkpoint**: API chat WS pronta. Parar. Abrir novo chat só para Fase 2.

---

## Phase 2 — UI painel chat na AulasPage — CHAT 2

**Stories**: US1 + US2 — **só frontend**  
**Goal**: `AulaChatPanel` embutido na ficha; connect com `aula_id` + JWT query; lista só da conexão atual; autor = **nome**; erros/estado degradado PT; painel distinto de Live/VOD; aluno **não** ganha controles live/VOD.  
**Independent Test**: Dois browsers (aluno + professor) na mesma aula veem fan-out; reabrir ficha = painel vazio; >500/vazio rejeitado na UI; sem moderação na UI.  
**FORA DESTE CHAT**: novos handlers Go (salvo bug bloqueante mínimo); Terraform/`alb.tf`; README/AGENTS longos; moderação UI.

### Implementation

- [ ] T009 [P] [US1] Estender `frontend/src/api/client.ts` com helper de URL WS (`http(s)` → `ws(s)`) para `/api/v1/aulas/{id}/chat/ws?token=...` a partir de `VITE_API_URL` + token do auth — depends: Fase 1 Done — aceite: local `ws://localhost:8080/...`; AWS `wss://` via mesmo base URL; token só na query do WS (não logado no console em prod path)
- [ ] T010 [P] [US1] Criar `frontend/src/components/AulaChatPanel.tsx`: connect no mount/abertura; send `chat.send`; lista efêmera de `chat.message`; exibir `autor.nome`; tratar `chat.error` e desconexão em PT; keepalive `chat.ping` opcional (~2–4 min); limpar lista ao unmount/nova conexão — depends: T009 — aceite: reabrir = vazio (FR-009/SC-011); validação UI de vazio/>500 com feedback PT (SC-012); sem controles de moderação
- [ ] T011 [US1] Em `frontend/src/pages/AulasPage.tsx`: bloco **Chat** embutido na ficha, **distinto** de “Transmissão ao vivo” e “Gravação”; disponível para aula existente com usuário autenticado (independente de live) — depends: T010 — aceite: SC-010; datas BR se exibidas; sem rota/página “Chat” dedicada
- [ ] T012 [P] [US2] Confirmar `frontend/src/auth/auth.ts` e UI: chat **não** adiciona poderes; aluno continua sem ingest/upload VOD; admin/professor usam o mesmo painel — depends: T011 — aceite: US2 AC2/SC-007; sem novos helpers RBAC de escrita só por causa do chat
- [ ] T013 [US1] `npm run build` + smoke manual Compose (dois browsers): aluno + professor mesma aula → fan-out ≤ 5 s; outra aula isolada; F5 = painel vazio; vazio/>500 rejeitados — depends: T011, T012 — aceite: checklist Done da Fase 2 / [quickstart.md](./quickstart.md) §A

### Done — Fase 2 (obrigatório antes do Chat 3)

- [ ] Dois browsers na mesma aula veem fan-out com **nome** do autor
- [ ] Reabrir ficha = painel vazio (sem histórico)
- [ ] Mensagem vazia ou >500 rejeitada com feedback PT na UI
- [ ] Painel distinto de Live/VOD; aluno **sem** novos controles live/VOD
- [ ] Sem moderação na UI; build frontend OK

**Checkpoint**: UI P1 pronta. Parar. Abrir novo chat só para Fase 3.

---

## Phase 3 — Infra ALB + docs/runbook — CHAT 3

**Story**: US3 (ciclo efêmero) + docs US4  
**Goal**: `idle_timeout` ALB adequado; nota CloudFront WS + keepalive; README/AGENTS/quickstart no ciclo apply→publish→warm-up→demo chat→destroy; **sem** Lambda/API GW/Redis.  
**Independent Test**: `terraform plan` só muda idle (e docs); budget intacto; runbook incremental ≤ 15 min (SC-005); destroy remove API/chat com a stack.  
**FORA DESTE CHAT**: código de hub/UI novo (salvo nota/docs); redesenho 001–003; ElastiCache; API GW WS+Lambda; sticky TG; budget.

### Implementation

- [ ] T014 [US3] Em `infra/alb.tf`: subir `idle_timeout` do ALB da API de 60 → **3600** (≤ 4000); **sem** stickiness no target group; **sem** ElastiCache/API GW WS/Lambda; **sem** tocar `infra/budget/` — depends: Fase 2 Done — aceite: [contracts/chat-env.md](./contracts/chat-env.md) / research §4; `terraform plan` mostra mudança de idle (e nada de Redis/API GW); baseline 001–003 intacta
- [ ] T015 [P] [US3] Atualizar [quickstart.md](./quickstart.md) (e nota inline se preciso): caminho AWS `wss` via CloudFront `api_url`; keepalive vs idle CF ~10 min; ciclo apply→publish→warm-up→demo chat→destroy; **não** colar URLs com `?token=` — depends: T014 — aceite: avaliador segue smoke §C sem reabrir desenho 001–003
- [ ] T016 [P] [US3] Atualizar `README.md` (seção Demo AWS / runbook): passos incrementais de chat no ciclo efêmero; local = WS real; CI = testes sem AWS; sem NAT/Redis AWS/API GW WS no P1 — depends: T014 — aceite: SC-005 (≤ 15 min incremental); chat encaixado sem stack paralela
- [ ] T017 [P] [US4] Atualizar `AGENTS.md`: chat 004 deixa de ser “fora do escopo”; documentar hub in-memory, path WS, RBAC inalterado, moderação P2 ainda fora — depends: T014 — aceite: agente/humano não trata chat como proibido; baseline 001–003 e “sem Redis AWS” permanecem
- [ ] T018 [US3] Verificação final: `terraform plan` (idle only) + checklist runbook; confirmar budget `infra/budget/` intocado — depends: T015, T016, T017 — aceite: checklist Done da Fase 3

### Done — Fase 3 (feature P1 completa)

- [ ] `terraform plan` só idle ALB (+ docs); **sem** Redis/API GW/Lambda/sticky
- [ ] Budget `infra/budget/` intacto; destroy da sessão remove API/chat
- [ ] README/AGENTS/quickstart descrevem apply→…→demo chat→destroy + nota CloudFront/keepalive
- [ ] Moderação P2 **não** entregue; 001–003 **não** reabertos

**Checkpoint**: Feature 004 P1 completa. Parar.

---

## Dependencies & Execution Order

### Phase Dependencies

```text
Chat 1 (Fase 1: Hub/API/testes)
    │
    ▼  Done Fase 1
Chat 2 (Fase 2: UI AulasPage)
    │
    ▼  Done Fase 2
Chat 3 (Fase 3: ALB + docs)
    │
    ▼  Done Fase 3 = P1 pronto
(US5 moderação P2 — sem chat / sem tasks nesta rodada)
```

- **Fase 1**: sem dependência de UI/infra; bloqueia Fase 2
- **Fase 2**: depende do Done da Fase 1; bloqueia Fase 3 só por protocolo de chats (UI validável no Compose antes do idle AWS)
- **Fase 3**: depende do Done da Fase 2; única fase que toca Terraform + docs longas

### User Story Mapping

| Story | Onde |
|-------|------|
| US1 enviar/receber | Fase 1 (API) + Fase 2 (UI) |
| US2 professor/admin + RBAC | Fase 1 (nome/auth) + Fase 2 (UI sem novos poderes) |
| US3 ciclo efêmero | Fase 3 |
| US4 local/CI | Fase 1 (testes + Compose) + docs Fase 3 |
| US5 moderação | **Excluída** |

### Within Each Phase

- Models/hub antes de handlers; handlers antes de wire/`main`; testes após wire
- UI: client helper → componente → `AulasPage` → smoke
- Infra: `alb.tf` antes de docs que referenciam o valor

### Parallel Opportunities

**Fase 1:** T001 ∥ T006; após T001: T002 ∥ (início de T003); T007 após T005  
**Fase 2:** T009 ∥ (prep); T010 após T009; T012 ∥ revisão auth após T011  
**Fase 3:** T015 ∥ T016 ∥ T017 após T014  

**Entre fases:** **não** paralelizar Chats 1–3 (regra de ouro).

---

## Parallel Example: Fase 1

```text
# Em paralelo no início do Chat 1:
Task: T001 hub.go
Task: T006 docker-compose / env check

# Após T001:
Task: T002 hub_test.go
Task: T003 chat_handler handshake (depois T004 frames)
```

---

## Implementation Strategy

### MVP First (Fase 1 = Chat 1)

1. Completar **somente** Fase 1 (hub + WS + testes).
2. **STOP**: validar Done (go test + dois clientes Compose).
3. Só então Chat 2 (UI).

### Incremental Delivery

1. Fase 1 → API demonstrável via websocat/teste (MVP técnico).
2. Fase 2 → demo visual na ficha (MVP de produto local).
3. Fase 3 → demo AWS no ciclo destroy + docs (MVP acadêmico completo).

### Suggested MVP scope

**Chat 1 / Fase 1** (hub + handlers + testes). UI e ALB/docs vêm nos chats seguintes.

---

## Notes

- Formato checklist: `- [ ] Tnnn [P?] [USn?] … — depends: … — aceite: …` em **todas** as tasks de implementação.
- **Sem** tasks de moderação, Redis, API GW WS+Lambda, tabela `chat_messages`, sticky TG, redesign 001–003.
- Um chat = uma fase; Done antes da próxima.
- **Não** iniciar `/speckit-implement` além da Fase 1 até o Done correspondente.
