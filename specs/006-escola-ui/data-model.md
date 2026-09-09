# Data model: 006-escola-ui (superfícies UI)

**Nota:** Esta feature **não** cria entidades de backend. O mapa abaixo descreve superfícies de interface e onde vivem live / VOD / chat / certificado. Domínios de dados continuam definidos em 001–005.

## Superfícies

| Superfície | Rota | Papel típico | Conteúdo principal |
|------------|------|--------------|--------------------|
| Login | `/login` | visitante | Autenticação |
| Home / catálogo | `/` | autenticado | Lista de cursos (CourseCard); CRUD curso se admin |
| Ficha do curso | `/cursos/:id` | autenticado | Meta do curso; trilha de aulas (LessonRow); certificado |
| Agenda | `/aulas` | autenticado | Lista/filtro de aulas; CRUD aula se admin/professor |
| Ficha da aula | `/aulas/:id` | autenticado | Meta da aula; **live**; **VOD**; **chat** |
| Alunos | `/alunos` | admin | CRUD alunos |
| Usuários | `/usuarios` | admin | CRUD usuários |

Redirect: `/cursos` → `/`.

## Onde vivem os blocos 002–005

| Capacidade | Spec origem | Superfície UI | Controles por papel |
|------------|-------------|----------------|---------------------|
| Transmissão ao vivo | 003 | **Ficha da aula** | Leitura: todos autenticados com acesso à aula. Gestão (agendar/cancelar/iniciar/encerrar) + ingest OBS: admin, professor |
| Gravação VOD | 002 | **Ficha da aula** | Assistir: autenticados. Publicar/substituir/remover: admin, professor. Sem download MP4. Sem página Biblioteca |
| Chat da aula | 004 | **Ficha da aula** | Enviar/receber: admin, professor, aluno. Sem histórico ao reabrir. Sem moderação |
| Certificado | 005 | **Ficha do curso** | Aluno: próprio status/PDF se elegível. Gestão elegibilidade/invalidar: **só admin**. Professor: sem gestão. Sem página Certificados |

## Conceitos de apresentação (não são tabelas novas)

- **Shell**: navegação Início · Agenda · (admin) Alunos / Usuários.
- **Empty / loading / erro**: mesmos padrões do design system da escola em todas as superfícies, inclusive blocos 002–005.
- **Status de transmissão (UI)**: derivado de `AulaLive` / API live (`inativa` \| `agendada` \| `ao_vivo` \| `encerrada`), distinto do campo `Aula.status` de CRUD da aula.
- **Status de aula (CRUD)**: `agendada` \| `ao_vivo` \| `encerrada` — metadado da entidade Aula; Agenda/trilha podem exibir; **não** substitui o estado live na ficha de transmissão.

## Relacionamentos de navegação

```text
Login → Home
Home → CursoPage (/cursos/:id)
CursoPage (trilha) → AulaPage (/aulas/:id)
Agenda → AulaPage
CursoPage ← certificado (painel no próprio curso)
AulaPage ← live + VOD + chat (painéis na própria aula)
```

## Fora deste modelo

- Entidades/tabelas novas, migrations, seeds, contratos REST novos.
- Página Biblioteca; página Certificados; portal separado de live.
