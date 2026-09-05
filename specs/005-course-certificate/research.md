# Research: 005-course-certificate

Phase 0 — decisões técnicas para materializar a spec (decisões de produto **não** reabertas). Baseline 001–004 intacta.

## 1. Biblioteca PDF em Go

**Decision:** gerar o PDF com **`github.com/go-pdf/fpdf`** (fork mantido do clássico gofpdf) + **fonte TrueType embutida** (ex. DejaVu Sans sob `backend/assets/certs/` ou equivalente) para acentos em português (`ç`, `ã`, `é`).

Layout P1: **uma página A4** com título fixo (“Certificado de Conclusão”), nome do usuário-aluno, título do curso e data de emissão em `dd/mm/yyyy` (ou `dd/mm/yyyy HH:mm`). Sem HTML→PDF, sem Chromium, sem wkhtmltopdf.

**Rationale:** pure Go, sem CGO, adequado a Fargate/CI; maturidade alta; UTF-8 via TTF resolve nomes BR; YAGNI (template rico = P2).

**Alternatives considered:**
- `github.com/signintech/gopdf` — também puro Go + TTF; aceitável, menos “padrão FPDF” na comunidade.
- `github.com/gpdf-dev/gpdf` — moderno e rápido; rejeitado no P1 por ser superfície nova demais para um PDF de 3 campos (risco desnecessário).
- `jung-kurt/gofpdf` (original) — arquivado; preferir o fork `go-pdf/fpdf`.
- HTML/Chrome headless ou wkhtmltopdf — binário extra na imagem ECS; rejeitado (custo/ops).
- Artefato HTML-only — fora da clarificação (PDF obrigatório).

## 2. Persistência do artefato: on-the-fly vs S3 da sessão

**Decision:** **geração on-the-fly** a cada download. O banco guarda só o **registro** do certificado (aluno, curso, status, data). O PDF é montado a partir desses campos (+ joins `users.nome` / `cursos.titulo` capturados ou lidos na emissão). **Sem** bucket S3 novo, **sem** prefixo no bucket VOD, **sem** disco local de PDFs.

Na **primeira** solicitação elegível sem ativo: criar registro `valido` **e** gerar/streamar o PDF na mesma operação; se a geração falhar, **não** deixar registro válido (transação / rollback). Downloads seguintes: ler registro ativo e regenerar o mesmo conteúdo (mesmos campos persistidos).

**Rationale (FR-006 + constitution I):** PDF é determinístico e minúsculo; S3/IAM/Terraform extras não pagam o ganho; `destroy` da sessão já remove o RDS — zero artefato órfão; local e AWS compartilham o mesmo caminho; Fargate 256/512 aguenta gerar 1 página sob demanda na demo.

**Alternatives considered:**
- S3 efêmero (bucket próprio ou prefix `certs/` no VOD) — alinhado ao padrão 002, mas custo/ops e wiring Terraform sem benefício para PDF regenerável; rejeitado no P1.
- Gravar PDF em volume Compose / EFS — rejeitado (EFS fora do baseline; volume local não espelha AWS).
- Presigned URL de objeto — desnecessário sem objeto.

**Implicação infra:** **nenhuma** mudança Terraform obrigatória para certificado no P1. Budget `infra/budget/` intocado. Destroy documentado = “registros somem com o RDS; não há objetos S3 de certificado”.

## 3. Superfície de endpoints (API)

**Decision:** recursos aninhados em **`/api/v1/cursos/:id/...`** (contexto do curso; um domínio `certificado_*`). Auth Bearer JWT existente. Detalhe: [contracts/certificado-api.md](./contracts/certificado-api.md).

| Método | Path | Quem | Efeito |
|--------|------|------|--------|
| `GET` | `/cursos/:id/certificado` | aluno (próprio); admin pode consultar por `?user_id=` | Status: elegível? ativo? invalidado? |
| `GET` | `/cursos/:id/certificado/pdf` | aluno (próprio) | Download PDF; **lazy emit** se elegível sem ativo; reutiliza ativo |
| `PUT` | `/cursos/:id/certificados/elegibilidade` | **só admin** | Marca/reabilita elegibilidade `{ "user_id": N }` (papel `aluno`) — **não** emite PDF |
| `GET` | `/cursos/:id/certificados` | **só admin** | Consulta elegibilidades + certificados do curso (mínimo demo) |
| `POST` | `/cursos/:id/certificados/:certId/invalidar` | **só admin** | Status → `invalidado` **e** remove elegibilidade do par (bloqueia nova emissão até reabilitar) |

**Rationale:** espelha aninhamento VOD/live/chat sob o recurso pai; UI no contexto do curso; professor sem rotas de gestão.

**Alternatives considered:**
- `/api/v1/certificados` flat + página dedicada — rejeitado (clarificação UI).
- Emitir PDF no `PUT` de elegibilidade — rejeitado (lazy na solicitação do aluno).
- Soft-delete de elegibilidade com flag vs delete da linha — ver §4; presença da linha = elegível.

## 4. Nomes de tabelas / models

**Decision:**

| Model Go | Tabela GORM | Papel |
|----------|-------------|--------|
| `CertificadoElegibilidade` | `certificado_elegibilidades` | Par `(user_id, curso_id)` **único**; existência = elegível |
| `Certificado` | `certificados` | Emissão; status `valido` \| `invalidado`; no máximo **um** `valido` por par |

Campos canônicos (detalhe em [data-model.md](./data-model.md)):
- Elegibilidade: `user_id` → `users`, `curso_id` → `cursos`, timestamps.
- Certificado: `user_id`, `curso_id`, `status`, `emitido_em`, snapshots `aluno_nome` / `curso_titulo` (congelam o PDF), timestamps.

`user_id` = **User** com `role = aluno` (login). **Não** FK para tabela `alunos` (CRUD Alunos).

**Rationale:** plural snake_case alinhado ao GORM default; nomes explícitos evitam colisão com “Aluno” CRUD; snapshots garantem PDF estável se o nome do curso mudar depois.

**Alternatives considered:**
- Uma tabela só com flags `elegivel` + `certificado_*` — mistura estados; rejeitado.
- FK para `alunos` — rejeitado na clarificação.
- Soft-flag `ativa` na elegibilidade — possível; P1 prefere **delete da linha** na invalidação e **recreate** no reabilitar (menos campos).

## 5. Invalidação ↔ elegibilidade

**Decision:** `POST .../invalidar` é atômico:
1. Certificado alvo (deve estar `valido`) → `invalidado`.
2. Remover a linha de `certificado_elegibilidades` daquele par.
3. Download / nova emissão recusados até admin `PUT` elegibilidade de novo.
4. Após reabilitar, próxima `GET .../pdf` do aluno cria **novo** `Certificado` `valido` (o invalidado permanece no histórico da sessão).

**Rationale:** FR-007a; evita “invalidado mas ainda elegível” ambíguo.

## 6. UI (contexto do curso)

**Decision:** estender `frontend/src/pages/CursosPage.tsx` (lista/ficha de cursos) com painel mínimo **por curso**:
- **Aluno:** status + botão solicitar/baixar PDF.
- **Admin:** marcar/reabilitar elegibilidade (user_id / select alunos), consultar emissão, invalidar.
- Professor: **sem** controles de certificado.
- Sem rota `/certificados`; sem página dedicada.

Helper RBAC: `canManageCertificados(role) => role === 'admin'` em `auth/auth.ts`.

**Rationale:** clarificação Session 2026-09-05; espelha padrão “bloco na ficha” de VOD/live/chat (lá em aulas; aqui no curso).

## 7. Seed

**Decision:** migration gormigrate `004_seed_certificado_demo` (idempotente):
- Localiza usuário seed `aluno` + curso demo (`002_seed_demo`).
- Insere **só** `CertificadoElegibilidade` se ausente.
- **Não** cria `Certificado` nem PDF.

**Rationale:** clarificação seed; demo prova o fluxo lazy.

## 8. Env / infra

**Decision:** **nenhuma** variável de ambiente obrigatória nova no P1. Sem `CERT_S3_*`. Opcional futuro (não P1): versão de template. Ver [contracts/certificado-env.md](./contracts/certificado-env.md).

**Rationale:** on-the-fly; baseline 001–004 suficiente.

## 9. CI e testes

**Decision:** testes Go em `handler/certificado_handler_test.go` (SQLite `:memory:`): RBAC (admin vs professor vs aluno), elegibilidade sem emitir, lazy emit + reuso, invalidar + bloqueio + reabilitar + novo certificado, erros PT, PDF `Content-Type: application/pdf`. CI existente sem AWS.

**Rationale:** constitution V; SC-006.

## 10. Pico / live

**Decision:** certificado **fora** do caminho T−0; sem warm-up extra; sem cold start na abertura da aula. Runbook 003/004 inalterado quanto a IVS/chat.

**Rationale:** constitution II + FR edge case da spec.
