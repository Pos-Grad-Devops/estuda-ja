## Template pra preencher

### Projeto: `EstudaJá — plataforma de cursos ao vivo (aula magna) + VOD`
### Equipe: `Gustavo Santos Arruda, Pedro Lucas dos Santos Ribeiro, Renan Roseno dos Santos, Victor da Silva Neves`

Cinco features essenciais do produto (das quais três entram como prioridade):

1. Transmissão ao vivo em massa (aula magna)
2. Chat em tempo real da aula
3. Biblioteca de gravações (VOD)
4. Certificado de conclusão do curso
5. Catálogo, agenda e acesso autenticado (cursos/aulas + papéis)

---

### 🥇 Feature 1 — `Transmissão ao vivo em massa (aula magna)`

- **Problema real que ela resolve:**
  A aula tem horário fixo e milhares de alunos batem na porta no mesmo minuto. Se o player não abre em T−0, a aula simplesmente não aconteceu — não dá para “ir escalando aos poucos”. Sem essa peça, o EstudaJá deixa de ser aula magna e vira só mais um site de curso.

- **Critério(s) de prioridade que mais pesaram:**
  Valor central do produto (é a proposta de valor); risco técnico e de negócio (pico instantâneo + SLA 99,9% na janela ao vivo); irreversibilidade do horário — falhar nesse minuto não tem replay.

- **Em uma frase o que seria a aplicação utópica** (a versão completa, dos sonhos):
  Canal elástico para dezenas de milhares, qualidade adaptativa, DVR, tokenização de playback, fallback automático, multi-região, audiência em tempo real e a live virando VOD sozinha no encerramento.

- **E qual seria um MVP comercializavel?** (a menor versão possível, ponta a ponta, entregável em ~3 dias):
  Uma aula agendada, professor entra ao vivo (OBS + um canal de streaming), aluno autenticado assiste na ficha da aula; estados simples (agendada / ao vivo / encerrada). Um canal, sem DVR, sem multi-região, sem gravação automática.

##### Fatias E2E (1 entrega por dia)

Cada dia amplia o caminho do usuário. Nada de “dia 1 = banco, dia 2 = API, dia 3 = tela”.

| Dia | O que o usuário já faz de ponta a ponta | Ainda não entra |
| --- | --- | --- |
| **1 — Assistir agora** | Aluno abre a ficha da aula, vê que está ao vivo e assiste no player (aula seed já ligada; player pode ser stub/HLS fixo). | Painel do professor, agenda, ingest OBS. |
| **2 — Ligar e desligar** | Professor inicia e encerra na ficha; aluno vê o estado mudar (agendada → ao vivo → encerrada) e só reproduz quando está ao vivo. | Canal real, horário obrigatório, OBS. |
| **3 — Transmitir de verdade** | Professor agenda o horário, manda o sinal (OBS + um canal) e a turma assiste o stream real na mesma ficha. Fecha o MVP. | DVR, multi-canal, live→VOD, tokenização. |

No dia 1 o aluno **já assiste** a aula magna. Os dias 2 e 3 só deixam o professor no controle e trocam o stub pelo stream de verdade.

---

### 🥈 Feature 2 — `Chat em tempo real da aula`

- **Problema real que ela resolve:**
  Na aula magna o aluno não levanta a mão. Sem um canal de pergunta e reação no mesmo momento da live, a experiência vira TV passiva: o professor não sente a turma e o aluno não participa.

- **Critério(s) de prioridade que mais pesaram:**
  Experiência (diferencia aula de broadcast); engajamento no pico de T−0; esforço menor que a live em si, mas alto impacto percebido; depende da aula existir, então entra logo depois da transmissão.

- **O "elefante" dela** (a versão completa, dos sonhos):
  Moderação, fila de perguntas, recados do professor, reações, histórico persistente, chat sincronizado com o replay, salas paralelas e limites sofisticados de abuso.

- **A primeira fatia** (a menor versão possível, ponta a ponta, entregável em ~3 dias):
  Sala por aula via WebSocket: texto curto, todos os papéis enviam e recebem enquanto estão conectados. Sem histórico ao reabrir, sem moderação, sem emojis/reação.

##### Fatias E2E (1 entrega por dia)

| Dia | O que o usuário já faz de ponta a ponta | Ainda não entra |
| --- | --- | --- |
| **1 — Mandar recado** | Aluno (e professor) abre a ficha da aula, envia um texto e a turma que está na mesma tela lê na hora (mesmo que seja uma sala só e atualize por polling). | WebSocket, nome de quem falou, uma sala por aula, limite de tamanho. |
| **2 — Conversar ao vivo** | A mensagem aparece na hora para todo mundo conectado (WebSocket); cada recado mostra quem falou. A aula já “tem voz”. | Isolamento entre aulas, teto de caracteres, sumir ao sair. |
| **3 — Chat daquela aula** | Cada aula tem a própria sala; texto curto (ex.: 500 caracteres); ao sair/reabrir a conversa some. Fecha o MVP. | Moderação, histórico, reações, fila de perguntas, replay. |

No dia 1 a turma **já conversa** na aula. Os dias 2 e 3 só deixam isso instantâneo, identificado e isolado por aula — não inventam o chat no último dia.

---

### 🥉 Feature 3 — `Biblioteca de gravações (VOD)`

- **Problema real que ela resolve:**
  Quem perdeu o horário fixo ou quer revisar não tem segunda chance. Sem o gravado, o valor do curso morre quando a live acaba — e o aluno que não entrou em T−0 está perdido.

- **Critério(s) de prioridade que mais pesaram:**
  Valor comercial (acesso continua depois da aula); cobre o SLA mais frouxo fora do horário ao vivo (99%); reduz a perda de quem não consegue estar no minuto da transmissão.

- **O "elefante" dela** (a versão completa, dos sonhos):
  Pipeline live→VOD automático, transcodificação, legendas, busca no conteúdo, CDN de mídia, playlists, DRM e download controlado.

- **A primeira fatia** (a menor versão possível, ponta a ponta, entregável em ~3 dias):
  Professor envia um MP4 na ficha da aula; aluno assiste no mesmo lugar. Um arquivo por aula, com trocar/apagar. Sem transcode, sem live→VOD, sem página “Biblioteca” separada.

##### Fatias E2E (1 entrega por dia)

| Dia | O que o usuário já faz de ponta a ponta | Ainda não entra |
| --- | --- | --- |
| **1 — Assistir a gravação** | Aluno abre a ficha da aula e assiste o MP4 já publicado (vídeo seed). Quem perdeu a live **já revisa**. | Upload do professor, trocar/apagar, página Biblioteca. |
| **2 — Publicar o vídeo** | Professor envia um MP4 na mesma ficha; o aluno passa a ver esse arquivo no player. | Substituir, remover, transcode, CDN. |
| **3 — Trocar ou remover** | Professor substitui ou apaga a gravação; a ficha mostra se há ou não vídeo. Fecha o MVP. | Live→VOD automático, legendas, busca, download, DRM. |

No dia 1 o aluno **já assiste** o gravado. Os dias 2 e 3 só entregam a publicação e a gestão ao professor — o valor da biblioteca não espera o upload existir.

---

### Ficou de fora (e por quê)

Listem pelo menos 2 features que a equipe considerou e decidiu **não**
priorizar agora. Uma linha de justificativa basta.

| Feature descartada | Por que não entrou entre as 3 |
| --- | --- |
| Certificado de conclusão do curso | Só faz sentido depois que o curso aconteceu; não desbloqueia a aula magna nem o pico de T−0. Credencial importante, mas dá para emitir depois. |
| Catálogo, agenda e acesso autenticado (cursos/aulas + papéis) | É a base operacional, mas é commodity: o primeiro ciclo vive com 1 curso, 1 aula e logins seed. O risco e o diferencial estão no live, no chat e no VOD. |

---
