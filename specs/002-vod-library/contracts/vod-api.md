# Contract: API VOD

Superfície HTTP nova sob `/api/v1/`. Datas em JSON: `dd/mm/yyyy` ou `dd/mm/yyyy HH:mm`. Erros em português. Auth: Bearer JWT (middleware existente).

## Autorização

| Ação | admin | professor | aluno | anônimo |
|------|-------|-----------|-------|---------|
| Ler metadados / playback (publicado) | ✓ | ✓ | ✓ | ✗ |
| Upload / replace / delete | ✓ | ✓ | ✗ | ✗ |
| Ler rascunho (P2) | ✓ | ✓ | ✗ | ✗ |

Escrita: mesmo recorte de `RequireRoles(admin, professor)` das aulas. Qualquer professor pode gerir VOD de qualquer aula.

## Endpoints P1

### GET `/api/v1/aulas/:id/vod`

Metadados do VOD vigente **publicado** (aluno não recebe rascunho).

**200** exemplo:

```json
{
  "aula_id": 1,
  "status": "publicado",
  "content_type": "video/mp4",
  "size_bytes": 1048576,
  "updated_at": "04/09/2026 22:00"
}
```

**404** — aula inexistente, sem VOD, ou só rascunho (para aluno): mensagem clara em PT (ex.: “Gravação não encontrada”).

### GET `/api/v1/aulas/:id/vod/playback`

Emite acesso temporário de reprodução (~15 min).

**200** exemplo:

```json
{
  "playback_url": "https://...",
  "expires_at": "04/09/2026 22:15",
  "expires_in_seconds": 900
}
```

- AWS: URL presigned S3 GET.
- Local: URL do endpoint de content assinado/autenticado equivalente (mesmo JSON).

**401/403** — sem sessão ou sem permissão de leitura.  
**404** — sem VOD publicado.

NÃO retornar URL permanente do bucket. NÃO oferecer endpoint de “download attachment” com `Content-Disposition: attachment` como feature de produto.

### PUT `/api/v1/aulas/:id/vod`

Upload/replace (`multipart/form-data`, campo de arquivo obrigatório, ex. `file`).

- Valida tamanho ≤ 50 MB e MP4.
- Substitui vigente; apaga objeto antigo conforme research.
- P1: status resultante = `publicado`.

**200/201** — metadados do VOD (como GET).  
**400** — arquivo inválido/grande demais (PT).  
**403** — aluno ou não autenticado com papel adequado.  
**404** — aula inexistente.

### DELETE `/api/v1/aulas/:id/vod`

Remove metadados + objeto no storage.

**204** ou **200** com mensagem PT.  
**403/404** — como acima.

## Efeito colateral no CRUD de aulas

`DELETE /api/v1/aulas/:id` (já existente) MUST remover VOD associado (metadados + storage) antes/junto da exclusão da aula.

## Fora deste contrato

- Página/rota frontend “Biblioteca”
- Download explícito do MP4
- Streaming ao vivo, chat, OAuth
- Transcodificação / múltiplas qualities
