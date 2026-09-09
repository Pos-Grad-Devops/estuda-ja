# Contract: UI routes & surfaces (006-escola-ui)

**Tipo:** contrato de interface (rotas + o que cada ficha mostra + RBAC de controles).  
**Não** redefine APIs REST/WS — referências:

- VOD: [specs/002-vod-library/contracts/vod-api.md](../../002-vod-library/contracts/vod-api.md)
- Live: [specs/003-live-streaming-ivs/contracts/live-api.md](../../003-live-streaming-ivs/contracts/live-api.md)
- Chat: [specs/004-live-class-chat/contracts/chat-ws.md](../../004-live-class-chat/contracts/chat-ws.md)
- Certificado: [specs/005-course-certificate/contracts/certificado-api.md](../../005-course-certificate/contracts/certificado-api.md)

## Rotas

| Método UI | Path | Página | Auth |
|-----------|------|--------|------|
| view | `/login` | LoginPage | público |
| view | `/` | HomePage | autenticado |
| redirect | `/cursos` → `/` | — | autenticado |
| view | `/cursos/:id` | CursoPage | autenticado |
| view | `/aulas` | AgendaPage | autenticado |
| view | `/aulas/:id` | AulaPage | autenticado |
| view | `/alunos` | AlunosPage | admin |
| view | `/usuarios` | UsuariosPage | admin |

**Proibido como rota/página dedicada:** `/biblioteca`, `/certificados` (qualquer path equivalente de “Biblioteca” ou “Certificados”).

**Experiência principal:** não usar `AulasPage` / `CursosPage` monolíticas de planilha.

## Ficha do curso (`/cursos/:id`)

**Mostra:**
- Título, descrição, ações CRUD do curso se `canManageCursos` (admin)
- Trilha de aulas → links para `/aulas/:id`; criar aula se `canManageAulas`
- Painel de certificado (comportamento 005)

**Certificado — controles:**

| Controle | admin | professor | aluno |
|----------|-------|-----------|-------|
| Ver status / baixar próprio PDF se elegível | ✓ (próprio se aplicável) | ✓ (próprio se aplicável) | ✓ (próprio) |
| Marcar/reabilitar elegibilidade | ✓ | — | — |
| Listar elegibilidades / certificados do curso | ✓ | — | — |
| Invalidar certificado | ✓ | — | — |

## Ficha da aula (`/aulas/:id`)

**Mostra:**
- Meta da aula (título, curso, horário BR, etc.)
- Bloco **transmissão ao vivo** (003)
- Bloco **gravação VOD** (002)
- Bloco **chat** (004)
- **Não** mostra aviso de que player/live/VOD “não fazem parte do MVP”

**Live — controles:**

| Controle | admin | professor | aluno |
|----------|-------|-----------|-------|
| Ver status / player quando aplicável | ✓ | ✓ | ✓ |
| Agendar / cancelar / iniciar / encerrar | ✓ | ✓ | — |
| Ver ingest OBS | ✓ | ✓ | — |

**VOD — controles:**

| Controle | admin | professor | aluno |
|----------|-------|-----------|-------|
| Assistir gravação publicada | ✓ | ✓ | ✓ |
| Publicar / substituir / remover | ✓ | ✓ | — |
| Download do MP4 | — | — | — |

**Chat — controles:**

| Controle | admin | professor | aluno |
|----------|-------|-----------|-------|
| Enviar / receber na conexão atual | ✓ | ✓ | ✓ |
| Histórico ao reabrir / moderação | — | — | — |

## Home / Agenda / Login / Admin

- **Home:** catálogo; CRUD curso se admin; tom escola.
- **Agenda:** listagem de aulas; CRUD se `canManageAulas`.
- **Login:** único destino do visitante para áreas protegidas.
- **Alunos / Usuários:** só admin; visual escola.

## Cliente HTTP (frontend) — acréscimos mínimos

Além do client já existente em 002–005, a UI de fichas exige:

| Método | Path API | Uso UI |
|--------|----------|--------|
| `cursos.get(id)` | `GET /api/v1/cursos/:id` | CursoPage |
| `aulas.get(id)` | `GET /api/v1/aulas/:id` | AulaPage |

Sem novos endpoints de produto. Se `GET curso` não vier com `aulas[]`, a UI MAY usar `aulas.list(cursoId)` já existente.

## Visual (aceitação)

Quem só viu Home / Curso / Login deve reconhecer player, ingest OBS, VOD, chat e certificado como a **mesma** UI (mesmos primitivos e tokens). Proibido tema paralelo “planilha” só nos blocos 002–005.
