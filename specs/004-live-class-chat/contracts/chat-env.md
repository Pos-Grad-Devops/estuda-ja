# Contract: variáveis de ambiente / infra do chat

Extensão mínima sobre [001 api-env](../../001-aws-mvp-terraform/contracts/api-env.md). Chat P1 **não** introduz backend alternativo (`CHAT_BACKEND`) nem Redis.

## Variáveis de aplicação

| Variável | Obrigatória | Default | Notas |
|----------|-------------|---------|--------|
| _(nenhuma nova)_ | — | — | Limite 500 chars e path WS são constantes de código/contrato |
| `JWT_SECRET` | já existente | — | Mesmo segredo valida `?token=` no handshake |
| `CORS_ORIGIN` | já existente | — | WS no browser usa origin do front; manter alinhado ao 001 |
| `PORT` | já existente | `8080` | Mesmo listener HTTP+WS |

Opcional (só se a implement quiser configurar sem hardcode):

| Variável | Default sugerido | Notas |
|----------|------------------|--------|
| `CHAT_MAX_CHARS` | `500` | MUST permanecer 500 no P1 da spec; env só para testes |
| `CHAT_PING_INTERVAL` | `120s` | Keepalive; documentar se usado |

**Não** adicionar: `REDIS_URL`, `CHAT_AWS_*`, vars de API Gateway WebSocket.

## Infra Terraform (sessão demo)

| Recurso | Mudança P1 |
|---------|------------|
| `aws_lb.api` `idle_timeout` | **Sim** — de 60 → **3600** (ou valor documentado ≥ ping; ≤ 4000) |
| Target group stickiness | **Não** (`desired_count = 1`) |
| ElastiCache / API GW WS / Lambda | **Não** |
| SG / listener / CloudFront API | Sem mudança estrutural esperada; validar upgrade WS no quickstart |
| ECS task env | Sem secret novo obrigatório |
| `infra/budget/` | Intocado |

## Compose

Sem serviço Redis. Sem porta extra: WS na **8080** da API.

```yaml
# ilustrativo — sem vars novas obrigatórias
services:
  api:
    environment:
      PORT: "8080"
      # JWT_SECRET, DATABASE_URL, CORS_ORIGIN — já existentes
```

## Segredos e logs

- JWT continua fora do git (SSM / `.env` local gitignored).
- **Não** logar URI com `?token=`.
- **Não** commitar URLs de demo com token na query.

## Comportamento de falha

- Sem `JWT_SECRET` → boot já falha (baseline); chat não tem modo “aberto”.
- Task reiniciada → conexões caem; clientes reconectam com painel vazio (sem histórico).
