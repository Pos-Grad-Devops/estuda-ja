# Feature Specification: Certificado ao final do curso

**Feature Branch**: `005-course-certificate`

**Created**: 2026-09-05

**Status**: Draft

**Input**: User description: "Certificado ao final do curso — última peça do escopo de produto (DESCRICAO.md) após 001–004. Aluno autenticado solicita/baixa/visualiza certificado ao cumprir critérios mínimos; admin (e/ou professor se RBAC fizer sentido) vê emissão / reemite / invalida no mínimo da demo; UI no contexto do curso (sem portal separado); registro persistido + artefato (PDF ou equivalente) alinhado à stack efêmera; seed de demo; CI/testes; documentar destroy. Fora: blockchain, redesign 001–004, OAuth, microsserviços, histórico eterno entre destroys, moderação chat P2, live→VOD. Preferir regra de conclusão mais simples (YAGNI)."

## Clarifications

### Session 2026-09-05

- Q: Qual regra P1 autoriza o certificado? → A: **Admin marca o curso como concluído para o aluno** (flag de elegibilidade por par aluno–curso). Sem tracking de aulas; sem matrícula formal.
- Q: Formato P1 do artefato? → A: **PDF baixável**. Persistência do arquivo (S3 da sessão vs geração on-the-fly) fica para o plan, privilegiando custo baixo e destroy entre sessões.
- Q: Quem gerencia elegibilidade / emissão / invalidação no P1? → A: **Só admin**. Professor e aluno NÃO gerem certificados; professor pode no máximo o que a leitura de cursos já permite (sem controles de gestão de certificado na UI).
- Q: Quando o registro do certificado (e o PDF) passa a existir após elegibilidade? → A: **Elegibilidade desbloqueia**; a **primeira solicitação do aluno elegível** cria o registro válido e o PDF; solicitações seguintes reutilizam o mesmo certificado ativo (sem duplicar).
- Q: Após invalidar, o aluno pode obter novo certificado no P1? → A: **Sim, com reabilitação**: invalidar encerra o ativo; admin MUST **reabilitar elegibilidade**; só então a próxima solicitação do aluno cria um **novo** certificado válido.
- Q: O que o seed deixa pronto para o aluno seed? → A: **Só elegibilidade** no curso demo (sem certificado pré-emitido); o aluno faz a primeira solicitação e gera o PDF na demo.
- Q: A que identidade de “aluno” se vinculam elegibilidade e certificado? → A: **Usuário autenticado com papel aluno** (conta de login); o CRUD de Alunos permanece independente no P1.
- Q: Onde o admin gerencia elegibilidade/consulta/invalidação na UI P1? → A: **Painel mínimo no contexto do curso** (mesmo contexto onde o aluno solicita o PDF); sem página “Certificados” dedicada.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Aluno obtém certificado de um curso elegível (Priority: P1)

Um aluno autenticado, no contexto de um **curso** para o qual o **admin já marcou elegibilidade** (curso concluído para aquele aluno), solicita o certificado: na **primeira** solicitação a plataforma **emite** (cria registro válido + PDF) e entrega o download; nas seguintes, **reutiliza** o certificado ativo. A jornada acontece na UI do curso (ficha/contexto do curso), sem portal elaborado. O certificado identifica o aluno, o curso e a data de emissão em formato brasileiro.

**Why this priority**: Fecha o enunciado do produto (“emissão de certificado ao final do curso”); sem o caminho aluno → certificado na demo, a feature não entrega valor de negócio.

**Independent Test**: Com o aluno seed (ou equivalente) e elegibilidade já marcada (via admin ou seed), o aluno autentica-se, abre o curso, solicita/baixa o PDF e verifica nome, curso e data BR — sem passos manuais opacos além do documentado no seed/runbook.

**Acceptance Scenarios**:

1. **Given** um aluno autenticado e elegível **sem** certificado ainda, **When** solicita o certificado na UI do curso pela primeira vez, **Then** a plataforma cria o registro válido, gera o **PDF** e entrega o download com nome do aluno, curso e data de emissão em `dd/mm/yyyy` (ou `dd/mm/yyyy HH:mm` se houver hora).
2. **Given** um aluno autenticado e um curso para o qual **não** é elegível, **When** tenta solicitar o certificado, **Then** a ação é recusada com feedback claro em português e nenhum certificado válido é emitido.
3. **Given** um visitante sem sessão válida, **When** tenta emitir ou baixar certificado, **Then** o acesso é negado com feedback em português.
4. **Given** um aluno que já possui certificado **válido** para aquele curso, **When** solicita novamente baixar, **Then** obtém o PDF do mesmo certificado ativo (sem criar segundo registro ativo).

---

### User Story 2 - Admin habilita elegibilidade e gerencia emissão na demo (Priority: P1)

Um **admin** autenticado consegue, no mínimo necessário à demo, num **painel mínimo no contexto do curso**: **marcar/reabilitar** o curso como concluído para um usuário-aluno (elegibilidade — **não** emite o PDF por si só), **consultar** se há certificado emitido para o par, e **invalidar** um certificado. **Somente admin** vê esses controles; professor e aluno não. Sem portal/página “Certificados” dedicada.

**Why this priority**: Sem controle mínimo de elegibilidade, a demo depende de estado oculto; sem matrícula formal, a flag admin é a regra YAGNI; a emissão ocorre na primeira solicitação do aluno.

**Independent Test**: Admin marca elegibilidade para o aluno seed; aluno solicita e recebe o PDF (primeira emissão); admin invalida; aluno não baixa certificado válido; admin **reabilita** elegibilidade; aluno solicita de novo e recebe um **novo** PDF válido.

**Acceptance Scenarios**:

1. **Given** admin autenticado na UI do curso, **When** marca o curso como concluído para um usuário com papel aluno, **Then** esse usuário passa a poder solicitar o certificado; **nenhum** registro de certificado válido é criado só por marcar elegibilidade; os controles aparecem no **contexto do curso**, não numa página dedicada de certificados.
2. **Given** admin autenticado e um certificado válido existente, **When** invalida o certificado, **Then** o registro deixa de ser tratado como válido; a elegibilidade desse par **deixa de autorizar** nova emissão até reabilitação explícita; tentativas do aluno de baixar o invalidado falham com feedback em português.
3. **Given** certificado invalidado, **When** admin reabilita elegibilidade e o aluno solicita de novo, **Then** um **novo** certificado válido (e PDF) é criado; o anterior permanece invalidado.
4. **Given** um aluno, **When** tenta ações de gestão (marcar elegibilidade de outros, invalidar certificados), **Then** a ação é recusada e a UI não oferece esses controles.
5. **Given** professor autenticado, **When** tenta marcar elegibilidade, emitir ou invalidar certificado, **Then** a ação é recusada com feedback em português e a UI não oferece controles de gestão de certificado.

---

### User Story 3 - Operador: registro e artefato no ciclo efêmero apply → destroy (Priority: P1)

O operador inclui certificados no ciclo da sessão efêmera: após apply/publish, a demo emite e exibe certificado; ao **destroy**, registros e artefatos da sessão são removidos com a stack (dados efêmeros + seed na próxima sessão). O runbook documenta explicitamente o que o destroy faz com certificados. Sem domínio customizado salvo necessidade mínima comprovada; sem Cognito; sem Redis AWS; sem NAT; `us-east-1`; budget da conta intocado.

**Why this priority**: Constitution e 001 exigem destroy entre sessões; certificado não pode deixar artefato/custo contínuo esquecido.

**Independent Test**: apply → publish → emitir/ver certificado na demo → destroy; stack e artefatos da sessão removidos; budget ativo; runbook descreve o efeito do destroy nos certificados.

**Acceptance Scenarios**:

1. **Given** a baseline 001–004, **When** o operador provisiona/publica esta feature conforme o runbook, **Then** emitir/visualizar certificado funciona na sessão em `us-east-1`, sem NAT, sem Cognito, sem Redis AWS, e sem domínio customizado salvo necessidade mínima justificada no plan.
2. **Given** o fim da demo, **When** executa o destroy completo da sessão, **Then** registros e artefatos de certificado da sessão são removidos com a stack; o alerta de orçamento permanece.
3. **Given** o runbook, **When** um avaliador o consulta, **Then** encontra o que o destroy faz com certificados e os passos encaixados no ciclo apply → publish → demo → destroy, **sem** reabrir o desenho de 001–004.

---

### User Story 4 - Seed de demo e caminho local/CI (Priority: P1)

O seed de demo oferece um caminho **reproduzível**: pré-marca **elegibilidade** do aluno seed no curso demo (**sem** certificado pré-emitido). Localmente (Compose/API) o aluno seed solicita e emite o PDF na primeira vez; o CI cobre testes de handlers/contratos/RBAC **sem** provisionar AWS.

**Why this priority**: Avaliadores e operadores precisam de caminho previsível; constitution V exige testes ao mudar API; seed só com elegibilidade demonstra o fluxo completo de emissão.

**Independent Test**: Com seed aplicado, aluno seed autentica-se, abre o curso demo, solicita certificado (primeira emissão) e baixa o PDF; no CI, testes da feature passam sem Terraform/AWS.

**Acceptance Scenarios**:

1. **Given** ambiente local com seed de demo, **When** o operador/avaliador segue o caminho documentado com o aluno seed, **Then** encontra elegibilidade já marcada, **sem** certificado pré-existente, e conclui a primeira solicitação (emissão + PDF) sem passos não documentados.
2. **Given** o pipeline de CI, **When** rodam os testes da feature, **Then** handlers/contratos/RBAC e erros em português são verificados sem stack AWS.
3. **Given** a documentação, **When** consultada, **Then** deixa explícito: seed = elegibilidade apenas; local = fluxo real de emissão na primeira solicitação; CI = testes sem AWS; demo AWS = mesmo comportamento na sessão efêmera.

---

### User Story 5 - Template visual mais rico e lista do aluno (Priority: P2 — fora da entrega desta rodada)

Em fase futura: template visual mais elaborado do certificado, reemissão com versionamento explícito e lista agregada de certificados do aluno. **Fora do plan/tasks desta rodada**: o “pronto” é P1 (elegibilidade mínima + emissão/visualização + gestão mínima + seed + destroy documentado + CI).

**Why this priority**: Melhora percepção de qualidade, mas não bloqueia o MVP demonstrável; YAGNI nesta rodada.

**Independent Test**: (Futuro) Aluno vê lista dos seus certificados; gestor reemite com template rico.

**Acceptance Scenarios** *(referência futura; não exigidos nesta entrega)*:

1. **Given** aluno com um ou mais certificados válidos, **When** abre a lista de certificados, **Then** vê os cursos certificados de forma legível.
2. **Given** admin (e perfil autorizado), **When** reemite com template enriquecido, **Then** o aluno obtém a nova versão conforme a política futura.

---

### Edge Cases

- Sessão AWS destruída: URLs/artefatos e registros da sessão anterior são inválidos; novo `apply` + seed recomeça limpo.
- Aluno tenta certificado de curso inexistente ou sem vínculo permitido pela regra: recusa com erro em português.
- Certificado invalidado: download como válido é recusado; elegibilidade do par **não** autoriza nova emissão até o admin **reabilitar** explicitamente.
- Após invalidar + reabilitar elegibilidade: a próxima solicitação do aluno cria um **novo** certificado ativo; o invalidado permanece no histórico da sessão (sem voltar a válido).
- Emissão duplicada: NÃO criar múltiplos certificados **ativos** para o mesmo par aluno–curso; segunda solicitação (com ativo válido) reutiliza o ativo.
- Marcar elegibilidade sem solicitação do aluno: NÃO cria registro/PDF de certificado.
- Falha ao gerar o PDF na primeira solicitação: feedback em português; NÃO deixar registro como válido se o PDF não puder ser entregue.
- Professor: ações de gestão de certificado sempre recusadas no P1; UI sem controles de gestão.
- Expectativa de verificação pública em blockchain / portal externo: fora de escopo.
- Expectativa de histórico eterno entre destroys: fora; dados efêmeros + seed.
- Pico de aula ao vivo: certificado NÃO faz parte do caminho crítico T−0 da live; warm-up de live/chat permanece o de 003/004; esta feature NÃO MUST introduzir cold start na abertura da aula.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: A plataforma MUST permitir que um **aluno autenticado** e elegível solicite e **baixe** um certificado de conclusão em **PDF** de um **curso**, na **UI no contexto do curso** (sem portal/página “Certificados” separada no P1).
- **FR-001a**: Em P1, a **primeira** solicitação de um aluno elegível **sem** certificado ativo MUST **criar** o registro válido e o PDF e entregar o download; solicitações seguintes MUST **reutilizar** o mesmo certificado ativo (sem segundo registro ativo). Marcar elegibilidade **sozinha** NÃO MUST criar registro nem PDF.
- **FR-001b**: Em P1, a gestão admin (marcar/reabilitar elegibilidade, consultar emissão, invalidar) MUST viver num **painel mínimo no mesmo contexto do curso** onde o aluno solicita o PDF. NÃO MUST haver página dedicada de certificados no P1.
- **FR-002**: O certificado emitido MUST identificar, no mínimo: nome do aluno, nome do curso e data de emissão em formato brasileiro (`dd/mm/yyyy` ou `dd/mm/yyyy HH:mm`).
- **FR-003**: Visitantes sem sessão válida NÃO MUST emitir nem baixar certificados; tentativas MUST resultar em feedback claro em português.
- **FR-004**: Em P1, a elegibilidade MUST ser concedida quando o **admin marca o curso como concluído** para um **usuário com papel aluno** (flag/estado por par usuário-aluno–curso). NÃO MUST exigir matrícula formal, critério baseado em conclusão de aulas, nem vínculo obrigatório com o CRUD de Alunos. Elegibilidade desbloqueia a solicitação; **não** emite o certificado por si só.
- **FR-004a**: Elegibilidade e certificado MUST referir-se ao **usuário autenticado com papel `aluno`** (conta de login). O nome no PDF MUST ser o nome desse usuário. O CRUD de Alunos NÃO MUST ser pré-requisito da emissão no P1.
- **FR-005**: A plataforma MUST **persistir o registro** do certificado (pelo menos: aluno, curso, status válido/invalidado, data de emissão) de forma recuperável durante a vida da sessão de demo / ambiente local, a partir da primeira emissão bem-sucedida.
- **FR-006**: Em P1, o artefato do certificado MUST ser um **PDF baixável**. O plan MUST escolher persistência do arquivo (ex.: storage da sessão) versus geração on-the-fly, privilegiando custo baixo e destroy entre sessões; o registro no banco MUST existir independentemente dessa escolha.
- **FR-007**: **Somente admin** MUST poder, no mínimo necessário à demo: marcar / **reabilitar** elegibilidade (curso concluído para o aluno), **consultar** emissão do par aluno–curso e **invalidar** certificado válido. Professor e aluno NÃO MUST executar essas ações de gestão. Admin NÃO MUST ser obrigado a “emitir” o PDF em passo separado no P1 (emissão = primeira solicitação do aluno elegível sem ativo).
- **FR-007a**: Em P1, **invalidar** um certificado MUST torná-lo não utilizável e MUST **impedir** nova emissão para o mesmo par até o admin **reabilitar elegibilidade**. Após reabilitação, a próxima solicitação do aluno MUST criar um **novo** certificado válido (o anterior permanece invalidado).
- **FR-008**: Aluno NÃO MUST gerir elegibilidade/invalidação de certificados; professor NÃO MUST gerir certificados no P1; a UI NÃO MUST expor esses controles a aluno nem a professor. Controles de gestão MUST aparecer **apenas** para admin no painel do contexto do curso.
- **FR-009**: Datas exibidas no contexto do certificado MUST usar formato brasileiro; erros da API MUST estar em português.
- **FR-010**: Auth permanece a JWT + papéis RBAC existentes (`admin`, `professor`, `aluno`). Esta feature NÃO MUST introduzir OAuth/Cognito nem ampliar poderes de live/VOD/chat além do necessário ao certificado.
- **FR-011**: Certificados da demo MUST viver na **stack efêmera** (mesmo ciclo apply/destroy); após destroy, registros e artefatos da sessão NÃO MUST permanecer gerando custo contínuo material. Budget da conta permanece separado e intocado.
- **FR-012**: Destino de demo MUST permanecer AWS gerenciada em **`us-east-1`**, sem NAT Gateway, sem Redis/ElastiCache na AWS, sem domínio customizado **salvo** necessidade mínima justificada no plan (ex.: link estável do artefato). Auth própria JWT + bcrypt.
- **FR-013**: MUST existir **seed de demo** (e documentação) que pré-marque **elegibilidade** do aluno seed no curso demo **sem** pré-emitir certificado, permitindo emissão na primeira solicitação sem setup manual opaco.
- **FR-014**: CI (build + testes) MUST permanecer obrigatório; alterações de API/handlers relacionadas a certificado MUST incluir testes automatizados relevantes. Ambiente local MUST exercitar o fluxo de negócio sem stack AWS.
- **FR-015**: O runbook (README / quickstart da feature) MUST documentar o que o **destroy** faz com certificados da sessão e os passos no ciclo apply → publish → demo → destroy.
- **FR-016**: Segredos e credenciais MUST permanecer fora do código e de commits.
- **FR-017**: Auth, papéis e comportamentos de 001–004 (cursos, aulas, alunos, usuários, VOD, live, chat) MUST permanecer; esta feature NÃO MUST redesenhar live, VOD, chat nem a fundação AWS.
- **FR-018** (P2 futuro — **fora da entrega desta rodada**): template visual mais rico, lista agregada de certificados do aluno e reemissão elaborada. Plan/tasks desta feature NÃO MUST implementar FR-018 como obrigatório.
- **FR-019**: Blockchain, verificação pública elaborada, microsserviço de certificado separado, OAuth/Cognito, histórico eterno entre destroys, moderação de chat P2, pipeline live→VOD e reabertura de 001–004 NÃO MUST fazer parte desta feature.

### Key Entities

- **Certificado de conclusão**: Comprovante criado na **primeira solicitação** do usuário-aluno elegível; possui status (válido / invalidado), data de emissão e vínculo ao **usuário (papel aluno)** e ao curso; associado a um PDF baixável; no máximo um ativo por par usuário–curso.
- **Elegibilidade / conclusão**: Flag/estado por par **usuário-aluno–curso**, concedida **somente pelo admin**; **desbloqueia** a solicitação; não cria o certificado sozinha; após invalidação, MUST ser **reabilitada** pelo admin antes de nova emissão.
- **Artefato do certificado**: PDF entregue ao aluno no download (persistido na sessão ou gerado on-the-fly — detalhe no plan).
- **Usuário-aluno**: Conta de login com papel `aluno`; sujeito da elegibilidade e do certificado no P1 (independente do CRUD Alunos).
- **Sessão de demo**: Janela apply→destroy em que registros e artefatos existem; após destroy, não valem.

### Constraints *(constitution)*

- Custo-consciente: artefato/storage só na janela da sessão; destroy remove; sem NAT; sem Redis AWS; budget separado.
- Pronto para o pico: certificado fora do caminho crítico T−0 da live; não introduzir cold start na abertura da aula.
- Escopo mínimo / YAGNI: entrega desta rodada = **só P1**; elegibilidade = flag admin; PDF simples; P2 documentado como futuro.
- Segurança e RBAC: JWT próprio; gestão de certificado **só admin**; erros PT; datas BR.
- Qualidade testável: CI verde; testes ao alterar API/handlers.
- IaC: Terraform enxuto se houver recurso de artefato na sessão; `us-east-1`.
- Baseline 001–004: não reabrir.

### Out of Scope

- Blockchain / verificação pública elaborada / QR de autenticidade externa obrigatória
- Redesign ou reabertura de 001–004 (AWS, VOD, live IVS, chat WS)
- OAuth externo / Cognito
- Microsserviço ou produto separado só de certificados
- Domínio customizado / ACM custom (salvo necessidade mínima justificada no plan)
- NAT Gateway; ElastiCache/Redis na AWS
- Histórico eterno de certificados entre destroys de demos distintas
- Vínculo obrigatório certificado ↔ CRUD Alunos
- Matrícula/enrolment formal; critério de conclusão por aulas (N aulas / presença)
- Gestão de certificado por professor (P1 = só admin)
- Portal dedicado elaborado de “Meus certificados” (P2)
- Template visual rico / brand kit completo (P2)
- Artefato HTML-only como substituto do PDF no P1
- Moderação de chat P2; pipeline live→VOD
- App nativo

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: No caminho feliz (local ou demo AWS com seed), um aluno autenticado elegível obtém o **PDF** do certificado na UI do curso em ≤ 2 minutos após login (seguindo o runbook).
- **SC-002**: Em verificação amostral, 100% das tentativas de emitir/baixar certificado sem autenticação válida são recusadas com feedback em português.
- **SC-003**: Em verificação amostral, 100% das tentativas de aluno não elegível obter certificado são recusadas; 100% das tentativas de aluno ou professor gerir elegibilidade/invalidação são recusadas; após invalidação sem reabilitação, 100% das tentativas de nova emissão são recusadas.
- **SC-004**: Após o destroy documentado da sessão, não permanece artefato/registro de certificado da demo gerando custo contínuo material (expectativa alinhada ao 001: ~US$ 0 fora da janela, além de resíduos já documentados); budget permanece.
- **SC-005**: Um operador completa os passos extras de certificado no runbook (além do ciclo 001–004) em ≤ 15 minutos de leitura/execução incremental, incluindo a compreensão do efeito do destroy.
- **SC-006**: CI permanece verde (build + testes) no branch da feature; regressão nos testes de handlers/contratos/RBAC de certificado falha o pipeline.
- **SC-007**: Avaliador distingue, no contexto do curso, a ação de certificado das áreas de live/VOD/chat, sem treinamento além da UI e do quickstart.
- **SC-008**: Itens fora de escopo desta entrega (blockchain, Cognito, redesign 001–004, portal elaborado, template rico P2, histórico entre destroys) não são entregues nem apresentados como “pronto” desta rodada.
- **SC-009**: O PDF do certificado contém nome do aluno, nome do curso e data de emissão em formato brasileiro, verificáveis na demo em ≤ 30 segundos de inspeção visual.
- **SC-010**: Com seed aplicado (elegibilidade do aluno seed, sem certificado pré-emitido), o caminho primeira solicitação → PDF é concluído sem passos manuais não documentados no runbook/quickstart.

## Assumptions

- Features `001`–`004` estão concluídas e são baseline; esta feature só acrescenta certificado de conclusão de curso.
- Constituição 1.0.0 permanece válida; esta feature não a emenda.
- O modelo atual **não** possui matrícula/enrolment formal; elegibilidade P1 = **flag admin** “curso concluído” para **usuário com papel aluno** (clarificação 2026-09-05).
- Identidade do certificado = **usuário (papel aluno)**, não o CRUD Alunos (clarificação 2026-09-05).
- Artefato P1 = **PDF baixável** (clarificação 2026-09-05); se o PDF fica em storage da sessão ou é gerado on-the-fly fica para o plan (custo baixo + destroy).
- Emissão P1 = **lazy na primeira solicitação** do aluno elegível; elegibilidade sozinha não emite (clarificação 2026-09-05).
- Reemissão P1 após invalidar = exige **reabilitação de elegibilidade** pelo admin; depois nova solicitação cria novo certificado (clarificação 2026-09-05).
- Gestão P1 de elegibilidade/consulta/invalidação = **só admin** (clarificação 2026-09-05); professor sem controles de gestão.
- UI no **contexto do curso** para aluno (solicitar/baixar) e para admin (painel mínimo de gestão) — sem página “Certificados” dedicada no P1 (clarificação 2026-09-05).
- Dados de certificado são **efêmeros entre sessões AWS** (destroy + seed), como o restante da demo 001.
- Seed de demo MUST pré-marcar elegibilidade do aluno seed no curso demo e NÃO MUST pré-emitir certificado (clarificação 2026-09-05); o caminho feliz da demo é a primeira solicitação.
- Rebuild/publish do frontend por sessão permanece como no 001.
- Datas BR e erros em português permanecem obrigatórios.
- **Entrega desta rodada** = **só P1**; US5 / FR-018 documentados como futuro.
- Certificado não altera o warm-up T−15 de live/chat; não é requisito do minuto zero da aula magna.
