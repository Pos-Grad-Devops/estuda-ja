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
- Ciclo: apply → publish API → health → rebuild front → demo → **`terraform destroy` entre sessões**
- Budget/alerta da conta: stack **separada** em `infra/budget/` (limiar default **US$ 5**) — **não** destruir com a demo
- Segredos: SSM + `*.tfvars` / state locais gitignored — nunca commitados
- Runbook: [README.md](./README.md) (seção Demo AWS) e [specs/001-aws-mvp-terraform/quickstart.md](./specs/001-aws-mvp-terraform/quickstart.md)

| Prioridade | Escopo |
|------------|--------|
| **P1** | IaC + publish manual + destroy + docs |
| **P2** | Automação warm-up / EventBridge — **gap (Fase F / escolha b)**; checklist T−15 min no quickstart |
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
- **Biblioteca VOD P1** (`002-vod-library`): gravação por aula (upload/replace/delete + playback ~15 min); storage **local** (Compose) ou **S3** (demo AWS); UI na ficha da aula (`AulasPage`); seed `003_seed_vod_demo` + asset `backend/assets/vod/demo-aula.mp4`

### Fora do escopo (não implementar sem pedido)
- Streaming ao vivo
- Chat WebSocket
- OAuth externo (Cognito, etc.)
- Redis/ElastiCache na AWS
- Domínio customizado / certificados ACM custom
- EventBridge apply/destroy (P2) — só se o chat pedir explicitamente Fase F
- Página/rota “Biblioteca” dedicada; download do MP4; CloudFront de mídia; rascunho/título/duração VOD (P2 da feature 002)

### Permissões (RBAC)

| Recurso | admin | professor | aluno |
|---------|-------|-----------|-------|
| Cursos — leitura | ✓ | ✓ | ✓ |
| Cursos — escrita | ✓ | — | — |
| Aulas — leitura | ✓ | ✓ | ✓ |
| Aulas — escrita | ✓ | ✓ | — |
| VOD (gravação da aula) | leitura + escrita | leitura + escrita | só leitura |
| Alunos | CRUD | — | — |
| Usuários | CRUD | — | — |

Escrita VOD = mesmo recorte de aulas (`RequireRoles(admin, professor)`). **Não reabrir** a baseline `001-aws-mvp-terraform` (sem NAT, sem Redis AWS, destroy entre sessões, budget separado).

Admin inicial (migration `001_seed_admin` via gormigrate): `ADMIN_EMAIL` / `ADMIN_PASSWORD` (default `admin@estudaja.com` / `admin123`).

Seed demo (migration `002_seed_demo`): professor (`DEMO_PROFESSOR_EMAIL` / `DEMO_PROFESSOR_PASSWORD`, default `professor@estudaja.com` / `professor123`), aluno (`DEMO_ALUNO_*`, default `aluno@estudaja.com` / `aluno123`), 1 curso e 1 aula. Idempotente (create if missing). Defaults de **dev** apenas.

Seed VOD (migration `003_seed_vod_demo`): se a aula de demo existir e ainda não tiver VOD, copia `assets/vod/demo-aula.mp4` para o storage (`vod/aulas/{id}/current.mp4`) e cria metadados `publicado`. Idempotente. Domínio backend: `handler/vod_handler.go`, `repository/vod_repository.go`, `vodstorage/` (local \| s3).

## Backend (Go)

### Organização

```
backend/
├── cmd/api/main.go           # bootstrap, rotas, middleware
├── assets/vod/               # MP4 seed (demo-aula.mp4) — COPY no Dockerfile
├── internal/
│   ├── config/
│   ├── database/
│   ├── models/               # entidades GORM (incl. AulaVod)
│   ├── repository/           # um arquivo por domínio (*_repository.go)
│   ├── handler/              # um arquivo por domínio (*_handler.go)
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
├── api/client.ts       # fetch + token JWT
├── auth/               # AuthContext, permissions, ProtectedRoute
├── pages/              # uma page por recurso
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
2. Confirmar escopo antes de streaming, EventBridge (P2) ou features fora da lista
3. Na AWS: preferir destroy entre sessões; não sugerir NAT nem Redis AWS; CD = opcional (não substitui publish manual)
4. Manter RBAC e formato de datas ao tocar em API ou UI
5. Responder em português
