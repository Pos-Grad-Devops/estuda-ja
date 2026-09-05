# Contract: variáveis de ambiente Live

Extensão do contrato [001 api-env](../../001-aws-mvp-terraform/contracts/api-env.md) e [002 vod-env](../../002-vod-library/contracts/vod-env.md). Variáveis **adicionais** da API Go (`backend/internal/config`).

## Variáveis

| Variável | Obrigatória | Default local | Fonte AWS | Notas |
|----------|-------------|---------------|-----------|--------|
| `LIVE_BACKEND` | sim (comportamento) | `stub` | env task `ivs` | `stub` \| `ivs` |
| `IVS_INGEST_ENDPOINT` | se `ivs` | — | env ou SSM String | Hostname do canal (`….global-contribute.live-video.net`), **sem** obrigar o prefixo `rtmps://` no env — a API monta `rtmps://{host}:443/app/` |
| `IVS_STREAM_KEY` | se `ivs` | — | SSM SecureString | Key da sessão; **nunca** no git; estável até destroy |
| `IVS_PLAYBACK_URL` | se `ivs` | — | env task | HLS `.m3u8` do canal; **não** publicar em output de produto |
| `IVS_CHANNEL_ARN` | não | — | env task (output TF) | Logs/ops; opcional no P1 |

Credenciais IVS na AWS: **injetadas** (SSM + env), sem access key IVS no env e **sem** `ivs:*` na task role no P1.

VOD (`VOD_*`) permanece independente.

## Comportamento de falha

- `LIVE_BACKEND=ivs` sem `IVS_STREAM_KEY` ou ingest/playback → **fail-fast no boot** (preferível na AWS) **ou** 503 em `start`/ingest/playback; **nunca** marcar `ao_vivo` sem config. Preferir fail-fast no boot na task ECS (espelho `VOD_BACKEND=s3` sem bucket).
- `LIVE_BACKEND=stub` ignora `IVS_*` (podem estar vazios).
- `LIVE_BACKEND` inválido → tratar como erro de boot (PT nos logs) ou default `stub` só em dev; na AWS MUST ser `ivs`.

## Compose (referência)

```yaml
# trecho ilustrativo — aplicar na implement
environment:
  LIVE_BACKEND: stub
```

Sem volume extra. Sem Redis. Sem dependência de AWS no CI.

## Segredos

Não commitar `IVS_STREAM_KEY`, `*.tfvars`, state. Não logar a stream key. CloudWatch: evitar `printf` da key.
