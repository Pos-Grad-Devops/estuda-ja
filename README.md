# EstudaJá — Plataforma de Cursos ao Vivo

## Integrantes

- Gustavo Santos Arruda
- Pedro Lucas dos Santos Ribeiro
- Renan Roseno dos Santos
- Victor da Silva Neves

---

## Desenvolvimento

### Estrutura do repositório

```
estuda-ja/
├── backend/          # API Go (Fiber) + PostgreSQL
├── frontend/         # Vite + React + TypeScript
├── infra/            # Terraform AWS (demo efêmera) + scripts de publish
├── docs/             # Apresentação e diagramas
├── specs/            # Speckit (ex.: 001-aws-mvp-terraform)
├── docker-compose.yml
├── Makefile
├── AGENTS.md         # Guidelines para agentes de IA e devs
└── .github/workflows/ci.yml
```

### Pré-requisitos

- Go 1.25+
- Node.js 22+
- Docker e Docker Compose

### Subir infraestrutura local

```bash
make infra-up          # PostgreSQL + Redis
```

### Backend (API)

```bash
cd backend
go run ./cmd/api
```

Variáveis padrão:

| Variável | Default |
|----------|---------|
| `PORT` | `8080` |
| `DATABASE_URL` | `postgres://estudaja:estudaja@localhost:5432/estudaja?sslmode=disable` |
| `CORS_ORIGIN` | `http://localhost:5173` |
| `JWT_SECRET` | `dev-secret-change-me` |
| `JWT_EXPIRATION` | `24h` |
| `ADMIN_EMAIL` | `admin@estudaja.com` |
| `ADMIN_PASSWORD` | `admin123` |
| `DEMO_PROFESSOR_EMAIL` | `professor@estudaja.com` |
| `DEMO_PROFESSOR_PASSWORD` | `professor123` |
| `DEMO_ALUNO_EMAIL` | `aluno@estudaja.com` |
| `DEMO_ALUNO_PASSWORD` | `aluno123` |
| `LIVE_BACKEND` | `stub` (local/CI; na AWS = `ivs`) |

Lista completa em `backend/.env.example`. Frontend: `VITE_API_URL` (ver `frontend/.env.example`; rebuild ao mudar).

Seed automático na subida da API: admin (`001`) + professor/aluno + 1 curso + 1 aula (`002`) + VOD demo (`003_seed_vod_demo`). Defaults acima são só de **dev** local. Live **não** sobe `ao_vivo` no seed.

### Autenticação e permissões

Autenticação própria com JWT e senha hash (bcrypt).

| Perfil | Permissões |
|--------|------------|
| **admin** | CRUD completo de cursos, aulas, alunos e usuários; VOD e live (gestão) |
| **professor** | Visualiza cursos; cria/edita/exclui aulas; VOD e live (gestão) |
| **aluno** | Visualiza cursos, aulas, VOD e status/playback live (sem ingest) |

| Endpoint | Descrição |
|----------|-----------|
| `POST /api/v1/auth/login` | Login (público) |
| `GET /api/v1/auth/me` | Usuário autenticado |
| `GET/POST/PUT/DELETE /api/v1/users` | Gestão de usuários (admin) |

Demais endpoints exigem header `Authorization: Bearer <token>`.

### Endpoints (CRUD)

| Recurso | Base |
|---------|------|
| Cursos | `GET /api/v1/cursos` (todos) · escrita (admin) |
| Aulas | `GET /api/v1/aulas` (todos) · escrita (admin, professor) |
| Live | `GET .../aulas/:id/live` + `/playback` (autenticados) · `POST .../start\|stop` + `GET .../ingest` (admin, professor) |
| Alunos | CRUD (admin) |
| Usuários | CRUD (admin) |

### Frontend

```bash
cd frontend
cp .env.example .env
npm install
npm run dev
```

Acesse http://localhost:5173. Logins de demo (seed): admin, professor ou aluno (credenciais na tabela acima).

### Stack completa com Docker

```bash
docker compose up --build
```

- API: http://localhost:8080
- Frontend: http://localhost:5173
- Health: http://localhost:8080/health

### Testes e CI

```bash
make test
```

Pipeline em `.github/workflows/ci.yml`: testes do backend, build do frontend e build das imagens Docker. Publish AWS da demo no caminho mínimo = **manual** (seção abaixo). CD opcional (P3) = `.github/workflows/cd.yml` — ver subseção **CD opcional (P3)**.

### Fora do escopo atual (produto)

- Certificado / presença · pipeline live→VOD · tokenização IVS
- Moderação de chat (P2 da feature `004`)
- Estados live P2 (`agendada` ricos — Fase 5 de `003`) sem pedido explícito
- OAuth externo
- Redis/ElastiCache na AWS
- Domínio customizado
- Automação EventBridge de apply/destroy (P2)

Streaming ao vivo P1 (`003`) **está implementado** (stub local + IVS na demo AWS) — ver seção Demo AWS e [quickstart 003](./specs/003-live-streaming-ivs/quickstart.md).  
Chat da aula P1 (`004`) **está implementado** (WebSocket na API, hub em memória; local + AWS) — ver [quickstart 004](./specs/004-live-class-chat/quickstart.md).

---

## Demo AWS (MVP P1)

Caminho oficial de hospedagem acadêmica: **AWS gerenciado** via Terraform em `infra/` (sem NAT Gateway, sem Redis AWS). Self-hosted **não** é o destino da demo — Compose fica só para desenvolvimento local.

Runbook completo e critérios de sucesso: [specs/001-aws-mvp-terraform/quickstart.md](./specs/001-aws-mvp-terraform/quickstart.md).  
Streaming ao vivo (IVS + OBS): [specs/003-live-streaming-ivs/quickstart.md](./specs/003-live-streaming-ivs/quickstart.md).  
Chat da aula: [specs/004-live-class-chat/quickstart.md](./specs/004-live-class-chat/quickstart.md).  
Contratos: [api-env](./specs/001-aws-mvp-terraform/contracts/api-env.md) · [terraform-outputs](./specs/001-aws-mvp-terraform/contracts/terraform-outputs.md) · [frontend-publish](./specs/001-aws-mvp-terraform/contracts/frontend-publish.md) · [live-env](./specs/003-live-streaming-ivs/contracts/live-env.md) · [chat-env](./specs/004-live-class-chat/contracts/chat-env.md).

**Ordem fixa da sessão** (não pular passos):

1. **Billing** → 2. **Apply** → 3. **Push API** → 4. **CORS** → 5. **Health** → 6. **Rebuild front** → 7. **Warm-up** (incl. canal IVS + smoke live) → 8. **Demo** (VOD + ao vivo + **chat**) → 9. **Destroy**

### 0. Alerta de billing (uma vez por conta)

Antes da primeira demo (FR-022 / SC-012):

```bash
cd infra/budget
cp terraform.tfvars.example terraform.tfvars   # edite notification_email
terraform init && terraform apply
```

| Item | Valor |
|------|--------|
| Limiar mensal padrão | **US$ 5** (`budget_limit_usd`) |
| Alerta ACTUAL | **80%** do limiar (`threshold_percent`) |
| Alerta FORECASTED | 100% |
| Console | **Billing → Budgets** (confirmar nome `estudaja-…` e e-mail) |

Esta stack é **separada**. `terraform destroy` em `infra/` **não** remove o budget — ele deve permanecer na conta.

### 1. Apply (infra da demo)

```bash
cd infra
cp terraform.tfvars.example terraform.tfvars   # opcional; segredos omitidos são gerados no apply
terraform init
terraform apply
```

Região fixa: **us-east-1**. Outputs obrigatórios: `frontend_url`, `api_url`, `ecr_repository_url`, `ecs_cluster_name`, `ecs_service_name`, `s3_bucket_name`, `cloudfront_frontend_distribution_id`, **`vod_bucket_name`**, **`ivs_channel_arn`**, etc. (contratos terraform-outputs + [terraform-vod](./specs/002-vod-library/contracts/terraform-vod.md) + [terraform-ivs](./specs/003-live-streaming-ivs/contracts/terraform-ivs.md)).

`CORS_ORIGIN` no SSM é preenchido automaticamente com a origem HTTPS do CloudFront do frontend (salvo override em `cors_origin`).

**VOD (bucket efêmero):** o apply cria um bucket S3 privado distinto do frontend (`force_destroy = true`). A task ECS sobe com `VOD_BACKEND=s3`, `VOD_S3_BUCKET` e `VOD_PLAYBACK_TTL=15m`. Não há CloudFront de mídia nem NAT.

**Live (canal IVS efêmero):** o apply cria **um** canal IVS BASIC LOW-latency (`infra/ivs.tf`). A task sobe com `LIVE_BACKEND=ivs`, `IVS_INGEST_ENDPOINT`, `IVS_PLAYBACK_URL`, `IVS_CHANNEL_ARN` e secret SSM `IVS_STREAM_KEY`. Output só `ivs_channel_arn` — **sem** stream key nem playback URL em plaintext. Local/CI permanece `LIVE_BACKEND=stub` (Compose).

**Chat (WebSocket na API):** sem serviço extra. ALB `idle_timeout = 3600` (`infra/alb.tf`) para conexões longas; **sem** stickiness no TG; **sem** Redis/ElastiCache; **sem** API Gateway WebSocket / Lambda. Local = WS real na porta 8080; CI = `go test` sem AWS. Keepalive `chat.ping`/`chat.pong` evita idle ~10 min do CloudFront. Path: `GET /api/v1/aulas/:id/chat/ws?token=...` (não colar URLs com token).

### 2–5. Publish API → CORS → health → front

Scripts (PowerShell, a partir de `infra/`):

```powershell
.\publish-api.ps1          # build linux/arm64 → ECR :latest → force deploy ECS
# Aguardar: GET {api_url}/health → 200 (HTTPS CloudFront da API)
.\publish-frontend.ps1     # VITE_API_URL={api_url} → npm build → S3 sync → invalidate CF
```

Equivalente manual: ver quickstart §§2–3 e contrato frontend-publish. Se alterar `CORS_ORIGIN` depois do apply, atualize o parâmetro SSM e force novo deploy da task.

Na subida da API, a migration `003_seed_vod_demo` copia `backend/assets/vod/demo-aula.mp4` (incluído na imagem) para o storage e cria **um** `AulaVod` `publicado` na aula de demo — o aluno assiste **sem** upload manual (SC-005).

**Este caminho manual P1 permanece o runbook oficial** e funciona com ou sem CD habilitado.

### CD opcional (P3)

Workflow: [`.github/workflows/cd.yml`](./.github/workflows/cd.yml). Espelha `publish-api.ps1` / `publish-frontend.ps1` no Actions (ARM64→ECR+ECS e front com `VITE_API_URL`→S3+invalidate CF).

| Item | Detalhe |
|------|---------|
| Gate | Jobs `backend` + `frontend` (test/build) **obrigatórios**; se falharem, `deploy-api` / `deploy-frontend` **não rodam** (CI vermelho não promove) |
| Habilitar | Variável de repositório `AWS_CD_ENABLED` = `true` (sem isso, só o gate roda; nada é publicado) |
| Triggers | `workflow_dispatch` ou `push` em `main` |
| Credenciais | Secrets `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY` |
| Sessão (efêmera) | Secrets preenchidos com **outputs do apply atual** (ver tabela abaixo) |

Secrets / vars da sessão (atualizar após **cada** `terraform apply`; invalidam após `destroy`):

| Secret / var | Origem (output TF) |
|--------------|-------------------|
| `ECR_REPOSITORY_URL` | `ecr_repository_url` |
| `ECS_CLUSTER_NAME` | `ecs_cluster_name` |
| `ECS_SERVICE_NAME` | `ecs_service_name` |
| `VITE_API_URL` | `api_url` (HTTPS, sem barra final) |
| `S3_BUCKET_NAME` | `s3_bucket_name` |
| `CLOUDFRONT_FRONTEND_DISTRIBUTION_ID` | `cloudfront_frontend_distribution_id` |
| `AWS_REGION` (var, opcional) | default `us-east-1` |

Após `terraform destroy`: desative `AWS_CD_ENABLED` (ou limpe os secrets). Não reutilize `VITE_API_URL` / bucket / distribution de sessão anterior — mixed content, CORS e 404 são sintomas típicos (contrato frontend-publish).

CD **não** substitui apply/destroy nem o warm-up manual. Sem stack ligada, habilitar CD só gera falha nos jobs de deploy.

### 6–7. Warm-up e demo (manual — Fase F / escolha b)

**Manual vs automatizado (003 / US5):**

| Parte | Status | O quê |
|-------|--------|--------|
| Apply / destroy da stack **001** | **Manual** (gap EventBridge P2) | Não há jobs EventBridge neste repo para ligar/desligar a demo |
| Checklist T−15 / OBS / start live | **Manual** (runbook) | Operador marca T−15→T−0; professor inicia live e OBS |
| Health API + sinal IVS (opcional) | **Semi-auto** (script local) | `infra/check-live-warmup.ps1` — só `GET /health` e, se passar `-ChannelArn`, `aws ivs get-stream`. **Não** aplica nem destrói a stack |

Warm-up **não** é scale-from-zero: stack ligada, `desired_count = 1`, task running, `/health` estável. Antecedência: **T−15 min** antes da janela (RDS cold + primeiro migrate + **canal IVS + OBS**).

Checklist T−15 min (resumo — marcar na ordem; detalhe streaming em [003 quickstart §B3](./specs/003-live-streaming-ivs/quickstart.md)):

| Momento | Verificar |
|---------|-----------|
| **T−15** | Apply ok; RDS available; ECS `desired_count = 1` + task RUNNING; targets healthy; **`terraform output ivs_channel_arn` presente** |
| **T−10** | `GET {api_url}/health` = **200** estável (2–3 chamadas) **ou** `.\check-live-warmup.ps1 -ApiUrl {api_url}`; live da aula demo ainda inativa/agendada (ou estado conhecido) |
| **T−5** | Login admin + ≥1 curso; UI em `{frontend_url}`; amostra RBAC; **professor inicia live + OBS Live; smoke player gestor/aluno** (opcional: script com `-ChannelArn` após OBS Live) |
| **T−0** | Não zerar `desired_count`; não apply/destroy; encoder continua; destroy só após o encerramento |

Pular o warm-up = risco de cold start na abertura (infra **ou** sinal ao vivo) — inadequado para demo de pico. **Não** subir stack nem OBS no minuto da aula.

Credenciais seed (defaults de **dev**, iguais ao local): admin / professor / aluno — ver tabela em Desenvolvimento.

**Demo VOD (passos extras ≤ ~10 min):**

1. Login `aluno@estudaja.com` → **Aulas** → aula de demo → reproduzir gravação seed (sem upload prévio).
2. (Opcional) Login professor → substituir MP4 ≤ 50 MB na ficha → aluno vê o novo conteúdo.
3. Confirmar: **sem** página Biblioteca; **sem** download; visitante sem JWT não reproduz.
4. Validação ponta a ponta: [specs/002-vod-library/quickstart.md](./specs/002-vod-library/quickstart.md).

O seed **recria** a cada apply limpo: 1 curso, 1 aula, 1 VOD publicado (asset `demo-aula.mp4`). Uploads feitos na sessão anterior **não** voltam após destroy.

**Demo streaming ao vivo (passos extras ≤ ~15 min incremental — SC-005):**

1. Após publish: confirmar `LIVE_BACKEND=ivs` na task e `ivs_channel_arn` no output (sem key/URL em outputs).
2. (Opcional P2) Professor → **Agendar transmissão** (exige horário na aula) — aluno vê **Agendada** sem player.
3. No warm-up T−5: professor (ou admin) → **Aulas** → ficha → **Iniciar transmissão** → copiar servidor RTMPS + stream key → OBS (serviço Personalizado; H.264+AAC; keyframe 2 s; ≤ 3,5 Mbps).
4. Aluno em outra sessão → mesma ficha → distingue **agendada / ao vivo / encerrada** e **Gravação**; reproduz in-app só em **ao vivo** (IVS Player). Sem ingest/botões de gestão.
5. Encerrar live; se houver VOD 002, a ficha **pode** apontar a gravação (sem pipeline live→VOD). Reiniciar na mesma sessão **sem** novo apply. No máximo **uma** live `ao_vivo`.
6. Validação ponta a ponta: [specs/003-live-streaming-ivs/quickstart.md](./specs/003-live-streaming-ivs/quickstart.md).

Local/CI: `LIVE_BACKEND=stub` — estados/RBAC/erros PT **sem** vídeo real (player não finge sinal).

**Demo chat da aula (passos extras ≤ ~15 min incremental — SC-005):**

1. Após publish + health: dois browsers (aluno + professor/admin) em `{frontend_url}` → mesma aula → painel **Chat** (distinto de live/VOD).
2. Enviar texto ≤ 500 chars → fan-out com **nome** do autor; outra aula isolada; F5 = painel vazio (sem histórico).
3. WS via `wss` no host de `{api_url}` (CloudFront). Não usar `alb_dns_name` no browser. **Não** colar URLs com `?token=`.
4. Validação: [specs/004-live-class-chat/quickstart.md](./specs/004-live-class-chat/quickstart.md) §C. Local/CI: §A / §B (WS real no Compose; testes Go sem AWS).

### 8. Destroy (entre sessões)

```bash
cd infra
terraform destroy
```

**Não** rode destroy em `infra/budget/`. Após o destroy da demo (procedimento ≤ 15 min — SC-007):

| Verificar na console | Esperado |
|----------------------|----------|
| ECS, ALB, RDS, CloudFront ×2, S3 site, **S3 VOD** (`vod_bucket_name`), VPC/SG do projeto | Ausentes |
| **Canal IVS + stream key** (`ivs_channel_arn` / SSM) | Ausentes (custo IVS contínuo ≈ 0) |
| NAT Gateway | **Nenhum** (nunca provisionado → custo NAT = US$ 0) |
| ECR | Removido (`force_delete = true`) |
| Snapshots RDS | Nenhum (`skip_final_snapshot = true`) |
| Objetos VOD da sessão | Removidos com o bucket (`force_destroy`) |
| Hub/chat em memória | Some com a task/API (sem recurso TF extra) |
| Log groups órfãos | Apagar se restarem fora do TF |
| **Billing → Budgets** | Budget **ainda ativo** |

**Fora do escopo VOD P1:** página Biblioteca dedicada, download do arquivo, rascunho/metadados (P2). Streaming ao vivo é feature **003** (não reabre 002). Chat P1 é feature **004** (hub na API; moderação P2 fora). Certificado, pipeline live→VOD e tokenização IVS continuam fora.

Resíduos e detalhes: [research.md §10](./specs/001-aws-mvp-terraform/research.md).

### Custo por sessão (ordem de grandeza)

| Cenário | Estimativa |
|---------|------------|
| Sessão ~4 h ligada (sem Free Tier) | **~US$ 1–3** |
| Mesma janela com Free Tier RDS/ALB elegível | **~US$ 0–1** |
| Após destroy completo da stack de demo | **~US$ 0** (ALB/RDS/Fargate/CF) |
| NAT Gateway | **US$ 0** (não existe neste MVP) |

Dominantes com stack ligada: ALB (~US$ 0,02–0,03/h), RDS `db.t4g.micro`, Fargate 256/512 ARM. Números são guia — conferir Pricing Calculator na data da demo ([research §9](./specs/001-aws-mvp-terraform/research.md)).

### Gaps explícitos (não confundir com P1)

| Prioridade | Item | Status no caminho mínimo |
|------------|------|---------------------------|
| **P1** | IaC + publish manual + destroy + docs | Entregue |
| **P2** | Automação warm-up / EventBridge apply–destroy **001** | **Gap** — checklist T−15 + script health/GetStream streaming; sem EventBridge |
| **P3** | CD no GitHub Actions (ECR + S3/CF) | **Opcional entregue** (`.github/workflows/cd.yml`); publish manual P1 continua válido |

### Segredos e state

Não versionar: `*.tfvars` (exceto `*.tfvars.example`), `*.tfstate*`, `.terraform/`, `.env`. Secrets da sessão ficam no state local e no SSM — nunca no git (SC-008).

---

## Contexto do negócio

A **EstudaJá** é uma EdTech de aulas ao vivo em massa — estilo "aula magna" — com milhares de alunos simultâneos e conteúdo gravado sob demanda.

**O problema:** a aula começa em horário fixo e o pico de acesso é **instantâneo**. Não dá para escalar aos poucos: ou a plataforma aguenta no minuto zero, ou milhares de alunos ficam de fora.

> *19h. Maria abre o app para a aula de Direito Constitucional. Em segundos, mais 3.000 alunos fazem o mesmo. O player precisa carregar, o chat precisa responder, e nada pode cair — a aula não espera.*

**Escopo do MVP:** transmissão ao vivo, chat da aula, biblioteca básica de vídeos e emissão simples de certificado. Prazo curto — entregamos o essencial primeiro e evoluímos depois.

---

## SLA desejado

| Período | Disponibilidade | Observação |
|---------|-----------------|------------|
| **Durante transmissão ao vivo** | **99,9%** | Janela crítica: início da aula até o encerramento |
| **Fora do horário de transmissão** | **99%** | Biblioteca de vídeos, chat assíncrono, emissão de certificados |

**Implicações operacionais:**

- A infraestrutura de streaming e chat deve estar **pré-aquecida** antes do horário da aula — não há tempo para escalar gradualmente.
- Monitoramento e alertas são obrigatórios na janela de transmissão.
- Fora do pico, tolera-se maior latência de provisionamento e janelas de manutenção planejadas.

---

## Estratégia: MVP e evolução

### MVP (Fase 1 — entrega inicial)

Objetivo: validar o fluxo principal com o menor escopo possível, priorizando estabilidade na janela ao vivo.

| Funcionalidade | Escopo MVP | Fora do MVP |
|----------------|------------|-------------|
| Transmissão ao vivo | Uma aula por vez via provedor de streaming gerenciado | Multi-stream, DVR ao vivo, qualidade adaptativa avançada |
| Chat da aula | Chat em tempo real básico (WebSocket), sem moderação automática | Moderação por IA, filtros, silenciamento em massa |
| Biblioteca de vídeos | Upload manual e playback simples (VOD) | Transcodificação automática, legendas, busca full-text |
| Certificados | Emissão manual ou PDF simples ao concluir curso | Integração com blockchain, verificação pública |
| Autenticação | Login básico (e-mail/senha ou OAuth social) | SSO corporativo, MFA obrigatório |
| Infraestrutura | Containers com orquestração e autoscaling; CDN + streaming gerenciado | Microsserviços completos, multi-região |
| Observabilidade | Logs centralizados + alertas na janela ao vivo | APM, tracing distribuído, dashboards de negócio |
| CI/CD | Pipeline mínimo (build + testes + deploy) | GitOps, ambientes efêmeros, feature flags |

### Plano de evolução pós-MVP

```mermaid
timeline
    title Roadmap EstudaJá
    section Fase 1 — MVP
        Semanas 1-4 : Transmissão ao vivo + player
                      : Chat básico
                      : Auth e cadastro de alunos
    section Fase 2 — Consolidação
        Semanas 5-8 : Biblioteca VOD completa
                      : Transcodificação automática
                      : Pré-aquecimento de infra na janela ao vivo
    section Fase 3 — Certificação e escala
        Semanas 9-12 : Emissão automática de certificados
                       : Chat com moderação
                       : Testes de carga e chaos engineering
    section Fase 4 — Maturidade
        Semanas 13+ : Microsserviços por domínio
                      : Multi-região e failover
                      : Analytics e recomendação de conteúdo
```

| Fase | Foco | Entregáveis principais |
|------|------|------------------------|
| **2 — Consolidação** | VOD e resiliência | Upload/transcodificação, CDN otimizada, runbooks de incidente |
| **3 — Certificação e escala** | Completude do produto | Certificados automáticos, moderação de chat, testes de carga documentados |
| **4 — Maturidade** | Escala e operação avançada | Decomposição em microsserviços, observabilidade avançada, custo otimizado |

---

## Plano de desenvolvimento

### Visão geral das sprints (MVP — 4 semanas)

| Sprint | Objetivo | Entregas |
|--------|----------|----------|
| **S1** | Fundação | Repositório, CI básico, auth, modelagem de domínio (Curso, Aula, Aluno) |
| **S2** | Ao vivo | Integração com provedor de streaming, player embed, página da aula |
| **S3** | Interação | Chat WebSocket, presença online, testes de carga na janela simulada |
| **S4** | Estabilização | Monitoramento, alertas SLA, documentação, demo para apresentação |

### Riscos e mitigações

| Risco | Impacto | Mitigação |
|-------|---------|-----------|
| Pico instantâneo na abertura da aula | Indisponibilidade | Pré-aquecimento de containers; CDN; load test antes do go-live |
| Latência do chat com milhares de usuários | Experiência ruim | Redis Pub/Sub + WebSocket gateway ou serviço SaaS de tempo real |
| Cold start de containers | Atraso no início da aula | Autoscaling com réplicas mínimas elevadas na janela; scaling programado |
| Custo de streaming | Orçamento | Serviço pay-per-use; gravar e servir VOD via CDN após a live |

### Critérios de aceite do MVP

- [ ] Aluno consegue entrar na aula ao vivo no horário agendado
- [ ] Player exibe stream com latência aceitável (< 10 s)
- [ ] Chat funciona com pelo menos 500 usuários simultâneos (simulado)
- [ ] Disponibilidade ≥ 99,9% medida em janela de teste de 1 h
- [ ] Pipeline CI executa build e testes automatizados a cada push

---

## Stack proposta

Cada camada oferece uma opção **gerenciada (AWS)** e uma **self-hosted**, para flexibilidade entre velocidade de entrega e controle de custo.

| Camada | AWS (gerenciado) | Self-hosted | Justificativa |
|--------|------------------|-------------|---------------|
| **Frontend** | S3 + CloudFront ou Amplify | Vite (React + TypeScript) + nginx | Build estático, HMR, deploy via CDN |
| **Backend API** | ECS Fargate | Go (Fiber ou chi) em Docker Compose | API REST + WebSocket; alta concorrência com baixo consumo |
| **Streaming ao vivo** | AWS IVS | NGINX-RTMP ou Ant Media Server | Escala automática vs. controle total do stream |
| **VOD / armazenamento** | S3 + CloudFront | MinIO + CDN | Vídeos gravados e assets estáticos |
| **Chat em tempo real** | API Gateway WebSocket + Lambda | WebSocket (Go) + Redis Pub/Sub | Baixa latência; backend Go gerencia conexões em escala |
| **Banco de dados** | RDS PostgreSQL | PostgreSQL em container/VM | Relacional, ACID para matrículas e progresso |
| **Cache** | ElastiCache Redis | Redis em container/VM | Sessões, pub/sub do chat, rate limiting |
| **Auth** | Cognito | Keycloak ou autenticação própria | OAuth social sem implementar do zero |
| **Orquestração** | **ECS Fargate** + ALB | Docker Compose / Docker Swarm | ECS reduz complexidade e oferece autoscaling nativo |
| **Registry de imagens** | ECR | Docker Hub / registry privado | Imagens versionadas pelo CI |
| **Autoscaling** | ECS Service Auto Scaling + EventBridge | Scripts cron + docker scale | Pré-aquecimento antes da aula |
| **IaC** | Terraform / CloudFormation | Terraform + Ansible | Infra reproduzível e versionada |
| **CI/CD** | GitHub Actions | GitHub Actions / GitLab CI | Build, testes e deploy automatizados |
| **Monitoramento** | CloudWatch + Grafana | Prometheus + Grafana | Métricas SLA na janela ao vivo |
| **Logs** | CloudWatch Logs | Loki ou ELK | Centralização e correlação de incidentes |

### Orquestração e autoscaling (recomendado: ECS)

Para o MVP, recomendamos **ECS com Fargate** — menor complexidade que Kubernetes e autoscaling nativo, ideal para o pico instantâneo.

**Estratégia de autoscaling:**

1. **Scaling programado (T-15 min):** elevar réplicas dos serviços críticos (API e chat) antes do horário da aula.
2. **Target tracking:** escala por CPU, memória ou métricas customizadas (ex.: conexões WebSocket ativas).
3. **Scaling pós-aula:** reduzir réplicas após o encerramento para controlar custo.

### Evolução da stack (Fases 2–4)

- **Microsserviços:** separar auth, chat, catalog e certificate em services independentes
- **Event-driven:** filas (SQS, RabbitMQ ou Kafka) para eventos de presença
- **Serverless:** funções para transcodificação e geração de certificados
- **DevSecOps:** SAST/DAST no pipeline, gestão de secrets, scan de imagens

---

## Arquitetura (MVP)

```mermaid
flowchart TB
    subgraph Clientes
        ALuno[Aluno / Browser]
    end

    subgraph Edge
        CDN[CDN]
        WAF[WAF]
        LB[Load Balancer]
    end

    subgraph Aplicacao["Aplicação (containers)"]
        WEB[Frontend]
        API[API REST]
        WS[WebSocket Gateway]
    end

    subgraph Streaming
        STREAM[Provedor de Streaming]
    end

    subgraph Dados
        PG[(PostgreSQL)]
        REDIS[(Redis)]
        STORAGE[(Object Storage)]
    end

    subgraph Observabilidade
        OBS[Monitoramento / Logs]
    end

    ALuno --> WAF --> CDN --> LB
    LB --> WEB & API & WS
    WEB --> STREAM
    API --> PG
    API --> REDIS
    API --> STORAGE
    WS --> REDIS
    Aplicacao --> OBS
```

### Fluxo — início da aula ao vivo (pico instantâneo)

```mermaid
sequenceDiagram
    participant A as Aluno
    participant CDN as CDN
    participant APP as Frontend
    participant API as Backend API
    participant STREAM as Streaming
    participant WS as Chat WS
    participant R as Redis

    Note over API,STREAM: Pré-aquecimento (T-15 min)<br/>Réplicas elevadas<br/>Stream já ativo

    A->>CDN: Acessa página da aula
    CDN->>APP: Serve assets estáticos
    APP->>API: Valida matrícula / token
    API->>A: Autorizado
    APP->>STREAM: Conecta player
    APP->>WS: Conecta WebSocket
    WS->>R: Subscribe canal da aula
    A->>WS: Mensagem no chat
    WS->>R: Publish
    R->>WS: Broadcast
    WS->>A: Mensagens em tempo real
```

---

## Disciplinas do curso

Disciplinas marcadas são as **mais críticas** para cumprir o SLA e operar a EstudaJá — especialmente o pico instantâneo na abertura das aulas ao vivo.

- [ ] Ecossistemas de Startups
- [ ] Direito Digital e LGPD
- [ ] Fundamentos de Engenharia de Software
- [ ] Metodologias Ágeis em Gestão de Projetos
- [x] Desenvolvimento de Software Integrado – DevOps
- [ ] Design da Experiência do Usuário
- [ ] Controle de Versão e Gerenciamento de Configuração
- [ ] Gerenciamento de Produtos
- [x] Integração e Entrega Contínua
- [x] Orquestração de Contêineres e Gerenciamento de Cluster
- [x] Infraestrutura Automatizada
- [ ] Desenvolvimento de Software Seguro – DevSecOps
- [ ] Testes Automatizados e Contínuos
- [x] Arquitetura de Microsserviços e Escalabilidade
- [ ] Documentação Técnica
- [ ] Computação em Nuvem
- [ ] Computação sem Servidores
- [x] Monitoramento e Análise de Logs
- [ ] Tópicos Avançados em Engenharia de Software

---

## Apresentação (referência)

- Arquivo de slides: [docs/apresentacao.md](./docs/apresentacao.md)
- Tempo máximo: **20 minutos**
- Conteúdo sugerido: problema (pico instantâneo) → MVP → arquitetura → SLA → disciplinas críticas → demo (quando houver código)

Para exportar os slides:

```bash
cd docs && npm run slides:html
```

---

## Documentação

| Arquivo | Descrição |
|---------|-----------|
| [docs/apresentacao.md](./docs/apresentacao.md) | Slides da apresentação (Marp) |
| [docs/slides/diagrams/](./docs/slides/diagrams/) | Diagramas Mermaid (fonte `.mmd` e SVG) |
| [docs/package.json](./docs/package.json) | Scripts para gerar diagramas e exportar slides |
| [AGENTS.md](./AGENTS.md) | Guidelines e contexto para desenvolvimento (IA e equipe) |
| [specs/001-aws-mvp-terraform/quickstart.md](./specs/001-aws-mvp-terraform/quickstart.md) | Runbook operacional da sessão AWS |
| [specs/001-aws-mvp-terraform/](./specs/001-aws-mvp-terraform/) | Spec / plan / tasks / contracts da feature AWS MVP |
| [specs/003-live-streaming-ivs/quickstart.md](./specs/003-live-streaming-ivs/quickstart.md) | Runbook streaming ao vivo (IVS + OBS) no ciclo 001 |
| [specs/004-live-class-chat/quickstart.md](./specs/004-live-class-chat/quickstart.md) | Chat da aula (WS local + AWS; idle ALB / keepalive) |

---

## Próximos passos

- [x] Demo AWS efêmera (Terraform + publish manual) — P1
- [x] Warm-up P2: **gap documentado** (escolha b) + checklist T−15 min — Fase F
- [x] CD no GitHub Actions — P3 opcional (Fase G; gate por test/build; secrets de sessão)
- [x] Streaming ao vivo IVS P1 (`003-live-streaming-ivs`) — stub local + canal efêmero AWS + OBS + runbook
- [x] Chat da aula P1 (`004-live-class-chat`) — WS na API + UI na ficha + idle ALB
- [ ] Certificado / presença · live→VOD · moderação de chat P2 · estados live P2 (fora do P1 atual)

---

## Referências

- [DESCRICAO.md](./DESCRICAO.md) — enunciado do projeto
- [REQUISITOS.md](./REQUISITOS.md) — entregáveis e disciplinas do curso
- [docs/apresentacao.md](./docs/apresentacao.md) — slides da apresentação
