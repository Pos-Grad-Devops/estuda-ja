# Feature Specification: Biblioteca VOD (vídeos gravados sob demanda)

**Feature Branch**: `002-vod-library`

**Created**: 2026-09-04

**Status**: Draft

**Input**: User description: "Biblioteca VOD (vídeos gravados sob demanda) no caminho AWS, construída sobre a fundação 001-aws-mvp-terraform (P1–P3 concluída). Aluno/professor/admin listam e reproduzem conteúdo gravado associado a cursos/aulas; upload/gestão conforme RBAC; armazenamento/entrega na AWS com custo controlado; integração com o ciclo efêmero apply/destroy; front + back ponta a ponta; CI obrigatório; IaC enxuto. Fora de escopo: streaming ao vivo, chat, certificado, OAuth, Redis AWS, NAT, domínio customizado, redesign do 001."

## Clarifications

### Session 2026-09-04

- Q: Quais papéis podem subir/associar/remover vídeo gravado? → A: Admin **e** professor (espelha a escrita de aulas). Aluno não gere VOD. Não há “professor dono do curso”: qualquer professor autentificado pode gerir o VOD de qualquer aula, como já ocorre com o CRUD de aulas.
- Q: Qual é a associação mínima de um VOD ao domínio? → A: **Um** vídeo vigente por **aula**; o curso aparece só como contexto da aula. Publicar de novo na mesma aula substitui o vigente. Sem catálogo no curso independente da aula e sem vários vídeos simultâneos por aula.
- Q: Onde o aluno encontra e assiste no P1? → A: Só nas fichas **curso/aula já existentes** (indicar se há gravação e reproduzir ali). Sem página ou rota nova “Biblioteca”. A “biblioteca” desta feature é o conjunto de aulas que têm gravação publicada, descoberto na navegação atual.
- Q: Qual deve ser o teto máximo de tamanho do arquivo de vídeo na demo (publicação e seed)? → A: **50 MB**.
- Q: Qual formato de arquivo de vídeo a demo deve aceitar na publicação (e no seed)? → A: **Só MP4 (H.264/AAC)**; demais tipos rejeitados.
- Q: Depois que o aluno autenticado inicia a reprodução, por quanto tempo o acesso ao arquivo pode continuar válido sem nova checagem de sessão? → A: **~15 minutos**.
- Q: Ao substituir ou remover o VOD de uma aula durante a sessão, o arquivo antigo no armazenamento deve ser apagado na hora? → A: **Sim — apagar imediatamente** na troca/remoção.
- Q: Além de assistir no reprodutor da ficha da aula, o aluno pode baixar o arquivo de vídeo? → A: **Só reproduzir no produto** — sem botão/fluxo de download.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Aluno assiste a gravação na ficha da aula (Priority: P1)

Um aluno autenticado abre um curso e uma aula já existentes na plataforma, vê se aquela aula tem gravação publicada e a reproduz na própria ficha — sem transmissão ao vivo e sem uma tela nova de catálogo. Professor e admin, que já leem cursos e aulas, também assistem o mesmo conteúdo publicado nesse contexto.

**Why this priority**: Sem reprodução ponta a ponta na jornada que o avaliador já conhece (curso → aula), a feature não demonstra valor.

**Independent Test**: Com a plataforma no ar (local ou sessão AWS) e pelo menos uma aula com vídeo publicado (seed ou associação prévia), um aluno faz login, abre essa aula e inicia a reprodução até ver/ouvir o vídeo.

**Acceptance Scenarios**:

1. **Given** um aluno autenticado e uma aula visível com VOD publicado, **When** o aluno abre a ficha dessa aula (a partir do curso, como já navega hoje), **Then** identifica que há gravação, inicia a reprodução no próprio produto (reprodutor embutido; **sem** fluxo de download do arquivo) e vê título/identificação da aula/curso, com datas em formato brasileiro quando exibidas.
2. **Given** um professor ou admin autenticado, **When** abre a mesma aula com VOD publicado, **Then** também consegue ver e reproduzir (leitura de VOD segue a leitura de cursos/aulas).
3. **Given** um visitante sem sessão válida, **When** tenta obter ou reproduzir o VOD de uma aula, **Then** o acesso é negado de forma clara, em português, sem exposição de endereço interno de armazenamento.
4. **Given** uma aula sem gravação publicada, **When** o aluno abre essa aula, **Then** não vê um reprodutor vazio enganoso: a ausência de VOD é evidente (indicação de que ainda não há gravação).
5. **Given** um aluno autenticado que iniciou a reprodução, **When** decorrem cerca de 15 minutos sem nova autorização de acesso ao arquivo, **Then** o acesso temporário ao objeto expira e nova reprodução exige novamente usuário autenticado com permissão de leitura.

---

### User Story 2 - Admin ou professor publica a gravação da aula (Priority: P1)

Um admin ou professor envia um arquivo de vídeo gravado e o associa a **uma aula** (um vigente por aula). Um aluno passa a assistir na ficha dessa aula. O aluno não vê ações de publicação e é recusado se tentar pela API. Novo envio na mesma aula substitui o vídeo anterior.

**Why this priority**: Sem publicação, a demo depende só do seed; o avaliador precisa ver o ciclo “professor/admin sobe a gravação → aluno assiste na aula”.

**Independent Test**: Admin ou professor faz login, publica um vídeo numa aula, e um aluno (outra sessão) abre essa aula e reproduz; um aluno tenta a mesma escrita e é recusado.

**Acceptance Scenarios**:

1. **Given** um admin ou professor autenticado, **When** envia um vídeo gravado válido e o associa a uma aula, **Then** esse passa a ser o único VOD vigente da aula e fica disponível para leitura/reprodução na ficha da aula pelos perfis que já leem cursos/aulas.
2. **Given** um aluno autenticado, **When** tenta publicar, substituir ou remover o VOD de uma aula, **Then** a ação é recusada com mensagem em português e a interface da aula não oferece o formulário/botão correspondente.
3. **Given** um admin ou professor e uma aula que já tem gravação, **When** substitui ou remove a gravação, **Then** a ficha da aula passa a reproduzir o novo arquivo ou a indicar ausência de VOD, o arquivo antigo no armazenamento é **apagado imediatamente**, e o aluno não fica preso a um endereço antigo inválido.
4. **Given** um arquivo inválido (não MP4 H.264/AAC) ou com tamanho **acima de 50 MB**, **When** admin ou professor tenta publicar, **Then** recebe erro claro em português e a aula não passa a apresentar conteúdo quebrado como “publicado”.

---

### User Story 3 - Operador inclui VOD no ciclo efêmero apply → demo → destroy (Priority: P1)

O operador da conta acadêmica provisiona, na mesma sessão Terraform da demo 001, o armazenamento e a entrega necessários aos vídeos gravados; usa o runbook atualizado; ao destruir a sessão, não deixa biblioteca cobrável ociosa. Na subida seguinte, a aula de exemplo volta a ter um clipe publicado via seed (sem backup da sessão anterior).

**Why this priority**: A constitution e o 001 exigem destroy entre sessões e custo ~0 fora delas; VOD não pode furar esse contrato com um silo permanente “esquecido”.

**Independent Test**: Operador segue o runbook da sessão (apply → publicação da app → demo VOD na ficha da aula → destroy) e verifica que os recursos de VOD da sessão foram removidos; num novo apply, o aluno assiste o VOD de exemplo do seed nessa aula sem reenvio manual.

**Acceptance Scenarios**:

1. **Given** o código de infra da demo 001 como baseline, **When** o operador executa o provisionamento documentado desta feature, **Then** o armazenamento/entrega de VOD fica disponível na mesma região e sessão, sem NAT, sem cache gerenciado extra e sem domínio customizado.
2. **Given** o término da janela de demo, **When** o operador executa o destroy completo já usado no 001, **Then** objetos e recursos de VOD da sessão são removidos com a stack (custo contínuo esperado de armazenamento VOD fora de sessão ≈ 0), e o alerta de orçamento da conta permanece (stack de budget intocada).
3. **Given** um novo `apply` após destroy, **When** a aplicação sobe com seed/migração, **Then** a aula de demo tem exatamente um VOD de exemplo publicado, suficiente para a User Story 1 sem upload manual.
4. **Given** o runbook da demo, **When** um avaliador o consulta, **Then** encontra os passos extras de VOD (se houver) encaixados no ciclo apply → publish → demo → destroy, sem reabrir o desenho do 001.

---

### User Story 4 - Metadados e estado rascunho/publicado (Priority: P2)

O gestor (admin ou professor) informa título (e duração, quando disponível) da gravação da aula e pode manter o vídeo em rascunho até decidir publicar. Alunos só vêem e reproduzem o vigente **publicado** na ficha da aula. Se P1 entregar “upload = já visível”, esta história acrescenta controle fino sem bloquear a demo mínima.

**Why this priority**: Melhora a gestão pedagógica, mas a demo acadêmica já vale com publicação imediata + seed.

**Independent Test**: Admin ou professor salva um rascunho na aula; aluno abre a aula e não reproduz; gestor publica; aluno passa a ver título (e duração se houver) e reproduzir.

**Acceptance Scenarios**:

1. **Given** um admin ou professor, **When** associa um vídeo em estado rascunho à aula, **Then** alunos não reproduzem esse item na ficha da aula; o gestor ainda o encontra na gestão daquela aula.
2. **Given** um item em rascunho na aula, **When** o gestor o publica, **Then** o aluno passa a reproduzir na ficha da aula, vendo título e, quando informado, duração.
3. **Given** um item publicado na aula, **When** o gestor volta a rascunho ou remove, **Then** o aluno deixa de reproduzir aquele conteúdo nessa aula.

---

### Edge Cases

- Sessão AWS destruída: URLs e arquivos da sessão anterior são inválidos; o produto NÃO MUST depender deles após novo `apply` (vale o seed novo).
- Stack esquecida ligada: o alerta de billing do 001 continua sendo a rede de segurança; VOD NÃO MUST introduzir recurso cobrável 24/7 fora do destroy.
- Arquivo corrompido, **não MP4 (H.264/AAC)**, ou tamanho **acima de 50 MB**: publicação recusada ou a aula não marcada como assistível; mensagem em português.
- Exclusão da aula (ou do curso que a contém): o VOD vigente dessa aula não permanece reproduzível nem “órfão” na UI de outra aula; o arquivo associado MUST ser **apagado do armazenamento** junto com a remoção do vínculo.
- Substituição ou remoção do VOD: o objeto antigo MUST ser **apagado imediatamente** do armazenamento da sessão (não deixar órfãos até o destroy).
- Dois gestores publicando na mesma aula ao mesmo tempo: o aluno vê um só vídeo vigente (o último válido), nunca dois players para a mesma aula.
- Ambiente local sem conta AWS: ver a gravação na ficha da aula, publicar e assistir MUST continuar possíveis no fluxo de desenvolvimento local (equivalente de armazenamento local), para CI e desenvolvimento não dependerem da nuvem.
- Tentativa de usar a gravação da aula como transmissão ao vivo, chat ou certificado: recusada / fora desta feature.
- Acesso de reprodução concedido a usuário autenticado: válido por cerca de **15 minutos**; após isso, nova autorização é necessária. URL/endereço de armazenamento NÃO MUST permanecer utilizável publicamente de forma permanente.
- Aluno autenticado acessando aula sem VOD publicado, VOD inexistente ou (em P2) só rascunho: ausência clara ou “não encontrado” em português, sem vazar rascunhos.
- Expectativa de uma página “Biblioteca” no menu: fora do P1 desta feature; descoberta é só curso → aula.
- Pedido de download do arquivo de vídeo: fora desta feature; apenas reprodução no reprodutor da ficha da aula.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: A plataforma MUST permitir que usuários autenticados com permissão de leitura de cursos/aulas (admin, professor e aluno, como no baseline 001) vejam, na ficha da aula (e no contexto do curso pai), se existe VOD **publicado** vigente.
- **FR-002**: A plataforma MUST permitir que esses mesmos leitores **reproduzam** o VOD publicado **na ficha da aula** (reprodutor no produto), com identificação da aula/curso. P1 NÃO MUST exigir página, rota ou item de menu “Biblioteca”. O acesso ao arquivo para reprodução MUST ser **temporário (~15 minutos)** após autorização de um usuário autenticado com permissão de leitura; NÃO MUST haver URL/endereço interno permanente utilizável sem autenticação. A plataforma NÃO MUST oferecer botão ou fluxo de **download** do arquivo de vídeo (apenas reprodução no produto).
- **FR-020**: A experiência VOD MUST ser **somente reprodução in-app** na ficha da aula. Download explícito do MP4 (para aluno, professor ou admin) NÃO MUST fazer parte desta feature.
- **FR-003**: A gestão (envio, associação, substituição e remoção) do VOD da aula MUST ser permitida a **admin e professor** (mesmo recorte da escrita de aulas). Aluno NÃO MUST gerir VOD. Tentativas não autorizadas MUST ser recusadas com mensagem em português; a UI da aula MUST ocultar ações de escrita quando o perfil não pode. Qualquer professor autenticado MAY gerir o VOD de qualquer aula (não há dono de curso no modelo atual).
- **FR-004**: Cada aula MUST ter no máximo **um** VOD vigente. O curso NÃO MUST ter biblioteca própria independente da aula. Novo envio válido na mesma aula MUST substituir o vigente **e apagar imediatamente** o arquivo anterior no armazenamento. Remoção explícita do VOD ou exclusão da aula/curso MUST apagar o arquivo associado. Não criar catálogo paralelo desconectado de aula nesta feature.
- **FR-005**: A descoberta e a reprodução P1 MUST ocorrer apenas nas telas de curso/aula já existentes: o aluno chega à gravação abrindo a aula; a listagem de cursos/aulas atual continua sendo o caminho. Página dedicada “Biblioteca” fica fora desta feature salvo pedido futuro.
- **FR-006**: Armazenamento e entrega do arquivo gravado no caminho de demo MUST usar armazenamento de objetos na AWS, com rede de distribuição quando isso não aumente custo contínuo fora da sessão (caminho constitucional: objetos + CDN na mesma região da demo). Transmissão ao vivo NÃO MUST ser usada.
- **FR-007**: Os recursos de VOD da demo MUST fazer parte da **stack efêmera** do 001: `destroy` remove armazenamento/entrega/objetos da sessão; NÃO MUST haver bucket/biblioteca permanente fora dessa stack. A cada `apply`, seed/migração MUST recriar exatamente um VOD de exemplo publicado na aula de demo (arquivo pequeno, adequado a demo). Preservar gravações reais entre sessões está fora de escopo.
- **FR-008**: A jornada abrir aula / publicar (admin ou professor) / assistir MUST funcionar no ambiente local de desenvolvimento (sem exigir conta AWS), com o mesmo RBAC, datas brasileiras e erros em português.
- **FR-009**: Datas exibidas ou aceitas nesta feature MUST permanecer no formato brasileiro (`dd/mm/yyyy` ou `dd/mm/yyyy HH:mm`); mensagens de erro da API MUST permanecer em português.
- **FR-010**: Auth, papéis e permissões de cursos, aulas, alunos e usuários MUST permanecer os do baseline 001; esta feature apenas acrescenta o VOD vigente da aula sem redesenhar CRUD existente.
- **FR-011**: Destino de demo MUST continuar AWS gerenciada em **`us-east-1`**, hostnames gerados, sem NAT Gateway, sem Redis/cache gerenciado extra, sem domínio customizado e sem OAuth externo.
- **FR-012**: Infraestrutura nova de VOD MUST ser declarada no código de infra já versionado da demo (extensão enxuta, sem módulos elaborados) e destruída com o restante da sessão.
- **FR-013**: CI (build + testes) MUST continuar obrigatório a cada push; alterações de API/handlers de VOD MUST incluir testes automatizados relevantes. CD opcional do 001 MAY permanecer como está (não é pré-requisito desta feature).
- **FR-014**: Segredos de acesso ao armazenamento (se houver) MUST permanecer fora do código e de commits, no mesmo espírito do 001.
- **FR-015**: A plataforma MUST rejeitar publicação de arquivo de vídeo que **não seja MP4 (H.264 + AAC)** ou com tamanho **acima de 50 MB** (teto único para demo AWS e ambiente local), com mensagem em português. O seed MUST ser um MP4 **abaixo** desse teto. Limite de duração explícito NÃO é obrigatório no P1 se o teto de tamanho for aplicado.
- **FR-019**: O formato aceito nesta feature MUST ser **apenas MP4 com vídeo H.264 e áudio AAC** (extensão `.mp4` / tipo correspondente). Outros containers ou codecs NÃO MUST ser aceitos na publicação; NÃO MUST haver transcodificação no servidor.
- **FR-016** (P2): A plataforma SHOULD permitir título e duração como metadados visíveis ao aluno na ficha da aula quando houver VOD publicado, e estados rascunho vs publicado, de forma que rascunho não seja reproduzível pelo aluno. P1 MAY tratar “envio bem-sucedido = publicado e vigente da aula”.
- **FR-017**: Documentação operacional (README e/ou runbook da demo) MUST descrever: onde o VOD vive na sessão, o que o seed recria na aula de demo, o que o destroy apaga, e que streaming ao vivo / chat / certificado / página Biblioteca dedicada não fazem parte desta feature.
- **FR-018**: Streaming ao vivo, chat em tempo real, certificados de conclusão, OAuth externo, Redis na AWS, NAT Gateway, domínio customizado, página Biblioteca dedicada, **download explícito do vídeo**, e redesign da fundação 001 NÃO MUST fazer parte desta feature.

### Key Entities

- **Conteúdo VOD**: O único vídeo gravado vigente de uma aula, assistível após publicação, identificado pelo título da aula (e curso pai).
- **Publicação VOD**: Ato de tornar o arquivo gravado o vigente reproduzível da aula; em P1 pode coincidir com o envio; em P2 distingue-se de rascunho.
- **Biblioteca (conceito, não tela)**: Conjunto das aulas que têm gravação publicada; o aluno “percorre a biblioteca” ao navegar cursos e aulas — sem tela própria nesta feature.
- **Aula (alvo de associação)**: Recurso já existente no 001; cada aula tem zero ou um VOD vigente. O curso só contextualiza a aula.
- **Objeto de sessão**: Arquivo e metadados da demo que existem só enquanto a stack da sessão está no ar; recriados pelo seed após destroy/apply.
- **Gestor de VOD**: Admin ou professor autenticado.

### Constraints *(constitution)*

- Custo-consciente: sem armazenamento VOD cobrável ocioso 24/7; destroy da demo inclui VOD; clipes de seed/demo **≤ 50 MB**; sem NAT; sem Redis AWS; budget da conta permanece separado.
- Pronto para o pico: esta feature NÃO cobre a janela de transmissão ao vivo; VOD NÃO MUST ser usado como substituto de streaming ao vivo nem exigir aquecimento extra além do runbook 001.
- Escopo mínimo / YAGNI: sem transcodificação; só MP4 H.264+AAC; sem CDN extra permanente; sem catálogo desconectado da aula; sem página Biblioteca; sem as exclusões listadas.
- Segurança e RBAC desde o início; erros e comunicação em português; escrita VOD = escrita de aulas (admin + professor); acesso de reprodução temporário (~15 min).
- Qualidade testável: CI verde; testes ao alterar API.
- IaC: Terraform enxuto estendendo a infra da demo; `us-east-1`.
- Baseline 001: não reabrir decisões já fechadas (RDS mínimo, tasks com IP público, rebuild do front por sessão, dados efêmeros, alerta de billing).

### Out of Scope

- Streaming ao vivo / IVS (ou equivalente)
- Chat em tempo real / WebSocket
- Certificados de conclusão
- OAuth externo
- Redis/ElastiCache na AWS
- NAT Gateway / domínio customizado / certificado em domínio próprio
- Redesign ou reabertura da feature 001 (CRUD, auth, ciclo apply/destroy, CD opcional)
- Bucket ou biblioteca VOD **permanente** que sobreviva ao `destroy` da demo
- Backup/restore de gravações entre sessões
- Matrícula/enrolment (não existe no baseline; visibilidade = leitura de cursos/aulas)
- Página, rota ou menu **Biblioteca** dedicada (P1 e esta feature); vários VOD por aula; biblioteca no curso independente da aula
- Download explícito do arquivo de vídeo (botão/fluxo de baixar o MP4)
- Pipeline de transcodificação / múltiplas qualidades / legendas / formatos além de MP4 H.264+AAC
- Aplicativo nativo móvel dedicado
- Dono de curso / VOD restrito “só às aulas do professor” (exigiria modelo de titularidade inexistente)

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: No caminho feliz, um aluno autenticado abre a aula de demo (via curso, na navegação já existente) e inicia a reprodução do VOD publicado em ≤ 2 minutos (local ou demo AWS com seed).
- **SC-002**: Após um admin ou professor publicar um vídeo válido numa aula, um aluno em outra sessão abre essa aula e inicia a reprodução em ≤ 2 minutos.
- **SC-003**: Em verificação amostral de RBAC, 100% das tentativas de escrita de VOD por aluno são recusadas com feedback em português; 100% das tentativas de leitura/reprodução sem sessão válida são recusadas; admin e professor concluem publicação na aula no caminho feliz.
- **SC-004**: Após o destroy documentado da sessão, não permanece armazenamento de VOD da demo gerando custo contínuo material (expectativa alinhada ao 001: ~US$ 0 de compute/banco/balanceador **e** de objetos VOD da sessão, salvo resíduos já documentados no 001, p.ex. imagens de registro).
- **SC-005**: Após um `apply` limpo (pós-destroy) e publicação da app da sessão, um avaliador faz login como aluno seedado, abre a aula de exemplo e assiste o VOD **sem** upload manual prévio.
- **SC-006**: No reprodutor na ficha da aula, o vídeo de exemplo (arquivo curto) começa a tocar em ≤ 10 segundos após o aluno confirmar “assistir”, em rede típica de apresentação.
- **SC-007**: CI permanece verde (build + testes) no branch da feature antes da demo; regressão nos testes de VOD falha o pipeline.
- **SC-008**: Um operador completa os passos extras de VOD no runbook (além do ciclo 001 que já conhece) em ≤ 10 minutos de leitura/execução incremental, sem inventar provisionamento na console.
- **SC-009**: Itens fora de escopo (ao vivo, chat, certificado, OAuth, Redis AWS, NAT, domínio custom, persistência entre sessões, página Biblioteca, vários VOD por aula, download do arquivo, redesign 001) não são entregues nem apresentados como parte desta feature.
- **SC-010**: A jornada abrir aula + assistir (e publicar, para admin/professor) completa com sucesso no ambiente local documentado, sem conta AWS.
- **SC-011**: Nenhum segredo de acesso a armazenamento aparece em arquivos rastreados pelo controle de versão.
- **SC-012**: Um acesso de reprodução autorizado a usuário autenticado deixa de funcionar após cerca de **15 minutos** sem nova autorização; visitantes sem sessão nunca obtêm acesso utilizável ao objeto.

## Assumptions

- A feature `001-aws-mvp-terraform` está concluída e é a baseline: CRUD, auth JWT + RBAC, front/back, RDS, S3+CloudFront do site, ECS+ALB+CloudFront da API, seed, runbook apply/destroy, CI, CD opcional, `us-east-1`.
- Constituição 1.0.0 permanece válida; esta feature não a emenda.
- Não há matrícula no produto atual: quem lê cursos/aulas lê o VOD publicado da aula.
- **Ciclo de vida dos objetos (decisão desta spec):** efêmero + re-seed. Recursos de VOD entram na stack da sessão e saem no `destroy`. Não há bucket separado permanente. Motivo: constitution (custo ~0 fora de sessão, dados de demo descartáveis) e clarificações do 001 (destroy completo; seed a cada apply). Dentro da sessão, substituição/remoção MUST apagar o objeto antigo **na hora** (além do destroy final).
- O seed inclui um arquivo de vídeo **curto e pequeno** (obrigatoriamente **≤ 50 MB**, preferencialmente bem abaixo) como único VOD vigente da aula de exemplo já existente, para SC-001 e SC-005.
- Teto de upload/publicação: **50 MB** por arquivo (FR-015); aplicável na demo AWS e no fluxo local.
- P1 pode tratar envio bem-sucedido como publicado e vigente da aula; rascunho/duração ricos são P2 (FR-016).
- Formato aceito na demo: **apenas MP4 (H.264 + AAC)**; sem pipeline de transcodificação; demais formatos rejeitados na publicação.
- Desenvolvimento local usa um equivalente de armazenamento (disco ou serviço local), não a conta AWS.
- Rebuild/publish do frontend por sessão (URL da API) permanece como no 001; VOD não muda esse acoplamento.
- Carga de milhares de espectadores simultâneos é requisito da **aula ao vivo** (fora desta feature); a demo VOD valida dezenas de avaliadores, não o pico magna.
- Titularidade “professor dono do curso” não existe: professor gere VOD de qualquer aula, como no CRUD de aulas.
- Acesso de reprodução ao arquivo: janela temporária de **~15 minutos** após autorização (não vínculo permanente ao storage).
- Nome de produto “biblioteca VOD” descreve a capacidade (gravações sob demanda nas aulas), não uma tela a construir nesta feature.
- Consumo do VOD: **somente reprodução** na ficha da aula; sem download do arquivo.
