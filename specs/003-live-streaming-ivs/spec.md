# Feature Specification: Streaming ao vivo da aula (AWS IVS)

**Feature Branch**: `003-live-streaming-ivs`

**Created**: 2026-09-05

**Status**: Draft

**Input**: User description: "Streaming ao vivo da aula (AWS IVS) sobre 001 (fundação AWS) e 002 (VOD P1). Professor/admin inicia/encerra ou associa transmissão ligada a uma aula com horário; aluno autenticado assiste in-app na ficha da aula; IVS na stack efêmera Terraform; warm-up alinhado ao runbook 001; front+back ponta a ponta; custo controlado (IVS só na janela; destroy). Fora: chat, certificado, pipeline live→VOD, OAuth/Cognito, Redis AWS, NAT, domínio custom, app nativo, multi-qualidade/DVR avançado."

## Clarifications

### Session 2026-09-05

- Q: Modelo de canal / lives simultâneas? → A: **Um canal compartilhado** da sessão Terraform; **no máximo uma live ativa por vez** (associada à aula da demo). YAGNI e menor custo.
- Q: Fluxo de ingestão P1 do professor/admin? → A: **Exibir endpoint + stream key** na ficha da aula (só admin/professor) para encoder externo (ex. OBS).
- Q: Live exige horário na aula? → A: **Horário recomendado na UI**, mas a API **permite** iniciar live em aula sem data/hora (demo ad hoc). Estado “agendada” (P2) só faz sentido quando houver horário.
- Q: Após encerrar, pode reiniciar live na mesma sessão Terraform? → A: **Sim** — reiniciar na mesma aula ou em outra, sem novo `apply`; permanece no máximo **uma** live ativa por vez.
- Q: Como o aluno autenticado obtém o playback da live? → A: **Somente via API autenticada (JWT)** com leitura de aulas; sem URL pública permanente de playback. Tokenização/assinatura de curta duração fica fora do P1.
- Q: Stream key ao reiniciar live na mesma sessão? → A: **Mesma stream key** do canal da sessão até o `destroy`; sem rotação automática a cada “iniciar”.
- Q: No local/CI sem IVS real, o que “iniciar live” garante? → A: **Stub** — estados, RBAC e erros em português testáveis; **sem** sinal de vídeo real. Ingest/playback reais só na sessão AWS.
- Q: Em P1, “iniciar” e “associar” são a mesma ação? → A: **Ação única** — “iniciar live nesta aula” associa ao canal, marca como ao vivo e mostra ingest; estado “agendada” fica no P2.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Aluno assiste a transmissão ao vivo na ficha da aula (Priority: P1)

Um aluno autenticado abre um curso e uma aula já existentes, identifica que há transmissão ao vivo ativa e assiste no próprio produto (reprodutor embutido na ficha da aula) — sem download do fluxo e sem tela nova de “ao vivo”. Professor e admin, que já leem cursos e aulas, também assistem no mesmo contexto quando a live estiver ativa.

**Why this priority**: Sem reprodução ao vivo ponta a ponta na jornada curso → aula, a feature não demonstra o enunciado (aula magna / pico na abertura) nem o valor acadêmico.

**Independent Test**: Com a plataforma no ar (sessão AWS com streaming provisionado) e uma aula marcada como ao vivo com sinal disponível, um aluno faz login, abre essa aula e inicia a reprodução até ver/ouvir o fluxo ao vivo.

**Acceptance Scenarios**:

1. **Given** um aluno autenticado e uma aula com transmissão **ao vivo** ativa e sinal disponível, **When** o aluno abre a ficha dessa aula, **Then** identifica que há live, inicia a reprodução no reprodutor embutido (**sem** fluxo de download) e vê identificação da aula/curso, com datas em formato brasileiro quando exibidas.
2. **Given** um professor ou admin autenticado, **When** abre a mesma aula com live ativa, **Then** também consegue assistir (leitura de live segue a leitura de cursos/aulas).
3. **Given** um visitante sem sessão válida, **When** tenta obter dados de playback ou reproduzir a live de uma aula, **Then** o acesso é negado de forma clara, em português, sem exposição de credenciais de ingestão, endereços internos ou URL de playback utilizável sem autenticação.
4. **Given** uma aula **sem** transmissão ao vivo ativa, **When** o aluno abre essa aula, **Then** não vê um player de live vazio enganoso: a ausência (ou estado não ao vivo) é evidente.
5. **Given** uma aula que já tem VOD publicado (feature 002) e também live, **When** o aluno abre a ficha, **Then** a experiência distingue claramente “ao vivo agora” de “gravação sob demanda”, sem misturar os dois fluxos nem exigir pipeline live→VOD.

---

### User Story 2 - Professor ou admin inicia / encerra a live da aula (Priority: P1)

Um admin ou professor **inicia** a transmissão ao vivo de **uma aula** com **uma única ação** P1 (associa ao canal compartilhado, marca como ao vivo e exibe ingest — horário recomendado na UI; API permite sem data/hora), vê na ficha o **endpoint de ingestão e a stream key** para enviar o sinal com encoder externo (ex. OBS) e pode encerrar a live. No máximo **uma** live ativa na sessão (canal compartilhado). Alunos passam a assistir na ficha enquanto estiver ao vivo. Aluno não gere live e é recusado se tentar pela API. A UI oculta ações de escrita quando o perfil não pode. Fluxo em dois passos (associar depois iniciar) e estado “agendada” são P2.

**Why this priority**: Sem o ciclo “gestor liga a live → aluno assiste”, a demo depende só de configuração opaca; o avaliador precisa ver o controle pedagógico na aula.

**Independent Test**: Admin ou professor faz login, inicia live numa aula, um aluno (outra sessão) abre a aula e assiste; o gestor encerra; o aluno deixa de ver live ativa; um aluno tenta a mesma escrita e é recusado.

**Acceptance Scenarios**:

1. **Given** um admin ou professor autenticado e infraestrutura de streaming da sessão disponível, **When** executa a ação única **iniciar live** numa aula, **Then** a aula fica **ao vivo**, a ficha exibe **endpoint de ingestão + stream key** para o encoder externo, sem expor esses segredos a alunos. A stream key é a do **canal da sessão** (estável até o destroy; reinícios reutilizam a mesma key).
2. **Given** um aluno autenticado, **When** tenta iniciar ou encerrar a live de uma aula (ou obter credenciais de ingestão), **Then** a ação é recusada com mensagem em português e a interface não oferece o controle correspondente.
3. **Given** uma live ativa numa aula, **When** admin ou professor encerra a transmissão, **Then** a ficha da aula deixa de apresentar a live como ativa para o aluno (estado encerrada ou equivalente), sem exigir destroy da stack inteira.
4. **Given** tentativa de operação sem infra de streaming disponível (sessão local sem equivalente, ou stack sem recurso), **When** o gestor tenta iniciar, **Then** recebe erro claro em português e a aula não fica marcada como ao vivo de forma enganosa.
5. **Given** já existe uma live ativa em outra aula, **When** admin ou professor tenta iniciar live numa segunda aula, **Then** a ação é recusada (ou a anterior deve ser encerrada primeiro) com mensagem em português — no máximo uma live ativa por sessão.
6. **Given** uma aula **sem** data/hora, **When** admin ou professor inicia a live, **Then** a API permite a operação; a UI SHOULD recomendar preencher o horário, sem bloquear a demo ad hoc.
7. **Given** a live de uma aula foi **encerrada** na mesma sessão Terraform (sem destroy), **When** admin ou professor inicia de novo nessa aula ou em outra, **Then** a operação é permitida (sem novo `apply`), desde que continue no máximo uma live ativa.

---

### User Story 3 - Operador inclui IVS no ciclo efêmero apply → warm-up → demo → destroy (Priority: P1)

O operador da conta acadêmica provisiona, na **mesma** sessão Terraform da demo 001 (+ VOD 002), o mínimo de recursos de streaming ao vivo necessário; segue o runbook atualizado (incluindo preparação antes do horário); ao destruir a sessão, não deixa canais/recursos de streaming cobráveis ociosos. Budget da conta permanece intocado.

**Why this priority**: Constitution e 001 exigem destroy entre sessões e custo ~0 fora delas; streaming NÃO pode furar o contrato com canal 24/7 esquecido.

**Independent Test**: Operador segue apply → publish → warm-up (checklist alinhado ao 001) → demo de live na ficha → destroy; verifica remoção dos recursos de streaming da sessão; budget ainda ativo.

**Acceptance Scenarios**:

1. **Given** o código de infra da demo 001/002 como baseline, **When** o operador executa o provisionamento documentado desta feature, **Then** os recursos mínimos de streaming ao vivo ficam disponíveis na mesma região (`us-east-1`) e sessão, sem NAT, sem Redis AWS, sem domínio customizado e sem Cognito.
2. **Given** o término da janela de demo, **When** o operador executa o destroy completo já usado no 001, **Then** recursos de streaming da sessão são removidos com a stack (custo contínuo esperado de IVS/streaming fora de sessão ≈ 0), e o alerta de orçamento da conta permanece.
3. **Given** o runbook da demo, **When** um avaliador o consulta, **Then** encontra passos de streaming encaixados no ciclo apply → publish → warm-up → demo → destroy, sem reabrir o desenho do 001/002.
4. **Given** uma demo com horário de aula simulado, **When** o operador conclui o warm-up documentado **antes** da abertura (antecedência alinhada ao T−15 do 001, com checagens extras de streaming se houver), **Then** na abertura da janela a live não depende de cold start da stack (API saudável + recurso de streaming já provisionado).

---

### User Story 4 - Estados da live na UI da aula (Priority: P2)

A ficha da aula exibe de forma explícita o estado da transmissão: **agendada**, **ao vivo** ou **encerrada** (ou equivalente claro). Alunos e gestores entendem se ainda vai começar, se podem assistir agora, ou se a live já terminou — eventualmente apontando para VOD existente (002) como caminho futuro/paralelo, **sem** implementar conversão live→VOD.

**Why this priority**: Melhora a clareza pedagógica e a demo, mas P1 já vale com “ativa vs não ativa”; estados ricos não bloqueiam o MVP de streaming.

**Independent Test**: Gestor coloca a aula em cada estado; aluno abre a ficha e vê indicação coerente; em encerrada com VOD, vê caminho para gravação se já existir (sem pipeline novo).

**Acceptance Scenarios**:

1. **Given** uma aula com live **agendada** (ainda não iniciada), **When** o aluno abre a ficha, **Then** vê indicação de agendada (com horário em formato BR quando exibido) e **não** um player de live ativo enganoso.
2. **Given** a mesma aula **ao vivo**, **When** o aluno abre a ficha, **Then** vê indicação de ao vivo e pode iniciar a reprodução (User Story 1).
3. **Given** a live **encerrada**, **When** o aluno abre a ficha, **Then** vê indicação de encerrada; se existir VOD publicado da aula (002), MAY indicar a gravação como alternativa — **sem** exigir que a live tenha gerado o VOD automaticamente.

---

### User Story 5 - Automação de warm-up só da parte de streaming (Priority: P2/P3)

Além do checklist manual T−15 do 001, a equipe **pode** automatizar apenas o aquecimento/verificação relacionado ao streaming (ex.: garantir recurso ativo e sinal de saúde antes do horário), sem reabrir EventBridge apply/destroy completo do 001 se isso permanecer gap.

**Why this priority**: Constitution exige pronto para o pico; automação total já foi gap no 001 — aqui só o pedaço de streaming, se necessário, e sem bloquear P1 manual.

**Independent Test**: Com P1 manual funcionando, avaliador consulta doc do que é automatizado vs checklist; se automação existir, um passo documentado valida saúde do streaming antes da janela.

**Acceptance Scenarios**:

1. **Given** que a automação de streaming ainda não está implementada, **When** um avaliador consulta a documentação, **Then** encontra o checklist de warm-up de streaming (manual) alinhado ao runbook 001 e o que é P2/P3.
2. **Given** automação de warm-up de streaming habilitada (se entregue), **When** se aproxima a janela da aula de demo, **Then** a verificação/preparação documentada ocorre sem exigir cliques ad hoc na console além do que o runbook permitir.

---

### Edge Cases

- Sessão AWS destruída: endpoints de playback/ingest da sessão anterior são inválidos; o produto NÃO MUST depender deles após novo `apply`.
- Stack esquecida ligada com canal de streaming: o alerta de billing do 001 continua sendo a rede de segurança; esta feature NÃO MUST introduzir recurso cobrável 24/7 fora do destroy.
- Live marcada como ao vivo sem sinal do encoder: aluno vê estado ao vivo mas feedback claro de indisponibilidade/rebuffer — não tela em branco silenciosa sem mensagem.
- Duas aulas “ao vivo” ao mesmo tempo: **não suportado** — canal compartilhado da sessão; no máximo uma live ativa; segunda tentativa recusada até encerrar a anterior.
- Encerrar live sem destroy: alunos param de assistir; recurso de infra da sessão MAY permanecer até o destroy (custo de sessão), mas estado da aula reflete encerrada. Após encerrar, admin/professor **pode reiniciar** live na mesma aula ou em outra na mesma sessão (sem novo `apply`), mantendo no máximo uma live ativa.
- Ambiente local sem conta AWS: jornada de UI/API e RBAC MUST permanecer testável no desenvolvimento/CI via **stub** (estados da live, permissões, erros em português); **sem** exigir sinal de vídeo real. Ingest/playback reais MAY depender da sessão AWS documentada.
- Aluno tenta “baixar” o ao vivo: fora de escopo; apenas reprodução in-app.
- Expectativa de chat durante a live: fora desta feature (004).
- Expectativa de certificado ao fim: fora desta feature (005).
- Expectativa de gravar a live automaticamente como VOD: fora de escopo; apenas menção futura/link se VOD já existir via 002.
- Pico sem warm-up / apply no horário: runbook MUST deixar explícito que cold start na abertura viola o critério de preparação (constitution II).
- Credenciais de ingestão vazadas na UI do aluno ou em logs commitados: NÃO MUST ocorrer; só perfis com escrita de aulas.
- Playback sem autenticação: dados de reprodução da live NÃO MUST ser obtidos sem JWT válido e permissão de leitura de aulas; URL pública permanente de playback NÃO MUST fazer parte do P1.
- Aula sem horário definido: API **permite** iniciar live; UI **recomenda** preencher data/hora; estado P2 “agendada” só quando houver horário.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: A plataforma MUST permitir que usuários autenticados com permissão de leitura de cursos/aulas (admin, professor e aluno, como no baseline 001) vejam, na ficha da aula, se existe transmissão ao vivo **ativa** (e, em P2, o estado agendada / ao vivo / encerrada).
- **FR-002**: A plataforma MUST permitir que esses leitores **reproduzam** a live ativa **na ficha da aula** (reprodutor no produto). P1 NÃO MUST exigir página ou menu “Ao vivo” dedicado. A plataforma NÃO MUST oferecer download do fluxo ao vivo. Dados de playback da live MUST ser obtidos **somente via API autenticada (JWT)** com leitura de aulas; P1 NÃO MUST expor URL pública permanente de playback nem exigir tokenização/assinatura de curta duração além do JWT.
- **FR-003**: A gestão (**iniciar** / **encerrar**) da live da aula MUST ser permitida a **admin e professor** (mesmo recorte da escrita de aulas). Em P1, **iniciar** MUST ser **ação única**: associa a aula ao canal compartilhado, marca a live como **ao vivo** e disponibiliza credenciais de ingestão. Fluxo em dois passos (associar depois iniciar) e estado “agendada” NÃO MUST ser exigidos em P1 (P2). Aluno NÃO MUST gerir live. Tentativas não autorizadas MUST ser recusadas com mensagem em português; a UI MUST ocultar ações de escrita quando o perfil não pode. Qualquer professor autenticado MAY gerir a live de qualquer aula (sem dono de curso no modelo atual).
- **FR-004**: A transmissão ao vivo MUST estar **ligada a uma aula** existente (com contexto de curso). NÃO MUST existir “canal solto” na UI desconectado de aula para o aluno.
- **FR-005**: A sessão de demo MUST provisionar **um único canal de streaming compartilhado**. Em qualquer momento MUST haver no máximo **uma** live ativa (associada a uma aula). Tentar iniciar live numa segunda aula enquanto outra estiver ativa MUST ser recusado com mensagem em português (ou exigir encerrar a anterior primeiro). Após encerrar, admin/professor MUST poder **reiniciar** live na mesma aula ou em outra **na mesma sessão Terraform** (sem novo `apply`), desde que permaneça no máximo uma live ativa. Canais por aula e lives simultâneas NÃO MUST fazer parte desta feature.
- **FR-006**: O gestor (admin/professor) MUST ver, na ficha da aula **após iniciar** (enquanto a live estiver ativa sob sua gestão), o **endpoint de ingestão e a stream key** do serviço gerenciado de streaming ao vivo na AWS (**IVS**, conforme tabela do README), para uso com encoder externo (ex. OBS). Aluno NÃO MUST ver essas credenciais. Em P1 a stream key MUST ser a do **canal compartilhado da sessão** e permanecer a **mesma** entre reinícios de live até o `destroy`; rotação automática a cada “iniciar” NÃO MUST fazer parte desta feature.
- **FR-021**: Horário da aula é **recomendado** na UI para live e para o estado P2 “agendada”, mas a API MUST **permitir** iniciar live mesmo se a aula não tiver data/hora (demo ad hoc).
- **FR-007**: Os recursos de streaming da demo MUST fazer parte da **stack efêmera** do 001: `destroy` remove canais/recursos cobráveis de streaming da sessão; NÃO MUST haver canal IVS permanente 24/7 fora dessa stack. Custo contínuo esperado de streaming fora de sessão ≈ 0.
- **FR-008**: A preparação para o pico MUST estar documentada no runbook (extensão do checklist T−15 do 001): stack e recurso de streaming provisionados **antes** do horário da aula; cold start na abertura da janela de demo é inaceitável. Automação desse warm-up de streaming é P2/P3 (FR-016); P1 MAY ser 100% manual.
- **FR-009**: Destino de demo MUST continuar AWS gerenciada em **`us-east-1`**, sem NAT Gateway, sem Redis/ElastiCache, sem domínio customizado, sem OAuth/Cognito. Auth permanece JWT + bcrypt do baseline.
- **FR-010**: Infraestrutura nova de streaming MUST ser declarada no Terraform enxuto já versionado da demo (extensão, sem módulos elaborados) e destruída com o restante da sessão. Budget (`infra/budget/`) permanece separado.
- **FR-011**: Datas exibidas ou aceitas nesta feature MUST permanecer no formato brasileiro (`dd/mm/yyyy` ou `dd/mm/yyyy HH:mm`); mensagens de erro da API MUST permanecer em português.
- **FR-012**: Auth, papéis e permissões de cursos, aulas, alunos, usuários e VOD (002) MUST permanecer; esta feature apenas acrescenta transmissão ao vivo da aula sem redesenhar CRUD existente nem a biblioteca VOD.
- **FR-013**: CI (build + testes) MUST continuar obrigatório a cada push; alterações de API/handlers de live MUST incluir testes automatizados relevantes.
- **FR-014**: Segredos de ingestão e chaves de streaming MUST permanecer fora do código e de commits; NÃO MUST ser expostos a alunos nem em URLs públicas permanentes indevidas. Endereço/dados de **playback** também NÃO MUST ser entregues a visitantes sem sessão; o gate de acesso em P1 é a API autenticada (JWT + leitura de aulas).
- **FR-015**: A jornada de leitura de estado / RBAC / erros MUST funcionar no ambiente local de desenvolvimento e no CI com **stub** (máquina de estados da live, permissões e mensagens em português), **sem** exigir sinal de vídeo real. Ingest/playback reais na nuvem MAY ser opcionais localmente desde que o comportamento sem sinal e o stub estejam documentados e testáveis.
- **FR-016** (P2/P3): A documentação ou automação SHOULD cobrir warm-up específico de streaming além do checklist 001, sem obrigar EventBridge apply/destroy completo se esse gap do 001 permanecer.
- **FR-017** (P2): A ficha da aula SHOULD exibir estados **agendada / ao vivo / encerrada** de forma inequívoca para o usuário.
- **FR-018**: Chat em tempo real, certificados, pipeline automático live→VOD, OAuth/Cognito, Redis AWS, NAT, domínio customizado, app nativo, multi-qualidade elaborada e DVR avançado além do mínimo do serviço NÃO MUST fazer parte desta feature.
- **FR-019**: A feature NÃO MUST reabrir decisões fechadas de 001/002 (destroy entre sessões, tasks com IP público, seed efêmero, VOD um-por-aula, sem página Biblioteca, sem download de VOD).
- **FR-020**: Após encerrar a live, a plataforma MAY indicar VOD publicado existente (002) como conteúdo sob demanda; NÃO MUST implementar gravação/conversão automática da live em VOD nesta feature.

### Key Entities

- **Transmissão ao vivo (live da aula)**: Associação entre uma aula e uma sessão de streaming assistível; possui estado (no mínimo ativa/não ativa; em P2 agendada/ao vivo/encerrada) e vínculo temporal com o horário da aula quando aplicável.
- **Sessão de streaming (recurso de demo)**: **Um** canal compartilhado de ingestão + reprodução na sessão AWS efêmera; destruído com o `destroy`; no máximo uma live ativa por vez.
- **Credencial de ingestão**: Endpoint + stream key do canal da sessão, exibidos na ficha da aula ao gestor; visível só a quem pode escrever aulas; estável até o destroy (sem rotação automática por live).
- **Aula (alvo)**: Recurso já existente; live sempre ligada a uma aula; horário recomendado mas não obrigatório na API.
- **Espectador**: Usuário autenticado com leitura de cursos/aulas assistindo in-app.
- **Gestor de live**: Admin ou professor autenticado.

### Constraints *(constitution)*

- Custo-consciente: IVS/streaming só na janela da sessão; destroy remove recursos cobráveis; sem NAT; sem Redis AWS; budget separado.
- Pronto para o pico: warm-up antes do horário; cold start na abertura inaceitável na demo.
- Escopo mínimo / YAGNI: sem chat, certificado, live→VOD, multi-qualidade/DVR avançado, app nativo.
- Segurança e RBAC: escrita live = escrita de aulas (admin + professor); JWT próprio; erros PT; segredos fora do git.
- Qualidade testável: CI verde; testes ao alterar API.
- IaC: Terraform enxuto estendendo `infra/`; `us-east-1`.
- Baseline 001/002: não reabrir.

### Out of Scope

- Chat WebSocket / tempo real (feature 004)
- Certificado de conclusão (feature 005)
- Pipeline live → VOD / gravação automática da transmissão
- Redesign ou reabertura de 001 e 002
- OAuth externo / Cognito
- Redis/ElastiCache na AWS
- NAT Gateway / domínio customizado
- Aplicativo nativo móvel
- Multi-qualidade elaborada / DVR avançado além do mínimo do serviço gerenciado
- Página/menu dedicado “Ao vivo” desconectado da ficha da aula (P1)
- Canal de streaming permanente fora do ciclo apply/destroy
- Múltiplos canais / lives simultâneas em aulas diferentes
- Exigir horário de aula como pré-condição rígida na API para iniciar live
- Fluxo P1 em dois passos (associar credenciais e só depois marcar ao vivo); “ao vivo” só quando o encoder conecta (P1 usa ação única + estado ao vivo mesmo sem sinal, com feedback de indisponibilidade)

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: No caminho feliz da demo AWS, um aluno autenticado abre a aula com live ativa e inicia a reprodução ao vivo in-app em ≤ 2 minutos (após login).
- **SC-002**: Após um admin ou professor iniciar a live numa aula (com sinal disponível), um aluno em outra sessão abre essa aula e inicia a reprodução em ≤ 2 minutos.
- **SC-003**: Em verificação amostral de RBAC, 100% das tentativas de gestão de live por aluno são recusadas com feedback em português; 100% das tentativas de assistir sem sessão válida são recusadas; admin e professor concluem iniciar/encerrar no caminho feliz.
- **SC-004**: Após o destroy documentado da sessão, não permanece recurso de streaming da demo gerando custo contínuo material (expectativa: ~US$ 0 de IVS/canal da sessão fora da janela, além dos resíduos já documentados no 001).
- **SC-005**: Um operador completa os passos extras de streaming no runbook (além do ciclo 001/002) em ≤ 15 minutos de leitura/execução incremental, sem inventar provisionamento permanente na console.
- **SC-006**: Com warm-up concluído conforme runbook **antes** da abertura, na T−0 da demo a API responde saudável e o recurso de streaming já está provisionado (zero dependência de “subir a stack no minuto da aula”).
- **SC-007**: No reprodutor na ficha da aula, o fluxo ao vivo começa a exibir vídeo/áudio em ≤ 15 segundos após o aluno confirmar “assistir”, em rede típica de apresentação (sinal já sendo enviado).
- **SC-008**: CI permanece verde (build + testes) no branch da feature antes da demo; regressão nos testes de live falha o pipeline.
- **SC-009**: Itens fora de escopo (chat, certificado, live→VOD, Cognito, Redis AWS, NAT, domínio custom, app nativo, redesign 001/002) não são entregues nem apresentados como parte desta feature.
- **SC-010**: Encerrar a live faz com que ≥ 95% dos alunos que reabrem/atualizam a ficha em seguida deixem de ver a live como ativa (sem precisar destroy).
- **SC-011**: Nenhum segredo de ingestão aparece em arquivos rastreados pelo controle de versão nem na UI apresentada ao perfil aluno; tentativa sem JWT de obter dados de playback da live é recusada.
- **SC-012**: Um avaliador consegue distinguir, na ficha da aula, live ativa vs ausência de live (e, se P2 entregue, agendada vs ao vivo vs encerrada) sem treinamento além da UI.

## Assumptions

- Features `001-aws-mvp-terraform` e `002-vod-library` estão concluídas e são baseline; esta feature só acrescenta streaming ao vivo.
- Constituição 1.0.0 permanece válida; esta feature não a emenda.
- Serviço de streaming ao vivo na AWS é **IVS**, conforme tabela de decisões do README (não reabrir escolha NGINX-RTMP self-hosted para a demo).
- Não há matrícula no produto: quem lê cursos/aulas pode assistir a live ativa da aula.
- Titularidade “professor dono do curso” não existe: professor gere live de qualquer aula, como no CRUD de aulas.
- Consumo: **somente reprodução in-app** na ficha da aula; sem download.
- Chat e certificado ficam para 004/005; menção a “após live, VOD” é apenas orientação futura / link se VOD 002 já existir.
- Warm-up P1 é checklist manual estendido do 001; automação EventBridge apply/destroy do 001 permanece gap salvo pedido explícito; automação só de streaming é opcional P2/P3.
- Desenvolvimento local prioriza RBAC, estados e contratos testáveis via **stub** (sem vídeo real); ingest/playback reais MAY depender da sessão AWS.
- Rebuild/publish do frontend por sessão (URL da API) permanece como no 001.
- Pico de milhares de espectadores é requisito de arquitetura/enunciado; a demo acadêmica valida o caminho IVS + warm-up, não necessariamente load test de milhares nesta feature.
- Datas BR e erros em português permanecem obrigatórios.
- **Canal único da sessão** + no máximo uma live ativa (clarificação 2026-09-05).
- **Ingest P1** = credenciais na ficha para OBS/encoder externo (clarificação 2026-09-05).
- **Horário** recomendado na UI, opcional na API (clarificação 2026-09-05).
- **Reinício após encerrar** na mesma sessão Terraform permitido (mesma aula ou outra), com no máximo uma live ativa (clarificação 2026-09-05).
- **Playback P1** só via API autenticada (JWT); sem URL pública permanente nem signed URL obrigatória (clarificação 2026-09-05).
- **Stream key** estável do canal da sessão até destroy; reinícios reutilizam a mesma key (clarificação 2026-09-05).
- **Local/CI** = stub de estados/RBAC/erros; sem sinal de vídeo real (clarificação 2026-09-05).
- **P1 ação única** “iniciar live” = associar ao canal + marcar ao vivo + mostrar ingest; “agendada” / dois passos = P2 (clarificação 2026-09-05).
