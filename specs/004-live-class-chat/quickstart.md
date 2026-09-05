# Quickstart: Chat da aula (004)

Validação ponta a ponta **sem implementar aqui** — guia para após as fases 1–3. Contrato: [chat-ws.md](./contracts/chat-ws.md). Modelo: [data-model.md](./data-model.md).

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

Esperado: testes de handler/hub/RBAC/erros PT verdes. **Não** exige dois browsers no pipeline.

## C — AWS (atrás do ALB / CloudFront)

Mesmo protocolo; URL:

```text
wss://<api_url CloudFront>/api/v1/aulas/<id>/chat/ws?token=<JWT>
```

`<api_url>` = output/`VITE_API_URL` da sessão (HTTPS). **Não** usar `alb_dns_name` no browser (mixed content / fora do baseline).

### Notas de infra

- ALB: `idle_timeout` elevado (plan/research; ex. 3600 s).
- CloudFront: WS suportado no distribution da API; keepalive `chat.ping` evita idle ~10 min.
- `desired_count = 1` → hub em memória coerente; sem Redis.
- Sem API Gateway WebSocket / Lambda.
- Ciclo: `terraform apply` → publish API/front → warm-up (`/health` + checklist 003 se live) → demo chat dois perfis → `terraform destroy`.
- Budget `infra/budget/` permanece.
- **Não** colar logs/URLs com `?token=` em issues.

### Smoke mínimo AWS

1. Após publish, health OK.
2. Dois logins na URL do CloudFront do front.
3. Mesma aula → enviar/receber como no Compose.
4. Destroy → endpoints da sessão anteriores inválidos.

## Fora deste quickstart

- Moderação (P2)
- Load test de milhares de conexões
- Histórico ao reabrir
- EventBridge apply/destroy do 001
