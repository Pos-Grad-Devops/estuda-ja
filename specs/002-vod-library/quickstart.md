# Quickstart: Biblioteca VOD (002)

Validação ponta a ponta da feature sobre o ciclo **001** (apply → publish → demo → destroy). Detalhes de API: [contracts/vod-api.md](./contracts/vod-api.md). Modelo: [data-model.md](./data-model.md).

## Pré-requisitos

- Feature `001-aws-mvp-terraform` operacional (Terraform, publish scripts, seed admin/professor/aluno/curso/aula).
- Credenciais AWS da conta acadêmica; região `us-east-1`.
- Asset seed: `backend/assets/vod/demo-aula.mp4` (após fase 4 de implement).
- Local: Docker Compose ou API+Postgres; `VOD_BACKEND=local`.

## A. Desenvolvimento local (sem AWS)

1. Subir infra local (`make infra-up` / Compose) com volume VOD e env `VOD_*` (ver [vod-env.md](./contracts/vod-env.md)).
2. Subir API e frontend.
3. Login `aluno@estudaja.com` → abrir **Aulas** → aula de demo → deve indicar gravação e reproduzir (seed).
4. Login `professor@estudaja.com` → na mesma aula, substituir por outro MP4 ≤ 50 MB → aluno (outra sessão) vê o novo.
5. Tentar upload como aluno → recusado (UI sem botão + API 403 PT).
6. Upload > 50 MB ou `.webm` → 400 PT; aula não fica quebrada.

**Esperado:** SC-001, SC-003, SC-010 localmente.

## B. Demo AWS (passos extras no ciclo 001)

Seguir o runbook 001; inserir os passos VOD abaixo.

### B1. Apply

```powershell
cd infra
terraform apply
```

**Verificar:** output `vod_bucket_name` presente; bucket vazio ou a ser preenchido pelo seed na subida da API.

### B2. Publish API (+ front)

Como no 001 (`publish-api.ps1` → healthy → `publish-frontend.ps1` com `VITE_API_URL` = `api_url`).

**Verificar:** task com `VOD_BACKEND=s3` e `VOD_S3_BUCKET`; seed criou objeto sob `vod/aulas/.../current.mp4`.

### B3. Demo funcional

1. Abrir `frontend_url` → login aluno seed → **Aulas** → reproduzir VOD (≤ 2 min até iniciar; áudio/vídeo ≤ 10 s após “assistir”).
2. Login professor → upload MP4 válido → aluno confirma novo conteúdo.
3. Confirmar: **sem** página Biblioteca; **sem** botão download; visitante sem JWT não reproduz.
4. (Opcional) Aguardar ~15 min sem renovar playback → URL antiga falha; novo GET `/vod/playback` com JWT funciona.

### B4. Destroy

```powershell
cd infra
terraform destroy
```

**Verificar:** bucket VOD removido; sem armazenamento VOD cobrável ocioso; stack `infra/budget/` intacta.

### B5. Novo apply

Após destroy, `apply` + publish de novo → aluno assiste VOD de seed **sem** upload manual (SC-005).

## Critérios rápidos de aceite

| ID | Check |
|----|--------|
| SC-001/006 | Aluno toca seed na ficha da aula |
| SC-002 | Após publish do professor, aluno toca |
| SC-003 | Aluno não escreve; anônimo não lê |
| SC-004 | Destroy zera VOD |
| SC-005 | Re-seed automático |
| SC-012 | Playback ~15 min |

## Fora desta validação

Streaming ao vivo, chat, certificado, OAuth, Redis AWS, NAT, domínio custom, página Biblioteca, download do arquivo, P2 rascunho (salvo fase 5).
