# Contract: Terraform / outputs VOD

Extensão enxuta da stack em `infra/` (baseline 001). Região **us-east-1**. Sem NAT, sem Redis AWS, sem domínio customizado. Budget em `infra/budget/` **não** faz parte deste contrato.

## Recursos novos (P1)

| Recurso | Nome / padrão | Requisitos |
|---------|---------------|------------|
| S3 bucket VOD | `${project_name}-vod-${account_id}` | Privado; SSE AES256; block public; `force_destroy = true`; **sem** OAC/CloudFront no P1 |
| IAM (task role) | policy anexa a `ecs_task` | `s3:PutObject`, `s3:GetObject`, `s3:DeleteObject` em `arn:aws:s3:::bucket/vod/*`; `s3:ListBucket` no bucket com prefix `vod/` se necessário |
| ECS task env | plain env | `VOD_BACKEND=s3`, `VOD_S3_BUCKET`, `VOD_PLAYBACK_TTL=15m`, `AWS_REGION=us-east-1` |

Arquivo sugerido: `infra/s3_vod.tf` (+ edits em `iam.tf`, `ecs.tf`, `outputs.tf`).

## Outputs novos

| Output | Descrição |
|--------|-----------|
| `vod_bucket_name` | Nome do bucket S3 de mídia VOD (sessão) |

Outputs 001 existentes (`frontend_url`, `api_url`, `s3_bucket_name` do **frontend**, etc.) permanecem.

## Destroy

`terraform destroy` da stack demo MUST remover o bucket VOD e objetos (`force_destroy`). Não deixar biblioteca cobrável ociosa. Não destruir `infra/budget/`.

## Fora deste contrato

- Bucket permanente fora da stack
- CloudFront de mídia (P1)
- ElastiCache, NAT, ACM custom
- Alterar CD P3 além de passar env VOD se o workflow já injeta task env (opcional na implement; publish manual P1 basta)
