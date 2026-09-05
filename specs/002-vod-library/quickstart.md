# Quickstart: Biblioteca VOD (002)

Validação ponta a ponta da feature sobre o ciclo **001** (apply → publish → demo → destroy). Detalhes de API: [contracts/vod-api.md](./contracts/vod-api.md). Modelo: [data-model.md](./data-model.md). Env: [contracts/vod-env.md](./contracts/vod-env.md).

## Pré-requisitos

- Feature `001-aws-mvp-terraform` operacional (Terraform, publish scripts, seed admin/professor/aluno/curso/aula).
- Credenciais AWS da conta acadêmica; região `us-east-1` (só para caminho B).
- Asset seed versionado: `backend/assets/vod/demo-aula.mp4` (≤ 2 MB; teto 50 MB).
- Local: Docker Compose (ou API+Postgres); `VOD_BACKEND=local`, volume `./data/vod` → `/data/vod`.

## A. Desenvolvimento local (sem AWS)

1. Na raiz: `docker compose up --build` (ou `make infra-up` + API/front). Compose já define `VOD_BACKEND=local`, `VOD_LOCAL_DIR=/data/vod`, `VOD_PLAYBACK_TTL=15m`.
2. Na subida da API, migrations `001`→`003` rodam: aula de demo recebe **um** VOD `publicado` (objeto sob `/data/vod/vod/aulas/{id}/current.mp4`).
3. Login `aluno@estudaja.com` / `aluno123` → **Aulas** → “Abertura da aula magna” → bloco Gravação → reproduzir (seed; **sem** upload manual).
4. Login `professor@estudaja.com` / `professor123` → na mesma aula, substituir por outro MP4 ≤ 50 MB → aluno (outra sessão) vê o novo.
5. Aluno: UI **sem** botões de escrita; `PUT`/`DELETE` VOD → 403 PT.
6. Upload > 50 MB ou `.webm` → 400 PT; aula não fica quebrada.

**Esperado:** SC-001, SC-003, SC-005 (seed), SC-010 localmente.

## B. Demo AWS (passos extras no ciclo 001)

Seguir o runbook do [README (Demo AWS)](../../README.md#demo-aws-mvp-p1) / [001 quickstart](../001-aws-mvp-terraform/quickstart.md); inserir os passos VOD abaixo. **Não** redesenhar a stack 001; **não** tocar `infra/budget/`.

### B1. Apply

```powershell
cd infra
terraform apply
```

**Verificar:** `terraform output vod_bucket_name` presente; bucket privado (SSE, block public, `force_destroy`).

### B2. Publish API (+ front)

```powershell
.\publish-api.ps1
# Aguardar GET {api_url}/health → 200
.\publish-frontend.ps1
```

**Verificar:** task com `VOD_BACKEND=s3` e `VOD_S3_BUCKET`; migration `003_seed_vod_demo` criou objeto `vod/aulas/.../current.mp4` no bucket (seed na imagem via `COPY assets/vod`).

### B3. Demo funcional

1. Abrir `frontend_url` → login aluno seed → **Aulas** → reproduzir VOD (≤ 2 min até iniciar; áudio/vídeo ≤ 10 s após “assistir”).
2. Login professor → upload MP4 válido → aluno confirma novo conteúdo.
3. Confirmar: **sem** página Biblioteca; **sem** botão download; visitante sem JWT não reproduz.
4. (Opcional) Aguardar ~15 min sem renovar playback → URL antiga falha; novo `GET /api/v1/aulas/:id/vod/playback` com JWT funciona (S3 presigned).

### B4. Destroy

```powershell
cd infra
terraform destroy
```

**Verificar:** bucket VOD e objetos removidos; sem armazenamento VOD cobrável ocioso; stack `infra/budget/` intacta.

### B5. Novo apply

Após destroy, `apply` + publish de novo → aluno assiste VOD de seed **sem** upload manual (SC-005). Uploads da sessão anterior **não** voltam.

## Critérios rápidos de aceite

| ID | Check |
|----|--------|
| SC-001/006 | Aluno toca seed na ficha da aula |
| SC-002 | Após publish do professor, aluno toca |
| SC-003 | Aluno não escreve; anônimo não lê |
| SC-004 | Destroy zera VOD |
| SC-005 | Re-seed automático (1 VOD na aula demo) |
| SC-012 | Playback ~15 min |

## Fora desta validação

Streaming ao vivo, chat, certificado, OAuth, Redis AWS, NAT, domínio custom, página Biblioteca, download do arquivo, P2 rascunho/título/duração (Fase 5).
