# Data Model: 005-course-certificate

Extensão mínima sobre o domínio existente. Identidade = **User** (`role = aluno`). **Não** vincular ao CRUD `Alunos`. Baseline 001–004 intacta.

## Entidades existentes (inalteradas no essencial)

### User
Campos atuais (`id`, `nome`, `email`, `password_hash`, `role`, timestamps).  
Papel relevante: `aluno` — sujeito da elegibilidade e do certificado. Nome no PDF = `users.nome` (snapshot na emissão).

### Curso
Campos atuais (`id`, `titulo`, `descricao`, timestamps).  
Contexto da UI e pai lógico do certificado. Título no PDF = snapshot na emissão.

### Aluno (CRUD)
**Fora** do fluxo P1 de certificado. Sem FK obrigatória.

---

## Entidade nova: CertificadoElegibilidade

Flag de “curso concluído” por par usuário-aluno–curso. **Só admin** cria/recria. Existência da linha = elegível. **Não** cria certificado.

| Campo | Tipo | Obrigatório | Notas |
|-------|------|-------------|-------|
| `id` | uint / PK | sim | Surrogate |
| `user_id` | uint | sim | FK → `users`; MUST ter `role = aluno` na concessão |
| `curso_id` | uint | sim | FK → `cursos` |
| `created_at` | DateTime BR | sim | `timeutil.DateTime` |
| `updated_at` | DateTime BR | sim | |

**Tabela:** `certificado_elegibilidades`  
**UNIQUE:** `(user_id, curso_id)`

### Relacionamentos

```text
User (aluno) 1 ─── * CertificadoElegibilidade * ─── 1 Curso
```

### Regras de validação

1. Admin só concede a usuário com `role = aluno`; caso contrário erro PT (ex.: “Usuário não é aluno”).
2. Curso e usuário MUST existir; senão 404 PT.
3. Marcar elegibilidade **idempotente**: se já existir, 200 sem duplicar; **não** emite PDF.
4. Na **invalidação** de certificado: **apagar** a elegibilidade do par (bloqueia nova emissão).
5. **Reabilitar** = novo `PUT` que recria a linha (após invalidação ou primeira vez).

### Transições

```text
(nenhum) --admin PUT elegibilidade--> elegível
elegível --admin invalidar certificado--> (nenhum)  [linha removida]
(nenhum) --admin PUT reabilitar--> elegível
elegível --destroy sessão / delete curso--> (nenhum)
```

---

## Entidade nova: Certificado

Comprovante criado na **primeira** solicitação bem-sucedida do aluno elegível sem ativo. Artefato PDF = **on-the-fly** a partir dos campos (sem storage de arquivo).

| Campo | Tipo | Obrigatório | Notas |
|-------|------|-------------|-------|
| `id` | uint / PK | sim | Surrogate (`certId` nas rotas) |
| `user_id` | uint | sim | FK → `users` (aluno) |
| `curso_id` | uint | sim | FK → `cursos` |
| `status` | string | sim | `valido` \| `invalidado` |
| `emitido_em` | DateTime BR | sim | Data/hora da primeira emissão bem-sucedida |
| `aluno_nome` | string | sim | Snapshot do nome na emissão (PDF estável) |
| `curso_titulo` | string | sim | Snapshot do título do curso na emissão |
| `created_at` | DateTime BR | sim | |
| `updated_at` | DateTime BR | sim | |

**Tabela:** `certificados`

**Índice / invariante:** no máximo **um** registro com `status = valido` por `(user_id, curso_id)`. Podem existir vários `invalidado` (histórico da sessão).

Constantes sugeridas:

```text
CertificadoStatusValido     = "valido"
CertificadoStatusInvalidado = "invalidado"
```

### Relacionamentos

```text
User (aluno) 1 ─── * Certificado * ─── 1 Curso
CertificadoElegibilidade ──(desbloqueia)──> emissão lazy de Certificado
```

Elegibilidade **não** é FK do certificado; é pré-condição de negócio.

### Regras de validação

1. Criar `valido` só se: autenticado como o próprio `user_id`, papel aluno, elegibilidade existe, **não** há outro `valido` do par, curso existe.
2. Se já existe `valido`: download reutiliza (não cria segundo ativo).
3. Se só há `invalidado` e **sem** elegibilidade: recusar emissão/download (PT).
4. Se PDF falhar na primeira emissão: **não** persistir como `valido` (rollback).
5. Invalidar: só admin; só se status atual `valido`; depois remove elegibilidade do par.
6. Snapshots `aluno_nome` / `curso_titulo` preenchidos na emissão; PDF usa snapshots (não relê nomes mutáveis depois).

### Transições de estado

```text
(nenhum) --1ª GET pdf (elegível) OK--> valido
valido   --admin invalidar--> invalidado  (+ remove elegibilidade)
invalidado --(nunca)--> valido          [imutável; novo certificado é outro id]
(após reabilitar elegibilidade)
(nenhum ativo) --nova GET pdf OK--> valido (novo id; invalidado permanece)
```

### Artefato PDF (não é tabela)

| Aspecto | Valor |
|---------|--------|
| Formato | `application/pdf` |
| Conteúdo mínimo | título fixo + `aluno_nome` + `curso_titulo` + `emitido_em` (BR) |
| Persistência | **Nenhuma** — gerado on-the-fly (research §2) |
| Lifecycle sessão | Some com o registro no destroy do RDS |

---

## Seed

- Migration `004_seed_certificado_demo` (nome sugerido; idempotente).
- Pré-marca elegibilidade: aluno seed (`DEMO_ALUNO_*`) × curso demo (`Aula Magna — Direito Constitucional`).
- **Não** cria linha em `certificados`.

---

## Cascade / destroy

- `DELETE` curso: remover elegibilidades e certificados daquele `curso_id` (integridade).
- `DELETE` user: remover elegibilidades/certificados daquele `user_id` (se o fluxo de users permitir delete).
- Demo AWS: `terraform destroy` destrói RDS → tabelas somem; **sem** objetos S3 de certificado.
