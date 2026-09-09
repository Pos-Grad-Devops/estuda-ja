# Quickstart: 006-escola-ui

Validação visual + jornadas seed após a implementação (frontend). Baseline API 001–005 já deve estar no ar (Compose local ou sessão AWS).

## Pré-requisitos

- Branch git: `feat/aws-speckit`
- API local (`make infra-up` + `make backend`) **ou** demo AWS já publicada
- Frontend: `cd frontend && npm install && npm run dev` (ou rebuild/`publish-frontend` na AWS)
- Contas seed (defaults de dev):
  - admin: `admin@estudaja.com` / `admin123`
  - professor: `professor@estudaja.com` / `professor123`
  - aluno: `aluno@estudaja.com` / `aluno123`

## Build

```bash
cd frontend
npm run build
```

Esperado: build OK (Tailwind + gsap/ogl). Testes Go **não** precisam ser reexecutados por esta feature, salvo regressão acidental de path (não prevista).

## Checklist visual (aceitação)

- [ ] Tom **escola**, não planilha/CRUD tabular como UX principal
- [ ] Shell: Início · Agenda · (admin) Alunos / Usuários
- [ ] Rotas `/cursos/:id` e `/aulas/:id` funcionam
- [ ] **0** aviso na ficha da aula de que player/live/VOD “não fazem parte do MVP”
- [ ] **0** página/rota “Biblioteca”
- [ ] **0** página/rota “Certificados”
- [ ] Live, VOD, chat e certificado usam os **mesmos** primitivos/tokens que Home/Curso/Login (sem tema B)

## Jornada aluno

1. Login como aluno → Home (catálogo).
2. Abrir um curso → `/cursos/:id` (trilha de aulas).
3. Abrir uma aula → `/aulas/:id`.
4. Live: ver status; se `ao_vivo`, player utilizável conforme backend.
5. VOD: assistir se houver gravação; **sem** download MP4; **sem** botões de upload/remover.
6. Chat: enviar/receber na sessão; reabrir ficha → sem histórico prévio.
7. Voltar ao curso: se elegível (seed 005), solicitar/baixar PDF; se não, feedback claro. Sem gestão de elegibilidade.

## Jornada professor

1. Login como professor.
2. Mesmas leituras do aluno (catálogo, fichas, live/VOD/chat).
3. CRUD de aulas (Agenda / CursoPage / modal) disponível; CRUD de cursos/alunos/usuários **não**.
4. Na ficha da aula: agendar/iniciar/encerrar live + ingest OBS quando aplicável; publicar/substituir/remover VOD.
5. Chat utilizável.
6. Na ficha do curso: **sem** controles de gestão de certificado.

## Jornada admin

1. Login como admin.
2. Tudo do professor **mais** CRUD cursos, alunos, usuários.
3. Na ficha do curso: marcar/reabilitar elegibilidade, listar, invalidar (005).
4. Confirmar aluno seed consegue baixar PDF após elegibilidade.

## Visitante

- Sem sessão: só `/login` para áreas protegidas.

## Regressão rápida 001–005

- [ ] Login JWT continua
- [ ] Live stub/IVS conforme ambiente (comportamento API inalterado)
- [ ] VOD playback/gestão conforme RBAC
- [ ] Chat WS na ficha
- [ ] Certificado PDF na ficha do curso
- [ ] Datas na UI em `dd/mm/yyyy` ou `dd/mm/yyyy HH:mm`

## Referências

- Rotas/RBAC UI: [contracts/ui-routes.md](./contracts/ui-routes.md)
- Superfícies: [data-model.md](./data-model.md)
- APIs: contratos em `specs/002-*` … `specs/005-*` (não reescritos aqui)
