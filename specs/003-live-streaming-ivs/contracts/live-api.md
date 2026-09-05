# Contract: API Live

Superfície HTTP nova sob `/api/v1/`. Datas em JSON: `dd/mm/yyyy` ou `dd/mm/yyyy HH:mm`. Erros em português. Auth: Bearer JWT (middleware existente).

## Autorização

| Ação | admin | professor | aluno | anônimo |
|------|-------|-----------|-------|---------|
| Ler status | ✓ | ✓ | ✓ | ✗ |
| Playback (aula `ao_vivo`) | ✓ | ✓ | ✓ | ✗ |
| Iniciar / encerrar | ✓ | ✓ | ✗ | ✗ |
| Agendar / cancelar agendada (P2) | ✓ | ✓ | ✗ | ✗ |
| Ingest (endpoint + stream key) | ✓ | ✓ | ✗ | ✗ |

Escrita: mesmo recorte de `RequireRoles(admin, professor)` das aulas. Qualquer professor pode gerir live de qualquer aula.

## Endpoints P1

Base: `/api/v1/aulas/:id/live`

### GET `/api/v1/aulas/:id/live`

Status da transmissão **desta** aula. Sem credenciais de ingestão.

**200** — com ou sem live prévia:

```json
{
  "aula_id": 1,
  "status": "inativa",
  "modo": "stub",
  "iniciada_em": null,
  "encerrada_em": null
}
```

`status`: `inativa` \| `agendada` \| `ao_vivo` \| `encerrada`.  
`modo`: `stub` \| `ivs` (espelha `LIVE_BACKEND`).

**401** — sem JWT.  
**404** — aula inexistente.

Nunca incluir `stream_key` neste payload.

### POST `/api/v1/aulas/:id/live/schedule` (P2)

Associa a live como **agendada**. **Exige** `agendada_em` na aula. Idempotente se já `agendada` nesta aula.

**200** — `status: "agendada"`.  
**400** — aula sem horário (`agendada_em`).  
**401/403** — sem sessão ou aluno.  
**404** — aula inexistente.  
**409** — esta aula já está `ao_vivo`.

### POST `/api/v1/aulas/:id/live/cancel` (P2)

Cancela live **agendada** (remove a linha → resposta `inativa`). Não aplica a `ao_vivo`/`encerrada`.

**200** — `status: "inativa"`.  
**401/403/404** — como acima.  
**409** — esta aula não está `agendada`.

### POST `/api/v1/aulas/:id/live/start`

Ação única P1 (ainda válida): associa ao canal da sessão, marca `ao_vivo`, habilita ingest. **Não** exige `agendada_em`. Também cobre transição P2 `agendada` → `ao_vivo`.

**200/201** — status como GET, com `status: "ao_vivo"` e `iniciada_em` preenchido.

**400** — pedido inválido (PT).  
**401/403** — sem sessão ou aluno.  
**404** — aula inexistente.  
**409** — já existe live `ao_vivo` em **outra** aula (ou nesta, se a API tratar start idempotente: preferir **200** se já `ao_vivo` **nesta** aula — YAGNI: 200 idempotente nesta aula **ou** 409 “já ao vivo aqui”; escolher **idempotente 200 nesta aula** na implement).  
**503** — `LIVE_BACKEND=ivs` sem infra/credenciais; **não** persiste `ao_vivo`.

### POST `/api/v1/aulas/:id/live/stop`

Encerra a live desta aula (`encerrada`). Não destrói o canal. Permite novo `start` depois.

**200** — `status: "encerrada"`.  
**401/403/404** — como acima.  
**409** — esta aula não está `ao_vivo`.

### GET `/api/v1/aulas/:id/live/playback`

Dados de reprodução **somente** se esta aula estiver `ao_vivo`. Gate = JWT + leitura de aulas. **Sem** URL permanente publicada fora desta API.

**200** modo `ivs`:

```json
{
  "aula_id": 1,
  "modo": "ivs",
  "protocolo": "hls",
  "player": "ivs",
  "playback_url": "https://….playback.live-video.net/….m3u8"
}
```

**200** modo `stub`:

```json
{
  "aula_id": 1,
  "modo": "stub",
  "protocolo": null,
  "player": null,
  "playback_url": null,
  "mensagem": "Ambiente local: não há sinal de vídeo real."
}
```

**401/403** — anônimo / sem leitura.  
**404** — aula inexistente **ou** live não `ao_vivo` (“Não há transmissão ao vivo nesta aula.”).

P1 **não** inclui `expires_at` / playback token IVS.

A URL HLS **não** MUST ser cacheada de forma permanente no frontend (reconsultar status/playback ao abrir a ficha). Residual: URL IVS não autorizada é reproduzível se vazada até o destroy — ver research §9.

### GET `/api/v1/aulas/:id/live/ingest`

Credenciais OBS. Só gestor e só enquanto **esta** aula está `ao_vivo`.

**200** modo `ivs`:

```json
{
  "aula_id": 1,
  "modo": "ivs",
  "ingest_server": "rtmps://{ingest_endpoint}:443/app/",
  "stream_key": "sk_us-east-1_…",
  "observacao": "A mesma stream key vale até o destroy da sessão Terraform."
}
```

**200** modo `stub`:

```json
{
  "aula_id": 1,
  "modo": "stub",
  "ingest_server": null,
  "stream_key": null,
  "mensagem": "Ingestão real só na sessão AWS."
}
```

**401/403** — aluno ou anônimo (mensagem PT; **sem** body com key).  
**404** — aula inexistente ou live desta aula não `ao_vivo`.

## Efeito colateral no CRUD de aulas

`DELETE /api/v1/aulas/:id` MUST remover `AulaLive` associado. **Não** altera o canal IVS.

`Aula.Status` do CRUD **não** é atualizado por start/stop no P1.

## Fora deste contrato

- Página/rota frontend “Ao vivo”
- Playback token IVS / signed URL
- Chat, recording, multi-canal
- Pipeline live→VOD automático
- EventBridge apply/destroy do 001
