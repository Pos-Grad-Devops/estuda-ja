# Quickstart: Chat da aula (004)

Validação ponta a ponta **sem implementar aqui** — guia para após as fases 1–3. Contrato: [chat-ws.md](./contracts/chat-ws.md). Env/infra: [chat-env.md](./contracts/chat-env.md). Modelo: [data-model.md](./data-model.md).

## Pré-requisitos

- Baseline 001–003 operacional (API + front + seed demo).
- Dois usuários autenticáveis (ex.: `aluno@estudaja.com` / `professor@estudaja.com` — defaults de seed).
- Uma aula existente (seed demo ou criada na UI).
- Protocolo WS implementado conforme contracts.

## A — Compose (dois browsers)

### Setup

```bash
make infra-up    # Postgres (+ Redis Compose se existir — chat P1 NÃO usa Redis)
make backend     # API com hub WS
make frontend    # Vite; VITE_API_URL=http://localhost:8080
```

Ou `docker compose up --build` equivalente.

Local = **WebSocket real** na mesma porta HTTP da API (`8080`). Sem vars `REDIS_URL` / `CHAT_AWS_*` / API Gateway.

### Passos

1. **Browser 1** — login como aluno → abrir ficha da aula demo → painel **Chat** visível (com ou sem live `ao_vivo`).
2. **Browser 2** (janela anônima) — login como professor → mesma aula → painel Chat.
3. Em um browser, enviar texto válido (≤ 500 chars).
4. No outro, a mensagem aparece em ≤ 5 s com **nome** do autor (sem reload manual obrigatório).
5. Abrir **outra** aula no browser 1 e confirmar que a mensagem da aula demo **não** aparece.
6. Fechar a ficha / F5 → painel inicia **vazio** (sem histórico).
7. Tentar mensagem vazia e texto > 500 → erro em português; sem broadcast.

### Auth negativa (opcional)

- Abrir WS sem `token` ou com token inválido (DevTools / `websocat`) → conexão rejeitada.
- **Não** colar a URL completa com `?token=` em issues, logs ou slides.

### Esperado

| Check | Resultado |
|-------|-----------|
| Fan-out mesma aula | OK |
| Isolamento entre aulas | OK |
| Reabrir = sem histórico | OK |
| Validação 500 / vazio | Erro PT |
| Aluno sem controles live/VOD extras | Intactos (002/003) |

## B — CI (sem AWS)

```bash
cd backend && go test ./...
cd frontend && npm run build
```

Esperado: testes de handler/hub/RBAC/erros PT verdes. **Não** exige dois browsers no pipeline. **Não** exige conta AWS.

## C — AWS (atrás do ALB / CloudFront)

Mesmo protocolo da §A; URL no browser usa o **`api_url` HTTPS** da sessão (output Terraform / `VITE_API_URL`), convertido para `wss`:

```text
wss://<host do api_url>/api/v1/aulas/<id>/chat/ws?token=<JWT>
```

- `<host do api_url>` = host do CloudFront da API (não o `alb_dns_name` — mixed content / fora do baseline).
- O front já monta essa URL a partir de `VITE_API_URL` + JWT da sessão.
- **Não** colar URLs com `?token=` em issues, PRs, screenshots ou runbooks compartilhados.

### Notas de infra (P1)

| Item | Valor |
|------|--------|
| ALB `idle_timeout` | **3600** s (`infra/alb.tf`) — conexões WS longas; sem stickiness no target group (`desired_count = 1`) |
| CloudFront | Encaminha upgrade WS no distribution da API; idle tipicamente ~**10 min** sem tráfego |
| Keepalive | Cliente envia `chat.ping` (~2–4 min); servidor responde `chat.pong` — evita corte por idle do CloudFront |
| Hub | Em memória na task ECS; **sem** Redis/ElastiCache; **sem** API Gateway WebSocket / Lambda |
| Budget | `infra/budget/` **intocado** (não destruir com a demo) |

### Ciclo da sessão (incremental ≤ ~15 min — SC-005)

Ordem fixa (não redesenhar 001–003):

1. `terraform apply` (stack demo; idle ALB já 3600)
2. Publish API + frontend (`publish-api.ps1` / `publish-frontend.ps1`)
3. Warm-up: `GET {api_url}/health` estável (+ checklist IVS/OBS se for demo live)
4. **Demo chat**: dois perfis (aluno + professor/admin) na mesma aula → enviar/receber; outra aula isolada; F5 = painel vazio
5. `terraform destroy` — remove API/task/ALB/CloudFront; **chat some com a API** (sem recurso extra)

### Smoke mínimo AWS

1. Após publish, health OK.
2. Dois logins na URL do CloudFront do front.
3. Mesma aula → enviar/receber como no Compose (fan-out com **nome**).
4. Destroy → endpoints da sessão anteriores inválidos; budget permanece.

## Fora deste quickstart

- Moderação (P2 / US5)
- Load test de milhares de conexões
- Histórico ao reabrir
- EventBridge apply/destroy do 001
- ElastiCache / API Gateway WebSocket + Lambda
