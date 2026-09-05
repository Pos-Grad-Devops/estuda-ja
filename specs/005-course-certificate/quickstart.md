# Quickstart: Certificado ao final do curso (005)

Validação ponta a ponta **após** implementação das fases 1–3 — este arquivo é o guia; não implementa. Contratos: [certificado-api.md](./contracts/certificado-api.md), [certificado-env.md](./contracts/certificado-env.md). Modelo: [data-model.md](./data-model.md).

## Pré-requisitos

- Baseline **001–004** operacional (API + front + seed demo + VOD/live/chat conforme já entregue).
- Contas seed: admin, `aluno@estudaja.com` / `aluno123`, curso demo “Aula Magna — Direito Constitucional”.
- Feature 005 implementada conforme plan (PDF on-the-fly; sem S3 de certificado).

## A — Compose (caminho feliz)

### Setup

```bash
make infra-up
make backend
make frontend
```

Ou `docker compose up --build` equivalente. Sem vars `CERT_*` / S3 de certificado.

### Passos

1. **Admin** — login → abrir **Cursos** → no painel do curso demo: confirmar elegibilidade do aluno seed (seed já marca) **ou** marcar `user_id` do aluno → **não** deve existir certificado emitido ainda.
2. **Aluno** (janela anônima) — login `aluno@estudaja.com` → mesmo curso → solicitar/baixar certificado.
3. Verificar download **PDF** com nome “Aluno Demo”, título do curso e data `dd/mm/yyyy` (ou com hora).
4. Baixar de novo → mesmo certificado ativo (sem segundo registro válido).
5. **Admin** — invalidar o certificado do par → aluno tenta baixar → recusa em português; elegibilidade bloqueada.
6. **Admin** — reabilitar elegibilidade → aluno solicita de novo → **novo** PDF/registro válido; o anterior permanece invalidado.
7. **Professor** — login → UI **sem** controles de gestão de certificado; tentativas de API de gestão → 403 PT.

### Esperado

| Check | Resultado |
|-------|-----------|
| Seed = só elegibilidade | Sem PDF pré-emitido |
| 1ª solicitação aluno | Cria registro + PDF |
| 2ª solicitação | Reutiliza ativo |
| Invalidar | Bloqueia até reabilitar |
| Reabilitar + solicitar | Novo certificado |
| Professor / aluno gestão | Sem UI; API 403 |
| Live/VOD/chat | Intactos |

## B — CI (sem AWS)

```bash
cd backend && go test ./...
cd frontend && npm run build
```

Esperado: testes de handlers/RBAC/lazy emit/invalidação/erros PT verdes. **Não** exige Terraform nem conta AWS.

## C — AWS (sessão efêmera)

Mesmo comportamento da §A na API atrás do `api_url` da sessão (`us-east-1`).

```text
apply → publish-api → publish-frontend (VITE_API_URL) → demo certificado → destroy
```

- **Sem** bucket/recursos novos de certificado (PDF on-the-fly).
- **Sem** NAT / Redis AWS / Cognito / domínio custom.
- Budget (`infra/budget/`) **não** destruir com a demo.

### Destroy — o que acontece com certificados

```bash
cd infra
terraform destroy
```

| Artefato | Após destroy |
|----------|----------------|
| Linhas `certificado_elegibilidades` / `certificados` | Removidas com o RDS |
| Arquivos PDF em S3 | N/A (não existem no P1) |
| Budget / alerta da conta | Permanece (`infra/budget/`) |

Próxima sessão: `apply` + seed → elegibilidade do aluno seed de novo; primeira solicitação emite PDF fresco.

## Tempo alvo

- Aluno elegível obtém PDF ≤ 2 min após login (SC-001).
- Operador lê/executa passos extras de certificado ≤ 15 min incremental (SC-005).
