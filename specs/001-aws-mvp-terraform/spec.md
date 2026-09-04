# Feature Specification: MVP AWS com custo controlado

**Feature Branch**: `001-aws-mvp-terraform`

**Created**: 2026-09-04

**Status**: Draft

**Input**: User description: "MVP ponta a ponta na AWS com custo controlado (Terraform) — endurecer entrega existente (front + back + auth + persistência), caminho de deploy AWS de baixo custo (frontend estático + API em containers gerenciados + PostgreSQL gerenciado + IaC versionada + segredos fora do código), estratégia Free Tier/scale-to-zero/warm-up para aula simulada, CI obrigatório com CD faseado, demo acadêmica sem streaming/chat."

## Clarifications

### Session 2026-09-04

- Q: Fora das janelas de demo ou aula simulada, a infra AWS deve ser totalmente destruída via Terraform e recriada só quando necessário, ou pode permanecer ligada com a menor capacidade possível? → A: Destroy completo (`terraform destroy`) após cada uso; `apply` só quando for demo/aula
- Q: Depois de destruir a infra, os dados da demo (cursos, aulas, usuários de teste) devem ser recriados como em cada novo `apply`, ou a equipe precisa preservar dados entre sessões? → A: Dados efêmeros: a cada `apply`, seed/migração recria admin e dados de demo
- Q: A API em containers pode ter IP público (com firewall restrito) para evitar o custo de um NAT Gateway, ou a equipe exige tasks só em rede privada com NAT? → A: Tasks com IP público permitido; sem NAT Gateway no MVP
- Q: Como o frontend deve descobrir a URL da API depois de cada `apply`, já que os hostnames gerados pela AWS mudam quando a stack é recriada? → A: Em cada sessão: após `apply`, rebuild/publish do frontend com a URL da API daquela sessão
- Q: A feature deve exigir alerta de orçamento/billing na conta AWS (aviso antes dos créditos acabarem), ou isso fica só como dica opcional no guia? → A: Obrigatório: alerta de billing/orçamento documentado e configurado (ex. limiar baixo em USD)

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Demo acadêmica na plataforma hospedada (Priority: P1)

Admin, professor e aluno acessam a plataforma já em execução no destino de nuvem gerenciada e realizam as operações do MVP conforme seus papéis: login, leitura/escrita de cursos e aulas, e gestão de alunos/usuários (quando permitido). A interface e a API permanecem coerentes (mesmo ambiente de demo).

**Why this priority**: Sem a jornada ponta a ponta demonstrável, a feature não cumpre o objetivo acadêmico nem valida o caminho AWS.

**Independent Test**: Com a stack já provisionada e a aplicação publicada, um avaliador faz login com cada perfil e executa pelo menos uma operação permitida e uma bloqueada pelo RBAC.

**Acceptance Scenarios**:

1. **Given** a plataforma publicada em URLs HTTPS acessíveis, **When** um admin faz login com credenciais válidas, **Then** obtém acesso às funções administrativas (gestão de usuários/alunos e CRUD de cursos/aulas) e vê datas no formato brasileiro.
2. **Given** um professor autenticado, **When** tenta criar ou editar uma aula e apenas ler cursos, **Then** as ações permitidas pelo RBAC sucedem e as proibidas são recusadas com mensagem em português.
3. **Given** um aluno autenticado, **When** lista cursos e aulas, **Then** visualiza o conteúdo permitido e não consegue criar, editar ou excluir recursos restritos.
4. **Given** credenciais inválidas ou sessão expirada, **When** o usuário tenta uma ação protegida, **Then** o acesso é negado de forma clara e sem exposição de segredos.

---

### User Story 2 - Provisionamento reproduzível da infra mínima (Priority: P1)

Um operador da equipe, a partir do repositório, provisiona a infraestrutura mínima necessária para hospedar frontend estático, API em containers orquestrados gerenciados, balanceador, banco PostgreSQL gerenciado e armazenamento de segredos — usando infraestrutura como código versionada, sem depender de cliques manuais permanentes na console.

**Why this priority**: A constitution exige IaC e proíbe provisionamento manual permanente; sem isso não há demo AWS reproduzível nem controle de custo documentável.

**Independent Test**: Em conta AWS vazia (ou dedicada ao projeto), o operador aplica o código de infra uma vez e obtém endpoints utilizáveis; uma segunda aplicação (idempotente) não exige recriação manual.

**Acceptance Scenarios**:

1. **Given** credenciais de conta AWS válidas e o código de infra no repositório, **When** o operador executa o provisionamento documentado, **Then** frontend estático (CDN), API (containers + balanceador), banco PostgreSQL gerenciado e mecanismo de segredos ficam disponíveis sem passos manuais permanentes na console.
2. **Given** a infra já provisionada, **When** o operador reaplica o mesmo código sem alterações materiais, **Then** o resultado é previsível (sem drift exigindo “clique” para corrigir) e o estado permanece versionado.
3. **Given** a aplicação implantada sobre essa infra, **When** o operador acessa a URL pública do frontend e a URL da API, **Then** ambas respondem via HTTPS e o frontend consegue consumir a API no mesmo ambiente de demo — desde que o frontend tenha sido publicado com a URL da API **daquela** sessão.
4. **Given** um novo `apply` após destroy, **When** a API sobe e as migrações/seed rodam, **Then** admin e dados de demo necessários à apresentação já existem (sem depender de backup da sessão anterior).

---

### User Story 3 - Endurecimento da entrega local/CI como baseline (Priority: P1)

A equipe mantém a entrega ponta a ponta já existente (CRUD, auth JWT + RBAC, front, back, persistência) funcional localmente e com CI de build + test verde, aplicando apenas os ajustes necessários para a app ser configurável e implantável no alvo AWS (ex.: URLs, segredos, variáveis de ambiente) sem reinventar o produto.

**Why this priority**: O código atual é a baseline; falha local/CI invalida a demo e viola qualidade testável da constitution.

**Independent Test**: Pipeline CI executa build e testes com sucesso; `docker compose` (ou fluxo local documentado) sobe a app e permite login + listagem de cursos.

**Acceptance Scenarios**:

1. **Given** um push no repositório, **When** o CI roda, **Then** build e testes automatizados passam (falha de teste/handler bloqueia o verde).
2. **Given** o ambiente local documentado, **When** a equipe sobe a stack, **Then** login e CRUD conforme RBAC funcionam como hoje, com datas BR e erros em português.
3. **Given** a necessidade de apontar o frontend para a API hospedada, **When** a configuração de ambiente é definida fora do código-fonte versionado com segredos, **Then** a mesma base de código serve local e nuvem sem hardcode de credenciais.
4. **Given** um novo ciclo após destroy (URLs da API potencialmente novas), **When** o operador republica o frontend com a URL da API da sessão, **Then** o browser consome a API correta sem depender de hostname da sessão anterior.

---

### User Story 4 - Ciclo apply → warm-up → destroy (Priority: P2)

Antes de uma demo ou “aula simulada”, o operador sobe a stack completa via Terraform (`apply`) com antecedência suficiente para banco e API ficarem saudáveis (warm-up operacional); durante a janela, a capacidade mínima documentada evita cold start. Ao terminar, o operador destrói a stack completa (`destroy`) para zerar custo contínuo. Automação de agendamento na nuvem pode ficar documentada/parcial nesta feature.

**Why this priority**: Atende “pronto para o pico” e custo-consciente com conta Free de créditos limitados; automação total pode ser faseada sem bloquear a demo P1.

**Independent Test**: Operador segue o runbook apply → verificação de saúde → (opcional) ajuste de capacidade → destroy; confere na console que recursos cobráveis foram removidos.

**Acceptance Scenarios**:

1. **Given** uma aula simulada ou demo agendada, **When** o operador executa `apply`, publica a API, **rebuild/publish do frontend com a URL da sessão** e conclui o warm-up operacional com antecedência acordada, **Then** a API responde saudável e a UI hospedada consome essa API na abertura da janela sem depender de cold start.
2. **Given** o término da janela, **When** o operador executa o destroy completo documentado, **Then** não permanece infra cobrável ociosa (API, balanceador, banco e correlatos removidos conforme o código de infra).
3. **Given** que a automação completa de agendamento ainda não está implementada, **When** um avaliador consulta a documentação, **Then** encontra o ciclo apply/warm-up/destroy e o que é P2 (automação) versus já entregue.

---

### User Story 5 - Pipeline de deploy contínuo (Priority: P3)

Após P1 (IaC + app deployável), a equipe pode automatizar a publicação da aplicação (imagem/API e assets do frontend) a partir do CI, de forma que um merge aprovado atualize o ambiente de demo sem passos manuais recorrentes.

**Why this priority**: CD completo agrega valor operacional, mas a constitution permite fasear deploy automatizado; P1 já exige app implantável e CI de qualidade.

**Independent Test**: Com CD habilitado, um commit na branch de deploy atualiza frontend e/ou API no ambiente; sem CD, o caminho manual/documentado de publicação ainda funciona (P1).

**Acceptance Scenarios**:

1. **Given** CI verde e credenciais de publicação configuradas no pipeline, **When** um artefato aprovado é promovido, **Then** a versão nova da API e/ou do frontend fica disponível nas URLs de demo sem republicação manual ad hoc.
2. **Given** falha de build ou teste, **When** o pipeline tenta promover, **Then** o deploy é bloqueado e o ambiente de demo permanece na versão anterior estável.

---

### Edge Cases

- Conta AWS sem Free Tier / limites acadêmicos esgotados: o operador deve conseguir seguir com tamanho mínimo pago, com custo estimado atualizado na documentação.
- Falha parcial de provisionamento (ex.: banco sobe, API não): o estado deve ser recuperável via reaplicação do IaC ou rollback documentado, sem deixar segredos em arquivos locais commitados.
- Segredo rotacionado ou ausente: a API não inicia “pela metade”; falha de forma explícita e logs mínimos permitem diagnóstico.
- Frontend publicado apontando para API errada ou HTTP inseguro: a configuração de ambiente deve tornar o descompasso detectável na verificação pós-deploy.
- Pico simulado sem warm-up / sem `apply` antecipado: o runbook deve deixar claro o risco de cold start e que isso viola o critério de preparação para pico.
- Esquecer o `destroy` após a demo: custo contínuo volta a consumir créditos; o runbook MUST incluir verificação pós-destroy na console e o alerta de billing MUST notificar se o gasto subir sem sessão planejada.
- Expectativa de dados da sessão anterior após novo `apply`: inválida — o banco é efêmero; apenas o que o seed/migração recria estará disponível.
- Uso de NAT Gateway “por padrão de produção”: rejeitado nesta feature por custo; rede MUST seguir o modelo sem NAT.
- Frontend ainda apontando para URL de API de sessão anterior após novo `apply`: falha de integração; o runbook MUST exigir rebuild/publish com a URL nova antes da demo.
- Tentativa de usar self-hosted como destino de demo/produção: fora de escopo; apenas Docker Compose local para desenvolvimento.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: A plataforma MUST continuar oferecendo login com autenticação própria (credencial + sessão baseada em token) e RBAC `admin`, `professor`, `aluno` na API e na interface, sem introduzir OAuth externo nesta feature.
- **FR-002**: A plataforma MUST manter CRUD de cursos, aulas e alunos (e gestão de usuários para admin) conforme as permissões já definidas, como baseline — sem redesenhar o domínio de negócio.
- **FR-003**: Datas exibidas e aceitas na API/UI MUST permanecer no formato brasileiro (`dd/mm/yyyy` ou `dd/mm/yyyy HH:mm`); mensagens de erro da API MUST permanecer em português.
- **FR-004**: O destino de demo/produção desta feature MUST ser nuvem AWS gerenciada; self-hosted NÃO MUST ser aceito como destino de demo desta feature (desenvolvimento local com containers permanece permitido).
- **FR-005**: O frontend MUST ser servido como site estático via armazenamento de objetos + CDN na AWS (caminho constitucional: S3 + CloudFront).
- **FR-006**: A API MUST ser executada em containers orquestrados gerenciados com balanceador na AWS (caminho constitucional: ECS Fargate + ALB).
- **FR-007**: Os dados persistentes MUST residir em **RDS PostgreSQL** clássico na menor classe viável para demo (preferência `db.t4g.micro` / elegibilidade Free Tier quando disponível). Justificativa: simplicidade e menor custo contínuo no MVP acadêmico; scale-down avançado (Aurora Serverless etc.) fica fora desta feature.
- **FR-008**: Redis/cache gerenciado NÃO MUST ser provisionado nesta feature salvo justificativa explícita de uso no MVP; na ausência de uso, permanece fora de escopo / futuro.
- **FR-009**: Toda a infraestrutura AWS desta feature MUST ser declarada e versionada como código no repositório com Terraform enxuto (CloudFormation só se justificado e documentado); provisionamento manual permanente na console é proibido.
- **FR-010**: Segredos (JWT, senha de admin inicial, credenciais de banco, chaves AWS de pipeline) MUST permanecer fora do código-fonte e de commits; o mecanismo padrão desta feature é o mais simples e barato adequado ao MVP (Parameter Store com valor seguro preferido a Secrets Manager, salvo necessidade demonstrada de rotação avançada).
- **FR-011**: A estratégia de custo MUST priorizar conta Free / créditos / Free Tier; o modo operacional padrão fora de demo/aula simulada MUST ser **destroy completo** da infra provisionada (sem stack cobrável ociosa 24/7). Over-provisioning permanente e “deixar ligado por precaução” são proibidos.
- **FR-012**: A feature MUST documentar o ciclo **apply → warm-up operacional → destroy**; durante a janela ativa, capacidade mínima documentada MUST evitar cold start. Automação de agendamento na nuvem MAY ser P2 e, se não implementada, MUST constar como gap documentado.
- **FR-013**: CI MUST continuar obrigatório em todo push (build + testes); pipeline de deploy contínuo (CD) MAY ser entregue em fase P3 e NÃO isenta o CI.
- **FR-014**: Observabilidade nesta feature MUST limitar-se ao suficiente para diagnosticar falhas na demo/janela simulada (logs e métricas básicas na nuvem); APM/tracing distribuído caro está fora de escopo.
- **FR-015**: A região AWS de implantação MUST ser **`us-east-1`**, documentada no IaC e no guia operacional (prioridade Free Tier/preço sobre latência BR nesta feature).
- **FR-016**: O acesso HTTPS público MUST usar **apenas hostnames gerados pelos serviços AWS** (CloudFront para o frontend; ALB para a API), sem domínio customizado nem certificado em domínio próprio nesta feature.
- **FR-017**: Documentação do repositório (README e/ou guia operacional da feature) MUST incluir: como provisionar (`apply`), como publicar API e **republicar frontend com URL da sessão**, como destruir (`destroy` + verificação), **como configurar alerta de billing**, estimativa de custo por sessão, e o que está fora de escopo.
- **FR-018**: Streaming ao vivo, chat em tempo real, certificados de conclusão, OAuth externo, microsserviços e troca de stack NÃO MUST fazer parte desta feature (apenas exclusões explícitas).
- **FR-022**: Antes da primeira demo na conta, o operador MUST configurar e documentar **alerta de billing/orçamento** com limiar baixo em USD (suficiente para avisar consumo anormal ou stack esquecida ligada). Ausência de alerta NÃO é aceitável para declarar a feature pronta.

### Key Entities

- **Ambiente de demo**: Instância lógica efêmera da plataforma hospedada (frontend + API + banco + segredos) usada para avaliação; recriada por sessão com seed.
- **Infraestrutura versionada**: Conjunto de recursos AWS descritos em código (rede, CDN, orquestração de containers, balanceador, banco, armazenamento de segredos) aplicável de forma reproduzível.
- **Perfil de usuário**: Admin, professor ou aluno com permissões RBAC inalteradas em relação ao baseline atual.
- **Janela de aula simulada / demo**: Intervalo em que a stack está provisionada (`apply`) e warm-up concluído; fora dela a stack MUST estar destruída.
- **Segredo operacional**: Credencial ou chave sensível injetada em runtime, nunca versionada em texto claro no repositório.
- **Artefato publicável**: Build do frontend estático e imagem/pacote da API prontos para o ambiente de demo.

### Constraints *(constitution)*

- Custo-consciente: serviços pagos contínuos exigem justificativa; Free Tier / conta acadêmica preferidos; sem NAT Gateway no MVP.
- Pronto para o pico: cold start na abertura da aula simulada é inaceitável; warm-up obrigatório (operacional ou automatizado).
- Escopo mínimo / YAGNI: não implementar streaming, chat, certificados ou OAuth nesta feature.
- Segurança e RBAC desde o início; comunicação e erros em português.
- Qualidade testável: CI build+test obrigatório; testes de handler ao alterar API.
- IaC: Terraform enxuto; sem módulos elaborados sem necessidade.
- Stack fechada: Go/Fiber, React/Vite/TS, PostgreSQL, JWT+bcrypt, ECS Fargate + ALB, S3 + CloudFront.

### Out of Scope

- Streaming ao vivo / IVS
- Chat WebSocket
- Emissão de certificados
- OAuth externo (Cognito etc.)
- Self-hosted como destino de produção/demo
- Observabilidade cara (APM/tracing pesado); CloudWatch básico OK
- Microsserviços / troca de stack
- Redis gerenciado (salvo justificativa futura)
- Automação completa de apply/warm-up/destroy via agendamento na nuvem (desejável P2, não bloqueia P1)
- Ambiente AWS ligado 24/7 com capacidade mínima (rejeitado: destroy completo entre sessões)
- Persistência/backup de dados de demo entre ciclos destroy/apply (dados efêmeros + seed)
- NAT Gateway / tasks apenas em subnet privada com NAT (rejeitado por custo)
- Proxy no CDN (`/api`) ou config.json dinâmico como mecanismo principal de descoberta da API (rejeitado: rebuild/publish por sessão)
- CD completo a partir do CI (P3; P1 exige caminho de deploy reproduzível)
- Domínio customizado + certificado em domínio próprio (demo usa hostnames gerados)

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Um operador consegue provisionar a infra mínima a partir do repositório e obter URLs HTTPS do frontend e da API em no máximo uma sessão de trabalho documentada (meta: ≤ 90 minutos na primeira vez, seguindo o guia).
- **SC-002**: Reaplicar o provisionamento sem mudanças materiais conclui com sucesso e sem correção manual na console (reprodutibilidade verificável).
- **SC-003**: Na demo hospedada, login + listagem de cursos completa com sucesso para os três perfis em ≤ 2 minutos por perfil (caminho feliz).
- **SC-004**: Em verificação de RBAC amostral, 100% das ações proibidas testadas são negadas com feedback em português.
- **SC-005**: CI permanece verde em build + testes no branch da feature antes da demo; regressão de teste falha o pipeline.
- **SC-006**: Documento de custo existe e estima gasto por sessão de demo (apply ligado N horas + destroy), deixando explícito que o custo contínuo esperado fora de sessão é ~US$ 0 de compute/banco/ALB (salvo resíduos documentados como imagens no registry) e que NAT Gateway não faz parte da conta.
- **SC-007**: Runbook apply → warm-up → destroy está publicado; um operador consegue completar o ciclo de encerramento (destroy + checagem) em ≤ 15 minutos sem inventar passos.
- **SC-008**: Nenhum segredo de produção/demo aparece em arquivos rastreados pelo controle de versão (verificação por revisão + ausência de credenciais hardcoded).
- **SC-009**: Itens fora de escopo (streaming, chat, certificados, OAuth, self-hosted demo) não são entregues nem apresentados como parte desta feature.
- **SC-010**: Após um `apply` limpo (pós-destroy), um avaliador consegue fazer login com o admin seedado e ver pelo menos um curso de exemplo sem cadastro manual prévio.
- **SC-011**: Após republish do frontend com a URL da API da sessão, login pela UI hospedada completa com sucesso (prova de que front e API da mesma sessão estão acoplados).
- **SC-012**: Existe alerta de billing/orçamento ativo na conta de demo, com limiar documentado no guia operacional (verificável na console de billing/budgets).

## Assumptions

- O código atual (CRUD, auth JWT + RBAC, frontend React, backend Go, Postgres local, CI) é a baseline funcional; esta feature foca no que falta para AWS + IaC + endurecimento de configuração/deploy.
- Desenvolvimento local continua via Docker Compose; isso não substitui o destino AWS da demo.
- Redis no Compose local pode permanecer para compatibilidade futura, mas não implica provisionar Redis na AWS nesta feature.
- Conta AWS Free (plano ~6 meses / créditos) dedicada ao trabalho acadêmico estará disponível em `us-east-1`; se Free Tier de RDS ou algum serviço estiver indisponível no plano Free, usa-se Paid + créditos ou a menor instância paga, com custo por sessão justificado no documento de custo.
- Fora de demo/aula, a expectativa é conta “fria”: stack destruída via Terraform, não ambiente mínimo permanente.
- Dados na AWS são descartáveis entre sessões; seed/migração na subida cobre a demo (sem pipeline de backup/restore no escopo).
- Preservação de dados entre sessões (snapshot RDS, dump S3, etc.) está fora de escopo.
- NAT Gateway fora do desenho do MVP; tasks Fargate com IP público + security groups são aceitáveis para custo acadêmico.
- Acoplamento front↔API por sessão: rebuild/publish do frontend após cada `apply` com `VITE_API_URL` (ou equivalente) da API corrente; proxy CDN `/api` e config.json dinâmico ficam fora desta decisão.
- Alerta de billing/orçamento é requisito de prontidão operacional (não apenas dica).
- Tamanhos mínimos de compute (CPU/memória da tarefa da API) serão definidos no plano técnico, privilegiando o menor custo que mantenha a demo estável em `us-east-1`.
- HTTPS via certificados gerenciados dos serviços AWS nos hostnames padrão (CloudFront/ALB) é suficiente; domínio customizado está fora de escopo desta feature.
- Pré-aquecimento pode começar como procedimento operacional documentado (P2 parcial) sem bloquear P1.
- CD automatizado (P3) não é pré-requisito para a primeira demo se o caminho de publicação manual/semi-manual estiver documentado e reproduzível.
- CloudFormation não será usado salvo justificativa explícita no plano; Terraform é o padrão.
- Parameter Store (valor seguro) é a escolha default de segredos por simplicidade e custo; Secrets Manager só se o plano demonstrar necessidade.
