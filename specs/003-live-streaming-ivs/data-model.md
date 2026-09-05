# Data Model: 003-live-streaming-ivs

Extensão mínima sobre o domínio 001 + VOD 002. **Não** redesenhar `Curso` / `Aula` / `AulaVod` além do vínculo e cascade da live.

## Entidades existentes (inalteradas no essencial)

### Curso
Campos atuais. Continua pai de `Aula`. Sem “canal solto” na UI.

### Aula
Campos atuais (`id`, `curso_id`, `titulo`, `descricao`, `agendada_em`, `status`, timestamps).

**`Aula.Status` (`agendada` \| `ao_vivo` \| `encerrada`)** permanece o status **pedagógico/CRUD** já existente (formulário de aula). **NÃO** é o estado da transmissão IVS. A feature de live **não** MUST sincronizar esse campo no P1 (evita redesign e colisão com a US4).

**`agendada_em`:** recomendado na UI para live; **opcional** na API ao iniciar (demo ad hoc). Estado P2 `agendada` da **live** só faz sentido se houver horário.

**Regra nova:** no máximo **um** `AulaLive` por aula. Ao **DELETE** aula (ou curso com suas aulas), remover o `AulaLive` associado. **Não** apagar o canal IVS da sessão (infra).

### AulaVod (002)
Independente. Live e VOD **coexistem** na ficha; encerrar live **não** cria VOD; VOD existente MAY ser indicado na UI após encerrar (P2), sem pipeline.

## Entidade nova: AulaLive

Associação entre **uma aula** e a **sessão de streaming** (canal compartilhado da demo). O canal em si **não** é linha de banco — vê [contracts/live-env.md](./contracts/live-env.md).

| Campo | Tipo | Obrigatório | Notas |
|-------|------|-------------|-------|
| `id` | uint / PK | sim | Surrogate |
| `aula_id` | uint | sim | **UNIQUE** — uma linha vigente por aula |
| `status` | string | sim | P1: `ao_vivo` \| `encerrada`. Ausência de linha = inativa. P2: + `agendada` |
| `iniciada_em` | DateTime BR | se `ao_vivo` (P1) | `timeutil.DateTime`; preenchido no start/reinício |
| `encerrada_em` | DateTime BR | se `encerrada` | nulo enquanto `ao_vivo` |
| `created_at` | DateTime BR | sim | |
| `updated_at` | DateTime BR | sim | |

**Não persistir:** stream key, ingest endpoint, playback URL (env/SSM da sessão).

### Relacionamentos

```text
Curso 1 ─── * Aula 1 ─── 0..1 AulaLive
                 └── 0..1 AulaVod   (002, inalterado)

Sessão AWS ── 1 canal IVS ── no máximo 1 AulaLive.status = ao_vivo
```

- Leitura de status / playback (quando `ao_vivo`): quem já lê aulas (admin, professor, aluno).
- Start / stop / ingest: admin + professor (qualquer aula; sem dono de curso).
- Anônimo: nenhuma leitura de playback/ingest.

### Invariante global (P1)

Em qualquer instante, **no máximo uma** linha com `status = ao_vivo` (canal compartilhado).

Implementação: checagem no repositório/handler **antes** do start + índice único parcial em Postgres quando possível (`UNIQUE (status) WHERE status = 'ao_vivo'`). Testes SQLite: mesma regra em código.

### Regras de validação

1. Start em aula inexistente → 404 PT.
2. Start com outra aula `ao_vivo` → **409** PT (“Já existe uma transmissão ao vivo. Encerre-a antes de iniciar.”). Não encerrar a anterior implicitamente.
3. Start **sem** `agendada_em` → **permitido**.
4. Start com `LIVE_BACKEND=ivs` e credenciais ausentes → **503** PT; **não** persistir `ao_vivo`.
5. Start com `LIVE_BACKEND=stub` → persiste `ao_vivo` (sem sinal).
6. Stop sem live `ao_vivo` nesta aula → 409/404 PT claro.
7. Reinício (após `encerrada`) na mesma aula ou em outra, mesma sessão de app/infra → permitido se o invariante (2) valer.
8. Aluno nunca recebe `stream_key` / ingest (nem no JSON de status).
9. Playback só se `status = ao_vivo` **desta** aula; senão 404 PT (“Não há transmissão ao vivo nesta aula.”).

### Transições de estado

**P1** (ação única “iniciar”; sem `agendada`)

```text
(sem linha | encerrada) --POST start--> ao_vivo
ao_vivo --POST stop--> encerrada
encerrada --POST start (mesma ou outra aula)--> ao_vivo   # se nenhuma outra ao_vivo
```

API de leitura para ficha **sem** linha: tratar como **inativa** (não ao vivo) — 200 com `status: "inativa"` no contrato, sem inventar linha.

**P2 (opcional)**

```text
(sem linha) --agendar (exige horário)--> agendada
agendada --iniciar--> ao_vivo
ao_vivo --encerrar--> encerrada
agendada --cancelar--> (inativa / sem linha)
```

Aluno **não** vê player ativo em `agendada` nem `encerrada`. Em `encerrada` + `AulaVod` publicado, UI MAY apontar a gravação (002).

## Sessão de streaming (não é tabela)

| Aspecto | Valor |
|---------|--------|
| Cardinalidade | 1 canal / sessão Terraform |
| Região | `us-east-1` |
| Lifecycle | Criado no `apply`; **destruído** no `destroy` |
| Stream key | Uma por canal; estável até destroy |
| Lives simultâneas | Não suportado |

## Seed

**Não** marcar a aula de demo como `ao_vivo` no migrate (evita live “fantasma” no local e na AWS sem OBS). Seed 001/002 (curso, aula, VOD) permanece. Operador/professor inicia a live na demo.
