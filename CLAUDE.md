# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Leia primeiro

**[AGENTS.md](./AGENTS.md) é a fonte principal de guidelines deste repositório** (escopo do produto, stack, RBAC, convenções, o que está fora de escopo). Este arquivo complementa com comandos e arquitetura; não repita aqui o que já está no AGENTS.md — leia os dois.

Responda em **português** ao usuário. Não commite a menos que explicitamente pedido (ver seção Git no AGENTS.md).

## Comandos

### Setup local
```bash
make infra-up              # sobe Postgres + Redis (Docker) para dev local
cd backend && go run ./cmd/api      # API em :8080 (lê backend/.env.example)
cd frontend && cp .env.example .env && npm install && npm run dev   # Vite em :5173
```

### Stack completa via Docker
```bash
docker compose up --build   # API :8080, frontend :5173, health em /health
```

### Testes
```bash
make test                          # go test ./... (backend) + npm run build (frontend)
cd backend && go test ./...        # todos os testes Go
cd backend && go test ./internal/handler -run TestNomeDoTeste -v   # um teste específico
cd backend && go test ./internal/chat/...                          # um pacote específico
```
Testes de backend usam SQLite `:memory:` (não precisam do Postgres rodando). Não há suíte de testes automatizados no frontend — `npm run build` (tsc + vite build) é o check de CI.

### Lint / build frontend
```bash
cd frontend && npm run lint     # oxlint
cd frontend && npm run build    # tsc -b && vite build
```

### Infra AWS (demo efêmera)
```bash
cd infra && terraform apply     # sobe a demo (região fixa us-east-1)
cd infra && terraform destroy   # destrói entre sessões (NÃO destrói infra/budget/)
# Publish: infra/publish-api.ps1 e infra/publish-frontend.ps1 (PowerShell)
```
Runbook completo, ordem dos passos e checklist de warm-up: seção "Demo AWS" do README.md e `specs/001-aws-mvp-terraform/quickstart.md`.

## Arquitetura

### Visão geral
Monorepo com API Go separada de um frontend Vite/React, comunicando via REST + WebSocket sob `/api/v1/`. PostgreSQL (GORM) é o único armazenamento relacional; não há Redis em produção (só disponível no Compose local, sem uso ativo hoje). Terraform em `infra/` provisiona uma demo AWS efêmera (ECS Fargate + ALB + RDS + S3/CloudFront + um canal IVS), destruída entre sessões.

### Backend (`backend/`)
Organização por domínio, não por camada técnica dentro de cada pasta — cada recurso (curso, aula, aluno, user, live, chat, vod, certificado) tem seu próprio `*_handler.go` e `*_repository.go`:

```
cmd/api/main.go       # bootstrap: config, DB, migrations, rotas, middleware
internal/config/      # variáveis de ambiente (VOD_*, LIVE_BACKEND, IVS_*)
internal/database/    # conexão + migrations/ (gormigrate, versionadas e com seed)
internal/models/      # entidades GORM
internal/repository/  # um arquivo por domínio
internal/handler/     # um arquivo por domínio; helpers comuns em handler.go
internal/chat/        # hub WebSocket em memória, uma sala por aula_id
internal/certpdf/     # geração de PDF on-the-fly (fpdf + TTF embutida)
internal/vodstorage/  # storage de vídeo: local (dev) | s3 (AWS), Put/Open/Delete/Presign
internal/auth/        # emissão/validação de JWT
internal/middleware/  # Authenticate + RequireRoles
internal/timeutil/    # datas em formato BR (dd/mm/yyyy)
```

Pontos-chave para quem for mexer no backend:
- **Migrations fazem seed automático e idempotente**: `001_seed_admin`, `002_seed_demo` (professor/aluno/curso/aula), `003_seed_vod_demo` (copia `assets/vod/demo-aula.mp4`), `004_seed_certificado_demo` (só elegibilidade, sem PDF). São reexecutadas a cada subida da API — não assumir que o banco começa vazio.
- **`LIVE_BACKEND` (`stub` | `ivs`) e `VOD_BACKEND` (`local` | `s3`) trocam a implementação em runtime** via `internal/config`; `main.go` faz fail-fast se `ivs`/`s3` estiverem incompletos. Local/CI sempre usam `stub`/`local`.
- **Estado da live (`AulaLive.Status`) é independente de `Aula.Status`** (CRUD do curso) — não confundir os dois ao alterar qualquer um.
- **Chat é um hub in-memory por processo** (`internal/chat`), sem persistência e sem Redis — mensagens somem ao reiniciar a API e não há histórico ao reabrir a conexão.
- **Certificado é emitido lazily**: elegibilidade (seed/admin) e emissão do PDF são eventos separados; o PDF só existe após a primeira requisição do aluno.
- RBAC é aplicado via `RequireRoles` nas rotas em `main.go`, não dentro dos handlers.

### Frontend (`frontend/src/`)
```
api/client.ts     # wrapper fetch com JWT (inclui live, vod, chat, certificado)
auth/             # AuthContext, ProtectedRoute, helpers de permissão (canManageCursos, etc.)
components/       # AppShell, ui/, LivePlayer, AulaChatPanel, CursoCertificadoPanel
pages/            # Home, CursoPage (/cursos/:id), AulaPage (/aulas/:id — live+VOD+chat), Agenda, Login, Alunos, Usuarios
utils/date.ts     # datas BR, espelha internal/timeutil do backend
```
UI esconde (não apenas desabilita) ações que o perfil logado não pode executar, usando os helpers de `auth/auth.ts`. Não existem páginas dedicadas de "Biblioteca" (VOD) ou "Certificados" — ambos vivem no contexto da aula/curso.

### Infra (`infra/`)
Terraform "flat" (sem modules) para uma demo efêmera, não para produção contínua: `infra/ivs.tf` (canal IVS), `infra/alb.tf` (ALB com `idle_timeout=3600` para WebSocket do chat), scripts PowerShell de publish. `infra/budget/` é uma stack Terraform **separada** para alerta de billing — nunca é destruída junto com a demo.

### Specs (`specs/`)
Documentação estilo Speckit por feature (spec/plan/tasks/contracts/quickstart), uma pasta por número: `001-aws-mvp-terraform`, `002-vod-library`, `003-live-streaming-ivs`, `004-live-class-chat`, `005-course-certificate`, `006-escola-ui`. Ao investigar o comportamento esperado de uma feature, o `quickstart.md` e `contracts/` da pasta correspondente são a referência mais confiável — mais atual que o código às vezes.
