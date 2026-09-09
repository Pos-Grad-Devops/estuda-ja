# AGENTS.md — EstudaJá

Contexto e guidelines para humanos e agentes de IA trabalharem neste repositório.

## Projeto

**EstudaJá** — plataforma EdTech de cursos ao vivo em massa (aula magna) + VOD.

- Pico de acesso **instantâneo** no horário da aula (restrição crítica de arquitetura)
- SLA alvo: **99,9%** na janela ao vivo; **99%** fora dela
- Responder em **português** na comunicação com o usuário

Documentação de negócio: [DESCRICAO.md](./DESCRICAO.md) · Entregáveis do curso: [REQUISITOS.md](./REQUISITOS.md)

## Estrutura do repositório

```
estuda-ja/
├── backend/              # API Go (Fiber) + PostgreSQL
├── frontend/             # Vite + React + TypeScript
├── infra/                # Terraform AWS (demo) + publish-*.ps1 + budget/
├── docs/                 # Apresentação Marp, diagramas, scripts npm de slides
├── specs/                # Speckit (ex.: 001-aws-mvp-terraform)
├── README.md             # Documentação principal + runbook AWS
├── DESCRICAO.md          # Enunciado do projeto
├── REQUISITOS.md         # Entregáveis acadêmicos
├── docker-compose.yml
├── Makefile
└── .github/workflows/ci.yml
```

**Raiz enxuta:** só README, DESCRICAO, REQUISITOS e pastas de código/docs/specs. Não mover documentação de slides para a raiz.

## Stack (decisões fechadas)

| Camada | Tecnologia |
|--------|------------|
| Backend | **Go** + **Fiber** |
| Frontend | **Vite** + **React** + **TypeScript** |
| Banco | **PostgreSQL** (GORM) · na AWS: **RDS** `db.t4g.micro` |
| Cache | **Redis** só no Compose local (uso futuro) — **sem Redis/ElastiCache na AWS** neste MVP |
| Orquestração (demo) | **ECS Fargate** ARM64 256/512 + ALB · `desired_count = 1` |
| Frontend estático (demo) | **S3** + **CloudFront** (API também atrás de CloudFront HTTPS) |
| Auth | **JWT** + **bcrypt** (autenticação própria) |
| IaC | Terraform flat em `infra/` (sem modules no MVP) |
| CI | GitHub Actions (build + test; **CD opcional P3** em `cd.yml`) |

**Caminho obrigatório da demo:** AWS gerenciado. Self-hosted / Compose **não** é destino de demo ou “produção acadêmica” — só desenvolvimento local. Não sugerir VPS self-hosted, NAT Gateway, ElastiCache ou domínio customizado no escopo P1.

## Demo AWS (operação)

- Região: **us-east-1** · **sem NAT Gateway** (tasks com IP público)
- Ciclo: apply → publish API → health → rebuild front → **warm-up (incl. IVS/OBS)** → demo → **`terraform destroy` entre sessões**
- Budget/alerta da conta: stack **separada** em `infra/budget/` (limiar default **US$ 5**) — **não** destruir com a demo
- Segredos: SSM + `*.tfvars` / state locais gitignored — nunca commitados
- Runbook: [README.md](./README.md) (seção Demo AWS), [specs/001-aws-mvp-terraform/quickstart.md](./specs/001-aws-mvp-terraform/quickstart.md), streaming [specs/003-live-streaming-ivs/quickstart.md](./specs/003-live-streaming-ivs/quickstart.md), chat [specs/004-live-class-chat/quickstart.md](./specs/004-live-class-chat/quickstart.md), certificado [specs/005-course-certificate/quickstart.md](./specs/005-course-certificate/quickstart.md)

| Prioridade | Escopo |
|------------|--------|
| **P1** | IaC + publish manual + destroy + docs (+ live IVS efêmero + chat WS) |
| **P2** | Automação warm-up / EventBridge apply–destroy **001** — **gap**; streaming: checklist T−15 + `infra/check-live-warmup.ps1` (health/GetStream, sem apply/destroy) |
| **P3** | CD GitHub Actions — **opcional** (`.github/workflows/cd.yml`; publish manual `infra/publish-*.ps1` permanece) |

## Escopo atual vs futuro

### Implementado
- CRUD: Cursos, Aulas, Alunos
- Auth + RBAC: `admin`, `professor`, `aluno`
- Gestão de usuários (admin)
- Datas em formato BR: `dd/mm/yyyy` ou `dd/mm/yyyy HH:mm`
- Docker Compose local, pipeline CI
- Demo AWS efêmera (Terraform + publish manual P1)
- CD opcional P3 (Actions; gate test/build; secrets de sessão efêmera)
- **Biblioteca VOD P1** (`002-vod-library`): gravação por aula (upload/replace/delete + playback ~15 min); storage **local** (Compose) ou **S3** (demo AWS); UI na ficha da aula (`AulaPage` `/aulas/:id`); seed `003_seed_vod_demo` + asset `backend/assets/vod/demo-aula.mp4`
- **Streaming ao vivo P1+P2** (`003-live-streaming-ivs`): `AulaLive` + handlers `live_*`; **stub** local/CI (`LIVE_BACKEND=stub`); **IVS** na demo AWS (`LIVE_BACKEND=ivs`); UI bloco **Transmissão ao vivo** + `LivePlayer` na `AulaPage`; estados `agendada` \| `ao_vivo` \| `encerrada` (agendada exige horário); warm-up streaming: checklist manual + `infra/check-live-warmup.ps1` (sem EventBridge 001)
- **Chat da aula P1** (`004-live-class-chat`): hub em memória por `aula_id` (`internal/chat`); `GET /api/v1/aulas/:id/chat/ws?token=...`; UI `AulaChatPanel` na ficha da aula; ALB `idle_timeout = 3600`; local WS real; CI = testes Go sem AWS; **sem** Redis AWS / API GW WS / Lambda; **sem** histórico ao reabrir
- **Certificado P1** (`005-course-certificate`): elegibilidade admin (`CertificadoElegibilidade`) + emissão lazy na 1ª `GET .../certificado/pdf` (PDF on-the-fly via `certpdf`/fpdf+TTF); UI `CursoCertificadoPanel` na ficha do curso (`CursoPage` `/cursos/:id`); gestão **só admin**; seed `004_seed_certificado_demo` = **só** elegibilidade (sem PDF pré-emitido); **sem** S3/tf/`CERT_*` de certificado; destroy = linhas somem com o RDS
- **UI escola P1** (`006-escola-ui`): AppShell + catálogo em `/`; fichas `/cursos/:id` e `/aulas/:id`; Agenda `/aulas`; **sem** páginas Biblioteca/Certificados; live/VOD/chat na aula; certificado no curso

### Fora do escopo (não implementar sem pedido)
- Moderação de chat P2 / IVS Chat (US5 de `004`)
- Certificado P2 (template rico, lista agregada, reemissão elaborada — US5 de `005`)
- Pipeline **live→VOD** (gravação IVS→S3 automática)
- Tokenização / playback authorization IVS; multi-canal; DVR
- Automação EventBridge apply/destroy da stack **001** (gap P2 / Fase F — só se pedido explícito)
- OAuth externo (Cognito, etc.)
- Redis/ElastiCache na AWS
- Domínio customizado / certificados ACM custom
- Página/rota “Biblioteca” dedicada; download do MP4; CloudFront de mídia; rascunho/título/duração VOD (P2 da feature 002)
- Página/rota “Certificados” dedicada; gestão de certificado por professor; S3 de PDF de certificado

### Permissões (RBAC)

| Recurso | admin | professor | aluno |
|---------|-------|-----------|-------|
| Cursos — leitura | ✓ | ✓ | ✓ |
| Cursos — escrita | ✓ | — | — |
| Aulas — leitura | ✓ | ✓ | ✓ |
| Aulas — escrita | ✓ | ✓ | — |
| VOD (gravação da aula) | leitura + escrita | leitura + escrita | só leitura |
| Live (status/playback) | leitura | leitura | leitura |
| Live (schedule/cancel/start/stop/ingest) | ✓ | ✓ | — |
| Chat (WS enviar/receber na aula) | ✓ | ✓ | ✓ |
| Certificado (status/PDF próprio) | ✓ | ✓ | ✓ (próprio) |
| Certificado (elegibilidade / list / invalidar) | ✓ | — | — |
| Alunos | CRUD | — | — |
| Usuários | CRUD | — | — |

Escrita VOD e gestão live = mesmo recorte de aulas (`RequireRoles(admin, professor)`). Chat = papéis com leitura de aulas (não amplia ingest/VOD). Gestão de certificado = **só admin** (`canManageCertificados` / `RequireRoles(admin)`). **Não reabrir** a baseline `001-aws-mvp-terraform` (sem NAT, sem Redis AWS, destroy entre sessões, budget separado).

Admin inicial (migration `001_seed_admin` via gormigrate): `ADMIN_EMAIL` / `ADMIN_PASSWORD` (default `admin@estudaja.com` / `admin123`).

Seed demo (migration `002_seed_demo`): professor (`DEMO_PROFESSOR_EMAIL` / `DEMO_PROFESSOR_PASSWORD`, default `professor@estudaja.com` / `professor123`), aluno (`DEMO_ALUNO_*`, default `aluno@estudaja.com` / `aluno123`), 1 curso e 1 aula. Idempotente (create if missing). Defaults de **dev** apenas.

Seed VOD (migration `003_seed_vod_demo`): se a aula de demo existir e ainda não tiver VOD, copia `assets/vod/demo-aula.mp4` para o storage (`vod/aulas/{id}/current.mp4`) e cria metadados `publicado`. Idempotente. Domínio backend: `handler/vod_handler.go`, `repository/vod_repository.go`, `vodstorage/` (local \| s3).

Seed certificado (migration `004_seed_certificado_demo`): se aluno seed + curso demo existirem e ainda não houver elegibilidade do par, cria **só** `CertificadoElegibilidade`. **Não** cria `Certificado` (emissão lazy na 1ª solicitação PDF). Idempotente. Domínio: `handler/certificado_handler.go`, `repository/certificado_repository.go`, `certpdf/` + TTF em `assets/certs/`. Contratos: [certificado-api](./specs/005-course-certificate/contracts/certificado-api.md), [certificado-env](./specs/005-course-certificate/contracts/certificado-env.md) (zero `CERT_*`; sem S3 de cert).

Live (feature `003`): modelo `AulaLive` (`handler/live_handler.go`, `repository/live_repository.go`); status `agendada` \| `ao_vivo` \| `encerrada` (ausência = `inativa`); schedule/cancel/start/stop; env `LIVE_BACKEND` (`stub` \| `ivs`) + `IVS_*` ([contracts/live-env](./specs/003-live-streaming-ivs/contracts/live-env.md)). Stub ignora `IVS_*`; AWS fail-fast se `ivs` incompleto. `Aula.Status` CRUD **não** é o estado da live. Sem seed `ao_vivo`. IaC: `infra/ivs.tf` + SSM/ECS; output só `ivs_channel_arn`. Warm-up: [check-live-warmup.ps1](./infra/check-live-warmup.ps1).

Chat (feature `004`): hub in-process (`internal/chat`) por `aula_id`; `handler/chat_handler.go`; path `GET /api/v1/aulas/:id/chat/ws?token=...` (JWT na query; não logar URI com token); texto ≤ 500; sala independente do status live; mensagens efêmeras à conexão; UI `AulaChatPanel` na `AulaPage`. ALB `idle_timeout = 3600` (`infra/alb.tf`); sem stickiness; sem Redis/API GW WS. Moderação P2 **fora**. Contratos: [chat-ws](./specs/004-live-class-chat/contracts/chat-ws.md), [chat-env](./specs/004-live-class-chat/contracts/chat-env.md).

Certificado (feature `005` P1): models `CertificadoElegibilidade` + `Certificado` (`status` `valido`\|`invalidado`); admin marca elegibilidade **sem** emitir; aluno elegível solicita PDF → lazy emit + stream `application/pdf`; invalidar remove elegibilidade; reabilitar + nova solicitação = **novo** id; UI no contexto do curso (`CursoCertificadoPanel` em `CursoPage`); **sem** página “Certificados”; **sem** Terraform/S3 de certificado; P2 (template/lista) **fora**.

## Backend (Go)

### Organização

```
backend/
├── cmd/api/main.go           # bootstrap, rotas, middleware
├── assets/vod/               # MP4 seed (demo-aula.mp4) — COPY no Dockerfile
├── assets/certs/             # TTF embutida (DejaVuSans) para PDF — COPY no Dockerfile
├── internal/
│   ├── config/               # VOD_*, LIVE_BACKEND, IVS_* (sem CERT_*)
│   ├── database/
│   ├── models/               # entidades GORM (incl. AulaVod, AulaLive, Certificado*)
│   ├── repository/           # um arquivo por domínio (*_repository.go)
│   ├── handler/              # um arquivo por domínio (*_handler.go; live/chat/certificado)
│   ├── chat/                 # hub WebSocket em memória por aula_id
│   ├── certpdf/              # PDF on-the-fly (fpdf + TTF)
│   ├── vodstorage/           # Storage local | S3 (Put/Open/Delete; Presign no S3)
│   ├── middleware/
│   ├── auth/
│   └── timeutil/             # datas BR (DateTime JSON + parse)
```

**Convenções:**
- Separar handlers e repositories **por domínio** (não monolito em um arquivo)
- Helpers compartilhados em `handler.go` (`parseID`, `handleError`)
- Mensagens de erro da API em **português**
- Endpoints sob `/api/v1/`
- Rotas protegidas: middleware `Authenticate` + `RequireRoles` em `main.go`
- Testes em `*_test.go` com SQLite `:memory:`

### Datas

- Usar `timeutil.DateTime` nos models para JSON
- Layouts: `02/01/2006` (data) · `02/01/2006 15:04` (data+hora)
- Parse/format em `internal/timeutil/datetime.go`

## Frontend (React)

### Organização

```
frontend/src/
├── api/client.ts       # fetch + token JWT (incl. live + vod + chat + certificado)
├── auth/               # AuthContext, permissions, ProtectedRoute
├── components/         # AppShell; ui/; LivePlayer; AulaChatPanel; CursoCertificadoPanel
├── pages/              # Home, CursoPage (/cursos/:id), AulaPage (/aulas/:id + live/VOD/chat), Agenda, Login, Alunos, Usuarios
└── utils/date.ts       # datas BR (espelhar backend)
```

**Convenções:**
- `import type` para tipos (verbatimModuleSyntax)
- Permissões UI via helpers em `auth/auth.ts` (`canManageCursos`, etc.)
- Esconder formulários/botões quando o perfil não pode escrever
- `VITE_API_URL` aponta para a API (default `http://localhost:8080`; na AWS = `api_url` da sessão — rebuild obrigatório)

## Comandos úteis

```bash
make infra-up      # Postgres + Redis (local)
make backend       # API local
make frontend      # Vite dev
make test          # go test + npm build
docker compose up --build

cd infra && terraform apply    # demo AWS
cd infra && terraform destroy  # entre sessões
# Publish: infra/publish-api.ps1 ; infra/publish-frontend.ps1
# CD opcional: .github/workflows/cd.yml (AWS_CD_ENABLED + secrets da sessão)

cd docs && npm run slides:html   # exportar apresentação
```

## Git e commits

- **Não commitar** unless o usuário pedir explicitamente
- **Não commitar** `.env`, `*.tfvars` (exceto `*.tfvars.example`), `*.tfstate*`, `.terraform/`, secrets, `node_modules/`, builds (`dist/`, `apresentacao.html`)
- Mensagens de commit em português, focadas no *porquê*

## Princípios de código

1. **Escopo mínimo** — só o pedido; sem refatorar o unrelated
2. **Seguir convenções existentes** — nomes, estrutura de pastas, estilo
3. **Sem over-engineering** — nada de abstrações prematuras
4. **Testes** — backend ao alterar handlers; não adicionar testes triviais
5. **Documentação** — atualizar README/AGENTS quando mudar stack, auth, estrutura ou caminho AWS

## Apresentação

- Slides em `docs/apresentacao.md` (Marp)
- Diagramas: `docs/slides/diagrams/*.mmd` → SVG via `npm run slides:diagrams`
- Marp CLI v4 **não renderiza Mermaid** inline — usar SVGs gerados

## Ao iniciar um chat novo

1. Ler este arquivo e, se necessário, [README.md](./README.md) (Desenvolvimento + Demo AWS)
2. Confirmar escopo: streaming `003`, chat `004` e certificado `005` P1 **completos** — não tratar como proibidos; **não** implementar certificado P2, moderação de chat P2, live→VOD ou EventBridge apply/destroy do 001 sem pedido explícito
3. Na AWS: preferir destroy entre sessões; não sugerir NAT nem Redis AWS; CD = opcional (não substitui publish manual); warm-up T−15 inclui canal IVS + OBS **antes** de T−0; script `check-live-warmup.ps1` só health/sinal; chat some com a API no destroy; certificados somem com o RDS (sem S3 cert)
4. Manter RBAC e formato de datas ao tocar em API ou UI (live = mesmo recorte de aulas; chat não amplia ingest/VOD; certificado gestão = só admin)
5. Responder em português
6. Escopo de produto [DESCRICAO.md](./DESCRICAO.md) (live + chat + VOD + certificado P1) **fechado** — próximos itens só sob pedido explícito (P2 / gaps)
