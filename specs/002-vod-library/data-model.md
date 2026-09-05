# Data Model: 002-vod-library

Extensão mínima sobre o domínio 001. **Não** redesenhar `Curso` / `Aula` além do vínculo e cascade.

## Entidades existentes (inalteradas no essencial)

### Curso
Campos atuais (`id`, `titulo`, `descricao`, timestamps). Continua pai de `Aula`. Sem biblioteca própria de VOD.

### Aula
Campos atuais (`id`, `curso_id`, `titulo`, `descricao`, `agendada_em`, `status`, timestamps).  
**Regra nova:** no máximo **um** `AulaVod` vigente associado. Ao **DELETE** aula (ou curso com suas aulas), remover metadados VOD + objeto no storage.

## Entidade nova: AulaVod

Representa o único conteúdo gravado vigente da aula.

| Campo | Tipo | Obrigatório | Notas |
|-------|------|-------------|-------|
| `id` | uint / PK | sim | Surrogate |
| `aula_id` | uint | sim | **UNIQUE** — FR-004 |
| `storage_key` | string | sim | Ex.: `vod/aulas/{aula_id}/current.mp4` |
| `content_type` | string | sim | Sempre `video/mp4` no P1 |
| `size_bytes` | int64 | sim | ≤ 52 428 800 |
| `status` | string | sim | P1: sempre `publicado`. P2: `rascunho` \| `publicado` |
| `titulo` | string | não | P2 — se vazio, UI usa título da aula |
| `duracao_segundos` | int | não | P2 — opcional |
| `created_at` | DateTime BR | sim | `timeutil.DateTime` |
| `updated_at` | DateTime BR | sim | |

### Relacionamentos

```text
Curso 1 ─── * Aula 1 ─── 0..1 AulaVod
```

- Leitura de VOD **publicado**: quem já lê aulas (admin, professor, aluno).
- Escrita de VOD: admin + professor (qualquer aula; sem dono de curso).

### Regras de validação

1. Não criar segundo `AulaVod` para a mesma `aula_id` — replace atualiza o registro existente.
2. Upload: rejeitar se tamanho > 50 MB ou não MP4 (magic/`video/mp4`) — mensagens em português.
3. Aluno **nunca** vê/reproduz item com `status = rascunho` (P2); P1 só persiste `publicado`.
4. `storage_key` MUST existir no backend de storage enquanto o registro estiver assistível; após delete, key removida.

### Transições de estado

**P1**

```text
(nenhum) --upload válido--> publicado
publicado --replace válido--> publicado (objeto antigo apagado se key distinta / sobrescrita)
publicado --DELETE vod / aula--> (nenhum) + Delete storage
```

**P2 (opcional)**

```text
(nenhum) --upload--> rascunho ou publicado
rascunho --publish--> publicado
publicado --unpublish--> rascunho  (aluno deixa de reproduzir)
* --DELETE--> (nenhum) + Delete storage
```

## Objeto de storage (não é tabela)

| Aspecto | Valor |
|---------|--------|
| Key canônica | `vod/aulas/{aula_id}/current.mp4` |
| Backend AWS | Bucket efêmero (ver terraform-vod) |
| Backend local | Arquivos sob `VOD_LOCAL_DIR` com o mesmo path relativo |
| Lifecycle sessão | Apagado no replace/remoção; bucket destruído no `terraform destroy` |

## Seed

- Usa a **aula de demo** criada em `002_seed_demo`.
- Cria exatamente um `AulaVod` `publicado` + copia `backend/assets/vod/demo-aula.mp4` para o storage.
- Idempotente: se já existir VOD para essa aula, não duplicar.
