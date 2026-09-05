# Research: 002-vod-library

Phase 0 — decisões técnicas para materializar a spec (decisões de produto **não** reabertas).

## 1. Layout de keys no S3 (e espelho local)

**Decision:** um objeto vigente por aula sob prefixo estável:

```text
vod/aulas/{aula_id}/current.mp4
```

Metadados no Postgres guardam `storage_key` (mesmo path relativo no disco local). Em substituição: upload para key temporária `vod/aulas/{aula_id}/upload-{uuid}.mp4` → validação → `Copy`/`Rename` para `current.mp4` **ou** Put direto em `current.mp4` após validar stream em buffer/temp — preferir **escrever temp → validar → substituir current → apagar temp e antigo se houver**. Como só existe um vigente, Put em `current.mp4` após validar em arquivo temporário local (API) e Delete da key anterior se a key mudou; se a key for estável (`current.mp4`), Put sobrescreve e não há órfão de key — ainda assim Delete explícito em remoção de VOD / exclusão de aula.

**Rationale:** key previsível simplifica IAM prefix (`vod/*`), seed e debug; um objeto = FR-004.

**Alternatives considered:**
- Keys só com UUID sem `current` — mais órfãos se o delete falhar; rejeitado.
- Prefix por curso — desnecessário (associação é à aula).

## 2. Upload: multipart pela API vs presigned PUT

**Decision:** **multipart `multipart/form-data` pela API** (campo arquivo + opcionalmente metadados P2). Limite de body **50 MB** (Fiber/`BodyLimit` + checagem `size`). Stream: gravar em temp → validar → `storage.Put` → atualizar metadados → apagar objeto/temp anterior.

**Rationale:** unifica local e AWS; validação de magic bytes/MIME no servidor antes de marcar publicado; Fargate 512 MiB tolera um upload de 50 MB por vez na demo; YAGNI (sem CORS S3, sem política de PostObject).

**Alternatives considered:**
- Presigned PUT browser→S3 — melhor em escala; exige confirm endpoint, CORS no bucket, validação pós-fato; complexidade extra para demo acadêmica.
- Presigned só na AWS + multipart local — dois caminhos; rejeitado no P1.

## 3. Playback: CloudFront signed URL vs S3 presigned

**Decision:** **S3 GetObject presigned URL com TTL ~15 minutos** (`VOD_PLAYBACK_TTL=15m`). API `GET .../vod/playback` (JWT + papel leitor) gera a URL. **Sem** distribution CloudFront dedicada a mídia no P1. Local: URL da API que streama o arquivo com token de curta duração **ou** path autenticado equivalente (`/api/v1/aulas/:id/vod/content?token=...` com JWT/exp claim ~15 min) — preferir **mesmo contrato**: response JSON `{ "playback_url", "expires_at" }` onde local `playback_url` aponta para endpoint de content autenticado/assinado.

**Rationale:** constitution I — CF de mídia extra = recurso contínuo na sessão sem ganho obrigatório para clipes curtos e dezenas de espectadores; S3 presigned atende “sem endereço interno permanente” e SC-012; implementável com AWS SDK v2.

**Alternatives considered:**
- CloudFront + OAC + signed URLs (key group) — melhor CDN/cache; custo/ops e chave RSA a mais; adiar se demo mostrar necessidade.
- Proxy permanente de bytes pela API sem URL temp — simples localmente, mas carrega a task Fargate em toda reprodução na AWS; usar só como **modo local** se presign S3 não se aplica.

## 4. Validação MP4 / tamanho

**Decision:**
1. Rejeitar se `Content-Length` / tamanho lido **> 50 MiB** (52 428 800 bytes).
2. Content-Type declarado deve ser `video/mp4` (ou vazio + inferência pela extensão `.mp4`).
3. Magic: primeiros bytes compatíveis com ISO BMFF (`ftyp` em offset típico).
4. **Não** exigir ffprobe/codecs no servidor no P1 (sem binário extra na imagem). Documentar que o seed e uploads de demo devem ser H.264+AAC; rejeição “profunda” de codec fica best-effort / fora se magic+MIME passarem mas o browser falhar (edge: mensagem PT na UI).

**Rationale:** FR-015/FR-019 sem pipeline de transcode; magic+MIME cobre o caso acadêmico; imagem Fargate permanece leve.

**Alternatives considered:**
- ffprobe no container — rejeitado (imagem maior, YAGNI).
- Aceitar qualquer vídeo “que o browser toque” — rejeitado na clarificação (só MP4).

## 5. Delete imediato do objeto antigo

**Decision:** em replace bem-sucedido, remoção de VOD, ou cascade ao deletar aula/curso: chamar `storage.Delete(storage_key)` **antes ou logo após** commit dos metadados (ordem: preferir atualizar/remover metadados + Delete; se Delete falhar, logar erro em PT nos logs e ainda assim não expor URL antiga no GET — best effort alinhado a “apagar na hora”, com destroy como rede final). Em replace com a **mesma** key `current.mp4`, Put sobrescreve; Delete só na remoção total.

**Rationale:** clarificação Q7; evita órfãos cobráveis na sessão.

**Alternatives considered:**
- Só limpar no destroy — rejeitado na clarificação.
- GC assíncrono — over-engineering.

## 6. Seed asset no repositório

**Decision:** incluir um MP4 **curto e pequeno** (alvo ≤ 2 MB, hard cap 50 MB) em:

```text
backend/assets/vod/demo-aula.mp4
```

Migration gormigrate `003_seed_vod_demo` (idempotente): se a aula seed do `002_seed_demo` existir e não houver VOD publicado, copia o asset para o storage e cria metadados `publicado`. Dockerfile da API deve `COPY` o diretório `assets/vod`. Compose: volume opcional ` ./data/vod:/data/vod` + `VOD_LOCAL_DIR=/data/vod`.

**Rationale:** SC-005 sem upload manual; asset versionado e reproduzível no CI/local/AWS.

**Alternatives considered:**
- Baixar clipe em runtime (URL externa) — frágil offline/CI.
- Gerar MP4 sintético no seed — possível fallback se binário for indesejado; preferir arquivo real mínimo no repo.

## 7. Extensão Terraform enxuta do `infra/` existente

**Decision:**
- Novo `infra/s3_vod.tf`: bucket `${project_name}-vod-${account_id}`, privado, SSE AES256, **sem** versioning obrigatório (ou Suspended), `force_destroy = true`, block public access. **Sem** website, **sem** OAC/CF no P1.
- `iam.tf`: policy na **task role** `ecs_task`: `s3:PutObject`, `GetObject`, `DeleteObject` em `arn:.../bucket/vod/*` (+ `ListBucket` no bucket com prefix condition se necessário).
- `ecs.tf`: env plain `VOD_BACKEND=s3`, `VOD_S3_BUCKET=<name>`, `VOD_PLAYBACK_TTL=15m`, `AWS_REGION=us-east-1` (SDK default na task).
- `outputs.tf`: `vod_bucket_name`.
- Não misturar com bucket `frontend`.
- Budget stack em `infra/budget/` **intocada**.

**Rationale:** mínimo alinhado a constitution VIII e US3; destroy remove bucket+objetos.

**Alternatives considered:**
- Reusar bucket frontend com prefix `vod/` — acopla sync SPA e mídia; permissões CF front dariam GetObject de vídeo se mal configurado; rejeitado.
- Segundo CloudFront — ver §3.

## 8. Custo por sessão (ordem de grandeza)

**Decision (estimativa operacional, não billing exato):**

| Item | Durante sessão curta (horas) | Após destroy |
|------|------------------------------|--------------|
| S3 VOD (poucos MB + requests) | centavos / ~0 material | **0** (bucket gone) |
| CF mídia | **não provisionado** | — |
| ECS/RDS/ALB/CF front+API | igual ao 001 | 0 (exceto resíduos já documentados no 001, ex. ECR vazio se force_delete) |
| Transfer S3→browser (presigned) | data out S3; clipes curtos = baixo | 0 |

**Rationale:** clarifica SC-004; operador não precisa de CF de mídia para manter custo ~0 fora da sessão.

**Alternatives considered:** CF mídia “por performance” — adiado.

## 9. Superfície de domínio na API

**Decision:** recursos aninhados em `/api/v1/aulas/:id/vod` (ver [contracts/vod-api.md](./contracts/vod-api.md)); handler/repository `vod_*`; não redesenhar CRUD de aulas além de cascade delete de VOD.

**Rationale:** FR-004/FR-010; um arquivo por domínio (AGENTS.md).

## 10. UI

**Decision:** estender `frontend/src/pages/AulasPage.tsx` (lista/ficha já usada): bloco “Gravação” com estado vazio / player / upload. Sem novas rotas em `App.tsx`. Permissão: `canManageAulas`.

**Rationale:** clarificação Q3; YAGNI.
