# Feature Specification: UI da escola (visual P1)

**Feature Branch**: `006-escola-ui` (trabalho permanece em `feat/aws-speckit`)

**Created**: 2026-09-09

**Status**: Draft

**Input**: User description: "UI da escola (visual P1) sobre a baseline 001–005 já entregue. Trazer só o visual/navegação da feat/frontend-escola para o frontend desta branch, sem mudar API, RBAC, backend, Terraform, publish, seeds nem contratos 001–005. Vestir com o mesmo visual live, VOD, chat e certificado. Não mergear branches inteiras; não apagar comportamento 002–005."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Navegação e catálogo no visual de escola (Priority: P1)

Um usuário autenticado deixa de ver a interface no tom de planilha/CRUD genérico e passa a navegar um **catálogo de cursos** (home), **ficha de curso** com trilha de aulas, **agenda de aulas** e **ficha de aula**, além de login e (conforme papel) gestão de alunos e usuários — tudo no mesmo visual de “escola”. As rotas de ficha de curso e de aula são o destino natural da jornada.

**Why this priority**: Sem o visual e a navegação de escola, a feature não entrega o valor pedido; é a base sobre a qual live, VOD, chat e certificado são recolocados.

**Independent Test**: Com contas seed, autenticar-se e percorrer home/catálogo → curso → aula; confirmar tom de escola (não planilha), fichas utilizáveis e login/alunos/usuários no visual novo — sem depender ainda de live/VOD/chat/certificado para validar a navegação.

**Acceptance Scenarios**:

1. **Given** um usuário autenticado (aluno, professor ou admin), **When** acessa a área autenticada, **Then** vê catálogo/home de cursos no visual de escola (não uma tabela/planilha de CRUD como experiência principal).
2. **Given** um curso existente, **When** abre a ficha do curso, **Then** vê a trilha de aulas e consegue seguir para a ficha de uma aula.
3. **Given** a agenda de aulas, **When** o usuário a consulta, **Then** consegue abrir a ficha da aula correspondente.
4. **Given** visitante sem sessão, **When** tenta acessar áreas autenticadas, **Then** só encontra o fluxo de login (sem catálogo/fichas protegidas).
5. **Given** admin (ou perfil autorizado), **When** usa gestão de alunos e usuários, **Then** essas telas estão no mesmo visual de escola, sem reabrir papéis ou permissões.

---

### User Story 2 - Live, VOD e chat na ficha da aula (Priority: P1)

Na **ficha da aula**, o aluno assiste à transmissão ao vivo e à gravação quando existirem e usa o chat da aula; professor e admin fazem o mesmo e, além disso, gerenciam live e VOD conforme o RBAC já entregue (002–004). Qualquer aviso de que o player “não faz parte do MVP” desaparece: live, VOD e chat entram no visual da escola. O chat continua efêmero à conexão atual (sem histórico ao reabrir; sem moderação).

**Why this priority**: Fecha o gap entre o visual do colega (sem 002–004) e o produto real; sem isso a demo perde live/VOD/chat no novo layout.

**Independent Test**: Na ficha de uma aula seed/demo, com contas aluno e professor/admin: verificar status live, player, ações de gestão live (só admin/professor), player e gestão VOD (escrita só admin/professor), e painel de chat — tudo no visual da escola e no mesmo recorte de permissões de 002–004.

**Acceptance Scenarios**:

1. **Given** aula com live em estado conhecido (`agendada`, `ao_vivo` ou `encerrada`), **When** aluno, professor ou admin abre a ficha da aula, **Then** vê o status e, quando aplicável, o player de transmissão no visual da escola.
2. **Given** admin ou professor na ficha da aula, **When** gerencia a live (agendar, cancelar, iniciar, encerrar) e, quando aplicável, obtém dados de ingest para OBS, **Then** as ações disponíveis batem com o RBAC atual e o visual é o da escola.
3. **Given** aluno na ficha da aula, **When** tenta ações de gestão de live ou de escrita VOD, **Then** a UI não oferece esses controles (apenas leitura/assistir).
4. **Given** aula com gravação publicada, **When** qualquer papel autenticado autorizado a ler abre a ficha, **Then** consegue assistir ao VOD; admin/professor podem publicar, substituir ou remover o arquivo conforme RBAC; **não** há download do MP4 nem página “Biblioteca”.
5. **Given** usuários autenticados na ficha da aula, **When** usam o chat, **Then** enviam/recebem mensagens só da conexão atual; ao reabrir a ficha, não há histórico prévio; sem controles de moderação.
6. **Given** a ficha da aula no visual da escola, **When** um avaliador a inspeciona, **Then** **não** encontra aviso de que player/live/VOD “não fazem parte deste MVP”.

---

### User Story 3 - Certificado na ficha do curso (Priority: P1)

Na **ficha do curso**, o aluno elegível solicita/baixa o próprio PDF; o admin gerencia elegibilidade e invalidação no mesmo contexto. Professor não gerencia certificado. Não existe página ou rota dedicada “Certificados”.

**Why this priority**: Completa o vestir das superfícies 005 no destino natural da escola (ficha do curso), mantendo o recorte P1 já aceito.

**Independent Test**: Com aluno seed elegível e admin: na ficha do curso, aluno baixa PDF; admin marca/reabilita elegibilidade e invalida; professor não vê gestão de certificado.

**Acceptance Scenarios**:

1. **Given** aluno autenticado e elegível, **When** solicita/baixa o certificado na ficha do curso, **Then** obtém o PDF conforme o comportamento já entregue em 005.
2. **Given** admin na ficha do curso, **When** gerencia elegibilidade ou invalida certificado, **Then** os controles aparecem no contexto do curso, no visual da escola, sem página “Certificados”.
3. **Given** professor autenticado, **When** abre a ficha do curso, **Then** **não** vê controles de gestão de certificado.
4. **Given** a navegação da aplicação, **When** um avaliador procura uma rota/página “Certificados”, **Then** ela não existe.

---

### User Story 4 - Jornadas por papel sem mudar o produto (Priority: P1)

As jornadas de aluno, professor e admin permanecem as do produto 001–005, apenas com visual e navegação novos. Nenhum papel novo, nenhum fluxo de negócio novo, nenhuma mudança de API/RBAC.

**Why this priority**: Critério de aceite da demo e garantia de não regressão de 001–005.

**Independent Test**: Percorrer as três jornadas seed e confirmar permissões e capacidades iguais às de antes, só com UI de escola.

**Acceptance Scenarios**:

1. **Given** aluno seed, **When** percorre catálogo → curso → aula, **Then** assiste live/VOD quando houver, usa chat e baixa certificado na ficha do curso se elegível — sem CRUD de cursos/aulas/alunos/usuários nem gestão de certificado/live/VOD além da leitura.
2. **Given** professor seed, **When** usa a aplicação, **Then** tem as mesmas leituras do aluno **mais** CRUD de aulas e gestão de live/VOD e chat; **sem** gestão de certificado e **sem** CRUD de cursos/alunos/usuários.
3. **Given** admin seed, **When** usa a aplicação, **Then** tem tudo do professor **mais** CRUD de cursos, alunos e usuários **e** gestão de certificado no curso.
4. **Given** a demo após esta feature, **When** se exercitam os fluxos 001–005 via UI, **Then** nenhum contrato ou capacidade de API necessária à demo deixa de funcionar por causa do visual.

---

### Edge Cases

- Merge cego ou substituição que apague blocos de live/VOD/chat/certificado: **proibido** — o visual da escola MUST incorporar essas superfícies; exclusão de comportamento 002–005 é falha de aceite.
- Aula sem live, sem VOD ou sem mensagens de chat: a ficha permanece utilizável com empty states coerentes com o visual da escola.
- Aluno não elegível ao certificado: feedback claro na ficha do curso; sem emissão.
- Usuário autenticado com papel sem permissão de escrita: controles de gestão (live, VOD, certificado, CRUD) não aparecem ou são recusados como hoje.
- Datas exibidas na UI MUST permanecer em formato BR (`dd/mm/yyyy` ou `dd/mm/yyyy HH:mm`).
- Trabalho fora de `feat/aws-speckit` ou merge git completo de `feat/frontend-escola` em `feat/aws-speckit`: fora do escopo desta feature.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: A interface autenticada MUST apresentar catálogo/home de cursos, ficha de curso (com trilha de aulas), agenda de aulas e ficha de aula no visual de “escola”, não como planilha/CRUD genérico como experiência principal.
- **FR-002**: Login, gestão de alunos e gestão de usuários MUST adotar o mesmo visual de escola, preservando quem pode acessar cada área segundo o RBAC atual.
- **FR-003**: Visitante não autenticado MUST ter acesso apenas ao fluxo de login para áreas protegidas.
- **FR-004**: A ficha da aula MUST exibir transmissão ao vivo (status `agendada` / `ao_vivo` / `encerrada`, player quando aplicável) e ações de gestão live + ingest OBS apenas para admin e professor, no mesmo recorte de 003.
- **FR-005**: A ficha da aula MUST exibir gravação VOD (assistir para papéis com leitura; publicar/substituir/remover apenas admin/professor), sem página “Biblioteca” e sem download do arquivo de mídia, no mesmo recorte de 002.
- **FR-006**: A ficha da aula MUST incluir o painel de chat da aula (mensagens apenas da conexão atual; sem histórico ao reabrir; sem moderação), no mesmo recorte de 004.
- **FR-007**: A ficha do curso MUST incluir certificado no contexto do curso (aluno solicita/baixa o próprio PDF se elegível; gestão de elegibilidade/invalidação só admin; professor sem gestão), sem página “Certificados”, no mesmo recorte de 005.
- **FR-008**: A UI MUST NÃO exibir aviso de que player, live ou VOD “não fazem parte deste MVP”.
- **FR-009**: Papéis, permissões e capacidades de negócio MUST permanecer exatamente os de `admin` / `professor` / `aluno` já entregues em 001–005; esta feature NÃO introduz papéis, fluxos de produto ou regras de negócio novas.
- **FR-010**: Datas apresentadas na UI MUST usar formato brasileiro `dd/mm/yyyy` ou `dd/mm/yyyy HH:mm`.
- **FR-011**: Esta feature MUST NÃO alterar API, contratos, seeds, backend, Terraform, publish, CI (além do necessário para o frontend buildar) nem reabrir o desenho das features 001–005.
- **FR-012**: A incorporação do visual MUST preservar todo o comportamento de produto já entregue em 002–005 (live, VOD, chat, certificado); não é aceitável “adotar a escola” apagando essas superfícies.
- **FR-013**: Fora de escopo MUST permanecer: P2 de 002–005 (moderação de chat, certificado P2, live→VOD, página Biblioteca, etc.), OAuth, Redis AWS, NAT, domínio customizado, EventBridge apply/destroy do 001, redesign de produto e merge git completo entre `feat/frontend-escola` e `feat/aws-speckit`.

### Key Entities

- **Catálogo / Home de cursos**: superfície de entrada autenticada que lista cursos no tom de escola.
- **Ficha do curso**: contexto do curso com trilha de aulas e painel de certificado (leitura/solicitação para aluno elegível; gestão só admin).
- **Ficha da aula**: contexto da aula com live, VOD e chat no visual da escola.
- **Agenda de aulas**: listagem/navegação temporal para abrir fichas de aula.
- **Papéis (RBAC)**: `admin`, `professor`, `aluno` — inalterados; visitante só login.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Com as três contas seed, um avaliador percorre em até 10 minutos as jornadas aluno, professor e admin e confirma visual de escola (não planilha) nas telas principais (home/catálogo, ficha de curso, ficha de aula, login e, no admin, alunos/usuários).
- **SC-002**: Em 100% das jornadas seed testadas, live, VOD, chat e certificado permanecem usáveis no mesmo recorte de permissões de 002–005, agora nas fichas de aula e curso do visual da escola.
- **SC-003**: 0 regressões de aceite nos fluxos de produto 001–005 exercitáveis pela UI (login, CRUD permitido por papel, live, VOD, chat, certificado) atribuíveis a esta mudança visual.
- **SC-004**: 0 páginas novas “Biblioteca” ou “Certificados”; 0 avisos na ficha da aula de que player/live/VOD estão fora do MVP.
- **SC-005**: 100% das datas visíveis nas telas tocadas por esta feature seguem o formato BR (`dd/mm/yyyy` ou `dd/mm/yyyy HH:mm`).
- **SC-006**: Após a mudança, na sessão de demo o usuário autentica e navega as fichas de curso/aula sem perder nenhuma capacidade de produto já demonstrável em 001–005.

## Assumptions

- O trabalho de implementação permanece na branch git `feat/aws-speckit` (como 001–005); o diretório Speckit é `specs/006-escola-ui`.
- A fonte do visual a incorporar é o frontend já desenvolvido em `feat/frontend-escola` (referência de commit/ideia: redesenho “escola, não planilha”); a abordagem prevista no plan é copiar/selecionar apenas artefatos de interface — sem merge git completo das branches e sem puxar alterações fora do frontend.
- Live, VOD, chat e certificado já existem e funcionam na baseline atual; esta feature só os recoloca/veste no layout de escola (ficha da aula / ficha do curso).
- Contas e dados seed atuais bastam para a demo de aceite.
- “Visual de escola” significa shell, tipografia/componentes de apresentação, cards, modais, status pills e empty states alinhados ao tom educacional — não um redesign de marca ou de produto além do que a fonte do colega já cobriu.
- Build do frontend pode exigir ajustes mínimos de integração (rotas, composição de painéis); qualquer mudança fora do frontend só se for estritamente necessária para o build, e deve ser justificada no plan/tasks.
- P2 e itens explicitamente fora de escopo nesta especificação não entram em “próximo passo implícito”.
