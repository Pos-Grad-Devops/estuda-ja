---
description: "Task list — Certificado ao final do curso (fases = chats isolados)"
---

# Tasks: Certificado ao final do curso

**Input**: Design documents from `/specs/005-course-certificate/`

**Prerequisites**: [plan.md](./plan.md), [spec.md](./spec.md), [research.md](./research.md), [data-model.md](./data-model.md), [contracts/](./contracts/), [quickstart.md](./quickstart.md), constitution

**Baseline**: `001`–`004` — **NÃO reabrir** (destroy, sem NAT, sem Redis AWS, VOD/live/chat intactos, budget separado, `desired_count = 1`).

**Tests**: Obrigatórios na Fase 1 (handlers/RBAC/lazy emit/invalidação — constitution V / FR-014). E2E Compose/UI = [quickstart.md](./quickstart.md) §A (Fase 2). CI **sem** AWS.

**Organization**: Fases **1–3** = **um chat `/speckit-implement` cada**. Não misturar API + UI + seed/docs no mesmo chat.

**FORA DESTA ENTREGA**: US5 / FR-018 (template rico, lista agregada, reemissão elaborada) — **sem tasks**; S3 de certificado; página/rota “Certificados”; gestão por professor; blockchain; Cognito; redesign 001–004; live→VOD; moderação chat P2.

---

## Formato e protocolo de chats

### Formato da task

`- [ ] [ID] [P?] [Story?] Descrição com path — depends: … — aceite: …`

- **[P]**: paralelizável (arquivos distintos, sem depender de task incompleta)
- **[USn]**: user story do [spec.md](./spec.md) (US1 aluno PDF · US2 admin gestão · US3 ciclo efêmero · US4 seed/local/CI · **US5 P2 = fora**)
- Setup/foundational embutidos na Fase 1 (sem fases Setup/Foundational separadas)

### Regra de ouro (NÃO negociável)

1. Abrir **um** chat `/speckit-implement` **somente** para a fase atual.
2. Completar **todas** as tasks da fase + checklist **Done** verificável.
3. Só então abrir o próximo chat na fase seguinte.
4. **Não** começar Fase 2+ sem Done da 1; **não** misturar UI (2) com API (1), nem seed/docs (3) com API/UI.

### Ordem dos chats

| Chat | Fase | Escopo | Prioridade |
|------|------|--------|------------|
| 1 | **1** | Models + repository + `certpdf` + handlers + testes Go | P1 / US1+US2+US4 (API) |
| 2 | **2** | UI painel certificado em `CursosPage` | P1 / US1+US2 (UI) |
| 3 | **3** | Seed elegibilidade + docs/runbook (sem tf cert) | P1 / US3+US4 (seed/docs) |

**Começar implementação pela Fase 1 apenas.** Não implementar neste artefato.

---

## Phase 1 — Model/API + PDF local + testes Go — CHAT 1

**Stories**: US1 (lazy emit/PDF) + US2 (RBAC gestão admin) + US4 (CI/testes) — **só backend**  
**Goal**: Models `CertificadoElegibilidade` + `Certificado`; repositório; `certpdf` (fpdf+TTF); handlers status / pdf lazy / elegibilidade / list / invalidar; RBAC só admin na gestão; erros PT; datas BR; testes Go. **Sem UI. Sem seed. Sem docs longas.**  
**Independent Test**: `go test ./...` verde; curl: admin marca elegibilidade (sem emitir); aluno 1ª `GET .../pdf` cria registro + PDF; 2ª reutiliza; invalidar bloqueia; reabilitar + nova solicitação = novo id; professor 403 em gestão.  
**FORA DESTE CHAT**: UI (`CursosPage` / painel); migration seed `004_seed_certificado_demo`; README/AGENTS longos; Terraform/S3 de cert; página “Certificados”; gestão professor; P2.

### Implementation

- [ ] T001 [P] [US1] Criar models `CertificadoElegibilidade` (`certificado_elegibilidades`, UNIQUE `(user_id, curso_id)`) e `Certificado` (`certificados`: `status` `valido`\|`invalidado`, `emitido_em`, snapshots `aluno_nome`/`curso_titulo`, timestamps `timeutil.DateTime`) em `backend/internal/models/` (arquivos dedicados ou extensão clara) — conforme [data-model.md](./data-model.md) — depends: nenhuma — aceite: alinhado ao data-model; sem FK para CRUD `Alunos`; constantes de status exportadas
- [ ] T002 [US1] Registrar AutoMigrate de ambos os models em `backend/internal/database/database.go` (e índice/invariante “no máximo um `valido` por par” se aplicável via unique parcial ou enforcement no repositório) — depends: T001 — aceite: `Migrate` cria tabelas; baseline models 001–004 intactos
- [ ] T003 [P] [US1] Criar `backend/internal/repository/certificado_repository.go`: CRUD elegibilidade (upsert/get/delete por par); find ativo `valido`; listar por curso; criar `valido` com snapshots; invalidar + delete elegibilidade atômico — depends: T001 — aceite: regras FR-001a/FR-007a; marcar elegibilidade não cria certificado
- [ ] T004 [P] [US1] Criar pacote `backend/internal/certpdf/` (wrapper `github.com/go-pdf/fpdf`) + TTF embutida sob `backend/assets/certs/` (ex. DejaVuSans) para acentos PT; API que gera 1 página A4 (título fixo + nome + curso + data BR) a partir dos snapshots — [research.md](./research.md) §1–2 — depends: nenhuma — aceite: bytes PDF válidos; sem S3/disco de PDF; `go get` da lib no `go.mod`
- [ ] T005 [P] [US4] Atualizar `backend/Dockerfile` para `COPY` `assets/certs` (espelhar padrão VOD) se a fonte não for só `embed` — depends: T004 — aceite: imagem Fargate/local gera PDF com acentos; **sem** volume de PDFs; **sem** vars `CERT_*`
- [ ] T006 [US2] Implementar em `backend/internal/handler/certificado_handler.go`: `GET /api/v1/cursos/:id/certificado` (aluno próprio; admin `?user_id=`) e `PUT /api/v1/cursos/:id/certificados/elegibilidade` (**só admin**, body `{user_id}`, só `role=aluno`, **não** emite) — [contracts/certificado-api.md](./contracts/certificado-api.md) — depends: T002, T003 — aceite: erros PT; 200 elegível sem criar certificado; professor/aluno gestão → 403
- [ ] T007 [US1] No mesmo `certificado_handler.go`: `GET /api/v1/cursos/:id/certificado/pdf` — lazy emit se elegível sem ativo (criar `valido` + stream PDF; rollback se PDF falhar); reutilizar ativo; 403 PT se não elegível/invalidado; `Content-Type: application/pdf` + `Content-Disposition` — depends: T004, T006 — aceite: FR-001a/FR-002; sem URL S3; só o próprio aluno baixa/emite
- [ ] T008 [US2] No mesmo `certificado_handler.go`: `GET /api/v1/cursos/:id/certificados` (lista admin) e `POST /api/v1/cursos/:id/certificados/:certId/invalidar` (status→`invalidado` **e** remove elegibilidade do par) — depends: T006 — aceite: após invalidar, PDF bloqueado até reabilitar; reabilitar + nova pdf = **novo** id; invalidado permanece
- [ ] T009 [US4] Wire rotas em `backend/cmd/api/main.go`: Authenticate + `RequireRoles(admin)` nas rotas de gestão; status/pdf com papéis de leitura de cursos (PDF restrito ao subject JWT); **sem** rotas flat `/certificados`; **sem** alterar VOD/live/chat — depends: T007, T008 — aceite: superfície = [certificado-api.md](./contracts/certificado-api.md); [certificado-env.md](./contracts/certificado-env.md) (zero `CERT_*`)
- [ ] T010 [US4] Testes Go em `backend/internal/handler/certificado_handler_test.go` (SQLite `:memory:`): admin marca sem emitir; aluno 1ª pdf cria; 2ª reutiliza mesmo id; invalidar bloqueia; reabilitar + pdf = novo id; professor 403 gestão; aluno 403 gestão; anônimo 401; PDF `application/pdf`; erros PT — depends: T009 — aceite: `go test ./...` verde; constitution V / FR-014; **sem** AWS
- [ ] T011 [US1] Smoke curl/Compose (sem UI): login admin → PUT elegibilidade → login aluno → 1ª GET pdf → 2ª GET pdf (mesmo id) → admin invalidar → aluno pdf 403 → PUT reabilitar → aluno pdf novo id → professor PUT/POST gestão 403 — depends: T010, T005 — aceite: checklist Done da Fase 1

### Done — Fase 1 (obrigatório antes do Chat 2)

- [ ] `go test ./...` verde (handlers/RBAC/lazy emit/invalidação)
- [ ] Admin marca elegibilidade **sem** criar certificado
- [ ] Aluno 1ª `GET .../pdf` cria registro `valido` + PDF; 2ª reutiliza o mesmo id
- [ ] Invalidar bloqueia download/nova emissão; reabilitar + nova solicitação = **novo** id
- [ ] Professor (e aluno) → 403 em gestão
- [ ] **Nenhuma** mudança UI; **nenhuma** migration seed; **nenhum** Terraform/S3 de cert; **sem** página “Certificados”

**Checkpoint**: API certificado pronta. Parar. Abrir novo chat só para Fase 2.

---

## Phase 2 — UI no contexto do curso — CHAT 2

**Stories**: US1 + US2 — **só frontend**  
**Goal**: Painel mínimo em `CursosPage` (ou `CursoCertificadoPanel`): aluno status + baixar PDF; admin marcar/reabilitar, consultar, invalidar; `canManageCertificados` (só admin); **sem** rota “Certificados”; professor sem controles de gestão.  
**Independent Test**: Aluno (elegível via API/admin) baixa PDF na UI do curso; admin invalida/reabilita pela UI; professor não vê gestão; live/VOD/chat intactos.  
**FORA DESTE CHAT**: novos handlers Go (salvo bug bloqueante mínimo); seed migration; README/AGENTS longos; Terraform; P2 lista/template rico.

### Implementation

- [ ] T012 [P] [US2] Adicionar `canManageCertificados(role) => role === 'admin'` em `frontend/src/auth/auth.ts` — depends: Fase 1 Done — aceite: só admin; professor/aluno false; **não** amplia poderes live/VOD/chat
- [ ] T013 [P] [US1] Estender `frontend/src/api/client.ts` com chamadas: status `GET .../certificado`, pdf `GET .../certificado/pdf` (blob/download), PUT elegibilidade, GET listagem admin, POST invalidar — depends: Fase 1 Done — aceite: alinhado a [certificado-api.md](./contracts/certificado-api.md); Bearer JWT; erros PT propagados
- [ ] T014 [US1] Criar `frontend/src/components/CursoCertificadoPanel.tsx` (ou bloco equivalente): aluno vê status + botão solicitar/baixar; feedback PT se não elegível; datas BR se exibidas — depends: T013 — aceite: FR-001/SC-001/SC-009; sem portal separado
- [ ] T015 [US2] No mesmo painel: controles **somente** se `canManageCertificados` — marcar/reabilitar elegibilidade (`user_id`), consultar emissão, invalidar; professor e aluno **não** veem gestão — depends: T012, T014 — aceite: US2 AC1–5 / FR-007/FR-008; SC-003 UI
- [ ] T016 [US1] Embutir painel em `frontend/src/pages/CursosPage.tsx` no **contexto do curso**; confirmar `App.tsx` **sem** rota/página “Certificados”; live/VOD/chat em `AulasPage` intactos — depends: T015 — aceite: SC-007; descoberta só via curso; sem redesign 001–004
- [ ] T017 [US1] `npm run build` + smoke manual Compose: aluno baixa PDF na UI; admin invalida → aluno bloqueado → admin reabilita → novo PDF; professor sem gestão — depends: T016 — aceite: checklist Done da Fase 2 / [quickstart.md](./quickstart.md) §A (UI)

### Done — Fase 2 (obrigatório antes do Chat 3)

- [ ] Aluno baixa PDF na UI do curso (status + download)
- [ ] Admin marca/consulta/invalida/reabilita pela UI do curso
- [ ] Professor **sem** controles de gestão de certificado
- [ ] Sem rota/página “Certificados”; `npm run build` OK
- [ ] Live / VOD / chat intactos

**Checkpoint**: UI P1 pronta. Parar. Abrir novo chat só para Fase 3.

---

## Phase 3 — Seed + docs/runbook — CHAT 3

**Stories**: US3 (ciclo efêmero / destroy) + US4 (seed + docs)  
**Goal**: Migration seed **só** elegibilidade aluno×curso demo (sem certificado pré-emitido); README/AGENTS/quickstart no ciclo apply→publish→demo certificado→destroy; confirmar **zero** Terraform/S3 de cert; destroy = RDS; budget intacto.  
**Independent Test**: Pós-seed: elegível sem PDF pré-emitido; runbook ≤ 15 min incremental (SC-005); destroy documentado; budget intacto.  
**FORA DESTE CHAT**: código de API/UI novo (salvo nota/docs); S3/tf de certificado; redesign 001–004; P2.

### Implementation

- [ ] T018 [US4] Migration gormigrate `004_seed_certificado_demo` em `backend/internal/database/migrations/` + registro em `migrations.go`: localiza aluno seed + curso demo (`002_seed_demo`); insere **só** `CertificadoElegibilidade` se ausente; **não** cria `Certificado` — depends: Fase 2 Done — aceite: idempotente; SC-010; FR-013
- [ ] T019 [P] [US3] Atualizar [quickstart.md](./quickstart.md): seed = elegibilidade apenas; fluxo A/B/C; tabela destroy (RDS remove linhas; N/A S3 cert); tempo SC-001/SC-005 — depends: T018 — aceite: avaliador segue sem reabrir 001–004
- [ ] T020 [P] [US3] Atualizar `README.md` (Demo AWS / runbook): passos incrementais de certificado no ciclo apply→publish→demo→destroy; local = emissão real na 1ª solicitação; CI = testes sem AWS; **sem** NAT/Redis/Cognito/S3 cert — depends: T018 — aceite: SC-005 (≤ 15 min incremental); FR-015
- [ ] T021 [P] [US4] Atualizar `AGENTS.md`: certificado 005 deixa de ser “fora do escopo”; documentar elegibilidade admin, lazy emit, PDF on-the-fly, RBAC só admin, seed só elegibilidade, P2 ainda fora; baseline 001–004 intacta — depends: T018 — aceite: agente/humano não trata cert como proibido; sem sugerir S3/página dedicada/gestão professor
- [ ] T022 [US3] Verificação final: confirmar **nenhum** `.tf` novo de certificado; `infra/budget/` intocado; [certificado-env.md](./contracts/certificado-env.md) respeitado; checklist Done da Fase 3 — depends: T019, T020, T021 — aceite: FR-011/FR-012/SC-004; destroy = registros somem com RDS

### Done — Fase 3 (feature P1 completa)

- [ ] Seed: aluno demo elegível **sem** certificado pré-emitido
- [ ] README/AGENTS/quickstart descrevem apply→…→demo certificado→destroy + efeito no RDS (sem S3 cert)
- [ ] Zero Terraform de certificado; budget `infra/budget/` intacto
- [ ] Runbook incremental ≤ 15 min (SC-005); P2 **não** entregue; 001–004 **não** reabertos

**Checkpoint**: Feature 005 P1 completa. Parar.

---

## Dependencies & Execution Order

### Phase Dependencies

```text
Chat 1 (Fase 1: Model/API/PDF/testes)
    │
    ▼  Done Fase 1
Chat 2 (Fase 2: UI CursosPage)
    │
    ▼  Done Fase 2
Chat 3 (Fase 3: Seed + docs)
    │
    ▼  Done Fase 3 = P1 pronto
(US5 template/lista P2 — sem chat / sem tasks nesta rodada)
```

- **Fase 1**: sem dependência de UI/seed/docs; bloqueia Fase 2
- **Fase 2**: depende do Done da Fase 1; bloqueia Fase 3 por protocolo (UI validável antes do seed/docs)
- **Fase 3**: depende do Done da Fase 2; única fase que adiciona seed + docs longas

### User Story Mapping

| Story | Onde |
|-------|------|
| US1 aluno obtém PDF | Fase 1 (API lazy) + Fase 2 (UI download) |
| US2 admin gestão | Fase 1 (API RBAC) + Fase 2 (painel admin) |
| US3 ciclo efêmero | Fase 3 (docs destroy; sem tf) |
| US4 seed/local/CI | Fase 1 (testes) + Fase 3 (seed + docs) |
| US5 template/lista | **Excluída** |

### Within Each Phase

- Models → AutoMigrate → repository; `certpdf` ∥ models; handlers após repo+pdf; wire → testes → smoke
- UI: auth helper ∥ client → painel → `CursosPage` → build/smoke
- Seed antes das docs que afirmam o caminho feliz com elegibilidade pré-marcada

### Parallel Opportunities

**Fase 1:** T001 ∥ T004; T003 após T001; T005 ∥ (após T004); T006 após T002+T003; T007 após T004+T006  
**Fase 2:** T012 ∥ T013; T014 após T013; T015 após T012+T014  
**Fase 3:** T019 ∥ T020 ∥ T021 após T018  

**Entre fases:** **não** paralelizar Chats 1–3 (regra de ouro).

---

## Parallel Example: Fase 1

```text
# Em paralelo no início do Chat 1:
Task: T001 models CertificadoElegibilidade + Certificado
Task: T004 certpdf + TTF em assets/certs

# Após T001:
Task: T002 AutoMigrate
Task: T003 certificado_repository.go

# Após T004:
Task: T005 Dockerfile COPY assets/certs (se necessário)
```

---

## Implementation Strategy

### MVP First (Fase 1 = Chat 1)

1. Completar **somente** Fase 1 (API + PDF + testes).
2. **STOP**: validar Done (`go test` + curl lazy/invalidar/RBAC).
3. Só então Chat 2 (UI).

### Incremental Delivery

1. Fase 1 → API demonstrável via curl/teste (MVP técnico).
2. Fase 2 → demo visual no contexto do curso (MVP de produto local).
3. Fase 3 → seed + runbook AWS/destroy (MVP acadêmico completo).

### Suggested MVP scope

**Chat 1 / Fase 1** (models + handlers + certpdf + testes). UI e seed/docs vêm nos chats seguintes.

---

## Notes

- Formato checklist: `- [ ] Tnnn [P?] [USn?] … — depends: … — aceite: …` em **todas** as tasks de implementação.
- **Sem** tasks de S3/cert bucket, página “Certificados”, gestão professor, template rico P2, lista agregada, blockchain, Cognito, redesign 001–004.
- Um chat = uma fase; Done antes da próxima.
- **Não** iniciar `/speckit-implement` além da Fase 1 até o Done correspondente.
- Identidade = `User` papel `aluno` (não CRUD Alunos).
