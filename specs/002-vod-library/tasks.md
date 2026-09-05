---
description: "Task list — Biblioteca VOD (fases = chats isolados)"
---

# Tasks: Biblioteca VOD (vídeos gravados sob demanda)

**Input**: Design documents from `/specs/002-vod-library/`

**Prerequisites**: [plan.md](./plan.md), [spec.md](./spec.md), [research.md](./research.md), [data-model.md](./data-model.md), [contracts/](./contracts/), [quickstart.md](./quickstart.md), constitution

**Baseline**: `001-aws-mvp-terraform` — **NÃO reabrir** (destroy, sem NAT, RBAC aulas, budget separado, etc.).

**Tests**: Obrigatórios quando handlers/API VOD mudarem (constitution V). E2E demo = [quickstart.md](./quickstart.md).

**Organization**: Fases **1–5** = **um chat `/speckit-implement` cada**. Não misturar API + UI + Terraform + docs no mesmo chat.

---

## Formato e protocolo de chats

### Formato da task

`- [ ] [ID] [P?] [Story?] Descrição com path — depends: … — aceite: …`

- **[P]**: paralelizável (arquivos distintos, sem depender de task incompleta)
- **[USn]**: user story do [spec.md](./spec.md) (US1 assiste · US2 publica · US3 ciclo efêmero · US4 P2 rascunho/metadados)
- Setup/foundational embutidos na Fase 1; Fase 5 é **opcional** (P2)

### Regra de ouro (NÃO negociável)

1. Abrir **um** chat `/speckit-implement` **somente** para a fase atual.
2. Completar **todas** as tasks da fase + checklist **Done** verificável.
3. Só então abrir o próximo chat na fase seguinte.
4. **Não** começar Fase 2+ sem Done da 1; **não** misturar UI (2) com API (1), Terraform (3) com seed/docs (4), nem P2 (5) antes de 1–4.

### Ordem dos chats

| Chat | Fase | Escopo | Prioridade |
|------|------|--------|------------|
| 1 | **1** | Modelo / API / storage local + testes Go | P1 / US1+US2 (API) |
| 2 | **2** | UI ficha aula (`AulasPage`) | P1 / US1+US2 (UI) |
| 3 | **3** | Terraform S3 + wire AWS (API modo S3) | P1 / US3 (IaC) |
| 4 | **4** | Seed asset MP4 + runbook/docs | P1 / US3 (seed+docs) |
| 5 | **5** | Rascunho/publicado + título/duração | **P2 / US4 — opcional** |

**Começar implementação pela Fase 1 apenas.**

---

## Phase 1 — Modelo / API / storage local + testes — CHAT 1

**Stories**: US1 (playback/metadados) + US2 (upload/replace/delete) — **só backend**  
**Goal**: `AulaVod`, `vodstorage` local, handlers upload/replace/delete/metadata/playback, validação MP4+50 MB, delete objeto antigo, RBAC admin+professor, testes Go.  
**Independent Test**: `go test ./...` verde; professor sobe MP4 local; aluno obtém `playback_url`; aluno não escreve; rejeita >50 MB / não-MP4.  
**FORA DESTE CHAT**: UI nova além do mínimo (não tocar `AulasPage`); Terraform/`infra/`; seed asset MP4 no repo; README/AGENTS longos; modo S3.

### Implementation

- [x] T001 [P] [US2] Estender `backend/internal/config/config.go` com `VOD_BACKEND`, `VOD_LOCAL_DIR`, `VOD_PLAYBACK_TTL` (e stubs `VOD_S3_BUCKET`/`AWS_REGION` só se necessário para boot local) conforme [contracts/vod-env.md](./contracts/vod-env.md) — depends: nenhuma — aceite: defaults locais `local` + dir + `15m`; sem secrets AWS
- [x] T002 [P] [US2] Em `docker-compose.yml` (e `Makefile` se preciso): volume `./data/vod` → `/data/vod` + env `VOD_BACKEND=local`, `VOD_LOCAL_DIR=/data/vod`, `VOD_PLAYBACK_TTL=15m` no serviço API — depends: T001 — aceite: Compose sobe API com dir VOD montado
- [x] T003 [P] [US1] Criar model `AulaVod` em `backend/internal/models/` (arquivo dedicado ou extensão clara) com `aula_id` UNIQUE, `storage_key`, `content_type`, `size_bytes`, `status` (`publicado` no P1), timestamps `timeutil.DateTime` — depends: nenhuma — aceite: alinhado a [data-model.md](./data-model.md); AutoMigrate/registro GORM inclui a tabela
- [x] T004 [US2] Criar pacote `backend/internal/vodstorage/` com interface `Storage` (Put/Get/Delete/Presign ou equivalente local) + implementação **local** sob `VOD_LOCAL_DIR` e key `vod/aulas/{aula_id}/current.mp4` — depends: T001 — aceite: Put/Delete/Get funcionam no disco; path relativo espelha research §1
- [x] T005 [US2] Criar `backend/internal/repository/vod_repository.go` (CRUD metadados por `aula_id`, replace upsert, delete) e registrar no `backend/internal/repository/repository.go` — depends: T003 — aceite: um VOD vigente por aula; sem segundo registro na mesma aula
- [x] T006 [US2] Validação de upload (tamanho ≤ 52 428 800 bytes, MIME/`video/mp4`, magic `ftyp`) em helper reutilizável (ex. `backend/internal/handler/` ou `vodstorage/`) — depends: nenhuma — aceite: rejeição com mensagem em português (FR-015/FR-019); sem ffprobe
- [x] T007 [US2] Implementar `backend/internal/handler/vod_handler.go`: `PUT /aulas/:id/vod` (multipart), `DELETE /aulas/:id/vod`, cascade delete storage ao apagar VOD; BodyLimit ≥ 50 MB na rota/app se necessário — depends: T004, T005, T006 — aceite: professor/admin publicam; replace apaga/sobrescreve objeto antigo; erros 400/403/404 em PT
- [x] T008 [US1] No mesmo `vod_handler.go`: `GET /aulas/:id/vod` (metadados publicados) e `GET /aulas/:id/vod/playback` retornando `{playback_url, expires_at, expires_in_seconds}` (~15 min); local: URL de content assinado/autenticado equivalente — depends: T004, T005 — aceite: contrato [vod-api.md](./contracts/vod-api.md); sem URL permanente de storage
- [x] T009 [US1] Endpoint local de content (ex. `GET /api/v1/aulas/:id/vod/content` com token de curta duração) **ou** stream autenticado documentado no handler — depends: T008 — aceite: `<video>`/curl consegue ler bytes com token válido; token expirado falha
- [x] T010 [US2] Wire em `backend/cmd/api/main.go`: rotas aninhadas sob `/api/v1/aulas/:id/vod*`; leitura autenticada (admin/professor/aluno); escrita `RequireRoles(admin, professor)`; injetar storage local no bootstrap — depends: T007, T008, T009 — aceite: RBAC espelha escrita de aulas; anônimo 401
- [x] T011 [US2] Em `backend/internal/handler/aula_handler.go` (e/ou repository): ao `DELETE` aula, remover metadados VOD + `storage.Delete` — depends: T005, T004 — aceite: exclusão de aula não deixa órfão no disco/DB
- [x] T012 [US1] Testes Go em `backend/internal/handler/vod_handler_test.go` (e auxiliares se preciso): GET metadados/playback; PUT válido; PUT >50 MB e não-MP4 → 400; aluno PUT/DELETE → 403; professor PUT OK; DELETE remove — depends: T010 — aceite: `go test ./...` verde; constitution V
- [x] T013 [US1] Verificação manual local (curl ou equivalente): login professor → upload MP4 pequeno → login aluno → GET playback → bytes OK; aluno sem escrita — depends: T012, T002 — aceite: checklist Done da Fase 1 abaixo

### Done — Fase 1 (obrigatório antes do Chat 2)

- [x] `go test ./...` verde (handlers VOD incluídos)
- [x] Professor sobe MP4 válido no storage local; metadados `publicado`
- [x] Aluno obtém `playback_url` e consegue ler o conteúdo
- [x] Aluno não escreve (403 PT); UI não é requisito deste chat
- [x] Rejeita arquivo > 50 MB e não-MP4 (400 PT); aula não fica “publicada” quebrada
- [x] Replace/delete apaga objeto antigo no disco
- [x] **Nenhuma** mudança Terraform; **nenhuma** página Biblioteca; **sem** modo S3 obrigatório

**Checkpoint**: API VOD local pronta. Parar. Abrir novo chat só para Fase 2.

---

## Phase 2 — UI ficha aula — CHAT 2

**Stories**: US1 + US2 — **só frontend**  
**Goal**: Em `AulasPage`: indicador VOD, `<video>` com URL de playback, upload/remover se `canManageAulas`; sem rota Biblioteca; sem download.  
**Independent Test**: Aluno assiste na ficha; professor publica pela UI; aluno sem botões de escrita.  
**FORA DESTE CHAT**: novos handlers Go (salvo bug bloqueante mínimo); Terraform; seed asset; docs longas.

### Implementation

- [x] T014 [P] [US1] Estender `frontend/src/api/client.ts` com tipos e chamadas: `getVod`, `getVodPlayback`, `uploadVod`, `deleteVod` alinhados a [contracts/vod-api.md](./contracts/vod-api.md) — depends: Fase 1 Done — aceite: JWT no header; erros de API legíveis na UI
- [x] T015 [US1] Em `frontend/src/pages/AulasPage.tsx`: bloco “Gravação” — ausência clara se 404/sem VOD; se houver, buscar playback e renderizar `<video controls>` (sem `download` / sem `Content-Disposition` de produto) — depends: T014 — aceite: aluno vê/toca VOD na ficha; datas BR se exibidas
- [x] T016 [US2] No mesmo `AulasPage.tsx`: formulário upload (input file) + botão remover **somente** se `canWrite` / `canManageAulas`; feedback de erro PT (tamanho/formato) — depends: T015 — aceite: professor/admin gerem VOD na UI; aluno não vê botões de escrita (SC-003 UI)
- [x] T017 [P] [US1] Confirmar `frontend/src/App.tsx` e rotas: **sem** nova rota/página “Biblioteca”; reutilizar `canManageAulas` de `frontend/src/auth/auth.ts` — depends: T015 — aceite: descoberta só curso → aula; SC-009 (sem Biblioteca)
- [x] T018 [US1] `npm run build` no frontend + smoke manual local (aluno assiste; professor substitui) — depends: T016, T017 — aceite: checklist Done da Fase 2

### Done — Fase 2 (obrigatório antes do Chat 3)

- [x] Aluno autenticado assiste VOD na ficha da aula (SC-001/SC-006 caminho local)
- [x] Professor/admin publicam/substituem/removem pela UI
- [x] Aluno **sem** botões de escrita VOD
- [x] Sem página/rota Biblioteca; sem fluxo de download
- [x] Build frontend OK

**Checkpoint**: UI P1 pronta. Parar. Abrir novo chat só para Fase 3.

---

## Phase 3 — Terraform S3 + wire AWS — CHAT 3

**Story**: US3 — armazenamento/entrega na sessão efêmera  
**Goal**: `s3_vod.tf` (`force_destroy`), IAM task, env ECS, outputs; API modo S3 (presigned GET); **sem** redesign 001; **sem** segundo CloudFront de mídia.  
**Independent Test**: `terraform apply` cria bucket; upload/playback na sessão; `destroy` remove bucket/objetos.  
**FORA DESTE CHAT**: seed MP4 no repo / migration 003; README/AGENTS completos; mudanças grandes de UI.

### Implementation

- [x] T019 [US3] Criar `infra/s3_vod.tf`: bucket `${project_name}-vod-${account_id}`, privado, SSE AES256, block public, `force_destroy = true`, **sem** website/OAC/CF — depends: Fase 2 Done — aceite: [contracts/terraform-vod.md](./contracts/terraform-vod.md); bucket distinto do frontend
- [x] T020 [US3] Estender `infra/iam.tf` (task role): `s3:PutObject`/`GetObject`/`DeleteObject` em `…/bucket/vod/*` (+ `ListBucket` com prefix `vod/` se necessário) — depends: T019 — aceite: task ECS consegue Put/Get/Delete sem access key no env
- [x] T021 [US3] Estender `infra/ecs.tf` (e `variables.tf` se preciso): env `VOD_BACKEND=s3`, `VOD_S3_BUCKET`, `VOD_PLAYBACK_TTL=15m`, `AWS_REGION=us-east-1` — depends: T019 — aceite: task plain env alinhado a vod-env
- [x] T022 [P] [US3] Output `vod_bucket_name` em `infra/outputs.tf` — depends: T019 — aceite: `terraform output` lista o bucket VOD
- [x] T023 [US3] Implementar backend S3 em `backend/internal/vodstorage/` (AWS SDK v2): Put/Delete/Presign GetObject TTL ~15 min; selecionar local vs s3 via `VOD_BACKEND` no bootstrap (`main.go`/`config`) — depends: T020, T021 — aceite: fail-fast se `s3` sem bucket; playback AWS = URL presigned (research §3)
- [x] T024 [US3] Testes Go que cubram seleção de backend / presign mockável **ou** testes unitários do wrapper S3 sem credenciais reais; regressão dos testes locais permanece verde — depends: T023 — aceite: `go test ./...` verde sem depender de conta AWS no CI
- [x] T025 [US3] Validar `terraform validate`/`plan` (apply se credenciais OK): bucket criado; publish API da sessão; upload+playback; `destroy` remove bucket — depends: T022, T023 — aceite: checklist Done da Fase 3; budget `infra/budget/` intocado

### Done — Fase 3 (obrigatório antes do Chat 4)

- [x] `terraform apply` cria bucket VOD efêmero
- [x] API com `VOD_BACKEND=s3` faz upload e playback presigned na sessão
- [x] `terraform destroy` remove bucket e objetos (`force_destroy`)
- [x] Sem NAT, sem Redis AWS, sem CF de mídia, sem redesign 001
- [x] `go test ./...` verde

**Checkpoint**: Wire AWS pronto. Parar. Abrir novo chat só para Fase 4.

---

## Phase 4 — Seed asset + runbook/docs — CHAT 4

**Story**: US3 — seed + documentação operacional  
**Goal**: MP4 seed no repo + migration idempotente; README/AGENTS/quickstart com ciclo VOD incremental.  
**Independent Test**: Pós-apply (ou local) aluno assiste **sem** upload manual; runbook extras ≤ 10 min (SC-008).  
**FORA DESTE CHAT**: redesenhar Terraform; P2 rascunho; novas features de UI além do necessário ao seed.

### Implementation

- [x] T026 [US3] Adicionar asset `backend/assets/vod/demo-aula.mp4` (curto, preferencialmente ≤ 2 MB, hard cap 50 MB) — depends: Fase 3 Done — aceite: arquivo versionado; reproduzível offline/CI
- [x] T027 [US3] Migration gormigrate `003_seed_vod_demo` em `backend/internal/database/migrations/` + registro em `migrations.go`: se aula do `002_seed_demo` existe e sem VOD, copia asset → storage e cria `AulaVod` `publicado` (idempotente) — depends: T026 — aceite: SC-005; re-run não duplica
- [x] T028 [P] [US3] Garantir `backend/Dockerfile` faz `COPY` de `assets/vod` (e path disponível no container) — depends: T026 — aceite: imagem API contém o MP4 de seed
- [x] T029 [US3] Teste Go da migration/seed VOD (ex. `backend/internal/database/migrations/…_test.go`) cobrindo idempotência — depends: T027 — aceite: `go test` verde; constitution V
- [x] T030 [P] [US3] Atualizar `README.md` (seção Demo AWS): passos extras VOD no ciclo apply → publish → demo → destroy; o que o seed recria; o que o destroy apaga; fora de escopo (Biblioteca, download, ao vivo) — depends: T027 — aceite: FR-017; SC-008
- [x] T031 [P] [US3] Atualizar `AGENTS.md` (menção VOD: storage local/S3, domínio `vod_*`, sem reabrir 001) — depends: T027 — aceite: agentes sabem escopo VOD
- [x] T032 [US3] Alinhar `specs/002-vod-library/quickstart.md` ao comportamento real entregue (local A + AWS B) — depends: T030 — aceite: quickstart executável sem inventar console
- [x] T033 [US3] Verificação: apply/publish (ou local compose) → login aluno → assiste seed sem upload manual; destroy (se AWS) zera VOD — depends: T027, T030 — aceite: checklist Done da Fase 4

### Done — Fase 4 (P1 completo)

- [x] Seed cria exatamente 1 VOD na aula de demo (SC-005)
- [x] Aluno assiste sem upload manual pós-apply/publish (ou compose local)
- [x] README/AGENTS/quickstart descrevem ciclo VOD incremental
- [x] Destroy continua zerando objetos VOD; budget intocado
- [x] CI/`go test` + build front verdes no branch

**Checkpoint**: P1 VOD entregue. Fase 5 só se pedir P2.

---

## Phase 5 — P2 rascunho/publicado + título/duração — CHAT 5 (OPCIONAL)

**Story**: US4 — metadados e estado rascunho/publicado  
**Goal**: Estados `rascunho`/`publicado`; título/duração visíveis; aluno não reproduz rascunho.  
**Independent Test**: Gestor salva rascunho → aluno não toca → publish → aluno vê título (e duração se houver) e reproduz.  
**FORA DESTE CHAT**: Biblioteca dedicada; transcode; CF de mídia; redesign 001.

### Implementation

- [ ] T034 [US4] Estender model/API: `status` `rascunho|publicado`, campos opcionais `titulo`, `duracao_segundos`; GET aluno só `publicado`; gestores leem rascunho — depends: Fase 4 Done — aceite: [data-model.md](./data-model.md) P2; 404 PT para aluno em rascunho
- [ ] T035 [US4] Endpoints/handlers: upload com status inicial; publish/unpublish (ou PATCH) em `backend/internal/handler/vod_handler.go` + testes Go — depends: T034 — aceite: aluno não reproduz rascunho; publish torna assistível; `go test` verde
- [ ] T036 [US4] UI em `AulasPage.tsx`: título/duração no bloco Gravação; controles rascunho/publicar só se `canManageAulas` — depends: T035 — aceite: Independent Test US4 passa na UI

### Done — Fase 5 (opcional)

- [ ] Aluno não reproduz rascunho
- [ ] Publish torna assistível com título (e duração se informada)
- [ ] Unpublish/remove remove reprodução do aluno
- [ ] Testes Go dos handlers atualizados

**Checkpoint**: P2 completo se executado.

---

## Dependencies & Execution Order

### Phase Dependencies

```text
Fase 1 (API local) ──► Fase 2 (UI) ──► Fase 3 (Terraform S3) ──► Fase 4 (seed+docs)
                                                                      │
                                                                      └──► Fase 5 (P2, opcional)
```

- **Fase 1**: sem dependência de UI/IaC/docs
- **Fase 2**: Depends Fase 1 Done
- **Fase 3**: Depends Fase 2 Done (demo UI+API antes do wire AWS; API S3 pode ser testada via publish)
- **Fase 4**: Depends Fase 3 Done (seed na imagem/AWS coerente com bucket)
- **Fase 5**: Depends Fase 4 Done; **não** bloqueia P1

### User Story Mapping

| US | Spec | Fases |
|----|------|-------|
| US1 | Aluno assiste na ficha | 1 (API) + 2 (UI); seed em 4 |
| US2 | Admin/professor publica | 1 (API) + 2 (UI) |
| US3 | Ciclo apply/demo/destroy | 3 (IaC) + 4 (seed+docs) |
| US4 | Rascunho/metadados | 5 (opcional) |

### Parallel Opportunities (dentro da mesma fase)

- **Fase 1**: T001 ∥ T003 ∥ T006; depois T004 ∥ T005; handlers sequenciais
- **Fase 2**: T014 ∥ preparação; T017 pode paralelizar com polish de T015
- **Fase 3**: T022 ∥ após T019; T023 após IAM/env
- **Fase 4**: T028 ∥ T030 ∥ T031 após asset/migration base
- **Entre fases**: **não** paralelizar chats — um chat = uma fase

### Parallel Example: Fase 1

```bash
# Em paralelo (arquivos distintos):
Task: "T001 config VOD_* em backend/internal/config/config.go"
Task: "T003 model AulaVod em backend/internal/models/"
Task: "T006 validação MP4/50MB"

# Depois, em paralelo:
Task: "T004 vodstorage local"
Task: "T005 vod_repository.go"
```

---

## Implementation Strategy

### MVP (P1) — Chats 1→4

1. Chat 1 = Fase 1 — API + storage local + testes
2. Chat 2 = Fase 2 — UI ficha aula
3. Chat 3 = Fase 3 — S3 Terraform + wire
4. Chat 4 = Fase 4 — seed + runbook
5. **STOP** — demo P1 completa (SC-001…SC-012 relevantes)

### Incremental opcional

5. Chat 5 = Fase 5 — só se P2 for pedido explicitamente

### Sugestão imediata

**Começar `/speckit-implement` somente na Fase 1 (Chat 1).** Não misturar UI, Terraform ou docs nesse chat.

---

## Notes

- Baseline 001 fechada: sem NAT, sem Redis AWS, sem OAuth, destroy entre sessões, budget separado, RBAC aulas = escrita VOD
- Playback AWS = S3 presigned (~15 min); **sem** CloudFront de mídia no P1
- Upload = multipart pela API (unifica local/AWS)
- Sem página Biblioteca; sem download; um VOD vigente por aula
- Mensagens API em português; datas BR
- Commitar só se o usuário pedir
