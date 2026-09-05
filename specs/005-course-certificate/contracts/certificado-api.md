# Contract: API Certificado

Superfície HTTP nova sob `/api/v1/`. Datas em JSON: `dd/mm/yyyy` ou `dd/mm/yyyy HH:mm`. Erros em português. Auth: Bearer JWT (middleware existente).

Identidade = **User** com papel `aluno`. Gestão = **somente admin**. Professor sem gestão. Baseline 001–004 intacta.

## Autorização

| Ação | admin | professor | aluno | anônimo |
|------|-------|-----------|-------|---------|
| Marcar / reabilitar elegibilidade | ✓ | ✗ | ✗ | ✗ |
| Listar/consultar emissões do curso | ✓ | ✗ | ✗ | ✗ |
| Invalidar certificado | ✓ | ✗ | ✗ | ✗ |
| Ver status do **próprio** certificado/elegibilidade | ✓* | ✗** | ✓ | ✗ |
| Baixar / emitir PDF (próprio) | ✗*** | ✗ | ✓ | ✗ |

\* Admin consulta via listagem ou `GET .../certificado?user_id=`.  
\*\* Professor não precisa de status de certificado no P1 (sem controles).  
\*\*\* Admin **não** “emite por” o aluno no P1; emissão = primeira solicitação do aluno. Admin pode consultar, não substitui o download do aluno.

Escrita de gestão: `RequireRoles(admin)`. Download/status aluno: autenticado + papéis com leitura de cursos, restringindo o PDF ao **próprio** `user_id`.

## Endpoints P1

### GET `/api/v1/cursos/:id/certificado`

Status no contexto do curso para o usuário autenticado (aluno = próprio). Admin MAY passar `?user_id=` para consultar um aluno.

**200** exemplo (aluno elegível sem emissão):

```json
{
  "curso_id": 1,
  "user_id": 3,
  "elegivel": true,
  "certificado": null
}
```

**200** com ativo:

```json
{
  "curso_id": 1,
  "user_id": 3,
  "elegivel": true,
  "certificado": {
    "id": 10,
    "status": "valido",
    "emitido_em": "05/09/2026 17:30",
    "aluno_nome": "Aluno Demo",
    "curso_titulo": "Aula Magna — Direito Constitucional"
  }
}
```

**200** após invalidação sem reabilitar:

```json
{
  "curso_id": 1,
  "user_id": 3,
  "elegivel": false,
  "certificado": {
    "id": 10,
    "status": "invalidado",
    "emitido_em": "05/09/2026 17:30",
    "aluno_nome": "Aluno Demo",
    "curso_titulo": "Aula Magna — Direito Constitucional"
  }
}
```

(Quando há vários invalidado + um valido, `certificado` = o ativo; se só invalidado(s), retornar o mais recente invalidado para feedback — ou `null` + flag; preferir **mais recente** do par.)

**401/403** — sem sessão / sem permissão.  
**404** — curso inexistente.

### GET `/api/v1/cursos/:id/certificado/pdf`

Download do PDF do **próprio** aluno autenticado (`Content-Type: application/pdf`, `Content-Disposition: attachment; filename="certificado-curso-{id}.pdf"`).

Comportamento:

1. Sem elegibilidade e sem ativo → **403** PT (ex.: “Você não está elegível para o certificado deste curso”).
2. Elegível, sem ativo → criar `Certificado` `valido` + gerar PDF; se PDF falhar → rollback, **500** PT (ex.: “Não foi possível gerar o certificado”).
3. Ativo `valido` → regenerar PDF a partir do registro (sem novo id).
4. Só invalidado / sem elegibilidade → **403** PT.

**401** — anônimo.  
**403** — professor tentando baixar “como aluno”; aluno não elegível; certificado invalidado sem reabilitação.  
**404** — curso inexistente.

NÃO exigir query `user_id` (sempre o subject do JWT). NÃO retornar URL S3.

### PUT `/api/v1/cursos/:id/certificados/elegibilidade`

**Só admin.** Body:

```json
{ "user_id": 3 }
```

Marca ou **reabilita** elegibilidade. **Não** cria certificado nem PDF.

**200** exemplo:

```json
{
  "curso_id": 1,
  "user_id": 3,
  "elegivel": true
}
```

**400** — body inválido; usuário não é `aluno` (PT).  
**403** — não admin.  
**404** — curso ou usuário inexistente.

### GET `/api/v1/cursos/:id/certificados`

**Só admin.** Lista mínima para a demo: elegibilidades + certificados do curso.

**200** exemplo:

```json
{
  "curso_id": 1,
  "elegibilidades": [
    { "user_id": 3, "user_nome": "Aluno Demo", "created_at": "05/09/2026 16:00" }
  ],
  "certificados": [
    {
      "id": 10,
      "user_id": 3,
      "status": "valido",
      "emitido_em": "05/09/2026 17:30",
      "aluno_nome": "Aluno Demo",
      "curso_titulo": "Aula Magna — Direito Constitucional"
    }
  ]
}
```

**403/404** — como acima.

### POST `/api/v1/cursos/:id/certificados/:certId/invalidar`

**Só admin.** Invalida o certificado `valido` e **remove** a elegibilidade do par `user_id`–`curso_id`.

**200** exemplo:

```json
{
  "id": 10,
  "status": "invalidado",
  "curso_id": 1,
  "user_id": 3
}
```

**400/409** — já invalidado ou id não pertence ao curso (PT).  
**403** — não admin.  
**404** — certificado ou curso inexistente.

Após isto, `GET .../certificado/pdf` do aluno falha até novo `PUT` elegibilidade; depois a próxima solicitação cria **novo** certificado.

## Fora deste contrato

- Página/rota “Certificados”
- Emissão forçada pelo admin (PDF)
- Gestão por professor
- Lista agregada de certificados do aluno (P2)
- Template visual rico / blockchain / URL pública permanente
- Endpoints sob `/alunos` (CRUD Alunos)
