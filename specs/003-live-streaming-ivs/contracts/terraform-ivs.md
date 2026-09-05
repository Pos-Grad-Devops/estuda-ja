# Contract: Terraform / outputs IVS

Extensão enxuta da stack em `infra/` (baseline 001 + bucket VOD 002). Região **us-east-1**. Sem NAT, sem Redis AWS, sem domínio customizado, sem Cognito, sem IVS Chat, sem recording configuration. Budget em `infra/budget/` **não** faz parte deste contrato.

## Recursos novos (P1)

| Recurso | Nome / padrão | Requisitos |
|---------|---------------|------------|
| IVS channel | `${project_name}-live` | `type = BASIC`, `latency_mode = LOW`, `authorized = false`; **sem** `recording_configuration_arn` |
| Stream key | associada ao canal | Uma key da sessão; valor → SSM SecureString |
| SSM | `${ssm_prefix}/IVS_STREAM_KEY` | SecureString; execution role já existente ganha este ARN no `GetParameters` |
| ECS task | env + secrets | `LIVE_BACKEND=ivs`; `IVS_INGEST_ENDPOINT`; `IVS_PLAYBACK_URL`; `IVS_CHANNEL_ARN`; secret `IVS_STREAM_KEY` |
| IAM task role | — | **Sem** policy `ivs:*` no P1 |

Arquivo sugerido: `infra/ivs.tf` (+ edits em `ssm.tf`, `iam.tf`, `ecs.tf`, `outputs.tf`). Provider AWS já `~> 5.0` (IVS suportado). Sem módulos novos.

## Outputs novos

| Output | Sensível | Descrição |
|--------|----------|-----------|
| `ivs_channel_arn` | não | ARN do canal da sessão (ops / checklist T−15) |

**Não** outputar `stream_key` nem `playback_url` (produto só via API JWT; key só SSM + endpoint ingest autenticado).

Outputs 001/002 (`frontend_url`, `api_url`, `vod_bucket_name`, …) permanecem.

## Destroy

`terraform destroy` da stack demo MUST remover o canal IVS e a stream key. Custo contínuo esperado de IVS ≈ **0**. Não destruir `infra/budget/`. Não deixar canal “de teste” criado na console.

## Fora deste contrato

- Canal permanente / multi-canal
- IVS Chat, Stages Real-Time, playback key pair, recording → S3
- CloudFront de mídia live
- ElastiCache, NAT, ACM custom
- Alterar CD P3 além de rebuild da imagem (env vem da task definition no apply; publish manual P1 basta)
