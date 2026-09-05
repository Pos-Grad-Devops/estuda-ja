# Research: 001-aws-mvp-terraform

Phase 0 — decisões técnicas restantes (decisões já fechadas na spec **não** são reabertas).

## 1. Layout Terraform no repositório

**Decision:** pasta `infra/` na raiz do monorepo, Terraform flat (sem módulos internos no MVP).

**Rationale:** mantém a raiz enxuta (`README`, `DESCRICAO`, `REQUISITOS`, código, `docs/`) e isola IaC; nome curto e comum. Um único root module (`main.tf`, `variables.tf`, `outputs.tf`, `providers.tf`) evita over-engineering (constitution VIII).

**Alternatives considered:**
- `terraform/` — equivalente; rejeitado só por preferência de nome mais curto.
- Módulos `modules/vpc`, `modules/ecs` — rejeitado no MVP (YAGNI); extrair só se o root virar ilegível.

## 2. Rede: VPC mínima vs default VPC

**Decision:** VPC custom mínima em `us-east-1` com **2 subnets públicas** (2 AZs), Internet Gateway, **sem NAT Gateway**; tasks Fargate com `assign_public_ip = true`; RDS na mesma subnet pública com security group restrito à SG das tasks.

**Rationale:** ALB exige ≥2 AZs; VPC própria é reproduzível entre contas e destruível por completo; alinhado à clarificação “IP público / sem NAT”. RDS acessível só pela SG das tasks (e opcionalmente SG do operador via variável, default fechado).

**Alternatives considered:**
- Default VPC — mais rápido no primeiro apply, mas depende do estado da conta e dificulta destroy limpo / documentação.
- Subnets privadas + NAT — rejeitado na spec (custo).

## 3. Tamanho da task Fargate

**Decision:** **256 CPU (0.25 vCPU) / 512 MiB**, arquitetura **linux/ARM64** (Graviton), **desired_count = 1** durante a janela de demo.

**Rationale:** menor combinação válida no Fargate; API Go/Fiber estática é leve; ARM reduz custo horário vs x86; 1 réplica cobre demo acadêmica (CRUD + login, sem streaming/chat). Warm-up = stack já ligada e `/health` OK (não cold start de scale-from-zero).

**Alternatives considered:**
- 256/1024 ou 512/1024 — margem extra sem necessidade demonstrada.
- Spot — risco de interrupção na demo; On-Demand para a janela.

## 4. Imagem da API (ECR) e publish manual P1

**Decision:**
1. Repositório ECR criado pelo Terraform (`force_delete = true` para destroy limpo).
2. Após o primeiro `apply` (ou quando o repositório existir): build local/CI ad-hoc `docker build --platform linux/arm64` → `docker push` tag `latest` (e opcionalmente tag de sessão/`git sha`).
3. Serviço ECS com `desired_count = 1` e deployment forçado após o push (`aws ecs update-service --force-new-deployment`).
4. Ordem do runbook: **apply infra → push imagem → force deploy → esperar healthy → rebuild front**.

**Rationale:** P1 exige caminho reproduzível sem CD; ECR é o registry constitucional; `force_delete` evita resíduo cobrável de storage de imagens após destroy.

**Alternatives considered:**
- ECR “bootstrap” fora do destroy — deixa resíduo permanente; rejeitado salvo necessidade.
- Imagem pública Docker Hub — menos alinhado ao caminho AWS e a segredos/conta acadêmica.
- CD no GitHub Actions — P3 (agora opcional em `.github/workflows/cd.yml`; P1 continua com publish manual).

## 5. HTTPS nos hostnames AWS (sem domínio customizado)

**Decision:**
- **Frontend:** CloudFront + S3 OAC; certificado gerenciado padrão do CloudFront (`*.cloudfront.net`).
- **API:** ALB em HTTP na porta 80 **atrás de um segundo distribution CloudFront** (origin = ALB DNS); HTTPS no hostname `*.cloudfront.net` da API. `VITE_API_URL` = URL HTTPS desse CloudFront da API (com path `/api/v1` implícito no client atual via `API_URL` base).

**Rationale:** ACM **não** emite certificado para `*.elb.amazonaws.com`. Frontend HTTPS + API HTTP gera **mixed content** no browser. CloudFront na frente do ALB resolve TLS sem domínio próprio e permanece no caminho “hostnames gerados pela AWS”.

**Alternatives considered:**
- Só ALB HTTP — quebra fetch a partir do CloudFront HTTPS.
- Domínio + ACM — fora do escopo (FR-016).
- API Gateway HTTP API como TLS — serviço extra sem ganho no MVP.

## 6. Segredos e env na task (Parameter Store)

**Decision:** SSM Parameter Store **SecureString** para `JWT_SECRET`, `ADMIN_PASSWORD`, senha RDS / `DATABASE_URL`; parâmetros String (ou outputs Terraform → env plain) para `CORS_ORIGIN`, `ADMIN_EMAIL`, `PORT`. Injeção via `secrets` / `environment` na task definition ECS; execution role com `ssm:GetParameters` + decrypt KMS.

**Rationale:** FR-010 e constitution IV; Parameter Store sem custo relevante no volume do MVP; Secrets Manager só agrega rotação/API paga sem necessidade.

**Alternatives considered:**
- Secrets Manager — rejeitado (custo/rotina sem benefício no ciclo destroy/apply).
- Env plaintext no task definition no state — evita para segredos (state ainda sensível; preferir SSM + state remote cuidadoso; local state: `.gitignore` + aviso no quickstart).

## 7. Seed / migração na subida do container

**Decision:** manter o fluxo atual: `main` → `database.Connect` → `RunMigrations` (gormigrate). Estender migrações para **seed de demo** (curso + aula + usuários professor/aluno de teste) além do admin (`001_seed_admin`). Dados efêmeros: banco novo a cada apply ⇒ seed sempre aplica.

**Rationale:** já implementado para admin; SC-010 exige admin + pelo menos um curso sem cadastro manual; não inventar job separado (ECS Exec / Lambda).

**Alternatives considered:**
- Script Terraform `null_resource` / local-exec — frágil e duplica lógica.
- Seed só via console — viola IaC/reprodutibilidade.

## 8. Estado Terraform e backend remoto

**Decision:** MVP com **state local** em `infra/` (arquivo `terraform.tfstate` no `.gitignore`); documentar backup opcional. Backend S3+DynamoDB MAY ser P2 se a equipe compartilhar a mesma conta.

**Rationale:** uma conta acadêmica / poucos operadores; remote state é serviço contínuo a mais; destroy + state local basta se o operador for o mesmo da sessão.

**Alternatives considered:**
- S3 backend desde P1 — melhor para equipe, mas custo/complexidade extra justificada só se houver conflito de state.

## 9. Estimativa de custo por sessão (ordem de grandeza, us-east-1)

**Decision:** documentar no quickstart/README uma faixa **~US$ 1–3 por sessão de ~4 h** (sem Free Tier RDS/ALB); com Free Tier elegível, compute/RDS podem cair para **~US$ 0–1** na mesma janela. Fora da sessão, após destroy completo: **~US$ 0** de ALB/RDS/Fargate/CloudFront (salvo resíduos listados abaixo). Limiar de budget sugerido: **US$ 5–10** mensal (alerta).

Componentes dominantes com stack ligada:
| Recurso | Ordem de grandeza |
|---------|-------------------|
| ALB | ~US$ 0,02–0,03/h + LCU baixo |
| RDS `db.t4g.micro` | Free Tier 750 h/mês se elegível; senão ~US$ 0,016/h |
| Fargate 0,25 vCPU / 0,5 GB ARM | ~US$ 0,01/h |
| CloudFront ×2 + S3 | centavos na demo |
| NAT | **US$ 0** (não provisionado) |

**Rationale:** SC-006; números são estimativa para guia, não cotação contratual — operador deve checar Pricing Calculator na data da demo.

**Alternatives considered:** deixar ligado 24/7 com mínimo — rejeitado na spec.

## 10. Resíduos após `terraform destroy` e limpeza

**Decision:** documentar checklist pós-destroy:
- Removidos pelo destroy (com `force_delete` onde necessário): VPC, ALB, ECS service/cluster/task def, RDS, S3 do site, CloudFront, SG, SSM params do projeto, log groups se gerenciados no TF.
- Podem restar se mal configurados: imagens ECR (mitigado com `force_delete`), log groups órfãos, snapshots RDS (desabilitar retention / `skip_final_snapshot = true` no MVP efêmero), alarmes/budget (budget **deve permanecer** — FR-022).
- State local e credenciais AWS na máquina do operador.

**Rationale:** SC-006/SC-007; budget não deve ser destruído junto com a demo.

## 10b. Fase F / P2 — warm-up: gap documentado (escolha b)

**Decision:** **não** implementar EventBridge (nem Lambda/CodeBuild) para apply, health-poll ou destroy. Entregar US4 via checklist operacional T−15 min no [quickstart.md](./quickstart.md) §4 e espelho no README. Automação de agendamento permanece gap explícito.

**Rationale:** constitution I (custo) + III (YAGNI) — agendar Terraform na conta acadêmica acrescenta IAM, runtime e superfície de falha sem ganho na demo pontual; com `desired_count = 1` o warm-up já é “stack ligada + `/health` OK”, não scale-from-zero. Alternativa (a) rejeitada neste MVP.

## 11. Observabilidade mínima

**Decision:** CloudWatch Logs no container da API (`awslogs`); métricas default ALB/ECS/RDS. Sem Container Insights pago extra, sem X-Ray/APM.

**Rationale:** constitution VI — suficiente para diagnosticar falha na janela.

## 12. CORS e acoplamento front↔API

**Decision:** após publish do front, `CORS_ORIGIN` na API = origem HTTPS do CloudFront do frontend (atualizar parâmetro/env e redeploy da task se a URL do front só for conhecida após apply). Runbook: apply → (URLs nos outputs) → push API → set CORS → force deploy → build front com `VITE_API_URL` → sync S3 → invalidação CloudFront.

**Rationale:** clarificação de rebuild por sessão; client já usa `import.meta.env.VITE_API_URL`.

## 13. Billing alert

**Decision:** AWS Budgets (ou alerta de billing SNS) com limiar baixo (ex. **US$ 5**); criado **uma vez** na conta (pode ser Terraform separado `infra/budget` ou passo manual documentado obrigatório antes da primeira demo). Não destruir no ciclo da stack de demo.

**Rationale:** FR-022 / clarificação; alerta é prontidão da conta, não recurso efêmero da sessão.

## NEEDS CLARIFICATION

Nenhum bloqueante restante. Valores numéricos de budget exato e senha de demo podem ser variáveis Terraform com defaults seguros só em Parameter Store (nunca commitados).
