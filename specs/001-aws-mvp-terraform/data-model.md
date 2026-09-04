# Data Model: 001-aws-mvp-terraform

Foco no que **muda ou aparece** para o ambiente AWS de demo. Entidades de negócio (`Curso`, `Aula`, `Aluno`, `User`) já existem e **não** são redesenhadas — apenas seed/config operacional.

## Entidades de negócio (baseline — sem alteração estrutural)

| Entidade | Campos relevantes | Notas |
|----------|-------------------|--------|
| **User** | nome, email, password_hash, role (`admin` \| `professor` \| `aluno`) | RBAC inalterado |
| **Curso** | titulo, descricao, timestamps BR | CRUD admin |
| **Aula** | curso_id, titulo, descricao, agendada_em, status | CRUD admin/professor |
| **Aluno** | nome, email | CRUD admin |

Relações existentes: `Aula` → `Curso`; papéis em `User` independentes da tabela `Aluno` (modelo atual preservado).

## Entidades / conceitos novos (operacionais AWS)

### Ambiente de demo

Instância lógica efêmera: CloudFront (front) + CloudFront/ALB (API) + ECS task + RDS + parâmetros SSM.

| Campo lógico | Origem | Persistência |
|--------------|--------|--------------|
| `frontend_url` | output Terraform (CloudFront) | Efêmero (muda a cada apply) |
| `api_url` | output Terraform (CloudFront → ALB) | Efêmero |
| `region` | `us-east-1` | Fixo |
| `session_id` | implícito (state/apply) | Não versionado |

**Regra:** após destroy, nenhum dado de app permanece; próximo apply recria banco vazio + seed.

### Segredo operacional

| Nome lógico | Store | Consumidor |
|-------------|-------|------------|
| `JWT_SECRET` | SSM SecureString | Task ECS |
| `ADMIN_PASSWORD` | SSM SecureString | Task (seed admin) |
| `DATABASE_URL` / senha RDS | SSM SecureString (ou composição no TF → SSM) | Task |
| `ADMIN_EMAIL` | SSM String ou env TF | Task (seed) |
| Credenciais AWS do operador / CI | Fora do repo | CLI / (P3) Actions secrets |

**Validação:** nunca em git; ausência → API não sobe (“fail loud” no `Connect` / `Load`).

### Artefato publicável

| Artefato | Conteúdo | Acoplamento |
|----------|----------|-------------|
| Imagem API | binário Go em ECR (`linux/arm64`) | Tag `latest` + force deploy |
| Bundle frontend | `dist/` estático no S3 | Build com `VITE_API_URL=<api_url da sessão>` |

### Janela de aula simulada

| Estado | Condição |
|--------|----------|
| `provisioning` | `terraform apply` em andamento |
| `warming` | recursos up; health checks / seed |
| `ready` | `/health` OK; front publicado com URL correta; CORS alinhado |
| `destroyed` | `terraform destroy` + checklist de resíduos |

## Seed / migração (dados efêmeros)

Fluxo na subida do container (já existente):

1. `AutoMigrate` / gormigrate schema
2. Migration `001_seed_admin` — cria admin se não houver
3. **Nova migration de demo (P1)** — se não existirem, criar no mínimo:
   - 1 usuário `professor` (credenciais documentadas só em guia / vars, não hardcoded em commit se sensíveis)
   - 1 usuário `aluno`
   - 1 `Curso` de exemplo
   - 1 `Aula` ligada ao curso

**Validação (SC-010):** login admin + listar ≥1 curso sem cadastro manual.

**Idempotência:** seeds “create if missing” (padrão do admin atual); banco novo a cada apply torna a condição trivialmente verdadeira.

## Configuração de runtime (não é tabela)

Espelha `backend/internal/config.Config`:

- `PORT`, `DATABASE_URL`, `CORS_ORIGIN`, `JWT_SECRET`, `JWT_EXPIRATION`, `ADMIN_EMAIL`, `ADMIN_PASSWORD`

`CORS_ORIGIN` MUST ser a origem HTTPS do CloudFront do frontend da sessão.

## State transitions (RDS / sessão)

```text
[inexistente] --apply--> [RDS vazio] --task start + migrate/seed--> [dados demo]
[dados demo] --destroy--> [inexistente]
```

Não há backup/restore entre sessões (fora de escopo).
