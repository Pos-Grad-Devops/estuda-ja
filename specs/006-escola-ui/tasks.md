---
description: "Task list — UI da escola visual P1 (fases = chats isolados)"
---

# Tasks: UI da escola (visual P1)

**Input**: Design documents from `/specs/006-escola-ui/`

**Prerequisites**: [plan.md](./plan.md), [spec.md](./spec.md), [research.md](./research.md), [data-model.md](./data-model.md), [contracts/ui-routes.md](./contracts/ui-routes.md), [quickstart.md](./quickstart.md), constitution

**Git branch**: `feat/aws-speckit` — **NÃO** criar/trocar branch. Spec dir = `006-escola-ui`.

**Baseline**: `001`–`005` — **NÃO reabrir** (API, RBAC, backend, Terraform, publish, seeds, contratos). Só `frontend/` (+ docs se nav mudar).

**Tests**: **Sem** testes Go novos; **sem** TDD frontend. Aceite = checklist Done + `npm run build` em `frontend/`. Smoke = [quickstart.md](./quickstart.md).

**Organization**: Fases **1–3** = **um chat `/speckit-implement` cada**. Não misturar.

**FORA DESTA ENTREGA / FORA DE TASKS**: merge git `feat/frontend-escola`; backend/infra/seeds/contratos API; testes Go; TDD front; P2 de 002–005; página Biblioteca/Certificados; live→VOD; OAuth; Redis AWS; NAT; EventBridge 001; redesign além da escola; **commitar** (só se o usuário pedir depois).

**Design system (fonte da verdade)**: `origin/feat/frontend-escola` @ `a99b6dc` — primitivos `AppShell`, `PageHeader`, `Button`, `StatusPill`, `EmptyState`, `Modal`, `CourseCard`, `LessonRow`, bits; tokens em `frontend/src/index.css`. Live/VOD/chat/cert **não** ganham tema B, paleta nova nem lib visual extra.

---

## Formato e protocolo de chats

### Formato da task

`- [ ] [ID] [P?] [Story?] Descrição com path — depends: … — aceite: …`

- **[P]**: paralelizável (arquivos distintos, sem depender de task incompleta)
- **[USn]**: user story do [spec.md](./spec.md) (US1 navegação escola · US2 live/VOD/chat na AulaPage · US3 certificado na CursoPage · US4 jornadas RBAC)
- Setup/foundational embutidos na **Fase 1** (como 004/005)

### Regra de ouro (NÃO negociável)

1. Abrir **um** chat `/speckit-implement` **somente** para a fase atual.
2. Completar **todas** as tasks da fase + checklist **Done** verificável.
3. Só então abrir o próximo chat na fase seguinte.
4. **Não** começar Fase 2+ sem Done da 1; **não** misturar port live/VOD/chat (2) com shell/US1 (1); **não** misturar certificado/polish (3) com Fase 2.
5. **Começar implementação só pela Fase 1**, em outro chat. Este artefato **não** implementa código.

### Ordem dos chats

| Chat | Fase | Escopo | Stories |
|------|------|--------|---------|
| 1 | **1** | Checkout seletivo + deps + fusão `client.ts`/`auth` + AppShell/rotas + páginas escola (**sem** portar live/VOD/chat/cert) | Setup + Foundational + **US1** |
| 2 | **2** | Port live+VOD+chat `AulasPage` → `AulaPage`; vestir `LivePlayer`/`AulaChatPanel` no DS escola | **US2** |
| 3 | **3** | `CursoCertificadoPanel` → `CursoPage`; jornadas RBAC; remover UX planilha; README/AGENTS se nav mudou | **US3** + **US4** + polish |

---

## Phase 1 — Setup + Foundational + US1 (shell/navegação escola) — CHAT 1

**Stories**: Setup/Foundational + **US1** (catálogo, fichas, login, alunos, usuários)  
**Goal**: Incorporar visual/navegação da escola via **checkout path-limited** de `frontend/` @ `a99b6dc`; deps Tailwind/gsap/ogl; fundir `client.ts` e `auth.ts` sem perder 002–005; AppShell + rotas `/`, `/cursos/:id`, `/aulas`, `/aulas/:id`; páginas Home/Curso/Aula/Agenda/Login/Alunos/Usuarios no DS escola. **AINDA SEM** portar live/VOD/chat/certificado. **NÃO** apagar `AulasPage`/`CursosPage` (referência para Fases 2–3). **NÃO** `git merge` da branch escola.  
**Independent Test**: Contas seed: login → Home → `/cursos/:id` → `/aulas/:id` e Agenda; tom escola; `npm run build` OK; `AulasPage`/`CursosPage` ainda no repo; client ainda tem vod/live/cert/chat helpers.  
**FORA DESTE CHAT**: blocos live/VOD/chat na AulaPage; certificado na CursoPage; deletar AulasPage/CursosPage; merge git; backend/infra; commit.

### Implementation

- [x] T001 Checkout/cópia **path-limited** a partir de `a99b6dc` (`origin/feat/frontend-escola`) **somente** sob `frontend/` (ex.: `components/layout|ui|course|forms|bits`, pages escola `HomePage`/`CursoPage`/`AulaPage`/`AgendaPage`/`LoginPage`/`AlunosPage`/`UsuariosPage`, `utils/course.ts`, `assets/` necessários, `index.css` escola) — **sem** `git merge`; **sem** puxar `backend/`/`infra/`/`specs/`/`docs/` — depends: nenhuma — aceite: arquivos escola presentes em `frontend/`; working tree permanece em `feat/aws-speckit`; nenhum path fora de `frontend/` veio da escola
- [x] T002 Alinhar tooling: `frontend/package.json` (+ lock) com deps escola (`tailwindcss@4`, `@tailwindcss/vite`, `gsap`, `ogl`); `frontend/vite.config.ts` com plugin Tailwind; `npm install` em `frontend/` — depends: T001 — aceite: `npm run build` consegue resolver deps; **sem** lib visual além da escola
- [x] T003 Fundir `frontend/src/api/client.ts`: **partir do client atual**; **manter** `vod`, `live`, `certificados`, `buildAulaChatWsUrl` e tipos; **acrescentar** `cursos.get`, `aulas.get` e `Curso.aulas?: Aula[]` — **não** sobrescrever com o client da escola — depends: T001 — aceite: métodos 002–005 intactos; `get` por id tipados; build TypeScript ok neste arquivo
- [x] T004 [P] Fundir `frontend/src/auth/auth.ts`: adotar o que a escola precisar **sem** perder `canManageCertificados` e helpers `canManageCursos`/`Aulas`/`Alunos`/`Users` da baseline — depends: T001 — aceite: `canManageCertificados(role) === (role === 'admin')`; papéis inalterados
- [x] T005 [US1] Wire `frontend/src/App.tsx` + `frontend/src/components/layout/AppShell.tsx`: rotas conforme [contracts/ui-routes.md](./contracts/ui-routes.md) (`/` Home, `/cursos`→`/`, `/cursos/:id`, `/aulas`, `/aulas/:id`, `/alunos`, `/usuarios`, `/login`); ProtectedRoute; nav Início·Agenda·(admin) Alunos/Usuários — depends: T002, T003, T004 — aceite: visitante só login; autenticado usa shell escola; **sem** rota Biblioteca/Certificados
- [x] T006 [P] [US1] Integrar páginas escola já checkoutadas: `frontend/src/pages/HomePage.tsx`, `AgendaPage.tsx`, `LoginPage.tsx`, `AlunosPage.tsx`, `UsuariosPage.tsx` com AppShell/PageHeader/Button/EmptyState/Modal/CourseCard e tokens de `frontend/src/index.css`; datas BR via `utils/date.ts` — depends: T005 — aceite: CRUD/listagens no visual escola; RBAC de escrita inalterado (admin cursos/alunos/users; professor aulas via Agenda/Curso)
- [x] T007 [US1] Integrar `frontend/src/pages/CursoPage.tsx`: carregar via `api.cursos.get`; trilha com `LessonRow`; se `curso.aulas` ausente, fallback `api.aulas.list(cursoId)` (**sem** backend); CRUD curso/aula com `CursoModal`/`AulaModal` + `canManageCursos`/`canManageAulas`; **ainda sem** `CursoCertificadoPanel` — depends: T003, T005 — aceite: `/cursos/:id` mostra trilha e navega para `/aulas/:id`; empty/loading com `EmptyState`
- [x] T008 [US1] Integrar `frontend/src/pages/AulaPage.tsx` **esqueleto** escola: meta da aula via `api.aulas.get`; `PageHeader`/`Button`/`StatusPill` só para metadado `Aula.status` CRUD se a escola já mostrar; **remover** o aviso “player ainda não faz parte deste MVP”; **ainda sem** portar live/VOD/chat — depends: T005 — aceite: `/aulas/:id` abre sem aviso MVP; placeholders/empty alinhados ao DS; `AulasPage.tsx` **permanece** no repo como referência
- [x] T009 [US1] Tema único: `frontend/src/index.css` da escola prevalece; deixar de importar/usar `App.css` (ou CSS planilha) como tema paralelo em `main.tsx`/`App.tsx`; **não** apagar ainda `AulasPage.tsx`/`CursosPage.tsx` — depends: T006, T007, T008 — aceite: 0 tema planilha na navegação principal; arquivos de referência ainda existem
- [x] T010 [US1] `npm run build` em `frontend/` + smoke US1: login seed → Home → curso → aula → Agenda; alunos/usuários (admin) no visual escola — depends: T009 — aceite: checklist Done da Fase 1

### Done — Fase 1 (obrigatório antes do Chat 2)

- [x] Checkout path-limited só `frontend/`; **sem** merge git da escola
- [x] `npm run build` OK (Tailwind + gsap/ogl)
- [x] Rotas `/`, `/cursos/:id`, `/aulas`, `/aulas/:id` funcionam no AppShell
- [x] Tom escola (não planilha) em Home/Curso/Aula/Agenda/Login/Alunos/Usuários
- [x] `client.ts` ainda tem vod/live/certificados/`buildAulaChatWsUrl` + `cursos.get`/`aulas.get`
- [x] `canManageCertificados` preservado
- [x] **Sem** live/VOD/chat/cert portados nesta fase; **sem** deletar `AulasPage`/`CursosPage`
- [x] **Sem** mudança backend/infra/seeds; **sem** commit neste chat (salvo pedido explícito do usuário)

**Checkpoint**: Shell + US1 prontos. Parar. Abrir novo chat só para Fase 2.

---

## Phase 2 — US2 live + VOD + chat na AulaPage — CHAT 2

**Story**: **US2**  
**Goal**: Portar comportamento de `frontend/src/pages/AulasPage.tsx` (blocos transmissão, gravação, chat) para `frontend/src/pages/AulaPage.tsx`; vestir `LivePlayer` e `AulaChatPanel` com primitivos/tokens da escola; status de transmissão = `api.live` / `LiveStatus` (**não** `Aula.status` CRUD); remover qualquer resto do aviso MVP.  
**Independent Test**: Aluno vê status/player live, VOD e chat; professor/admin gerem live+VOD + ingest OBS; aluno sem escrita; empty states DS; `npm run build` OK.  
**FORA DESTE CHAT**: certificado/`CursoCertificadoPanel`; deletar CursosPage (pode remover só AulasPage se port completo e build verde); README/AGENTS longos; backend; tema B / paleta nova.

### Implementation

- [x] T011 [US2] Em `frontend/src/pages/AulaPage.tsx`, bloco **Transmissão ao vivo**: portar lógica de `AulasPage.tsx` (`api.live.get` / schedule / cancel / start / stop / getPlayback / getIngest); exibir status com `StatusPill` (ou equivalente) alimentado por **`LiveStatus`** (`inativa`\|`agendada`\|`ao_vivo`\|`encerrada`); ações com `Button` (primário/secundário/perigo) só se `canManageAulas`; empty/loading com `EmptyState`; seção sob `PageHeader` — depends: Fase 1 Done — aceite: FR-004; status live ≠ `Aula.status` CRUD; aluno só leitura; 0 aviso “fora do MVP”
- [x] T012 [US2] Vestir `frontend/src/components/LivePlayer.tsx` com tokens de `frontend/src/index.css` (superfície, cantos, borda, cor `live` se aplicável); embutir no bloco live da `AulaPage`; **proibido** CSS planilha legado como tema do player e **proibido** paleta/componente visual novo “só player” — depends: T011 — aceite: player reconhecível como mesma UI de Home/Curso; playback quando `ao_vivo` conforme API
- [x] T013 [US2] Em `frontend/src/pages/AulaPage.tsx`, bloco **Gravação VOD**: portar de `AulasPage.tsx` (`api.vod.get` / getPlayback / upload / remove); player `<video>` ou equivalente; publicar/substituir/remover só `canManageAulas` via `Button`; empty com `EmptyState`; **sem** download MP4; **sem** rota Biblioteca — depends: Fase 1 Done — aceite: FR-005 / RBAC 002; visual com mesmos tokens/primitivos (`PageHeader` seção, `Button`, `EmptyState`, `Modal` se upload/confirm)
- [x] T014 [US2] Em `frontend/src/pages/AulaPage.tsx`, bloco **Chat**: embutir `frontend/src/components/AulaChatPanel.tsx` + `buildAulaChatWsUrl`; vestir painel com tokens/`Button`/`EmptyState` (sem tema B); mensagens só da conexão; sem moderação; sem histórico ao reabrir — depends: Fase 1 Done — aceite: FR-006 / 004; aluno/professor/admin usam o mesmo painel; chat não amplia ingest/VOD
- [x] T015 [P] [US2] Revisar Home/Agenda/`StatusPill` se destacarem “ao vivo”: não tratar `Aula.status === 'ao_vivo'` como substituto de live IVS; destaques de transmissão usam `api.live` quando a UI afirmar canal ao vivo — [research.md](./research.md) §5 — depends: T011 — aceite: sem confundir CRUD status com `AulaLive`
- [x] T016 [US2] `npm run build` + smoke US2 (seed): ficha `/aulas/:id` com live/VOD/chat no DS escola; professor gestão; aluno só leitura live/VOD write — depends: T012, T013, T014, T015 — aceite: checklist Done da Fase 2 / [quickstart.md](./quickstart.md) jornada aula

### Done — Fase 2 (obrigatório antes do Chat 3)

- [x] Live + VOD + chat na `AulaPage` com comportamento 002–004
- [x] Status transmissão = `api.live`; 0 aviso “fora do MVP”
- [x] `LivePlayer` + `AulaChatPanel` + bloco VOD no **mesmo** DS (AppShell/PageHeader/Button/StatusPill/EmptyState/Modal/tokens) — sem tema B
- [x] Aluno sem gestão live/VOD write; sem download MP4; sem Biblioteca
- [x] `npm run build` OK; **sem** certificado ainda (Fase 3); **sem** backend

**Checkpoint**: US2 pronta. Parar. Abrir novo chat só para Fase 3.

---

## Phase 3 — US3 certificado + US4 jornadas + polish — CHAT 3

**Stories**: **US3** + **US4** + polish  
**Goal**: Portar `CursoCertificadoPanel` para `CursoPage`; vestir no DS escola; validar jornadas aluno/professor/admin; remover UX planilha (`CursosPage`/`AulasPage` se ainda referenciadas); atualizar README/AGENTS **somente** se a navegação mudou o suficiente para o runbook.  
**Independent Test**: Quickstart completo ([quickstart.md](./quickstart.md)); 0 planilha como UX principal; 0 página Certificados/Biblioteca; RBAC igual 001–005; `npm run build` OK.  
**FORA DESTE CHAT**: novos endpoints; P2; merge git; redesign de marca; commit salvo pedido explícito.

### Implementation

- [x] T017 [US3] Vestir `frontend/src/components/CursoCertificadoPanel.tsx` com primitivos/tokens da escola (`Button`, `EmptyState`, `Modal` se confirmções, tipografia/espaçamento de `index.css`); **proibido** tema planilha ou paleta “só certificado” — depends: Fase 2 Done — aceite: painel visualmente da mesma UI que Home/Curso
- [x] T018 [US3] Embutir painel em `frontend/src/pages/CursoPage.tsx` (contexto do curso): aluno status/PDF se elegível; gestão só se `canManageCertificados`; professor **sem** gestão; confirmar `App.tsx` **sem** rota/página “Certificados” — depends: T017 — aceite: FR-007 / US3 AC; comportamento 005 inalterado
- [x] T019 [US4] Smoke jornadas seed (aluno / professor / admin) conforme [quickstart.md](./quickstart.md) e US4: permissões iguais à baseline; datas BR; visitante só login — depends: T018 — aceite: SC-001/SC-002/SC-003; 0 regressão de produto atribuível ao visual
- [x] T020 Remover UX planilha: desreferenciar/deletar `frontend/src/pages/CursosPage.tsx` e `frontend/src/pages/AulasPage.tsx` se o port estiver completo; limpar imports mortos; garantir que nenhuma rota aponta para elas — depends: T018 — aceite: experiência principal = escola; build sem imports quebrados
- [x] T021 [P] Atualizar `README.md` e/ou `AGENTS.md` **somente se** a navegação (rotas fichas / AppShell) exigir: documentar `/cursos/:id`, `/aulas/:id`, home catálogo; **não** reabrir 001–005 — depends: T020 — aceite: docs batem com [ui-routes.md](./contracts/ui-routes.md); se nav já coberta, task = N/A documentado
- [x] T022 [US4] Verificação final: `npm run build`; checklist visual do quickstart (0 planilha, 0 aviso MVP, 0 Biblioteca, 0 Certificados); confirmar frontend-only (diff sem backend/infra/seeds) — depends: T019, T020, T021 — aceite: Done Fase 3 = feature P1 completa

### Done — Fase 3 (feature P1 completa)

- [x] Certificado na `CursoPage` no DS escola; professor sem gestão; sem página Certificados
- [x] Jornadas seed aluno/professor/admin OK (US4)
- [x] UX planilha removida da experiência principal; `npm run build` OK
- [x] README/AGENTS atualizados se necessário
- [x] Baseline 001–005 (API/infra) intacta; P2 **não** entregue

**Checkpoint**: Feature 006 P1 completa. Parar.

---

## Dependencies & Execution Order

### Phase Dependencies

```text
Chat 1 (Fase 1: Setup + client/auth + AppShell + US1 páginas)
    │
    ▼  Done Fase 1
Chat 2 (Fase 2: US2 live/VOD/chat → AulaPage + DS)
    │
    ▼  Done Fase 2
Chat 3 (Fase 3: US3 cert + US4 jornadas + polish)
    │
    ▼  Done Fase 3 = P1 pronto
```

- **Fase 1**: bloqueia Fase 2; **não** porta 002–005 ainda; mantém AulasPage/CursosPage como referência
- **Fase 2**: depende Done Fase 1; **não** toca certificado
- **Fase 3**: depende Done Fase 2; única fase que remove páginas planilha e fecha US3/US4

### User Story Mapping

| Story | Onde |
|-------|------|
| US1 navegação/catálogo escola | **Fase 1** |
| US2 live/VOD/chat na ficha | **Fase 2** |
| US3 certificado na ficha curso | **Fase 3** |
| US4 jornadas RBAC | **Fase 3** (smoke) + validação cruzada após US2/US3 |

### Within Each Phase

- Fase 1: checkout → deps → client ∥ auth → AppShell/rotas → pages → tema → build
- Fase 2: live → LivePlayer → VOD → chat (VOD/chat ∥ após esqueleto live se arquivos distintos) → status Home → build
- Fase 3: vestir painel cert → CursoPage → jornadas → remover planilha → docs opcional → build final

### Parallel Opportunities

**Fase 1:** T003 ∥ T004 após T001; T006 ∥ (após T005) com cuidado em App; T007/T008 após T005  
**Fase 2:** T013 ∥ T014 após Fase 1 (arquivos/blocos distintos na AulaPage — coordenar se mesmo arquivo); T015 ∥ após T011  
**Fase 3:** T021 ∥ após T020  

**Entre fases:** **não** paralelizar Chats 1–3 (regra de ouro).

---

## Parallel Example: Fase 1

```text
# Após T001 (checkout):
Task: T003 fundir client.ts (manter vod/live/cert + get)
Task: T004 fundir auth.ts (manter canManageCertificados)

# Após T002 + T003 + T004:
Task: T005 App.tsx + AppShell + rotas

# Após T005 (coordenar se paralelo em pages distintas):
Task: T006 Home/Agenda/Login/Alunos/Usuarios
Task: T007 CursoPage (sem cert)
Task: T008 AulaPage esqueleto (sem live/VOD/chat)
```

---

## Implementation Strategy

### MVP First (Fase 1 = Chat 1)

1. Completar **somente** Fase 1 (shell + navegação escola).
2. **STOP**: validar Done (`npm run build` + smoke US1).
3. Só então Chat 2 (live/VOD/chat).

### Incremental Delivery

1. Fase 1 → demo visual catálogo/fichas (MVP visual).
2. Fase 2 → demo aula completa (live/VOD/chat).
3. Fase 3 → certificado + jornadas + limpeza planilha (MVP acadêmico completo).

### Suggested MVP scope

**Chat 1 / Fase 1** (US1). Começar `/speckit-implement` **apenas** na Fase 1.

---

## Notes

- Formato checklist: `- [ ] Tnnn [P?] [USn?] … — depends: … — aceite: …` em **todas** as tasks de implementação.
- Tasks de vestir 002–005 **citam** primitivos: `AppShell`, `PageHeader`, `Button`, `StatusPill`, `EmptyState`, `Modal`, tokens `frontend/src/index.css`.
- **Sem** tasks de merge git, backend, infra, seeds, contratos API, testes Go, TDD, P2, Biblioteca, Certificados, commit.
- Um chat = uma fase; Done antes da próxima.
- **Não** iniciar `/speckit-implement` além da Fase 1 até o Done correspondente.
