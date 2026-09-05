# Contract: protocolo WebSocket do chat da aula

**Path:** `GET /api/v1/aulas/:id/chat/ws?token=<JWT>`  
**Upgrade:** WebSocket (texto JSON UTF-8).  
**Uma aula por conexão** (`:id`). Sem frame `join`. Sem multiplex.

Ver [data-model.md](../data-model.md) e [chat-env.md](./chat-env.md).

## Handshake

### Query

| Param | Obrigatório | Descrição |
|-------|-------------|-----------|
| `token` | sim | JWT do login (`Authorization: Bearer` equivalente). Sem cookie. |

### Pré-condições (servidor)

1. `token` presente e assinatura/exp válidas → senão **rejeitar upgrade** (HTTP 401 se ainda HTTP, ou close imediato).
2. `:id` = aula existente → senão rejeitar (404).
3. Papel com leitura de aulas (`admin` \| `professor` \| `aluno`) → senão 403.
4. Resolver `User.Nome` para a conexão.

### Erros de handshake (português)

| Situação | Mensagem sugerida |
|----------|-------------------|
| Sem token / inválido | `não autenticado` / `token inválido` |
| Aula inexistente | `aula não encontrada` |
| Sem permissão | `sem permissão` |

**Logs:** não registrar a query string completa (omitir `token`).

## Frames cliente → servidor

### `chat.send`

```json
{
  "type": "chat.send",
  "texto": "string"
}
```

| Campo | Regra |
|-------|--------|
| `texto` | Obrigatório; após trim: 1..500 caracteres (contar runes Unicode); vazio/só espaços → erro; >500 → erro |

Outros `type` desconhecidos → `chat.error` (`tipo de mensagem inválido`) ou ignore documentado; P1 MUST rejeitar com erro claro.

**Não** enviar `aula_id`, `autor` ou `id` no send — servidor deriva da conexão.

## Frames servidor → cliente

### `chat.message` (broadcast na sala)

```json
{
  "type": "chat.message",
  "id": "uuid",
  "aula_id": 1,
  "autor": {
    "id": 2,
    "nome": "Maria Silva"
  },
  "texto": "Bom dia",
  "enviado_em": "05/09/2026 15:04"
}
```

- Enviado a **todos** os clientes conectados na mesma `aula_id` (incluindo o remetente).
- `enviado_em`: formato BR `dd/mm/yyyy HH:mm`.
- UI MUST usar `autor.nome` como rótulo.

### `chat.error`

```json
{
  "type": "chat.error",
  "error": "mensagem em português"
}
```

| Situação | `error` (exemplos) |
|----------|---------------------|
| Texto vazio / só espaços | `mensagem vazia` |
| Texto > 500 | `mensagem excede 500 caracteres` |
| Tipo inválido | `tipo de mensagem inválido` |
| Falha interna | `falha ao enviar mensagem` |

Erro de validação **não** gera `chat.message`.

### `chat.ping` / `chat.pong` (keepalive)

Cliente **ou** servidor MAY enviar:

```json
{ "type": "chat.ping" }
```

Resposta:

```json
{ "type": "chat.pong" }
```

Recomendado a cada ~2–4 minutos para idle CloudFront/ALB (ver [research.md](../research.md) §4–5). Não entram na lista de chat da UI.

## Isolamento

- Mensagem da aula A **nunca** é enviada a conexões da aula B.
- Nova conexão **não** recebe backlog.

## Fora do contrato P1

- `chat.delete`, `chat.mute`, `chat.history`, `chat.presence`, anexos, binário.
- REST paralelo `POST /aulas/:id/chat/messages` (opcional futuro; P1 = só WS).

## RBAC (resumo)

| Ação | admin | professor | aluno |
|------|-------|-----------|-------|
| Conectar / enviar / receber | ✓ | ✓ | ✓ |
| Gestão live / VOD | inalterado (003/002) | inalterado | só leitura |

Participar do chat **não** concede ingest/upload.
