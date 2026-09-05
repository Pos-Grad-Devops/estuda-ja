# Quickstart: Streaming ao vivo (003 / AWS IVS)

Validação ponta a ponta sobre o ciclo **001** (apply → publish → warm-up → demo → destroy) + VOD **002**. Detalhes: [contracts/live-api.md](./contracts/live-api.md) · [data-model.md](./data-model.md) · [live-env.md](./contracts/live-env.md) · [terraform-ivs.md](./contracts/terraform-ivs.md).

Este arquivo **não** substitui o [quickstart 001](../001-aws-mvp-terraform/quickstart.md). Encaixa **passos extras** de streaming. **Não** reabrir NAT/Redis/budget/VOD. Resumo operacional também em [README — Demo AWS](../../README.md#demo-aws-mvp-p1).

**Tempo incremental (SC-005):** leitura + execução dos passos extras deste arquivo ≤ **15 min** depois da stack 001 já aplicada/publicada. Cold start de canal/OBS na abertura da janela é **inaceitável** (SC-006 / FR-008) — warm-up **antes** de T−0.

## Pré-requisitos

- Features `001-aws-mvp-terraform` e `002-vod-library` operacionais.
- Conta AWS; região `us-east-1`; Terraform ≥ 1.5; AWS CLI; Docker ARM64; Node 22.
- **OBS Studio** (ou encoder RTMPS equivalente) na máquina do gestor da demo.
- Local/CI: Docker Compose ou API+Postgres — **stub** (sem vídeo).

## A. Desenvolvimento local (sem AWS / sem vídeo)

1. Na raiz: `docker compose up --build` (ou `make infra-up` + API/front). `LIVE_BACKEND=stub` (Compose já define).
2. Login `professor@estudaja.com` / `professor123` → **Aulas** → aula de demo.
3. (P2) **Agendar transmissão** (exige horário) → aluno vê **Agendada**, **sem** player ativo.
4. **Iniciar transmissão** → ficha mostra **Ao vivo**; ingest stub (**sem** stream key real). Player **não** finge sinal.
5. Outra sessão: `aluno@estudaja.com` / `aluno123` → mesma aula → vê “ao vivo”, **sem** ingest, **sem** botões; VOD 002 distinguível.
6. **Encerrar** → **Encerrada**; se houver VOD publicado, a ficha pode apontar a gravação (sem pipeline live→VOD).
7. Aluno `POST .../live/start` ou `GET .../ingest` → **403** PT. Agendar/cancelar também **403** para aluno.
8. Professor inicia live numa **segunda** aula sem encerrar a primeira → **409** PT.
9. Aula sem data/hora: API/UI permitem **iniciar** (UI avisa); **agendar** exige horário.

**Esperado:** SC-003 amostral; SC-012 (3 estados); FR-015/FR-017. `go test ./...` verde (handlers live).

## B. Demo AWS + OBS

Seguir README / [001 §1–3](../001-aws-mvp-terraform/quickstart.md): billing (já feito) → **apply** → **publish-api** → **publish-frontend**. Budget `infra/budget/` **não** se destroi.

### B1. Apply

```powershell
cd infra
terraform apply
```

**Verificar:** `terraform output ivs_channel_arn` presente; **um** canal; **não** há output de stream key/playback. Task definition do apply já inclui `LIVE_BACKEND=ivs` + env `IVS_*` (secret da key no SSM).

### B2. Publish API + front

```powershell
.\publish-api.ps1
# Aguardar GET {api_url}/health → 200
.\publish-frontend.ps1
```

**Verificar:** task em execução com `LIVE_BACKEND=ivs` + secrets SSM da stream key; VOD 002 ainda funciona.

### B3. Warm-up (T−15) — incluir streaming

Completar o checklist [001 §4 T−15](../001-aws-mvp-terraform/quickstart.md#checklist-t15-min-obrigatório-antes-da-janela-simulada) **e** os checks extras abaixo. Ordem fixa: **T−15 → T−10 → T−5 → T−0**.

| Momento | Check infra (001) | Check streaming (003) |
|---------|-------------------|------------------------|
| **T−15** | Apply ok; RDS available; ECS RUNNING; targets healthy | Canal existe: `terraform output ivs_channel_arn`; 1 canal; key **não** em outputs |
| **T−10** | `GET {api_url}/health` = 200 estável | Status live da aula demo = inativa (ou conhecido); sem stack/OBS frios pendentes |
| **T−5** | Login + UI + amostra RBAC | Professor inicia live; OBS **Live**; player gestor com sinal; aluno vê ao vivo |
| **T−0** | Não zerar `desired_count`; destroy só depois | Não apply/destroy; encoder continua; **não** “ligar OBS agora” |

**Não** subir a stack nem o OBS no minuto da aula (constitution II / SC-006).

#### Manual vs automatizado (US5)

| Parte | Como | Notas |
|-------|------|--------|
| Apply / destroy stack **001** | **Manual** | Gap EventBridge apply/destroy do 001 **permanece** — este quickstart **não** liga/desliga a stack |
| Checklist T−15 + OBS + start live | **Manual** | Operador percorre a tabela acima |
| Health API (+ GetStream opcional) | **Script** `infra/check-live-warmup.ps1` | Só checagem; **sem** terraform apply/destroy |

```powershell
cd infra
.\check-live-warmup.ps1 -ApiUrl (terraform output -raw api_url)
# Após OBS Live (T−5), opcional:
.\check-live-warmup.ps1 -ApiUrl (terraform output -raw api_url) -ChannelArn (terraform output -raw ivs_channel_arn)
```

### B4. OBS (gestor)

1. Login admin ou professor em `{frontend_url}` → ficha da aula → **Iniciar transmissão**.
2. Copiar **Servidor** (`rtmps://…:443/app/`) e **Stream key** (só esta UI; aluno não vê). A API monta o RTMPS a partir de `IVS_INGEST_ENDPOINT` (hostname).
3. OBS → Configurações → Stream: serviço **Personalizado**; colar servidor e key.
4. Output: H.264 + AAC; keyframe **2 s**; 720p ou 1080p com bitrate ≤ **3,5 Mbps** (canal BASIC).
5. Iniciar transmissão. Aguardar estado Live no OBS.
6. Na ficha: player IVS deve exibir áudio/vídeo em ≤ 15 s após “assistir” (SC-007).

### B5. Demo funcional (aluno)

1. Login aluno (outra sessão/browser) → **Aulas** → mesma aula → identificar **Transmissão ao vivo** (não confundir com **Gravação**).
2. Reproduzir in-app ≤ 2 min após login (SC-001/002). Sem download.
3. Visitante sem JWT: playback/ingest recusados (SC-011).
4. Professor **Encerra**; aluno atualiza a ficha → deixa de ver live ativa (SC-010). VOD seed, se existir, permanece.
5. Reiniciar live (mesma ou outra aula) **sem** novo apply; no máximo uma ativa; stream key **estável** até destroy.

### B6. Destroy

```powershell
cd infra
terraform destroy
```

**Verificar:** canal IVS e key ausentes; sem input IVS ocioso; VOD bucket e resto da demo 001/002 removidos; **budget ainda ativo**. Endpoints da sessão anterior **inválidos** — não reutilizar.

### B7. Novo apply

Novo canal e **nova** stream key. Aula **não** volta `ao_vivo` sozinha. OBS precisa das credenciais novas da ficha após iniciar.

## Critérios rápidos de aceite

| ID | Check |
|----|--------|
| SC-001/002 | Aluno toca live na ficha após gestor iniciar + OBS |
| SC-003 | Aluno não gere live; anônimo não obtém playback |
| SC-004 | Destroy zera IVS |
| SC-005 | Passos extras de streaming ≤ 15 min de leitura/execução incremental |
| SC-006 | Warm-up (canal + OBS) **antes** de T−0; cold start na abertura inaceitável |
| SC-007 | Imagem ≤ 15 s com sinal no ar |
| SC-008 | CI verde no branch |
| SC-009 | Sem chat, live→VOD, Cognito, NAT, Redis AWS |
| SC-011 | Sem stream key na UI aluno / no git / em outputs TF |
| SC-012 | Agendada vs ao vivo vs encerrada (e vs VOD) evidentes |

## Consistência rápida (docs ↔ código)

| Item | Esperado |
|------|----------|
| `LIVE_BACKEND` | Compose/local `stub`; ECS `ivs` |
| Output TF | só `ivs_channel_arn` (sem key/URL) |
| RBAC | schedule/cancel/start/stop/ingest = admin/professor; aluno lê status/playback |
| Live status | `inativa` \| `agendada` \| `ao_vivo` \| `encerrada` |
| VOD 002 | intacto na mesma ficha; blocos distintos; MAY link se encerrada |
| `Aula.Status` CRUD | **não** é estado da live (`AulaLive`) |
| Warm-up auto | só `check-live-warmup.ps1` (health/GetStream); **sem** EventBridge 001 |

## Fora desta validação

Chat (004), certificado (005), pipeline live→VOD, tokenização IVS, EventBridge apply/destroy do 001, app nativo, multi-qualidade/DVR.
