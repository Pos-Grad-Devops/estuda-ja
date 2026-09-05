# Research: 003-live-streaming-ivs

Phase 0 — decisões técnicas para materializar a spec (decisões de produto **não** reabertas). Baseline 001/002 **não** reaberta.

## 1. Canal IVS mínimo (tipo e latência)

**Decision:** **um** `aws_ivs_channel` por sessão Terraform, `us-east-1`:

| Atributo | Valor P1 |
|----------|----------|
| `type` | `BASIC` |
| `latency_mode` | `LOW` |
| `authorized` | `false` |
| `recording_configuration_arn` | **omitido** (sem live→VOD) |
| IVS Chat / stage Real-Time | **não provisionar** |

Ingest OBS: `rtmps://{ingest_endpoint}:443/app/` + stream key. Encoder: H.264 + AAC, keyframe **2 s**, bitrate ≤ **3,5 Mbps** se 720p/1080p (teto BASIC HD/Full HD) ou ≤ **1,5 Mbps** se 480p.

**Rationale:** BASIC é o mais barato (input **US$ 0,20/h** só enquanto há sinal) e cabe no Free Tier histórico de contas novas (5 h BASIC/mês — não depender dele). LOW + IVS Player atende SC-007 (≤ 15 s até imagem). `authorized=false` porque tokenização de playback de curta duração está **fora do P1** (clarificação); o gate de produto é JWT na API. Recording seria pipeline live→VOD, fora de escopo.

**Alternatives considered:**
- `STANDARD` / `ADVANCED_*` — ABR e transcode; input US$ 0,50–2,00/h; YAGNI para demo de dezenas de avaliadores.
- `NORMAL` latency — HLS clássico ~8–12 s+; pior para “aula magna”; rejeitado.
- Canal por aula — rejeitado na clarificação (custo e complexidade).
- IVS Real-Time (Stages) — outro produto/preço; não é o caminho do README (IVS low-latency).

## 2. Onde guardar a stream key (SSM)

**Decision:** Terraform cria o canal e obtém a stream key da sessão (`aws_ivs_stream_key` **ou** data source da key padrão do `CreateChannel` — na implement, **não** criar segunda key se a API já devolver uma). Persistir:

| Dado | Destino |
|------|---------|
| Stream key | SSM **SecureString** `${ssm_prefix}/IVS_STREAM_KEY` → secret da task `IVS_STREAM_KEY` |
| Ingest endpoint (hostname) | SSM String **ou** env plain `IVS_INGEST_ENDPOINT` |
| Playback HLS URL | env plain `IVS_PLAYBACK_URL` (não é senha, mas **não** vira output público de produto) |
| Channel ARN | env `IVS_CHANNEL_ARN` + output Terraform `ivs_channel_arn` |

**Não** gravar stream key no Postgres. **Não** commitar. **Não** expor em `terraform output` não-sensitive. Task **não** precisa de `ivs:GetStreamKey` no P1 (credenciais injetadas, padrão 001).

A key é **estável até o destroy** (clarificação); reiniciar live não rotaciona.

**Rationale:** Espelha JWT/DB no SSM; execution role já faz `GetParameter`; YAGNI de IAM IVS na task role.

**Alternatives considered:**
- API chama `ivs:GetStreamKey` em runtime — source of truth no IVS, mas IAM extra e latência; desnecessário com key estável.
- Output Terraform sensitive para o operador colar no `.env` — frágil no publish ECS; rejeitado.
- Guardar no DB — vazamento em backups/seeds; rejeitado.

## 3. Player no frontend: IVS Player vs `<video>` HLS

**Decision:** bloco **Ao vivo** usa **Amazon IVS Player Web SDK**. Caminho P1 recomendado: script oficial `https://player.live-video.net/{versão}/amazon-ivs-player.min.js` (WASM hospedado pela AWS) + `<video>` âncora. Fallback: mensagem PT se o browser não suportar. **VOD (002) permanece `<video>` nativo de MP4** — players **não** se misturam.

Não usar `<video src=".m3u8">` como caminho feliz da live (Chrome sem HLS nativo / latência alta). Service worker iOS Safari: **fora do P1** (demo acadêmica em desktop).

A `playback_url` só entra no player **depois** de `GET .../live/playback` com JWT.

**Rationale:** Documentação IVS: latência baixa **exige** o player oficial. SC-007 e constituição II.

**Alternatives considered:**
- HLS.js / Video.js sem tech IVS — latência e suporte piores; rejeitado no P1.
- `amazon-ivs-player` via npm + copiar WASM no `public/` — válido na implement se o CDN for indesejado; mais atrito Vite.
- Broadcast SDK in-browser (sem OBS) — fora da clarificação P1 (encoder externo).

## 4. Stub local / CI (sem vídeo real)

**Decision:** `LIVE_BACKEND=stub` (default Compose/local) vs `LIVE_BACKEND=ivs` (ECS após apply).

| Backend | `POST start` | Ingest | Playback | Sinal |
|---------|--------------|--------|----------|-------|
| `stub` | **Sucesso** se RBAC/invariante OK; persiste `ao_vivo` | JSON `modo: stub`, sem key real; mensagem PT | JSON `modo: stub`, `playback_url` nulo; UI **não** monta player vazio | Nenhum |
| `ivs` sem env completo | **Falha** 503 PT; aula **não** fica ao vivo | — | — | — |
| `ivs` ok | Sucesso + credenciais reais | endpoint + stream key | HLS URL da sessão | OBS → IVS |

CI testa **somente stub** (estados, 403/409, PT). Compose **não** sobe IVS.

**Rationale:** Harmoniza clarificação (stub testável) com US2 AC4 (stack sem recurso ≠ marcar ao vivo). Dois modos explícitos evitam “live falsa” na AWS mal configurada.

**Alternatives considered:**
- Sempre 501 no local — quebraria testes de máquina de estados; rejeitado.
- Fake HLS file local — esforço sem valor de demo IVS; rejeitado.

## 5. Extensão Terraform enxuta

**Decision:** arquivo novo `infra/ivs.tf` (flat, sem modules):

- `aws_ivs_channel` nome `${project_name}-live` (ou sufixo conta se colisão de nome for problema — ARN é o id estável).
- Stream key associada ao canal (um recurso; ver §2).
- SSM: stream key SecureString; ingest/playback String se não forem só env interpolado na task.
- `ecs.tf`: `LIVE_BACKEND=ivs` + `IVS_*`.
- `iam.tf` (execution): ARNs SSM novos no `GetParameters` existente.
- `outputs.tf`: **`ivs_channel_arn` apenas**. Sem `playback_url`, sem stream key.
- **Não** tocar `infra/budget/`, VPC, NAT, Redis, Cognito, CloudFront extra, `s3_vod.tf`.

Destroy da pasta `infra/` remove o canal (custo IVS contínuo ≈ 0).

**Rationale:** Constitution VIII + FR-010; mesmo padrão de `s3_vod.tf`.

**Alternatives considered:**
- Módulo `terraform-aws-ivs` — módulos elaborados proibidos no MVP.
- Canal permanente fora da stack — viola FR-007.

## 6. Custo por sessão (ordem de grandeza)

Preços listados AWS IVS Low-Latency (consulta 2026-09-05; **não** é cotação). Canal **ocioso** (sem ingest, sem viewers): **input e output ≈ US$ 0**. Cobrança começa quando o encoder envia (arredondado p/ minuto) e quando há output para espectadores.

| Item | Durante sessão curta | Após destroy |
|------|----------------------|--------------|
| Canal BASIC idle (apply até OBS) | ~ **US$ 0** IVS | **0** (recurso gone) |
| Input BASIC (OBS ligado) | **US$ 0,20 / h** | 0 |
| Output (1.º 10k h) | NA SD **US$ 0,036 / h/viewer**; América do Sul SD **US$ 0,042 / h/viewer** | 0 |
| Demo típica (~1 h OBS + 2–5 viewers HD/SD) | **centavos a ~US$ 1** de IVS | 0 |
| ECS/RDS/ALB/CF/S3 VOD | igual 001+002 (~US$ 1–3 / ~4 h) | 0 (resíduos já documentados no 001) |
| NAT / ElastiCache / CF mídia live | **não provisionado** | — |

Risco de custo: OBS esquecido **streaming** 24/7 = ~US$ 4,80/dia de input BASIC (+ output se houver viewers). Rede de segurança: **destroy** + budget US$ 5 do 001.

**Rationale:** SC-004; operador não precisa STANDARD nem IVS 24/7.

**Alternatives considered:** STANDARD “por qualidade” — rejeitado no §1.

## 7. Warm-up checklist (extensão do T−15 do 001)

**Decision:** P1 **100% manual**. Não reabrir EventBridge apply/destroy (gap 001). Acrescentar checagens de **streaming** no mesmo relógio T−15 / T−10 / T−5 / T−0. Detalhe operacional: [quickstart.md](./quickstart.md). Resumo:

| Momento | Extra streaming |
|---------|-----------------|
| **T−15** | Canal IVS existe (`ivs_channel_arn`); task com `LIVE_BACKEND=ivs`; SSM da stream key presente |
| **T−10** | Health API 200; `GET .../live` da aula demo responde (inativa ou estado conhecido) |
| **T−5** | Professor **inicia** live; OBS conecta (status Live); preview no player gestor; aluno (outra sessão) vê “ao vivo” — **não** deixar o primeiro `apply`/OBS para T−0 |
| **T−0** | Não destroy; não zerar ECS; encoder continua enviando |

P2/P3 (fase 5): MAY automatizar só `GetStream` / health do canal — **sem** ligar/desligar a stack 001.

**Rationale:** FR-008 / constitution II; SC-006.

## 8. Superfície de domínio na API

**Decision:** recursos aninhados em `/api/v1/aulas/:id/live` (ver [contracts/live-api.md](./contracts/live-api.md)); handler/repository `live_*`. **Não** reutilizar `Aula.Status` (`agendada`/`ao_vivo`/`encerrada` do CRUD 001) como estado da transmissão — colisão de significado; FR-012.

**Rationale:** Espelho VOD; um arquivo por domínio.

## 9. Playback JWT vs URL HLS do IVS (residual P1)

**Decision:** Produto **nunca** publica playback em output Terraform, README de aluno ou página anônima. API devolve a URL HLS **só** com JWT + leitura de aulas, e **só** se aquela aula estiver `ao_vivo`. P1 **não** liga `authorized=true` nem playback tokens IVS.

**Residual aceito:** quem interceptar a HLS URL pode reproduzir fora do app até o destroy (canal não autorizado). Mitigação: sessão efêmera + não logar a URL em commits + UI aluno sem “copiar link”. Tokenização = fora do P1 (clarificação).

**Alternatives considered:** IVS playback authorization — contradiz “tokenização curta fora do P1” e quebraria o player sem tokens.

## 10. UI na ficha (live vs VOD)

**Decision:** estender `AulasPage.tsx` com bloco **“Transmissão ao vivo”** **acima ou ao lado** do bloco **“Gravação”** (002), labels distintos. Sem novas rotas. Permissão de escrita: `canManageAulas`. Horário: aviso se `agendada_em` vazio, submit de iniciar **não** bloqueado.
