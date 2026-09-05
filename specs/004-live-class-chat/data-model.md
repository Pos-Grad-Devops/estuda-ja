# Data Model: 004-live-class-chat

Phase 1 — entidades lógicas do chat P1. **Sem migration** de mensagens: nada é persistido no Postgres além das entidades já existentes (`Aula`, `User`).

## Visão geral

```text
User ──(participa via WS)──► SalaChat(aula_id) ──(fan-out em memória)──► MensagemEmVoo
         ▲                         │
         │                         └── 1:1 com Aula (existe no DB)
         └── nome/role lidos no handshake (User no DB; JWT só id/email/role)
```

## Entidade: Sala de chat da aula (`SalaChat`)

| Campo / atributo | Tipo | Persistência | Notas |
|------------------|------|--------------|-------|
| `aula_id` | `uint` | **Não** (chave do mapa em memória) | = `Aulas.id`; sala só existe enquanto houver ≥1 conexão **ou** até o hub limpar sala vazia |
| `clientes` | set de conexões | memória | Isolamento: broadcast nunca cruza `aula_id` |
| Disponibilidade | — | — | **Sempre** que a aula existir no DB + usuário autenticado com leitura — **não** depende de `AulaLive.status` |

**Regras:**
- Criação implícita no primeiro connect válido para aquele `aula_id`.
- Remoção da entrada do mapa quando o último cliente desconecta (implementação MAY adiar GC curto; MUST não vazar salas indefinidamente).
- Reinício da task API → todas as salas somem (esperado).

**Relacionamentos:**
- 1 sala lógica ↔ 1 `Aula`.
- N participantes (conexões) por sala.
- Não há FK nova; validação = `FindAula(id)` no handshake.

## Entidade: Mensagem em voo (`MensagemEmVoo`)

Não há tabela `chat_messages` / `aula_chat_*` no P1.

| Campo | Tipo | Obrigatório | Validação / notas |
|-------|------|-------------|-------------------|
| `id` | string (UUID) | sim | Gerado no servidor no accept; só para UI da sessão atual |
| `aula_id` | `uint` | sim | Igual à sala da conexão; cliente **não** pode forçar outra aula no send |
| `autor_id` | `uint` | sim | De `claims.UserID` |
| `autor_nome` | string | sim | Lookup `User.Nome` no handshake (ou no send se preferir cache na conexão) |
| `autor_role` | enum | opcional no payload | `admin` \| `professor` \| `aluno` — MAY omitir na UI P1 |
| `texto` | string | sim | Após `strings.TrimSpace`: length ≥ 1 e ≤ **500**; Unicode = runes/código conforme Go `utf8.RuneCountInString` (documentar no contract) |
| `enviado_em` | instante | sim | Servidor; JSON exibido como `dd/mm/yyyy HH:mm` (timeutil / `utils/date.ts`) |

**Ciclo de vida:**
1. Cliente envia `chat.send` com `texto`.
2. Servidor valida → cria `MensagemEmVoo` → broadcast `chat.message` aos clientes **ainda conectados** na mesma sala (incluindo o remetente, para eco consistente).
3. Clientes guardam a mensagem **só no estado React** da conexão atual.
4. Disconnect / reabrir ficha → lista local vazia; servidor **não** reenvia histórico.

## Entidade: Participante (conexão)

| Campo | Tipo | Notas |
|-------|------|-------|
| `conn` | WebSocket | 1 aula por conexão |
| `user_id` | `uint` | Do JWT |
| `nome` | string | Cacheado na conexão após lookup |
| `role` | Role | Do JWT |
| `aula_id` | `uint` | Do path |

Não há lista “quem está online” persistida nem evento `presence` obrigatório no P1 (MAY omitir).

## Estado / transições

### Conexão

```text
[handshake] --token+aula OK--> Conectado --send válido--> (broadcast)
     |                              |
     +-- falha auth/aula -----------> Rejeitado (sem sala)
     |
Conectado --close/erro/rede--> Desconectado (remove do hub; sem replay)
```

### Mensagem

```text
(texto inválido) --> chat.error (PT); não broadcast
(texto válido)   --> chat.message → todos na sala conectados
```

**Não há** estados de moderação (apagada/silenciada) no P1.

## O que NÃO entra no modelo P1

- Tabela ou índice de histórico de chat
- Soft-delete / mute / ban
- Anexos, reações, threads
- Sala global / DM
- Dependência de `AulaLive` ou `AulaVod`
- Redis keys / streams

## Impacto em models existentes

| Model | Mudança P1 |
|-------|------------|
| `Aula` | Nenhuma (só leitura no handshake) |
| `User` | Nenhuma (só leitura de `Nome`) |
| `AulaLive` / `AulaVod` | Nenhuma |

AutoMigrate: **sem** novos models GORM para chat no P1.
