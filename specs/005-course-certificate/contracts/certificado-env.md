# Contract: Env / infra — Certificado (005)

P1 usa **PDF on-the-fly** (sem objeto em storage). Nenhuma variável nova é **obrigatória**. Baseline 001–004 permanece; budget intocado.

## Variáveis de ambiente

| Variável | Obrigatória | Default | Notas |
|----------|-------------|---------|--------|
| *(nenhuma `CERT_*`)* | — | — | Geração local na API a partir do registro |

Reutiliza o que já existe (DB, JWT, seed `DEMO_ALUNO_*`, etc.).

## Compose / local

- Sem volume de PDFs.
- Sem serviço extra.
- Fluxo: API gera `application/pdf` na resposta de `GET .../certificado/pdf`.

## AWS (demo efêmera)

| Item | Decisão P1 |
|------|-------------|
| S3 de certificados | **Não** provisionar |
| IAM task | Sem policy nova de certificado |
| ECS env | Sem `CERT_*` obrigatório |
| CloudFront / domínio custom | Sem mudança |
| NAT / Redis / Cognito | Proibidos (inalterados) |
| Região | `us-east-1` |
| Budget (`infra/budget/`) | **Intocado** |

## Destroy

```text
terraform destroy   # stack demo 001(+002/003/004)
```

Efeito em certificados:

- Registros em `certificado_elegibilidades` e `certificados` → removidos com o **RDS**.
- **Não** há objetos S3 de certificado a limpar.
- Próximo `apply` + migrations/seed recria elegibilidade do aluno seed (sem PDF pré-emitido).

## Fora

- `CERT_S3_BUCKET`, prefixo no bucket VOD, EFS, ACM custom para “URL estável do PDF”.
