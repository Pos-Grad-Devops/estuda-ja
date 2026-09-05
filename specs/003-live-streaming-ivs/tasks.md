---
description: "Task list — Streaming ao vivo IVS (fases = chats isolados)"
---

# Tasks: Streaming ao vivo da aula (AWS IVS)

**Input**: Design documents from `/specs/003-live-streaming-ivs/`

**Prerequisites**: [plan.md](./plan.md), [spec.md](./spec.md), [research.md](./research.md), [data-model.md](./data-model.md), [contracts/](./contracts/), [quickstart.md](./quickstart.md), constitution

**Baseline**: `001-aws-mvp-terraform` + `002-vod-library` — **NÃO reabrir** (destroy, sem NAT, sem Redis AWS, VOD um-por-aula, sem página Biblioteca, budget separado).

**Tests**: Obrigatórios quando handlers/API live mudarem (constitution V / FR-013). E2E AWS+OBS = [quickstart.md](./quickstart.md). Local/CI = stub.

**Organization**: Fases **1–4** obrigatórias + **5 opcional** = **um chat `/speckit-implement` cada**. Não misturar API + UI + Terraform + docs no mesmo chat.

---

## Formato e protocolo de chats

### Formato da task

`- [ ] [ID] [P?] [Story?] Descrição com path — depends: … — aceite: …`

- **[P]**: paralelizável (arquivos distintos, sem depender de task incompleta)
- **[USn]**: user story do [spec.md](./spec.md) (US1 assiste · US2 inicia/encerra · US3 ciclo efêmero · US4 estados P2 · US5 warm-up auto P2/P3)
- Setup/foundational embutidos na Fase 1; Fase 5 é **opcional** (depois de 1–4)

### Regra de ouro (NÃO negociável)

1. Abrir **um** chat `/speckit-implement` **somente** para a fase atual.
2. Completar **todas** as tasks da fase + checklist **Done** verificável.
3. Só então abrir o próximo chat na fase seguinte.
4. **Não** começar Fase 2+ sem Done da 1; **não** misturar UI (2) com API (1), Terraform (3) com docs (4), nem P2 (5) antes de 1–4.

### Ordem dos chats

| Chat | Fase | Escopo | Prioridade |
|------|------|--------|------------|
| 1 | **1** | Modelo / API / stub + testes Go | P1 / US1+US2 (API) |
| 2 | **2** | UI ficha aula (`AulasPage` + `LivePlayer`) | P1 / US1+US2 (UI) |
| 3 | **3** | Terraform IVS + wire AWS (`LIVE_BACKEND=ivs`) | P1 / US3 (IaC) |
| 4 | **4** | Runbook/docs + warm-up streaming | P1 / US3 (docs) |
| 5 | **5** | Estados `agendada`/`ao_vivo`/`encerrada` + auto warm-up só streaming | **P2 / US4+US5 — opcional** |

**Começar implementação pela Fase 1 apenas.**

---

## Phase 1 — Modelo / API / stub + testes — CHAT 1

**Stories**: US1 (status/playback) + US2 (start/stop/ingest) — **só backend**  
**Goal**: `AulaLive`, handlers start/stop/status/playback/ingest, invariante ≤1 live `ao_vivo`, RBAC, horário opcional na API, `LIVE_BACKEND=stub`, erros PT, testes Go.  
**Independent Test**: `go test ./...` verde; professor inicia/encerra/reinicia no stub; aluno 403 em escrita/ingest; 409 se segunda aula ao vivo; playback stub **sem** URL de vídeo; aula sem `agendada_em` inicia; `Aula.Status` CRUD inalterado.  
**FORA DESTE CHAT**: UI (`AulasPage` / `LivePlayer`); Terraform/`infra/ivs.tf`; README/AGENTS longos; modo IVS real.

### Implementation

- [x] T001 [P] [US2] Estender `backend/internal/config/config.go` com `LIVE_BACKEND` (default `stub`), stubs `IVS_INGEST_ENDPOINT` / `IVS_STREAM_KEY` / `IVS_PLAYBACK_URL` / `IVS_CHANNEL_ARN` conforme [contracts/live-env.md](./contracts/live-env.md) — depends: nenhuma — aceite: Compose/local boota com `stub`; `IVS_*` ignorados no stub; valor inválido de `LIVE_BACKEND` falha boot ou default documentado
- [x] T002 [P] [US2] Em `docker-compose.yml` (e `Makefile` se preciso): env `LIVE_BACKEND=stub` no serviço API — depends: T001 — aceite: Compose sobe API em stub sem credenciais AWS/IVS
- [x] T003 [P] [US1] Criar model `AulaLive` em `backend/internal/models/aula_live.go` com `aula_id` UNIQUE, `status` (`ao_vivo` \| `encerrada`), `iniciada_em` / `encerrada_em` (`timeutil.DateTime`), timestamps — depends: nenhuma — aceite: alinhado a [data-model.md](./data-model.md); AutoMigrate/registro GORM em `backend/internal/database/database.go` inclui a tabela; **sem** seed `ao_vivo`
- [x] T004 [US2] Criar `backend/internal/repository/live_repository.go`: GetByAulaID, GetAtiva (status=`ao_vivo`), Start (upsert → `ao_vivo`), Stop (`encerrada`), DeleteByAulaID; invariante ≤1 `ao_vivo` (checagem em código; índice parcial Postgres se trivial) — depends: T003 — aceite: segunda start em outra aula falha antes de persistir; reinício após `encerrada` OK; ausência de linha = inativa
- [x] T005 [US2] Implementar `backend/internal/handler/live_handler.go`: `GET /aulas/:id/live` (status sem ingest), `POST .../live/start`, `POST .../live/stop` — depends: T001, T004 — aceite: contrato [live-api.md](./contracts/live-api.md); start idempotente **200** se já `ao_vivo` **nesta** aula; **409** PT se outra aula `ao_vivo`; start **sem** `agendada_em` permitido; stub persiste `ao_vivo`; `ivs` sem credenciais → **503** PT e **não** marca ao vivo; erros PT; **não** altera `Aula.Status` CRUD
- [x] T006 [US1] No mesmo `live_handler.go`: `GET /aulas/:id/live/playback` — se `ao_vivo` desta aula: stub → `playback_url: null` + mensagem PT; ivs → HLS URL do env; senão **404** PT — depends: T005 — aceite: JWT + leitura de aulas; anônimo 401; stub **sem** URL de vídeo; sem `stream_key` no body
- [x] T007 [US2] No mesmo `live_handler.go`: `GET /aulas/:id/live/ingest` — só admin/professor e só com live `ao_vivo` desta aula; stub → key/server null + mensagem; ivs → `rtmps://{host}:443/app/` + stream key — depends: T005 — aceite: aluno/anônimo **403/401** PT **sem** key no body; key estável da sessão (env)
- [x] T008 [US2] Wire em `backend/cmd/api/main.go`: rotas `/api/v1/aulas/:id/live`, `/live/start`, `/live/stop`, `/live/playback`, `/live/ingest`; leitura autenticada (admin/professor/aluno); start/stop/ingest `RequireRoles(admin, professor)` — depends: T005, T006, T007 — aceite: RBAC espelha escrita de aulas; anônimo 401 em todos
- [x] T009 [US2] Em `backend/internal/handler/aula_handler.go` (e/ou repository): ao `DELETE` aula, remover `AulaLive` associado — depends: T004 — aceite: exclusão de aula não deixa órfão no DB; **não** toca canal IVS
- [x] T010 [US1] Testes Go em `backend/internal/handler/live_handler_test.go` (SQLite `:memory:`): GET status inativa; professor start/stop/reinício; start sem `agendada_em` OK; aluno start/stop/ingest → 403; 409 segunda aula ao vivo; playback stub sem URL; anônimo 401; stop sem ao vivo → 409/404 — depends: T008 — aceite: `go test ./...` verde; constitution V / FR-013
- [x] T011 [US2] Smoke manual local (curl ou Compose): login professor → start → GET ingest stub → login aluno → GET status ao vivo + playback stub + ingest 403 → start 2ª aula 409 → stop → start de novo OK — depends: T010, T002 — aceite: checklist Done da Fase 1

### Done — Fase 1 (obrigatório antes do Chat 2)

- [x] `go test ./...` verde (handlers live incluídos)
- [x] Professor inicia / encerra / reinicia no stub
- [x] Aluno 403 em escrita e ingest (PT); anônimo 401
- [x] 409 se tentar 2ª live ativa; playback stub **sem** URL de vídeo
- [x] Aula sem `agendada_em` inicia; `Aula.Status` CRUD **inalterado**
- [x] **Nenhuma** mudança UI; **nenhuma** mudança Terraform IVS

**Checkpoint**: API live stub pronta. Parar. Abrir novo chat só para Fase 2.

---

## Phase 2 — UI ficha aula — CHAT 2

**Stories**: US1 + US2 — **só frontend**  
**Goal**: Em `AulasPage`: bloco **Ao vivo** distinto de **Gravação**; `LivePlayer` (IVS Player na AWS / mensagem no stub); iniciar/encerrar + ingest só `canManageAulas`; UI recomenda horário sem bloquear; **sem** rota “Ao vivo”.  
**Independent Test**: Aluno assiste/distingue live vs VOD; aluno sem ingest/botões; professor inicia e vê credenciais OBS (stub ou real).  
**FORA DESTE CHAT**: novos handlers Go (salvo bug bloqueante mínimo); Terraform IVS; docs longas README/AGENTS.

### Implementation

- [x] T012 [P] [US1] Estender `frontend/src/api/client.ts` com tipos e chamadas: `getLive`, `startLive`, `stopLive`, `getLivePlayback`, `getLiveIngest` alinhados a [contracts/live-api.md](./contracts/live-api.md) — depends: Fase 1 Done — aceite: JWT no header; erros de API legíveis na UI
- [x] T013 [P] [US1] Criar `frontend/src/components/LivePlayer.tsx`: se `modo=ivs` + `playback_url` → Amazon IVS Player Web SDK (script CDN `player.live-video.net` + `<video>` âncora); se stub / sem URL → mensagem PT clara (**não** player vazio enganoso); feedback se live ativa sem sinal — depends: T012 — aceite: research §3 / SC-007 caminho; VOD continua `<video>` MP4 separado
- [x] T014 [US1] Em `frontend/src/pages/AulasPage.tsx`: bloco **“Transmissão ao vivo”** (acima ou ao lado de **Gravação**): status inativa/ao_vivo/encerrada; se `ao_vivo` buscar playback e montar `LivePlayer`; ausência evidente se não ao vivo — depends: T012, T013 — aceite: aluno distingue live vs VOD (US1 AC5 / SC-012); datas BR se exibidas; sem download do fluxo
- [x] T015 [US2] No mesmo `AulasPage.tsx`: botões Iniciar / Encerrar + painel ingest (servidor + stream key) **somente** se `canWrite` / `canManageAulas` e conforme estado; aviso se `agendada_em` vazio **sem** bloquear submit — depends: T014 — aceite: professor vê OBS creds quando ao vivo; aluno **não** vê ingest nem botões (SC-003 UI / FR-006)
- [x] T016 [P] [US1] Confirmar `frontend/src/App.tsx` e rotas: **sem** nova rota/página “Ao vivo”; reutilizar `canManageAulas` de `frontend/src/auth/auth.ts` — depends: T014 — aceite: descoberta só curso → aula; FR-002
- [x] T017 [US1] `npm run build` + smoke manual local (stub): professor inicia → vê mensagem stub/ingest stub; aluno vê “ao vivo” sem controles; VOD 002 intacto — depends: T015, T016 — aceite: checklist Done da Fase 2

### Done — Fase 2 (obrigatório antes do Chat 3)

- [x] Aluno autentica, abre ficha, distingue **ao vivo** vs **gravação** vs ausência
- [x] Aluno **sem** ingest e **sem** botões de gestão
- [x] Professor inicia/encerra e vê credenciais OBS (stub: mensagem; sem key real)
- [x] Player vazio enganoso **não** aparece sem live / no stub
- [x] Sem rota “Ao vivo”; build frontend OK

**Checkpoint**: UI P1 pronta. Parar. Abrir novo chat só para Fase 3.

---

## Phase 3 — Terraform IVS + wire AWS — CHAT 3

**Story**: US3 — canal efêmero + API modo `ivs`  
**Goal**: `ivs.tf` canal BASIC LOW; stream key → SSM SecureString; ECS `LIVE_BACKEND=ivs` + env/secrets `IVS_*`; outputs só ARN; destroy remove canal; budget intacto.  
**Independent Test**: `terraform apply` cria 1 canal; live real com OBS; aluno assiste; `destroy` zera IVS; budget intacto.  
**FORA DESTE CHAT**: README/AGENTS completos (Fase 4); P2 estados; redesign 001/002; segundo CloudFront / recording / IVS Chat.

### Implementation

- [x] T018 [US3] Criar `infra/ivs.tf`: `aws_ivs_channel` `${project_name}-live` (`type=BASIC`, `latency_mode=LOW`, `authorized=false`, **sem** recording); stream key da sessão (um recurso; sem segunda key desnecessária) — depends: Fase 2 Done — aceite: [contracts/terraform-ivs.md](./contracts/terraform-ivs.md); research §1
- [x] T019 [US3] Estender `infra/ssm.tf`: SecureString `${ssm_prefix}/IVS_STREAM_KEY`; String/env para ingest endpoint e playback URL se não interpolados só na task — depends: T018 — aceite: key **não** no git nem em output plaintext; [live-env.md](./contracts/live-env.md)
- [x] T020 [US3] Estender `infra/iam.tf` (execution role): incluir ARN(s) SSM novos no `GetParameters` existente — depends: T019 — aceite: task lê stream key sem `ivs:*` na task role (P1)
- [x] T021 [US3] Estender `infra/ecs.tf`: env `LIVE_BACKEND=ivs`, `IVS_INGEST_ENDPOINT`, `IVS_PLAYBACK_URL`, `IVS_CHANNEL_ARN`; secret `IVS_STREAM_KEY` do SSM — depends: T018, T019, T020 — aceite: fail-fast no boot se `ivs` sem credenciais (espelho VOD s3)
- [x] T022 [P] [US3] Output `ivs_channel_arn` **apenas** em `infra/outputs.tf` — depends: T018 — aceite: **sem** stream key / **sem** playback_url em outputs
- [x] T023 [US3] Confirmar backend já seleciona stub vs ivs via `LIVE_BACKEND` em `backend/internal/config` + handlers (`main.go` fail-fast se `ivs` incompleto); ajustar só o necessário para wire AWS — depends: T021 — aceite: CI/local permanece stub; AWS task usa ivs; `go test ./...` verde sem conta AWS
- [x] T024 [US3] Validar `terraform validate`/`plan` (apply se credenciais OK): 1 canal; publish-api; professor inicia → ingest real; OBS → aluno assiste com `LivePlayer`; `destroy` remove canal/key; **não** tocar `infra/budget/` — depends: T021, T022, T023 — aceite: checklist Done da Fase 3; SC-004

### Done — Fase 3 (obrigatório antes do Chat 4)

- [x] `terraform apply` cria **1** canal IVS BASIC LOW
- [x] Live real com OBS; aluno autentica e assiste in-app
- [x] `terraform destroy` remove canal/key (custo IVS contínuo ≈ 0)
- [x] Budget `infra/budget/` intacto; sem NAT / Redis AWS / CF mídia / recording
- [x] Outputs sem key/URL; `go test ./...` verde

**Checkpoint**: Wire AWS pronto. Parar. Abrir novo chat só para Fase 4.

---

## Phase 4 — Runbook / docs + warm-up streaming — CHAT 4

**Story**: US3 — documentação operacional  
**Goal**: README + AGENTS + quickstart 001/003: ciclo apply → publish → **warm-up streaming** → OBS → demo → destroy; checklist T−15 estendido.  
**Independent Test**: Operador percorre runbook ≤ 15 min incremental (SC-005); T−0 sem cold start (SC-006); AGENTS deixa de listar streaming como proibido sem spec.  
**FORA DESTE CHAT**: redesenhar Terraform; P2 estados ricos; automação EventBridge apply/destroy do 001.

### Implementation

- [x] T025 [P] [US3] Atualizar `README.md` (seção Demo AWS): passos extras IVS/OBS no ciclo apply→publish→warm-up→demo→destroy; apontar [quickstart.md](./quickstart.md) desta feature — depends: Fase 3 Done — aceite: avaliador encontra streaming encaixado sem reabrir 001/002
- [x] T026 [P] [US3] Atualizar `AGENTS.md`: streaming ao vivo **deix de ser** “fora do escopo / proibido sem pedido” para esta feature (`003`); manter chat/certificado/live→VOD fora; mencionar stub local vs IVS AWS; domínio `live_*` / `AulaLive` — depends: Fase 3 Done — aceite: agentes não tratam streaming como bloqueado após 003
- [x] T027 [US3] Alinhar `specs/003-live-streaming-ivs/quickstart.md` e, se necessário, trecho T−15 em `specs/001-aws-mvp-terraform/quickstart.md` (extensão **sem** reabrir decisões): tabela T−15/T−10/T−5/T−0 com checagens de canal/ARN/OBS — depends: T025 — aceite: warm-up **antes** da janela; cold start na abertura explicitamente inaceitável (FR-008 / SC-006)
- [x] T028 [US3] Revisar consistência docs ↔ código (env `LIVE_BACKEND`, output `ivs_channel_arn`, RBAC live=aulas, VOD intacto); smoke de leitura ≤ 15 min incremental — depends: T025, T026, T027 — aceite: checklist Done da Fase 4; SC-005

### Done — Fase 4 (obrigatório antes do Chat 5 opcional)

- [x] Runbook incremental streaming ≤ 15 min (SC-005)
- [x] T−15 estendido com canal IVS + smoke live antes de T−0 (SC-006)
- [x] `AGENTS.md` atualizado (streaming não “proibido” sem spec desta feature)
- [x] Baseline 001/002 e budget **não** reabertos

**Checkpoint**: P1 documentado. Parar. Fase 5 só se pedido explícito (P2).

---

## Phase 5 — P2 estados + automação warm-up só streaming — CHAT 5 (OPCIONAL)

**Stories**: US4 + US5 — **só após Done das Fases 1–4**  
**Goal**: API/UI com `agendada` / `ao_vivo` / `encerrada` (agendada **só** com horário); MAY apontar VOD 002 se encerrada; automação **somente** health/sinal de streaming (**sem** EventBridge apply/destroy do 001).  
**Independent Test**: Aluno distingue os 3 estados; docs deixam claro manual vs automatizado; gap EventBridge 001 permanece.  
**FORA DESTE CHAT**: chat (004), certificado (005), live→VOD, multi-canal, tokenização IVS.

### Implementation

- [x] T029 [P] [US4] Estender `AulaLive.status` + API em `backend/internal/models/aula_live.go`, `live_repository.go`, `live_handler.go`: transição `(sem linha)→agendada` exige `agendada_em`; `agendada→ao_vivo` / `ao_vivo→encerrada`; cancelar agendada → inativa; testes Go — depends: Fases 1–4 Done — aceite: [data-model.md](./data-model.md) P2; FR-017/FR-021; `go test ./...` verde
- [x] T030 [US4] Em `frontend/src/pages/AulasPage.tsx` (+ `LivePlayer` se preciso): labels inequívocos agendada / ao vivo / encerrada; player **só** em `ao_vivo`; se `encerrada` + VOD publicado (002), MAY linkar gravação **sem** pipeline live→VOD — depends: T029 — aceite: US4 AC1–3 / SC-012 P2
- [x] T031 [P] [US5] Documentar (README e/ou `specs/003-live-streaming-ivs/quickstart.md`) o que é **manual** vs **automatizado** no warm-up de streaming; se entregar automação: script/check mínimo (ex. health canal / `GetStream`) **sem** ligar/desligar stack 001 — depends: Fases 1–4 Done — aceite: US5 AC1–2; EventBridge apply/destroy 001 permanece gap
- [x] T032 [US4] Smoke: gestor coloca cada estado; aluno vê indicação coerente; build + `go test` verdes — depends: T029, T030, T031 — aceite: checklist Done da Fase 5

### Done — Fase 5 (opcional)

- [x] Aluno distingue agendada / ao vivo / encerrada
- [x] Agendada só com horário; player não engana em agendada/encerrada
- [x] Docs: manual vs auto; sem EventBridge apply/destroy do 001
- [x] CI verde

**Checkpoint**: P2 completo se entregue. Fim da feature.

---

## Dependencies & Execution Order

### Phase Dependencies

```text
Fase 1 (API stub) ──► Fase 2 (UI) ──► Fase 3 (Terraform IVS) ──► Fase 4 (docs)
                                                                      │
                                                                      └──► Fase 5 (opcional P2)
```

- **Fase 1**: sem dependências de código desta feature; baseline 001/002 já no repo
- **Fase 2**: **BLOCKED** até Done Fase 1
- **Fase 3**: **BLOCKED** até Done Fase 2 (UI existe para validar OBS/aluno na sessão)
- **Fase 4**: **BLOCKED** até Done Fase 3 (docs descrevem recurso real)
- **Fase 5**: **BLOCKED** até Done Fases 1–4; **não** iniciar junto com 1–4

### User Story Mapping

| Story | Fases |
|-------|-------|
| US1 Aluno assiste | 1 (playback/status API) + 2 (player/UI) |
| US2 Gestor inicia/encerra | 1 (start/stop/ingest API) + 2 (controles/UI) |
| US3 Ciclo efêmero IVS | 3 (IaC) + 4 (runbook) |
| US4 Estados ricos | 5 (opcional) |
| US5 Warm-up auto streaming | 5 (opcional) |

### Parallel Opportunities (dentro da mesma fase / mesmo chat)

- **Fase 1**: T001 ∥ T003; T006 ∥ T007 após T005
- **Fase 2**: T012 ∥ T013; T016 ∥ com T015 se App.tsx não conflitar
- **Fase 3**: T022 ∥ com T019–T021 após T018
- **Fase 4**: T025 ∥ T026
- **Fase 5**: T029 ∥ T031
- **Entre fases**: **não** paralelizar chats diferentes

### Parallel Example: Fase 1

```text
# Em paralelo no início do Chat 1:
T001 — config LIVE_BACKEND em backend/internal/config/config.go
T003 — model AulaLive em backend/internal/models/aula_live.go

# Depois sequencial:
T004 repository → T005 start/stop/status → T006∥T007 playback/ingest → T008 wire → T010 testes
```

---

## Implementation Strategy

### MVP (Chats 1–4)

1. **Chat 1** = Fase 1 → Done API stub + testes
2. **Chat 2** = Fase 2 → Done UI ficha
3. **Chat 3** = Fase 3 → Done IVS efêmero + OBS
4. **Chat 4** = Fase 4 → Done runbook
5. **STOP** — P1 demonstrável; Fase 5 só se pedido

### Incremental

Cada fase agrega valor sem quebrar a anterior: stub testável → UI → vídeo real AWS → docs operacionais → (opcional) estados/auto.

### Notas

- [P] = arquivos distintos, sem dependência de task incompleta na mesma fase
- Paths explícitos; testes Go quando handlers mudarem
- **NÃO** reabrir 001/002; **NÃO** misturar API+UI+Terraform+docs no mesmo chat
- Commitar só se o usuário pedir
