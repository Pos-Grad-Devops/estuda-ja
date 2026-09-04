# Contract: publish front ↔ API (por sessão)

Acoplamento explícito após cada `terraform apply` (clarificação da spec). Sem proxy `/api` no CDN e sem `config.json` dinâmico nesta feature.

## Pré-condições

1. Infra aplicada; outputs `api_url` e `frontend_url` disponíveis.
2. Imagem da API no ECR; serviço ECS com task **healthy** (`GET {api_url}/health` → 200).
3. `CORS_ORIGIN` da API = origem de `frontend_url` (scheme + host, sem path).

## Contrato de build do frontend

| Item | Valor |
|------|-------|
| Variável de build | `VITE_API_URL` |
| Valor | output `api_url` (HTTPS CloudFront da API), **sem** barra final preferencialmente |
| Consumo no código | `frontend/src/api/client.ts` → `import.meta.env.VITE_API_URL` |
| Artefato | conteúdo de `frontend/dist/` |
| Destino | bucket S3 (`s3_bucket_name`) com sync; OAC via CloudFront |
| Pós-deploy | invalidação CloudFront (`/*`) no distribution do frontend |

Exemplo de invocação (conceitual):

```bash
VITE_API_URL="$API_URL" npm run build
aws s3 sync dist/ "s3://$BUCKET/" --delete
aws cloudfront create-invalidation --distribution-id "$DIST_ID" --paths "/*"
```

## Verificação de acoplamento (SC-011)

1. Abrir `frontend_url` no browser.
2. Login admin (seed).
3. Listar cursos — sucesso implica front e API da **mesma** sessão.

## Falhas detectáveis

| Sintoma | Causa provável |
|---------|----------------|
| Mixed content / blocked | `VITE_API_URL` em HTTP ou ALB cru |
| CORS error | `CORS_ORIGIN` ≠ origem do CloudFront front |
| Login 404 / Network error | front ainda com URL de sessão anterior |
| 502 no `api_url` | task unhealthy / imagem não publicada |

## Fora deste contrato

- CD automatizado (P3)
- Domínio customizado
- Reutilização de `VITE_API_URL` entre destroys
