# Feature Specification: Chat da aula ao vivo (tempo real)

**Feature Branch**: `004-live-class-chat`

**Created**: 2026-09-05

**Status**: Draft

**Input**: User description: "Chat da aula ao vivo (tempo real) sobre 001–003 concluídos. Aluno/professor/admin autenticados enviam e recebem mensagens na sala por aula (preferencialmente na janela ao vivo); UI na ficha da aula; infra de chat na stack efêmera Terraform (apply→demo→destroy); RBAC mínimo alinhado aos papéis existentes; front+back ponta a ponta; CI; stub local/CI; custo baixo sem ElastiCache salvo necessidade; sem NAT; us-east-1; JWT própria. Fora: certificado (005), redesign 001–003, OAuth/Cognito, chat global, moderação avançada/anexos, histórico eterno entre destroys."

## Clarifications

### Session 2026-09-05

- Q: Quando a sala de chat fica disponível em P1? → A: **Sempre que a aula existe** e o usuário está autenticado com leitura de cursos/aulas — independente do estado da live (`ao_vivo` ou não).
- Q: Qual a retenção de mensagens em P1? → A: **Somente enquanto conectado** (efêmero na sessão de conexão); **sem** histórico ao reabrir a ficha. Persistência entre destroys de demos distintas continua fora; histórico recente na sessão fica fora do P1 (alinhado a esta escolha).
- Q: Desenho de transporte na AWS em P1? → A: **WebSocket no serviço de API já existente** da sessão (ECS Fargate da demo) — sem API Gateway WebSocket + Lambda separados no P1. Sem Redis/ElastiCache salvo necessidade demonstrada no plan.
- Q: No P1, como o ambiente local e o CI devem exercitar o chat? → A: **WebSocket real no Compose/local** (mesmo protocolo da sessão AWS, fan-out em memória); **CI** cobre testes de handlers/contratos/RBAC **sem** stack AWS. Não usar stub sem WebSocket no local; integração multi-cliente obrigatória no CI não é exigida.
- Q: Como o cliente autentica a conexão WebSocket do chat em P1? → A: **JWT na query do handshake** (`?token=...`); conexão sem token válido MUST ser rejeitada. Sem cookie HttpOnly nem mudança da auth baseline; plan MUST evitar logar a query completa com o token.
- Q: Qual o tamanho máximo de uma mensagem de texto do chat em P1? → A: **500 caracteres**; vazio ou só espaços rejeitados; acima do limite rejeitado com erro em português.
- Q: A entrega desta feature (plan/tasks) inclui moderação P2? → A: **Só P1** no escopo de entrega desta rodada; User Story 5 / FR-016 permanecem documentados como **P2 futuro**, sem tasks obrigatórias agora.
- Q: Como o cliente indica a qual aula a conexão de chat pertence? → A: **`aula_id` + JWT no handshake** (path e/ou query); aula inexistente ou sem permissão de leitura → rejeitar conexão. Sem `join` posterior nem multiplex multi-aula na mesma conexão em P1.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Aluno envia e recebe mensagens no chat da aula (Priority: P1)

Um aluno autenticado abre a ficha de **qualquer aula existente** e participa do **chat daquela aula**: envia mensagens de texto curtas e vê, em tempo quase real, mensagens de outros participantes da mesma sala — sem sair da ficha, sem app separado e sem chat global da plataforma. A disponibilidade do chat **não** depende da live estar `ao_vivo`.

**Why this priority**: O enunciado inclui chat da aula; sem envio/recebimento ponta a ponta na jornada curso → aula, a feature não demonstra interação na aula magna nem o pico de participação na abertura.

**Independent Test**: Dois alunos (ou aluno + professor) autenticados abrem a mesma aula; um envia mensagem; o outro a vê na UI do chat da ficha sem recarregar a página inteira de forma manual obrigatória.

**Acceptance Scenarios**:

1. **Given** um aluno autenticado e uma aula existente, **When** o aluno abre a ficha dessa aula (com ou sem live ativa), **Then** vê o painel de chat embutido (não uma aplicação separada), conecta-se à sala dessa aula via handshake com `aula_id` (e JWT), com identificação do contexto da aula/curso e datas em formato brasileiro quando exibidas.
2. **Given** dois usuários autenticados na mesma aula, **When** um envia uma mensagem de texto válida, **Then** o outro (ainda conectado à mesma sala) recebe/visualiza essa mensagem no chat da mesma aula em tempo quase real (sem precisar navegar para outro curso/aula).
3. **Given** um visitante sem sessão válida (ou token JWT ausente/inválido na conexão), **When** tenta conectar-se ao chat ou enviar mensagem, **Then** o acesso é negado com feedback claro em português.
4. **Given** um aluno autenticado na aula A, **When** envia mensagem, **Then** ela NÃO aparece no chat da aula B (sala isolada por aula).
5. **Given** falha de entrega ou conexão interrompida, **When** o aluno tenta enviar ou permanece na ficha, **Then** vê erro ou estado degradado compreensível em português — não silêncio absoluto sem indicação.
6. **Given** um aluno que enviou mensagens e depois **reabre** a ficha (nova conexão), **When** o painel de chat carrega, **Then** NÃO vê histórico das mensagens anteriores — apenas o que chegar enquanto estiver conectado de novo.

---

### User Story 2 - Professor e admin participam sem quebrar RBAC de live/VOD (Priority: P1)

Professor e admin autenticados também enviam e recebem mensagens no mesmo chat da aula. A participação no chat **não** concede ao aluno poderes de gestão de live (003) nem de VOD (002). Quem já podia gerir live/VOD continua podendo; quem só lia continua só lendo esses recursos.

**Why this priority**: Demo multi-perfil; regressão de RBAC quebraria a constitution e o trabalho já entregue.

**Independent Test**: Admin e professor postam no chat da aula; aluno posta no chat; aluno tenta ação de gestão live/VOD e continua recusado; UI do chat não expõe controles de ingestão/upload indevidos.

**Acceptance Scenarios**:

1. **Given** admin ou professor autenticado e sala disponível, **When** abre a ficha da aula, **Then** participa do chat (enviar/receber) no mesmo painel embutido.
2. **Given** um aluno que acabou de usar o chat, **When** tenta iniciar/encerrar live ou fazer upload/replace/delete de VOD, **Then** continua recusado (ou sem controles na UI) — chat NÃO amplia permissões de 002/003.
3. **Given** perfis distintos na mesma aula, **When** mensagens são exibidas, **Then** cada mensagem identifica de forma legível quem enviou pelo **nome** do usuário (campo já existente no produto), sem expor senhas ou tokens.

---

### User Story 3 - Operador inclui chat no ciclo efêmero apply → demo → destroy (Priority: P1)

O operador habilita o chat em tempo real **no serviço de API já existente** da sessão (WebSocket no backend da demo ECS), na **mesma** stack efêmera Terraform (001+002+003), documenta no runbook e, ao destruir a sessão, remove a API/task e quaisquer ajustes de infra do chat. Não adiciona API Gateway WebSocket + Lambda separados no P1. Custo fora da sessão ≈ 0. Budget da conta permanece intocado.

**Why this priority**: Constitution e 001 exigem destroy entre sessões; chat NÃO pode deixar serviço pago contínuo esquecido nem stack paralela desnecessária.

**Independent Test**: apply → publish → (warm-up se aplicável) → demo de chat na ficha → destroy; stack destruída; budget ainda ativo; sem recursos órfãos de chat separados.

**Acceptance Scenarios**:

1. **Given** a baseline 001–003, **When** o operador provisiona/publica esta feature conforme o runbook, **Then** o chat da demo funciona via o serviço de API da sessão na região (`us-east-1`), sem NAT Gateway, sem domínio customizado, sem Cognito, sem API Gateway WebSocket + Lambda no P1, e sem ElastiCache **salvo** justificativa explícita no plan (default: sem Redis AWS).
2. **Given** o fim da demo, **When** executa o destroy completo da sessão, **Then** a API/task e ajustes de chat da sessão são removidos com a stack; o alerta de orçamento (`infra/budget/`) permanece.
3. **Given** o runbook, **When** um avaliador o consulta, **Then** encontra passos de chat encaixados no ciclo apply → publish → warm-up → demo → destroy, **sem** reabrir o desenho de 001–003.

---

### User Story 4 - Ambiente local/CI testável sem sessão AWS (Priority: P1)

Desenvolvimento local (Compose/API) exercita o **mesmo WebSocket** da demo AWS (fan-out em memória no processo da API). O CI cobre handlers, contratos e RBAC com testes automatizados **sem** provisionar a stack AWS — sem exigir integração multi-cliente obrigatória no pipeline. Diferente do stub de vídeo do 003: aqui não há serviço de chat AWS separado a “simular”; o protocolo local é o real.

**Why this priority**: Constitution V (qualidade testável); sem caminho local real o chat só seria demonstrável na AWS.

**Independent Test**: No Compose, dois clientes (ou teste de handler) postam na mesma aula via WebSocket; no CI, testes de handlers/contratos/RBAC passam sem Terraform/AWS.

**Acceptance Scenarios**:

1. **Given** API local (Compose ou equivalente), **When** dois usuários autenticados conectam-se ao chat da mesma aula, **Then** envio/recebimento via WebSocket funciona com o mesmo protocolo esperado na AWS (sem stack AWS).
2. **Given** o pipeline de CI, **When** rodam os testes da feature, **Then** handlers/contratos/RBAC e erros em português são verificados **sem** provisionar AWS; integração multi-cliente no CI NÃO é obrigatória.
3. **Given** a documentação, **When** um avaliador a consulta, **Then** fica explícito: local = WebSocket real; CI = testes sem AWS; demo AWS = mesmo protocolo atrás do ALB da sessão.

---

### User Story 5 - Moderação simples na sala (Priority: P2 — fora da entrega desta rodada)

Admin e/ou professor **poderão** (em fase futura) apagar uma mensagem inadequada e/ou silenciar temporariamente um participante. **Fora do plan/tasks desta feature**: o “pronto” da rodada 004 é apenas P1 (enviar/receber + infra + local/CI). Histórico ao reabrir continua fora.

**Why this priority**: Melhora o enunciado de “chat básico”, mas não bloqueia o MVP demonstrável; YAGNI nesta rodada.

**Independent Test**: (Futuro) Gestor apaga mensagem visível a conectados; aluno não modera.

**Acceptance Scenarios** *(referência futura; não exigidos nesta entrega)*:

1. **Given** admin ou professor na sala (participantes ainda conectados), **When** remove uma mensagem, **Then** ela deixa de ser exibida como conteúdo ativo para os participantes conectados (feedback claro).
2. **Given** um aluno, **When** tenta moderar (apagar/silenciar), **Then** a ação é recusada com mensagem em português e a UI não oferece o controle.

---

### Edge Cases

- Sessão AWS destruída: conexões e endpoints de chat da sessão anterior são inválidos; novo `apply` começa sala(s) limpa(s).
- Stack esquecida ligada: alerta de billing do 001 continua sendo rede de segurança; esta feature NÃO MUST introduzir custo contínuo material fora do ciclo destroy.
- Conexão WebSocket sem JWT válido na query do handshake: rejeitada; visitante não participa.
- Handshake com `aula_id` inexistente ou sem permissão de leitura: conexão rejeitada; erro claro em português.
- Token JWT vazado em logs de access/proxy: plan/runbook MUST orientar a **não** registrar a query string completa do upgrade; segredos fora do git permanecem obrigatórios.
- Reabrir a ficha após sair: painel começa vazio (sem histórico); só mensagens novas enquanto conectado.
- Mensagem vazia, só espaços, ou com **mais de 500 caracteres**: rejeitada com erro em português.
- Reinício da task/API no meio da demo: conexões WebSocket caem; cliente informa falha ou reconecta; mensagens anteriores da conexão perdida NÃO MUST ser recuperadas (política efêmera).
- Rajada de mensagens / abuso leve: P1 MAY limitar tamanho/frequência de forma simples; rate limiting sofisticado e filtros ML são fora de escopo.
- Desconexão no meio da aula: cliente tenta recuperar ou informa falha; não corrompe salas de outras aulas.
- Pico na abertura: chat MUST estar preparado com a demo (warm-up/documentação); cold start do chat na T−0 da demo é inaceitável no caminho AWS.
- Expectativa de certificado ao fim da aula: fora (005).
- Expectativa de anexos, imagens, reações, threads: fora de P1 (P2+ se especificado depois).
- Chat entre usuários fora do contexto de uma aula (global/DM): fora de escopo.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: A plataforma MUST permitir que usuários autenticados com leitura de cursos/aulas (`admin`, `professor`, `aluno`) **enviem e recebam** mensagens de texto no **chat ligado a uma aula específica** (sala por aula), via painel embutido na **ficha da aula**. Em P1, cada mensagem MUST ter no máximo **500 caracteres**; mensagens vazias, só espaços ou acima do limite MUST ser rejeitadas com erro em português.
- **FR-002**: P1 NÃO MUST exigir página/app de chat separado nem chat global da plataforma. Mensagens de uma aula NÃO MUST vazar para outra. Em P1, cada conexão WebSocket MUST vincular-se a **uma** aula no **handshake** (`aula_id` em path e/ou query, junto com o JWT); aula inexistente ou sem permissão de leitura MUST rejeitar a conexão. Mensagem `join` posterior e multiplex de várias aulas na mesma conexão NÃO MUST fazer parte do P1.
- **FR-003**: Visitantes sem sessão válida NÃO MUST participar do chat. Em P1, a autenticação da conexão WebSocket MUST usar o **JWT na query do handshake** (`?token=...`), reutilizando o token já emitido pelo login (sem cookie HttpOnly nem OAuth). Conexão sem token válido MUST ser rejeitada; tentativas MUST resultar em feedback claro em português. O plan/runbook MUST evitar logging da query completa com o token.
- **FR-004**: Participar do chat NÃO MUST alterar permissões de gestão de live (003) nem de VOD (002). Escrita de live/VOD permanece admin+professor; aluno continua só leitura nesses recursos.
- **FR-005**: Datas exibidas no contexto do chat (quando houver) MUST usar formato brasileiro (`dd/mm/yyyy` ou `dd/mm/yyyy HH:mm`); erros da API/chat MUST estar em português.
- **FR-006**: O chat da demo MUST viver na **stack efêmera** Terraform existente (mesmo ciclo apply/destroy da API); custo contínuo esperado fora da sessão ≈ 0. Budget da conta permanece separado e intocado pelo destroy da demo.
- **FR-007**: Destino de demo MUST permanecer AWS gerenciada em **`us-east-1`**, sem NAT Gateway, sem domínio customizado, sem OAuth/Cognito. Auth permanece JWT + bcrypt do baseline. ElastiCache/Redis na AWS NÃO MUST ser adotado **salvo** necessidade demonstrada e justificada no plan desta feature.
- **FR-008**: Em P1, a sala de chat MUST estar disponível para qualquer **aula existente** quando o usuário está autenticado com leitura de cursos/aulas — **independentemente** do estado da live (`ao_vivo`, agendada, encerrada ou inativa).
- **FR-009**: Em P1, mensagens MUST ser **efêmeras à conexão**: visíveis aos participantes **enquanto conectados** à sala; ao reabrir a ficha (nova conexão) NÃO MUST haver histórico das mensagens anteriores. Persistência entre destroys de demos distintas NÃO MUST fazer parte desta feature. Histórico recente (mesmo só dentro da sessão Terraform) NÃO MUST ser exigido no P1.
- **FR-010**: Em P1, o transporte em tempo real na AWS MUST usar **WebSocket no serviço de API já existente** da sessão (backend na stack efêmera ECS). API Gateway WebSocket + Lambda **separados** NÃO MUST fazer parte do P1. Redis/ElastiCache na AWS NÃO MUST ser adotado salvo necessidade demonstrada e justificada no plan.
- **FR-011**: Ambiente local (Compose/API) MUST oferecer **WebSocket real** com o mesmo protocolo da sessão AWS (fan-out em memória no processo da API; `desired_count = 1`). O CI MUST cobrir testes de handlers/contratos/RBAC e erros em português **sem** stack AWS; integração multi-cliente obrigatória no CI NÃO MUST ser exigida. Stub sem WebSocket no local NÃO MUST ser o caminho P1 (diferente do stub de vídeo IVS do 003).
- **FR-012**: CI (build + testes) MUST permanecer obrigatório; alterações de API/handlers relacionadas ao chat MUST incluir testes automatizados relevantes.
- **FR-013**: Quaisquer ajustes de infra para o chat (ex.: suporte a upgrade WebSocket no balanceador, variáveis de ambiente) MUST permanecer no Terraform enxuto da demo, documentados no runbook (README / quickstart da feature), e destruídos com a sessão; sem módulos elaborados desnecessários. Warm-up: a API da sessão MUST estar saudável antes da abertura da demo (alinhado ao T−15 do 001).
- **FR-014**: Segredos e credenciais de conexão de chat MUST permanecer fora do código e de commits; NÃO MUST ser commitados em `.env` rastreados, `*.tfvars` de segredos ou state.
- **FR-015**: Auth, papéis e comportamentos de 001–003 (cursos, aulas, alunos, usuários, VOD, live) MUST permanecer; esta feature NÃO MUST redesenhar IVS, VOD nem a fundação AWS.
- **FR-016** (P2 futuro — **fora da entrega desta rodada**): Em fase posterior, admin e/ou professor SHOULD poder **apagar** mensagem e/ou **silenciar** participante na sala (para quem ainda está conectado); aluno NÃO MUST moderar. Plan/tasks da feature 004 NÃO MUST implementar FR-016.
- **FR-017**: Anexos ricos, reações, threads, chat global, DM, histórico ao reabrir, moderação por ML, filtros avançados, OAuth/Cognito, certificado (005), pipeline live→VOD, API Gateway WebSocket + Lambda no P1 e reabertura de 001–003 NÃO MUST fazer parte desta feature.

### Key Entities

- **Sala de chat da aula**: Canal de mensagens associado a **uma** aula; isolado de outras aulas; disponível para qualquer aula existente com usuário autenticado leitor (independente da live); vínculo estabelecido no **handshake** da conexão (`aula_id`).
- **Mensagem**: Texto curto (máx. **500 caracteres** em P1) enviado por um usuário autenticado; possui autor, conteúdo, instante (exibido em formato BR quando aplicável) e vínculo à sala/aula; ciclo de vida **efêmero à conexão** (sem histórico ao reabrir).
- **Participante**: Usuário autenticado (`admin` | `professor` | `aluno`) presente na sala da aula.
- **Moderação (P2)**: Ação de remover mensagem ou silenciar participante, restrita a perfis autorizados.
- **Sessão de demo**: Janela apply→destroy em que a API (e o chat nela) existem; após destroy, endpoints da sessão anterior não valem.

### Constraints *(constitution)*

- Custo-consciente: chat só na janela da sessão; destroy remove recursos cobráveis; sem NAT; sem Redis AWS por padrão; budget separado.
- Pronto para o pico: preparação/warm-up antes da abertura da aula na demo; cold start do chat na T−0 inaceitável no caminho AWS.
- Escopo mínimo / YAGNI: entrega desta rodada = **só P1** (postar/receber + WebSocket na API local/AWS + CI); moderação P2 documentada como futuro; sem histórico ao reabrir; sem ML/anexos; sem API GW WS+Lambda no P1.
- Segurança e RBAC: JWT próprio na query do handshake WebSocket; papéis existentes; erros PT; chat não amplia poderes de live/VOD; não logar token na query.
- Qualidade testável: CI verde; testes ao alterar API/handlers.
- IaC: Terraform enxuto estendendo a stack da demo; `us-east-1`.
- Baseline 001–003: não reabrir.

### Out of Scope

- Certificado de conclusão (feature 005)
- Redesign ou reabertura de 001 (AWS fundação), 002 (VOD), 003 (IVS live)
- OAuth externo / Cognito
- Domínio customizado / certificados ACM custom
- NAT Gateway
- ElastiCache/Redis na AWS (salvo justificativa forte no plan)
- Chat global da plataforma, DMs, canais fora da aula
- Moderação avançada, filtros ML, silenciamento em massa sofisticado
- Moderação simples (apagar/silenciar) — P2 futuro; **fora do plan/tasks desta rodada**
- Anexos ricos (imagem/arquivo), reações, threads, editar mensagem
- Histórico de mensagens ao reabrir a ficha (P1 = só enquanto conectado)
- Persistência de histórico entre destroys de demos distintas
- API Gateway WebSocket + Lambda como stack separada no P1 (README permanece referência histórica; decisão P1 = WebSocket na API ECS)
- App nativo / cliente de chat separado
- Load test obrigatório de milhares de conexões nesta feature (enunciado orienta arquitetura; demo valida caminho ponta a ponta)

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: No caminho feliz da demo AWS, um aluno autenticado abre uma aula existente e envia a primeira mensagem visível no chat embutido em ≤ 2 minutos (após login) — com ou sem live ativa.
- **SC-002**: Com dois participantes autenticados e conectados na mesma aula, a mensagem enviada por um torna-se visível para o outro em ≤ 5 segundos em rede típica de apresentação (caminho feliz, API aquecida).
- **SC-003**: Em verificação amostral, 100% das tentativas de participar do chat sem JWT válido (ausente/inválido na query do handshake) são recusadas com feedback em português; 100% das mensagens de aula A não aparecem na aula B no caminho feliz testado.
- **SC-004**: Após o destroy documentado da sessão, não permanece recurso da demo (incluindo a API/chat) gerando custo contínuo material (expectativa ~US$ 0 fora da janela, além dos resíduos já documentados no 001); sem stack órfã de API Gateway WebSocket + Lambda do P1.
- **SC-005**: Um operador completa os passos extras de chat no runbook (além do ciclo 001–003) em ≤ 15 minutos de leitura/execução incremental, sem provisionamento permanente ad hoc na console.
- **SC-006**: CI permanece verde (build + testes) no branch da feature; regressão nos testes de handlers/contratos/RBAC do chat falha o pipeline.
- **SC-007**: Em verificação de RBAC, aluno que usa o chat ainda não consegue gestão de live nem escrita de VOD; admin e professor concluem postagem no chat no caminho feliz.
- **SC-008**: Itens fora de escopo desta entrega (certificado, redesign 001–003, Cognito, chat global, ML/anexos, histórico ao reabrir, API GW WS+Lambda, **moderação P2**) não são entregues nem apresentados como parte do “pronto” desta rodada.
- **SC-009**: Com warm-up/preparação concluídos conforme runbook **antes** da abertura, na T−0 da demo a API (e o chat nela) já está saudável — sem “subir a stack no minuto da aula”.
- **SC-010**: Avaliador distingue, na ficha da aula, o painel de chat do player de live e do VOD, sem treinamento além da UI.
- **SC-011**: Após sair e reabrir a ficha da mesma aula, o painel inicia sem as mensagens da conexão anterior (política efêmera verificável na demo).
- **SC-012**: Mensagens com mais de 500 caracteres ou vazias/só espaços são rejeitadas no caminho de validação; mensagem válida ≤ 500 caracteres é aceita no caminho feliz.

## Assumptions

- Features `001-aws-mvp-terraform`, `002-vod-library` e `003-live-streaming-ivs` estão concluídas e são baseline; esta feature só acrescenta chat por aula.
- Constituição 1.0.0 permanece válida; esta feature não a emenda.
- Não há matrícula no produto: quem lê cursos/aulas pode entrar no chat de qualquer aula existente.
- UI do chat vive na ficha da aula (mesmo padrão de VOD/live), não em rota “Chat” dedicada no P1.
- **Sala sempre disponível** para aula existente + autenticado (clarificação 2026-09-05).
- **Retenção P1** = só enquanto conectado; sem histórico ao reabrir (clarificação 2026-09-05).
- **Transporte P1** = WebSocket no serviço de API ECS existente; não API Gateway WebSocket + Lambda no P1 (clarificação 2026-09-05). A tabela do README que cita API GW+Lambda fica como referência histórica / alternativa futura, não como decisão desta feature.
- Com `desired_count = 1` na demo, fan-out de mensagens em memória no processo da API é aceitável no P1; escala multi-task / Redis pub-sub fica fora salvo justificativa no plan.
- **Tamanho P1** = máximo **500 caracteres** por mensagem (clarificação 2026-09-05).
- **Auth WebSocket P1** = JWT na query do handshake (`?token=...`); rejeitar sem token válido; sem cookie HttpOnly; não logar query completa (clarificação 2026-09-05).
- **Local/CI** = WebSocket real no Compose/local; CI = handlers/contratos/RBAC sem AWS; sem stub sem-WS e sem multi-cliente obrigatório no CI (clarificação 2026-09-05).
- Pico de milhares de participantes no chat é requisito de arquitetura/enunciado; a demo acadêmica valida o caminho ponta a ponta + preparação, não necessariamente load test de milhares nesta feature.
- Rebuild/publish do frontend por sessão permanece como no 001.
- Datas BR e erros em português permanecem obrigatórios.
- **Handshake P1** = `aula_id` + JWT na conexão; uma aula por conexão; rejeitar aula inválida/sem leitura; sem join/multiplex (clarificação 2026-09-05).
- Autor da mensagem na UI P1 = **nome** do usuário (campo já existente); e-mail pode existir no token/perfil mas NÃO é o rótulo obrigatório no painel.
- **Entrega desta rodada** = **só P1**; moderação P2 (US5 / FR-016) documentada como futuro, sem tasks obrigatórias agora (clarificação 2026-09-05).
- Moderação P2 NÃO faz parte do MVP demonstrável desta rodada; P1 prioriza postar/receber + WebSocket na API (local e AWS).
