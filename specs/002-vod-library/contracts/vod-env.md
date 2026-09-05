# Contract: variáveis de ambiente VOD

Extensão do contrato [001 api-env](../../001-aws-mvp-terraform/contracts/api-env.md). Variáveis **adicionais** da API Go (`backend/internal/config`).

## Variáveis

| Variável | Obrigatória | Default local | Fonte AWS | Notas |
|----------|-------------|---------------|-----------|--------|
| `VOD_BACKEND` | sim (comportamento) | `local` | env task `s3` | `local` \| `s3` |
| `VOD_LOCAL_DIR` | se `local` | `./data/vod` ou `/data/vod` no Compose | — | Raiz dos arquivos; keys relativas iguais ao S3 |
| `VOD_S3_BUCKET` | se `s3` | — | env task (output Terraform) | Bucket efêmero da sessão |
| `VOD_PLAYBACK_TTL` | não | `15m` | env task | `time.ParseDuration`; clarificação ~15 min |
| `AWS_REGION` | se `s3` | — | env task `us-east-1` | SDK S3; região fixa da demo |

Credenciais S3 na AWS: **task role** IAM (sem access key no env). Local: filesystem apenas.

Segredos novos de VOD: **nenhum** além do já coberto pelo 001 (JWT etc.). Não commitar keys.

## Comportamento de falha

- `VOD_BACKEND=s3` sem `VOD_S3_BUCKET` → falha no boot ou na primeira operação de storage (preferir fail-fast no boot na AWS).
- `VOD_BACKEND=local` e diretório inexistente → criar no boot ou erro claro em PT.
- Upload > 50 MB / não MP4 → 400 PT; não marcar aula como publicada.

## Compose (referência)

```yaml
# trecho ilustrativo — aplicar na implement
environment:
  VOD_BACKEND: local
  VOD_LOCAL_DIR: /data/vod
  VOD_PLAYBACK_TTL: 15m
volumes:
  - ./data/vod:/data/vod
```
