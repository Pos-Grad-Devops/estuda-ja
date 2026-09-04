# Contract: variáveis de ambiente da API

Contrato de runtime da API Go (`backend/internal/config`). Não inventa endpoints REST novos — a superfície HTTP existente permanece sob `/api/v1/` e `/health`.

## Variáveis

| Variável | Obrigatória (AWS) | Default local | Fonte AWS | Notas |
|----------|-------------------|---------------|-----------|--------|
| `PORT` | sim | `8080` | env task | ALB target group → container 8080 |
| `DATABASE_URL` | sim | postgres local | SSM SecureString | `sslmode=require` típico no RDS |
| `CORS_ORIGIN` | sim | `http://localhost:5173` | env / SSM String | Origem exata do CloudFront do **frontend** (HTTPS) |
| `JWT_SECRET` | sim | `dev-secret-change-me` | SSM SecureString | Nunca default de dev na AWS |
| `JWT_EXPIRATION` | não | `24h` | env | Duração Go (`time.ParseDuration`) |
| `ADMIN_EMAIL` | sim (seed) | `admin@estudaja.com` | env / SSM | Usado na migration de admin |
| `ADMIN_PASSWORD` | sim (seed) | `admin123` | SSM SecureString | Apenas seed; rotacionar por sessão se desejado |

Redis **não** faz parte deste contrato na AWS.

## Comportamento de falha

- `DATABASE_URL` inválida / RDS inacessível → processo encerra no `Connect` (`log.Fatalf`).
- Migração/seed falha → processo encerra (`migrate`).
- Segredo ausente com fallback inseguro: na AWS, task definition MUST injetar valores reais (sem depender do default de JWT de desenvolvimento).

## Superfície HTTP estável (referência)

| Método | Path | Auth |
|--------|------|------|
| GET | `/health` | público |
| POST | `/api/v1/auth/login` | público |
| CRUD | `/api/v1/cursos`, `/aulas`, `/alunos`, `/users` | JWT + RBAC existente |

Datas JSON: `dd/mm/yyyy` ou `dd/mm/yyyy HH:mm`. Erros em português.
