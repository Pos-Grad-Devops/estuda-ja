# Research: 006-escola-ui

**Date**: 2026-09-09 | **Git branch**: `feat/aws-speckit` | **Visual source**: `origin/feat/frontend-escola` @ `a99b6dc`

## 1. Checkout seletivo vs merge git

**Decision:** Incorporar o visual com **cópia/checkout apenas de caminhos sob `frontend/`** a partir de `a99b6dc` (ex.: `git checkout a99b6dc -- frontend/src/components/...` e arquivos listados). **Proibido** `git merge feat/frontend-escola` (ou merge da branch inteira). **Proibido** puxar `backend/`, `infra/`, `specs/`, `docs/` da escola.

**Rationale:** A escola partiu de uma main **sem** VOD/live/chat/certificado. Merge completo apagaria ou conflitaria com `AulasPage`/`CursosPage` e o `client.ts` ricos da baseline. Path-limited copy + fusão controlada preserva 002–005.

**Alternatives considered:**
- Merge completo + resolver conflitos → rejeitado (risco de perder comportamento; user proibiu).
- Redesign do zero no visual da escola → rejeitado (fonte já existe; YAGNI).
- Cherry-pick de commits da escola → rejeitado como estratégia primária (commits misturam fora de `frontend/` e histórico divergente); checkout por path é mais previsível.

## 2. Mapeamento página antiga → nova

| Superfície atual (`feat/aws-speckit`) | Destino escola | Ação |
|---------------------------------------|----------------|------|
| Layout header flat + links | `AppShell` | Substituir |
| `CursosPage` lista/CRUD | `HomePage` + `CursoModal` | Adotar escola |
| `CursosPage` bloco certificado | **`CursoPage`** (`/cursos/:id`) | Portar `CursoCertificadoPanel` |
| `AulasPage` lista/CRUD | `AgendaPage` + `AulaModal` + trilha em `CursoPage` | Adotar escola |
| `AulasPage` detalhe (seleção) live/VOD/chat | **`AulaPage`** (`/aulas/:id`) | Portar blocos; remover aviso MVP |
| `LoginPage` planilha | `LoginPage` escola | Adotar |
| `AlunosPage` / `UsuariosPage` | versões escola | Adotar |
| `LivePlayer`, `AulaChatPanel`, `CursoCertificadoPanel` | manter componentes; restyle com tokens/primitivos escola | Não inventar tema B |

**Rotas (adotar escola):**

| Path | Página |
|------|--------|
| `/login` | LoginPage |
| `/` | HomePage (catálogo) |
| `/cursos` | redirect → `/` |
| `/cursos/:id` | CursoPage |
| `/aulas` | AgendaPage |
| `/aulas/:id` | AulaPage |
| `/alunos`, `/usuarios` | admin |

`AulasPage` / `CursosPage` **deixam de ser** a experiência principal (podem ser removidas após port completo).

## 3. O que portar de AulasPage / CursosPage

### De `AulasPage` → `AulaPage`

- Estado e ações **live**: `api.live.get` / schedule / cancel / start / stop / getPlayback / getIngest; status `agendada` \| `ao_vivo` \| `encerrada` (e ausência/`inativa`); `LivePlayer`; ingest OBS só admin/professor (`canManageAulas`).
- **VOD**: `api.vod.get` / getPlayback / upload / remove; player; empty state; escrita só admin/professor; **sem** download MP4 / página Biblioteca.
- **Chat**: `AulaChatPanel` + `buildAulaChatWsUrl`; mensagens só da conexão atual.
- Remover da `AulaPage` escola o texto: *“O player de transmissão ainda não faz parte deste MVP.”*

### De `CursosPage` → `CursoPage`

- `CursoCertificadoPanel` (aluno: status/PDF; admin: elegibilidade/list/invalidar).
- Visibilidade: aluno ou `canManageCertificados`; professor **sem** gestão.
- Sem rota/página “Certificados”.

### CRUD

- Cursos: Home + CursoModal (admin).
- Aulas: Agenda / trilha CursoPage + AulaModal (admin/professor).

## 4. Fusão de `client.ts`

**Decision:** Partir do `client.ts` **atual** (aws-speckit). Acrescentar apenas o que a escola precisa e que falta hoje:

- `api.cursos.get(id)` → `GET /api/v1/cursos/:id`
- `api.aulas.get(id)` → `GET /api/v1/aulas/:id`
- Tipo `Curso`: campo opcional `aulas?: Aula[]` (como na escola)

**Manter intactos:** tipos e métodos `vod`, `live`, `certificados`, `buildAulaChatWsUrl`, CRUD alunos/users, auth.

**Não** sobrescrever o arquivo inteiro com o da escola.

**Nested `aulas` no GET curso:** a `CursoPage` escola usa `curso.aulas ?? []`. Se a resposta do GET não trouxer nested, fallback **só no frontend**: `api.aulas.list(cursoId)` (já existe) — **sem** mudar backend.

## 5. Design system aplicado a live / VOD / chat / certificado

**Decision:** `feat/frontend-escola` é a **única** fonte visual. Blocos 002–005 **não** ganham tema próprio.

**Seguir:**
- Primitivos: `AppShell`, `PageHeader`, `Button`, `StatusPill`, `EmptyState`, `Modal`, `CourseCard`, `LessonRow`, bits (`FadeContent`, etc.)
- Tipografia, espaçamento, cantos, bordas, superfícies, cores (incl. token `live`)
- Empty / loading / erro iguais aos da escola
- Variantes de botão da escola (primário, secundário, perigo); RBAC só mostra/esconde

**Proibido:**
- CSS/layout planilha só nos blocos 002–005
- Tema A no catálogo e tema B no player/chat/cert
- Tokens/paleta/componentes visuais novos “só para” live/VOD/chat/cert
- Lib visual além de tailwind/gsap/ogl da escola
- Terceiro visual

**`index.css` / Tailwind:** sistema da escola prevalece. Remover ou deixar de carregar tema planilha (`App.css` legado) como paralelo. Reaproveitar CSS mínimo de player/chat **somente** se já for funcional e puder ser alinhado aos tokens (classes Tailwind preferíveis).

**`StatusPill` e “ao vivo”:** a escola usa `Aula.status` CRUD. No produto atual, estado de transmissão é **`AulaLive.status`** (API live). **Decision:** na ficha da aula e em destaques de live, pills/ações de transmissão usam **`LiveStatus`** / `api.live`; `Aula.status` permanece o campo CRUD da aula (Agenda/trilha podem continuar a mostrá-lo como metadado da aula, sem confundir com live). Home “ao vivo”: preferir sinal live (`api.live` / filtro coerente), não assumir que `Aula.status === 'ao_vivo'` = canal IVS.

## 6. Tooling / deps

**Decision:** Alinhar `frontend/package.json` e `vite.config.ts` à escola: `tailwindcss@4`, `@tailwindcss/vite`, `gsap`, `ogl`. Rodar install no `frontend/`. CI: `npm run build` deve passar; Go tests intocados.

## 7. Auth

**Decision:** Manter `canManageCertificados` e demais helpers da baseline. A escola não os tem — **não** sobrescrever `auth.ts` cegamente; merge: shell/rotas da escola + helpers aws-speckit.

## 8. Escopo negativo (confirmado)

Sem P2 002–005, OAuth, Redis AWS, NAT, EventBridge 001, live→VOD, página Biblioteca/Certificados, merge git completo, redesign de marca além da escola, mudanças de API/contratos/seeds/Terraform.
