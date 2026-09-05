# Research: 004-live-class-chat

Phase 0 — decisões técnicas para materializar a spec (decisões de produto **não** reabertas). Baseline 001–003 **não** reaberta.

## 1. Biblioteca WebSocket no Fiber

**Decision:** Usar **`github.com/gofiber/contrib/websocket`** (compatível com Fiber v2 / fasthttp), no padrão do recipe oficial Fiber “websocket-chat”: middleware que verifica `websocket.IsWebSocketUpgrade`, depois `websocket.New(handler)`.

**Rationale:** Stack fechada = Fiber; contrib é o caminho documentado pela própria Fiber; evita misturar `net/http` + gorilla num segundo listener. Fasthttp já é dependência transitiva do projeto.

**Alternatives considered:**
- `gorilla/websocket` atrás de `adaptor` Fiber — funciona, mas duplica stack HTTP e atrito com Locals/middleware.
- `nhooyr.io/websocket` / `coder/websocket` — ótimos em `net/http`; sem ganho no monólito Fiber atual.
- API Gateway WebSocket + Lambda — **rejeitado** na clarificação P1 (serviço extra, custo e ciclo destroy distintos).

## 2. Hub em memória por `aula_id`

**Decision:** Pacote `internal/chat` com um **Hub** de processo:

| Estrutura | Papel |
|-----------|--------|
| `map[uint]*Room` | salas indexadas por `aula_id` |
| `Room.clients` | set de conexões naquela aula |
| canais `register` / `unregister` / `broadcast` | serializar mutações do mapa (goroutine única do hub) |
| mutex por conexão | writes concorrentes seguros |

- Handshake registra o cliente **só** na sala do `aula_id` do path.
- Broadcast de uma mensagem **só** para clientes dessa sala.
- Sala vazia: liberar entrada do mapa (evitar leak de salas órfãs).
- Reinício da task: mapa zera → alinhado à política efêmera (FR-009).

**Rationale:** Spec + clarificação: `desired_count = 1`, sem Redis. Fan-out in-process atende SC-002 na demo. Multi-task exigiria pub-sub — fora do P1.

**Alternatives considered:**
- Redis Pub/Sub / ElastiCache — constitution + FR-007/010 proíbem salvo necessidade; com 1 task, desnecessário.
- Uma goroutine global sem salas — vaza mensagens entre aulas; rejeitado (FR-002).
- Persistência em Postgres — contradiz “só enquanto conectado”; rejeitado no P1.

## 3. Path WebSocket e handshake

**Decision:**

```http
GET /api/v1/aulas/:id/chat/ws?token=<JWT>
Upgrade: websocket
```

- `:id` = `aula_id` (uma aula por conexão; sem `join` posterior; sem multiplex).
- `token` = JWT emitido pelo login existente (mesmo segredo/`TokenService`).
- Ordem no upgrade (antes de aceitar o socket, ou imediatamente após com close se falhar):
  1. Token ausente/inválido → **rejeitar** (401/close com motivo PT).
  2. Aula inexistente → rejeitar.
  3. Papel sem leitura de aulas (só `admin`/`professor`/`aluno` existem e todos leem) — alinhar ao middleware de `GET /aulas/:id`.
  4. Carregar **`nome`** do usuário no DB (JWT atual **não** carrega `nome` — só `user_id`, `email`, `role`) e guardar no contexto da conexão para broadcasts.

Cliente local: `ws://localhost:8080/api/v1/aulas/{id}/chat/ws?token=...`  
Cliente AWS: `wss://{cloudfront_api}/api/v1/aulas/{id}/chat/ws?token=...` (derivar de `VITE_API_URL`).

**Rationale:** Clarificações de auth + handshake; path aninhado espelha VOD/live; lookup de `nome` evita ampliar claims JWT só para chat (YAGNI) e garante rótulo correto na UI.

**Alternatives considered:**
- `Sec-WebSocket-Protocol` com token — menos comum em demos; browsers/proxies variam; rejeitado no P1.
- Cookie HttpOnly — fora da clarificação; mudaria baseline auth.
- `nome` só no front (do `me`) — autor poderia spoofar no payload; servidor MUST fixar identidade.

## 4. ALB idle timeout e sticky sessions

**Decision:**

| Item | P1 |
|------|-----|
| `aws_lb.api.idle_timeout` | Subir de **60** (atual em `infra/alb.tf`) para **3600** s (1 h) — suficiente para demo; máximo ALB = 4000 |
| Sticky no target group | **Não** no P1 — WebSocket já é sticky após o upgrade; `desired_count = 1` elimina split de hub |
| Keepalive aplicativo | Ping/pong JSON leve (`chat.ping` / `chat.pong`) a cada **~2–4 min** (abaixo do idle CloudFront de **10 min** e do ALB) |

**Rationale:** Idle 60 s corta WS ociosos e quebra a demo. Aumentar idle é custo zero (atributo do ALB). Sticky explícito é ruído com uma task. CloudFront documenta idle WS de origem ≈ 10 min sem bytes — ping evita corte silencioso.

**Alternatives considered:**
- Idle 4000 s sem ping — ainda esbarra no idle CloudFront 10 min; rejeitado como única medida.
- Sticky `lb_cookie` “por precaução” — YAGNI com `desired_count = 1`; adicionar só se ECS escalar depois.
- Contornar CloudFront e falar WS direto no ALB HTTP — mixed content / URL diferente de `VITE_API_URL`; rejeitado (baseline 001: API via CloudFront HTTPS).

## 5. CloudFront e WebSocket

**Decision:** Manter o caminho existente **browser → CloudFront (HTTPS) → ALB (HTTP) → ECS :8080**. CloudFront **suporta** WebSocket sem recurso extra; a distribution atual usa `Managed-CachingDisabled` + `Managed-AllViewerExceptHostHeader` (adequado a forward de headers de upgrade).

**Rationale:** Sem segundo endpoint; mesmo `VITE_API_URL`; destroy continua único.

**Alternatives considered:**
- Origin request policy custom só para `/chat/ws` — só se a managed falhar na validação; não assumir no plan.
- Desabilitar compress no behavior — docs AWS recomendam cuidado com headers `Sec-WebSocket-*`; validar no quickstart AWS; se houver falha de upgrade, desligar compress nesse behavior na fase 3.

## 6. Custo (sem serviço extra)

**Decision:** P1 **não** adiciona linha de cobrança nova material:

| Componente | Impacto |
|------------|---------|
| Task ECS existente | Já paga na sessão; WS = conexões na mesma task |
| ALB | Já existe; idle timeout ≠ cobrança extra |
| CloudFront API | Já existe; bytes WS entram no tráfego da demo |
| Redis / API GW WS / Lambda | **Não provisionar** |

Custo fora da sessão ≈ 0 após `terraform destroy` (como 001–003). Budget separado intocado.

**Rationale:** Constitution I + FR-006/010.

**Alternatives considered:**
- API GW WS + Lambda (tabela histórica do README) — clarificação rejeitou no P1; README = referência futura, não decisão desta feature.
- IVS Chat — outro produto/preço; fora do escopo.

## 7. Logs sem token na query

**Decision:**

- **Não** habilitar access logs do ALB apontando query string completa no P1 (hoje não há `access_logs` no `alb.tf` — manter assim salvo necessidade).
- Logs da aplicação: ao registrar falha de upgrade, logar **path sem query** (ou só `aula_id` + motivo: token ausente/inválido) — **nunca** `c.OriginalURL()` / query crua com `token=`.
- CloudWatch da task: mesmos cuidados; não `fmt` do request URI completo no handler WS.
- Runbook: lembrar operador de não colar URLs com `?token=` em issues/prints.

**Rationale:** Clarificação + FR-003/014; JWT na query é compromisso consciente do P1 — mitigação = não persistir a query.

**Alternatives considered:**
- Token só em primeiro frame pós-upgrade — mais seguro contra access logs, mas fora da clarificação fechada (`?token=`).
- Rotacionar token de curta duração só para WS — YAGNI no P1.

## 8. Local vs CI

**Decision:**

| Ambiente | Comportamento |
|----------|----------------|
| Compose / API local | WebSocket **real**; mesmo path/protocolo; hub em memória |
| CI | Testes de handler (auth, validação 500 chars, isolamento de sala, erros PT); **sem** Terraform/AWS; multi-cliente E2E **não** obrigatório |
| AWS demo | Mesmo protocolo atrás de CloudFront/ALB |

**Rationale:** Clarificação US4; diferente do stub de vídeo do 003 (aqui não há backend AWS separado a simular).

**Alternatives considered:**
- Stub sem WS no local — rejeitado na clarificação.
- Testcontainers multi-browser no CI — fora do P1 (custo de pipeline).

## 9. Identidade na UI (`nome`)

**Decision:** Payload de broadcast inclui `autor.nome` resolvido no servidor no handshake (lookup `users` por `claims.UserID`). UI exibe **nome**, não e-mail como rótulo obrigatório.

**Rationale:** Spec US2 AC3 + assumption; JWT sem `nome` hoje.

**Alternatives considered:**
- Incluir `nome` no JWT no login — válido, mas muda contrato de auth baseline; só se implement preferir; plan default = lookup no handshake.
- Cliente envia `nome` no `chat.send` — inseguro; rejeitado.

## 10. Moderação P2

**Decision:** Documentada na spec (US5 / FR-016); **fora** de research de implementação desta rodada — sem frames `chat.delete` / `chat.mute` nos contracts P1.

**Rationale:** Clarificação explícita.
