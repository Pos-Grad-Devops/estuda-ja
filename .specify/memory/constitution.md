<!--
Sync Impact Report
- Version change: (template placeholders / sem versão) → 1.0.0
- Modified principles: N/A (primeira ratificação a partir do scaffold)
  - [PRINCIPLE_1_NAME] → I. Custo-consciente na AWS
  - [PRINCIPLE_2_NAME] → II. Pronto para o pico
  - [PRINCIPLE_3_NAME] → III. Escopo mínimo e YAGNI
  - [PRINCIPLE_4_NAME] → IV. Segurança e RBAC desde o início
  - [PRINCIPLE_5_NAME] → expandido em V–VIII (ver Added)
- Added sections:
  - V. Qualidade testável
  - VI. Observabilidade suficiente
  - VII. Documentação e Spec-Driven Development
  - VIII. Infraestrutura como código
  - Constraints de Stack e AWS
  - Workflow Speckit e Qualidade
  - Governance (ratificada)
- Removed sections: nenhum (scaffold genérico substituído)
- Follow-up TODOs: nenhum placeholder deferido
-->

# EstudaJá Constitution

## Core Principles

### I. Custo-consciente na AWS
A plataforma MUST priorizar custo baixo (conta acadêmica, Free Tier e
scale-to-zero quando possível) sem comprometer demo e horários de aula
simulados. Preferir serviços gerenciados baratos no caminho AWS.
Pré-aquecer somente na janela da aula; reduzir réplicas após; evitar
over-provisioning permanente. Qualquer serviço pago contínuo MUST ter
justificativa explícita no spec/plan ou README. Rationale: orçamento
acadêmico limitado e pico concentrado — pagar o tempo todo pelo pico é
desperdício.

### II. Pronto para o pico (NON-NEGOTIABLE)
A aula ao vivo tem horário fixo; o pico de acesso é INSTANTÂNEO no minuto
de abertura — não escala “aos poucos”. Arquitetura e operação MUST
contemplar warm-up, autoscaling e CDN antes do horário da aula. Cold start
na abertura da aula é inaceitável. SLA alvo: 99,9% na janela ao vivo; 99%
fora dela. Rationale: milhares de alunos no minuto zero; falha de
aquecimento quebra a demo e o enunciado do projeto.

### III. Escopo mínimo e YAGNI
Entregar o necessário para o trabalho acadêmico e a demo. Visão de produto
(transmissão ao vivo, chat, biblioteca VOD, certificado) MAY ser planejada
em fases; NÃO forçar tudo no MVP. Streaming, chat WebSocket, certificado e
deploy AWS só quando especificados via Speckit ou pedido explícito. Sem
over-engineering, abstrações prematuras ou refatoração fora do escopo do
pedido. Rationale: prazo curto; entrega ponta a ponta faseável vale mais
que feature completa não demonstrável.

### IV. Segurança e RBAC desde o início
Auth própria com JWT + bcrypt no MVP (sem OAuth externo). Papéis RBAC
obrigatórios: `admin`, `professor`, `aluno` — aplicados na API e refletidos
na UI. Segredos (JWT, senhas admin, credenciais AWS) MUST permanecer fora
do código e fora de commits. Mensagens de erro da API MUST estar em
português. Comunicação humana/agentes MUST ser em português. Rationale:
base multi-perfil e segredos corretos desde o dia um evitam retrabalho e
vazamento.

### V. Qualidade testável
Alterações em handlers/API MUST incluir ou atualizar testes Go relevantes.
CI (build + test) é obrigatório a cada push. Deploy automatizado na AWS
MAY ser faseado; a ausência de deploy NÃO isenta build/test no CI.
Rationale: regressão silenciosa na API quebra demo e RBAC.

### VI. Observabilidade suficiente
Logs estruturados e métricas mínimas MUST existir para diagnosticar falha
na janela ao vivo (ex.: CloudWatch no caminho AWS). Stack de
observabilidade cara (APM completo, tracing distribuído pesado) NÃO é
exigida no MVP. Rationale: sem telemetria mínima, o SLA na janela crítica
não é verificável.

### VII. Documentação e Spec-Driven Development
Mudanças de feature/comportamento MUST ser governadas por artefatos
Speckit (`spec.md`, `plan.md`, `tasks.md`) quando o fluxo Speckit estiver
em uso. `AGENTS.md` e `README.md` MUST acompanhar decisões de stack, auth,
estrutura e caminho AWS. Raiz do repositório permanece enxuta (README,
DESCRICAO, REQUISITOS + código/docs). Rationale: agentes e humanos
precisam da mesma fonte de verdade.

### VIII. Infraestrutura como código
Mudanças de infra AWS MUST passar por Terraform (alternativa aceitável:
CloudFormation, se justificada). Estado e código de infra versionados no
repositório; provisionamento manual permanente na conta é proibido (drift).
Terraform MUST permanecer enxuto no MVP — sem módulos ou complexidade
desnecessários. Rationale: reproduzibilidade acadêmica e custo
controlável; “clique na console” não escala nem documenta.

## Constraints de Stack e AWS

**Caminho obrigatório:** AWS gerenciado. Self-hosted como destino de
produção/demo NÃO é negociável nesta constitution (Docker Compose local
para desenvolvimento permanece permitido).

**Stack fechada (não trocar sem motivo forte + emenda):**
- Backend: Go + Fiber
- Frontend: Vite + React + TypeScript
- Banco: PostgreSQL (GORM)
- Cache: Redis somente quando justificado
- Auth MVP: JWT + bcrypt (própria)
- Orquestração alvo: ECS Fargate + ALB
- Frontend estático: S3 + CloudFront
- Datas BR: `dd/mm/yyyy` ou `dd/mm/yyyy HH:mm` (backend `timeutil`,
  frontend `utils/date.ts`)
- Endpoints sob `/api/v1/`; handlers/repositories um arquivo por domínio

**Permitido no MVP / fases:**
- Entrega ponta a ponta: frontend + backend + auth + persistência + CI
- Deploy AWS faseado (IaC primeiro; pipeline de deploy depois)
- Warm-up/autoscaling programado na janela de aula simulada
- Terraform enxuto versionado no repo

**Proibido sem pedido explícito / spec:**
- Troca de stack principal ou adoção de OAuth externo no MVP
- Streaming ao vivo, chat WebSocket, certificados (até especificados)
- Over-provisioning permanente “por segurança”
- Provisionamento manual permanente na conta AWS
- Módulos Terraform elaborados sem necessidade demonstrada
- Stack de observabilidade cara no MVP

## Workflow Speckit e Qualidade

1. **Specify** (`/speckit-specify`) — descrever a feature e restrições
   (custo, pico, RBAC, fases) antes de codificar.
2. **Plan** (`/speckit-plan`) — desenhar arquitetura alinhada a esta
   constitution (AWS, Terraform enxuto, Go/React).
3. **Tasks** (`/speckit-tasks`) — quebrar em tarefas ordenadas e testáveis.
4. **Implement** (`/speckit-implement`) — executar tasks; escopo mínimo;
   testes de handler ao mudar API; atualizar AGENTS/README se stack/auth/
   estrutura mudarem.

**NÃO fazer sem pedido ou spec:** streaming, chat, certificado, deploy
produção, OAuth, refatoração ampla, troca de banco/framework, IaC além do
escopo da feature em andamento.

**Gates de qualidade:** CI verde (build + test); erros API em português;
RBAC respeitado; datas BR; segredos fora do git; conformidade com
princípios I–VIII revisável em PR/review.

## Governance

Esta constitution prevalece sobre preferências ad hoc, comentários em
issues e “atalhos” de console AWS. Em conflito com README/AGENTS, a
constitution MUST ser seguida e os docs alinhados na mesma mudança ou em
seguida imediata.

**Emendas:** alteração de princípio ou constraint exige bump de versão
semântica + atualização de **Last Amended**:
- MAJOR — remoção/redefinição incompatível de princípio
- MINOR — novo princípio/seção ou expansão material
- PATCH — esclarecimentos, tipografia, refinamentos não semânticos

**Compliance:** PRs e reviews MUST verificar alinhamento aos Core
Principles e Constraints. Complexidade e serviços pagos contínuos MUST
ser justificados. Guidance operacional: `AGENTS.md` e `README.md`.

**Version**: 1.0.0 | **Ratified**: 2026-09-04 | **Last Amended**: 2026-09-04
