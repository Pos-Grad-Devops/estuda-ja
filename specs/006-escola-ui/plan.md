# Implementation Plan: UI da escola (visual P1)

**Branch git**: `feat/aws-speckit` (não criar/trocar branch) | **Spec feature**: `006-escola-ui` | **Date**: 2026-09-09 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/006-escola-ui/spec.md` + abordagem de incorporação visual (checkout seletivo de `frontend/` a partir de `origin/feat/frontend-escola` @ `a99b6dc`). Baseline **001–005** intacta. **Só frontend.**

## Summary

Substituir a UI “planilha/CRUD” da baseline por o **visual e navegação de escola** já existentes em `feat/frontend-escola`, **sem** merge git da branch inteira e **sem** alterar API/RBAC/backend/infra/seeds/contratos 001–005. Em seguida, **vestir** live, VOD, chat e certificado (comportamento já entregue) nas fichas `AulaPage` e `CursoPage` usando **exclusivamente** o design system da escola (AppShell, ui/, course/, forms/, bits/, tokens Tailwind). Critério: quem só viu Home/Curso/Login reconhece player, OBS, VOD, chat e certificado como a mesma UI.

## Technical Context

**Language/Version**: TypeScript (frontend Vite); Go/API **intocados**

**Primary Dependencies (frontend)**: React 19 + React Router; adotar da escola **Tailwind CSS v4** (`@tailwindcss/vite`), **gsap**, **ogl**. Sem lib visual extra. Sem dependências novas de backend.

**Storage**: N/A (sem mudanças de persistência)

**Testing**: `npm run build` no `frontend/` (obrigatório); testes Go **intocados**; validação manual das jornadas seed em [quickstart.md](./quickstart.md)

**Target Platform**: Browser (dev local Compose + demo AWS já publicada — só rebuild do front se necessário)

**Project Type**: Monorepo; escopo desta feature = pasta `frontend/` (+ docs Speckit / menção README/AGENTS se a navegação mudar)

**Performance Goals**: UI responsiva de demo; animações bits (gsap/ogl) não bloqueiam jornadas seed; sem impacto no warm-up T−0 da live (só apresentação)

**Constraints**: Sem merge `feat/frontend-escola`; só paths `frontend/`; preservar `api.vod` / `api.live` / chat WS / `api.certificados` + `canManageCertificados`; design system da escola é fonte da verdade para 002–005; sem tema paralelo planilha; sem página Biblioteca/Certificados; sem P2; datas BR; RBAC inalterado

**Scale/Scope**: Superfícies UI: Home, Curso (`/cursos/:id`), Agenda, Aula (`/aulas/:id`), Login, Alunos, Usuários; blocos portados: live + VOD + chat → AulaPage; certificado → CursoPage

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Princípio / constraint | Status | Como o plan atende |
|------------------------|--------|-------------------|
| I. Custo-consciente | PASS | Sem recurso AWS novo; só UI; rebuild front na sessão existente |
| II. Pronto para o pico | PASS | Não altera IVS/ALB/warm-up; live continua o mesmo caminho, só muda apresentação |
| III. Escopo mínimo / YAGNI | PASS | Só UI; sem backend/Terraform/seeds/contratos API novos; P2 fora |
| IV. Segurança e RBAC | PASS | Mesmos papéis e helpers; controles só mostram/escondem; `canManageCertificados` preservado |
| V. Qualidade testável | PASS | `npm run build`; smoke jornadas seed no quickstart; Go tests intocados |
| VI. Observabilidade | PASS | Sem telemetria nova |
| VII. Documentação / Speckit | PASS | Artefatos nesta pasta; README/AGENTS só se navegação/estrutura front mudar na implement |
| VIII. IaC | PASS | Terraform intocado |
| Stack fechada | PASS | Vite/React/TS; deps novas = só as da escola |
| Baseline 001–005 | PASS | API/infra/contratos/seeds intocados; comportamento 002–005 portado, não reescrito |
| Self-hosted como demo | PASS | Compose = dev; demo AWS inalterada no desenho |

**Post-design re-check:** PASS — research fecha checkout vs merge, mapeamento de páginas, fusão `client.ts` e aplicação do design system; data-model = mapa de superfícies (sem entidades backend novas); contracts = UI/rotas referenciando 002–005; quickstart = jornadas seed. Sem violação a justificar.

## Project Structure

### Documentation (this feature)

```text
specs/006-escola-ui/
├── plan.md                 # Este arquivo
├── research.md             # Phase 0
├── data-model.md           # Phase 1 — mapa de superfícies UI
├── quickstart.md           # Phase 1 — validação visual + jornadas
├── contracts/
│   └── ui-routes.md        # Paths, fichas, RBAC de controles
├── checklists/
│   └── requirements.md
└── tasks.md                # NÃO criado aqui → /speckit-tasks
```

### Source Code (escopo de mudança)

```text
frontend/
├── package.json / package-lock.json   # + tailwind v4, gsap, ogl; vite plugin
├── vite.config.ts                     # + @tailwindcss/vite
├── src/
│   ├── index.css                      # tokens/tema escola (prevalece)
│   ├── App.tsx                        # rotas + AppShell da escola
│   ├── api/client.ts                  # MANTER vod/live/chat/cert; + cursos.get, aulas.get; Curso.aulas?
│   ├── auth/auth.ts                   # MANTER canManageCertificados (+ helpers existentes)
│   ├── components/
│   │   ├── layout/AppShell.tsx        # DA ESCOLA
│   │   ├── ui/ course/ forms/ bits/   # DA ESCOLA (design system)
│   │   ├── LivePlayer.tsx             # COMPORTAMENTO atual; visual alinhado aos tokens
│   │   ├── AulaChatPanel.tsx          # idem
│   │   └── CursoCertificadoPanel.tsx  # idem → usado em CursoPage
│   ├── pages/
│   │   ├── HomePage / CursoPage / AulaPage / AgendaPage / LoginPage / Alunos / Usuarios
│   │   ├── AulaPage                   # + blocos live/VOD/chat (port de AulasPage)
│   │   └── CursoPage                  # + CursoCertificadoPanel (port de CursosPage)
│   └── utils/date.ts, course.ts
└── (remover ou deixar de usar como UX principal)
    pages/AulasPage.tsx, pages/CursosPage.tsx, App.css “planilha”
```

**Structure Decision:** Monorepo atual; **apenas** `frontend/`. Incorporação por **checkout/cópia seletiva** de paths da escola (`git checkout a99b6dc -- frontend/...` ou equivalente path-limited), depois fusão manual de `client.ts`, auth e port dos painéis 002–005. Backend/, infra/, specs de 001–005, seeds: **fora**.

## Complexity Tracking

> Nenhuma violação de constitution a justificar. Tabela omitida.
