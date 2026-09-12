# Usar WebSocket na API com hub em memória para o chat da aula

| Campo | Valor |
| --- | --- |
| Status | Aceita |
| Data | 2026-09-05 |
| Decisores | Gustavo Santos Arruda, Pedro Lucas dos Santos Ribeiro, Renan Roseno dos Santos, Victor da Silva Neves |
| Feature | Chat em tempo real da aula |
| Tags | chat, websocket, hub, demo-aws |

## Contexto e problema

Na aula magna o aluno não levanta a mão. Sem um canal de pergunta **na mesma ficha** da aula (`/aulas/:id`), a live vira TV passiva: o professor não sente a turma e o aluno não participa.

O MVP é um painel WebSocket com uma sala por aula, mensagens de até 500 caracteres, e participação de aluno, professor e admin **enquanto estão conectados**. A sala precisa funcionar **com ou sem** live ligada. O chat **não** amplia permissões de ingestão ou de VOD.

A tabela histórica do README e dos slides ainda apontava API Gateway WebSocket + Lambda (AWS) ou WebSocket + Redis Pub/Sub (self-hosted). A demo, porém, já tem uma API em ECS Fargate (`desired_count = 1`), ciclo `destroy` entre sessões, e a constituição do projeto proíbe Redis/ElastiCache na AWS neste MVP.

## Critérios de decisão

Qualquer opção precisava:

- Entregar envio e recebimento na ficha, em tempo quase real, com isolamento por aula.
- Reusar login JWT e os papéis já existentes (`admin`, `professor`, `aluno`).
- Não adicionar serviço cobrável contínuo (sem stack paralela de chat).
- Caber em **uma** task ECS; o destroy da API deve encerrar conexões e mensagens.
- Funcionar de verdade no Compose local (mesmo protocolo da demo); o CI cobre handlers sem AWS.
- Manter conexões longas atrás do Application Load Balancer (ALB) e do CloudFront já usados pela API.

## Opções consideradas

1. **API Gateway WebSocket + Lambda** — serviço extra, custo e ciclo de destroy distintos da stack 001; rejeitado na clarificação P1 da spec.
2. **Redis / ElastiCache Pub/Sub** — necessário se houvesse várias tasks compartilhando salas; com uma task, é custo e ops sem ganho. Constituição e spec proíbem Redis AWS neste MVP.
3. **IVS Chat** — outro produto AWS, outro preço; o chat da aula não deve depender do canal de vídeo.
4. **Persistir mensagens no PostgreSQL** — contradiz a política P1: só enquanto o usuário está conectado; sem histórico ao reabrir a ficha.
5. **WebSocket no mesmo serviço Fiber/ECS, hub em memória por `aula_id`** (escolhida).

## Decisão

O chat vive **na API já existente** (Go + Fiber, mesma task ECS da demo). O cliente abre:

```http
GET /api/v1/aulas/:id/chat/ws?token=<JWT>
Upgrade: websocket
```

Uma conexão = uma aula. Sem frame `join` e sem multiplex. O handshake valida o JWT, a existência da aula e o papel com leitura de aulas; o servidor resolve o **nome** do usuário no banco (o JWT não carrega `nome`) para o rótulo das mensagens. Textos vazios ou com mais de 500 caracteres são recusados em português.

O fan-out fica no pacote `internal/chat`: um **Hub** de processo com salas indexadas por `aula_id`. Broadcast só para clientes daquela sala. Reiniciar a task (ou dar `destroy`) zera o mapa — alinhado à retenção efêmera.

Na AWS não há Redis, stickiness no target group, API Gateway WebSocket nem Lambda. O ALB usa `idle_timeout = 3600` segundos (o default de 60 s cortaria conexões ociosas). Keepalive de aplicativo (`chat.ping` / `chat.pong` a cada 2 minutos) evita o idle de ~10 minutos do CloudFront na origem. O caminho permanece browser → CloudFront HTTPS → ALB → ECS :8080.

No Compose o WebSocket é **real** (porta 8080). O CI exercita auth, limite de 500 caracteres, isolamento de sala e erros em português — sem Terraform.

![Caminho do chat: browser na ficha, CloudFront da API, ALB com idle de 3600 s, task ECS Fiber e hub em memória com broadcast por aula](./diagrams/0002-chat.svg)

## Consequências

**Positivas**

- Um protocolo só, local e AWS; nada a “simular” como o stub de vídeo da live.
- Custo extra da feature ≈ 0: mesma task, mesmo ALB, mesmo CloudFront.
- Isolamento por aula e RBAC de live/VOD intactos (chat não dá ingest nem upload ao aluno).
- Independência da live: dá para conversar na ficha mesmo com a transmissão inativa.

**Negativas / aceitas de propósito**

- **Uma** task: se no futuro `desired_count` > 1, clientes em tasks diferentes não se veem (faria falta pub/sub).
- Sem histórico ao reabrir a ficha; mensagens morrem com a conexão, com o restart da API e com o destroy.
- JWT na query do handshake é um compromisso do P1: logs da aplicação **não** devem registrar a URI completa com `token=`.

**Fora desta decisão (e do P1)**

- Moderação, reações, fila de perguntas, anexos, chat no replay, escala horizontal.

## Como conferir no código

- Hub: [`backend/internal/chat/hub.go`](../../../backend/internal/chat/hub.go).
- Handshake, limite de 500 caracteres, ping/pong: [`backend/internal/handler/chat_handler.go`](../../../backend/internal/handler/chat_handler.go).
- Rota: [`backend/cmd/api/main.go`](../../../backend/cmd/api/main.go) (`GET /aulas/:id/chat/ws`).
- ALB: [`infra/alb.tf`](../../../infra/alb.tf) (`idle_timeout = 3600`).
- UI: [`frontend/src/components/AulaChatPanel.tsx`](../../../frontend/src/components/AulaChatPanel.tsx).

## Mais informações

- Problema e MVP da feature: [`docs/entregaveis/template-features.md`](../template-features.md) (Feature 2).
- Spec e research: [`specs/004-live-class-chat/spec.md`](../../../specs/004-live-class-chat/spec.md), [`research.md`](../../../specs/004-live-class-chat/research.md).
- Protocolo: [`specs/004-live-class-chat/contracts/chat-ws.md`](../../../specs/004-live-class-chat/contracts/chat-ws.md).
- Operação na demo: [`specs/004-live-class-chat/quickstart.md`](../../../specs/004-live-class-chat/quickstart.md).
- Runbook: seção Demo AWS em [`README.md`](../../../README.md) (chat WebSocket na API).
