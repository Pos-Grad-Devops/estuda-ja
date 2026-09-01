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
├── docs/                 # Apresentação Marp, diagramas, scripts npm de slides
├── README.md             # Documentação principal + plano
├── DESCRICAO.md          # Enunciado do projeto
├── REQUISITOS.md         # Entregáveis acadêmicos
├── docker-compose.yml
├── Makefile
└── .github/workflows/ci.yml
```

**Raiz enxuta:** só README, DESCRICAO, REQUISITOS e pastas de código/docs. Não mover documentação de slides para a raiz.

## Stack (decisões fechadas)

| Camada | Tecnologia |
|--------|------------|
| Backend | **Go** + **Fiber** |
| Frontend | **Vite** + **React** + **TypeScript** |
| Banco | **PostgreSQL** (GORM) |
| Cache | **Redis** (infra pronta; uso futuro) |
| Orquestração alvo | **ECS Fargate** (documentado; deploy pendente) |
| Auth | **JWT** + **bcrypt** (autenticação própria) |
| CI | GitHub Actions (build + test; **sem deploy** por enquanto) |

Alternativas self-hosted vs AWS estão no README — preferir o que já está implementado salvo pedido explícito.

## Escopo atual vs futuro

### Implementado
- CRUD: Cursos, Aulas, Alunos
- Auth + RBAC: `admin`, `professor`, `aluno`
- Gestão de usuários (admin)
- Datas em formato BR: `dd/mm/yyyy` ou `dd/mm/yyyy HH:mm`
- Docker Compose local, pipeline CI

### Fora do escopo (não implementar sem pedido)
- Streaming ao vivo
- Chat WebSocket
- Deploy / hospedagem em produção
- OAuth externo (Cognito, etc.)

### Permissões (RBAC)

| Recurso | admin | professor | aluno |
|---------|-------|-----------|-------|
| Cursos — leitura | ✓ | ✓ | ✓ |
| Cursos — escrita | ✓ | — | — |
| Aulas — leitura | ✓ | ✓ | ✓ |
| Aulas — escrita | ✓ | ✓ | — |
| Alunos | CRUD | — | — |
| Usuários | CRUD | — | — |

Admin inicial (migration `001_seed_admin` via gormigrate): `ADMIN_EMAIL` / `ADMIN_PASSWORD` (default `admin@estudaja.com` / `admin123`).

## Backend (Go)

### Organização

```
backend/
├── cmd/api/main.go           # bootstrap, rotas, middleware
├── internal/
│   ├── config/
│   ├── database/
│   ├── models/               # entidades GORM
│   ├── repository/           # um arquivo por domínio (*_repository.go)
│   ├── handler/              # um arquivo por domínio (*_handler.go)
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
- `VITE_API_URL` aponta para a API (default `http://localhost:8080`)

## Comandos úteis

```bash
make infra-up      # Postgres + Redis
make backend       # API local
make frontend      # Vite dev
make test          # go test + npm build
docker compose up --build

cd docs && npm run slides:html   # exportar apresentação
```

## Git e commits

- **Não commitar** unless o usuário pedir explicitamente
- **Não commitar** `.env`, secrets, `node_modules/`, builds (`dist/`, `apresentacao.html`)
- Mensagens de commit em português, focadas no *porquê*

## Princípios de código

1. **Escopo mínimo** — só o pedido; sem refatorar o unrelated
2. **Seguir convenções existentes** — nomes, estrutura de pastas, estilo
3. **Sem over-engineering** — nada de abstrações prematuras
4. **Testes** — backend ao alterar handlers; não adicionar testes triviais
5. **Documentação** — atualizar README/AGENTS quando mudar stack, auth ou estrutura

## Apresentação

- Slides em `docs/apresentacao.md` (Marp)
- Diagramas: `docs/slides/diagrams/*.mmd` → SVG via `npm run slides:diagrams`
- Marp CLI v4 **não renderiza Mermaid** inline — usar SVGs gerados

## Ao iniciar um chat novo

1. Ler este arquivo e, se necessário, [README.md](./README.md) (seção Desenvolvimento)
2. Confirmar escopo antes de implementar streaming, deploy ou features fora da lista
3. Manter RBAC e formato de datas ao tocar em API ou UI
4. Responder em português
