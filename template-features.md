## Template pra preencher

### Projeto: `EstudaJá — plataforma de cursos ao vivo (aula magna) + VOD`

### Equipe: `Gustavo Santos Arruda, Pedro Lucas dos Santos Ribeiro, Renan Roseno dos Santos, Victor da Silva Neves`

Cinco features essenciais do produto (das quais três entram como prioridade):

1. Transmissão ao vivo em massa (aula magna)
2. Chat em tempo real da aula
3. Biblioteca de gravações (VOD)
4. Certificado de conclusão do curso
5. Catálogo, agenda e acesso autenticado (cursos/aulas + papéis)

As três priorizadas estão **na mesma ficha da aula** (`/aulas/:id`): bloco Transmissão ao vivo, bloco Gravação e Chat da aula. Login JWT. A live **não** gera o MP4. O chat **não** depende da live estar ligada.

---



### 🥇 Feature 1 — `Transmissão ao vivo em massa (aula magna)`

- **Problema real que ela resolve:**
A aula tem horário fixo e concentra os acessos no mesmo minuto. Se a transmissão não estiver pronta em T−0, perde-se o momento principal da aula — não dá para preparar a infraestrutura aos poucos depois que ela começou. Sem essa peça, o EstudaJá deixa de atender à proposta de aula magna e se limita ao conteúdo gravado.
- **Critério(s) de prioridade que mais pesaram:**
Valor central do produto; risco técnico e de negócio causado pelo pico instantâneo; meta de 99,9% de disponibilidade na janela ao vivo; e necessidade de preparar a transmissão antes do horário marcado.
- **Em uma frase o que seria a aplicação utópica** (a versão completa, dos sonhos):
Canal elástico para dezenas de milhares, qualidade adaptativa, DVR, tokenização de playback, fallback automático, multi-região, audiência em tempo real e a live virando VOD sozinha no encerramento.
- **E qual seria um MVP comercializável?** (a menor versão possível, ponta a ponta, entregável em ~3 dias):
Professor ou admin agenda, inicia e encerra a transmissão na ficha da aula; o aluno autenticado acompanha no player da mesma tela. Na demo AWS, o sinal é enviado pelo OBS para um único canal IVS. A live pode ficar inativa, agendada, ao vivo ou encerrada — estado separado do status cadastral da aula. Não há DVR, lives simultâneas nem conversão automática para VOD.



##### Fatias E2E (1 entrega por dia)

Cada dia amplia o caminho do usuário. Nada de “dia 1 = banco, dia 2 = API, dia 3 = tela”.


| Dia                                   | O que o usuário já faz de ponta a ponta                                                                                                                                                                                                                                                                                                                   | Ainda não entra                                            |
| ------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------- |
| **1 — Controlar o estado localmente** | Professor ou admin inicia e encerra a live na ficha da aula. O aluno autenticado abre a mesma aula e vê o estado correto; o player de live só é oferecido quando ela está **ao vivo**. No ambiente local, o backend `stub` permite testar fluxo e permissões, mas não fornece vídeo real nem credenciais de ingestão. A live não começa ligada pelo seed. | Sinal real por OBS/IVS e agendamento da live.              |
| **2 — Transmitir na AWS**             | Na demo AWS, professor ou admin inicia a live, obtém servidor e stream key na ficha e envia o sinal pelo OBS para o canal IVS da sessão. O aluno assiste no player da aula. A chave permanece a mesma até o `destroy`; encerrar permite iniciar novamente sem recriar a stack. O sistema mantém no máximo uma aula ao vivo por vez.                       | Agendamento; DVR; lives simultâneas; live→VOD.             |
| **3 — Agendar e preparar a abertura** | Professor ou admin agenda a transmissão de uma aula com data e hora. O aluno distingue os estados **agendada**, **ao vivo** e **encerrada**; o gestor pode cancelar, iniciar, encerrar e reagendar. A API ainda permite iniciar diretamente sem agendamento. Se houver VOD, ele continua como gravação separada.                                          | DVR, múltiplos canais, live→VOD e tokenização de playback. |


No primeiro dia, o fluxo de controle e RBAC já pode ser validado de ponta a ponta no ambiente local, sem fingir que existe vídeo. O segundo acrescenta o sinal real na AWS; o terceiro completa os estados e a preparação operacional que já existem no projeto.

---



### 🥈 Feature 2 — `Chat em tempo real da aula`

- **Problema real que ela resolve:**
Na aula magna o aluno não levanta a mão. Sem um canal de pergunta no mesmo momento da live, a experiência vira TV passiva: o professor não sente a turma e o aluno não participa.
- **Critério(s) de prioridade que mais pesaram:**
Experiência e participação durante a aula; integração direta com a ficha já existente; alto impacto percebido com escopo menor que o streaming; e manutenção das permissões já definidas para aluno, professor e admin.
- **O "elefante" dela** (a versão completa, dos sonhos):
Moderação, fila de perguntas, recados do professor, reações, histórico persistente, chat sincronizado com o replay, salas paralelas e limites sofisticados de abuso.
- **A primeira fatia** (a menor versão possível, ponta a ponta, entregável em ~3 dias):
Painel WebSocket na ficha da aula, com uma sala por aula e mensagens de até 500 caracteres. Aluno, professor e admin participam enquanto estão conectados. A sala funciona independentemente do estado da live e não mantém histórico ao reabrir. Moderação, reações e anexos ficam de fora.



##### Fatias E2E (1 entrega por dia)

Cada dia amplia o caminho do usuário. Nada de “dia 1 = banco, dia 2 = API, dia 3 = tela”.


| Dia                              | O que o usuário já faz de ponta a ponta                                                                                                                                                                                                                                                      | Ainda não entra                                                                       |
| -------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------- |
| **1 — Conversar na ficha**       | Dois usuários autenticados abrem a mesma aula e trocam mensagens em tempo real pelo painel WebSocket. Cada mensagem mostra o nome do autor, e o painel começa vazio quando a ficha é reaberta. O chat funciona com ou sem live ativa.                                                        | Isolamento entre aulas, validação do texto e execução na AWS.                         |
| **2 — Isolar e proteger a sala** | A conexão usa o JWT do login e fica vinculada à aula informada no endereço WebSocket. Mensagens de uma aula não aparecem em outra; textos vazios ou com mais de 500 caracteres são recusados. Aluno, professor e admin participam, mas o chat não concede ao aluno controles de live ou VOD. | Operação pela infraestrutura da demo; moderação; histórico.                           |
| **3 — Executar na demo AWS**     | O mesmo WebSocket funciona pela API ECS da sessão, com hub em memória e sem Redis AWS. O ALB usa `idle_timeout` de 3600 segundos, e o protocolo envia `ping/pon g` para manter a conexão ativa. Ao destruir a demo ou reiniciar a API, conexões e mensagens desaparecem.                     | Escala horizontal, histórico, moderação, reações, fila de perguntas e chat no replay. |


No primeiro dia a turma já conversa na ficha. O segundo acrescenta as validações, o isolamento por aula e o RBAC que o código aplica. O terceiro leva o mesmo protocolo para a infraestrutura efêmera da demo, sem prometer carga ou persistência que não foram implementadas.

---



### 🥉 Feature 3 — `Biblioteca de gravações (VOD)`

- **Problema real que ela resolve:**
Quem perdeu o horário fixo ou quer revisar não tem segunda chance. Sem o gravado, o valor do curso morre quando a live acaba — e o aluno que não entrou em T−0 está perdido.
- **Critério(s) de prioridade que mais pesaram:**
Continuidade do acesso depois da aula; valor para quem perdeu o horário ou quer revisar; meta de 99% de disponibilidade fora da janela ao vivo; e reaproveitamento da ficha de aula já usada para live e chat.
- **O "elefante" dela** (a versão completa, dos sonhos):
Pipeline live→VOD automático, transcodificação, legendas, busca no conteúdo, CDN de mídia, playlists, DRM e download controlado.
- **A primeira fatia** (a menor versão possível, ponta a ponta, entregável em ~3 dias):
Um MP4 publicado por aula, reproduzido na própria ficha e sem fluxo de download. Professor ou admin pode enviar, substituir ou remover o arquivo; aluno apenas assiste. Não há página “Biblioteca” dedicada nem conversão automática da live em gravação.



##### Fatias E2E (1 entrega por dia)

Cada dia amplia o caminho do usuário. Nada de “dia 1 = banco, dia 2 = API, dia 3 = tela”.


| Dia                                     | O que o usuário já faz de ponta a ponta                                                                                                                                                                                                                         | Ainda não entra                                                                                      |
| --------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------- |
| **1 — Assistir à gravação**             | Aluno autenticado abre a ficha da aula, identifica se há gravação publicada e assiste no player da própria tela. A aula de exemplo já recebe um MP4 pelo seed. Live e VOD aparecem como blocos separados; não existe página “Biblioteca” nem botão de download. | Publicação e remoção pelo gestor; armazenamento S3 da demo.                                          |
| **2 — Publicar, substituir ou remover** | Professor ou admin envia um MP4 de até 50 MB na mesma ficha. Cada aula mantém no máximo um VOD vigente: um novo envio substitui o anterior e remove o arquivo antigo; a remoção explícita tira a gravação da ficha. O aluno não recebe esses controles.         | Armazenamento e reprodução temporária na AWS; metadados e rascunho.                                  |
| **3 — Executar na demo AWS**            | Na AWS, os vídeos ficam no bucket S3 privado da sessão e a API entrega acesso temporário para reprodução. Depois de um novo `apply` limpo e da subida da API, a migração repõe o clipe da aula de exemplo. O `destroy` remove bucket e objetos da sessão.       | Live→VOD, transcodificação, legendas, busca, download, DRM, CloudFront de mídia e página Biblioteca. |


No primeiro dia, quem perdeu a live já consegue revisar a aula pelo vídeo de exemplo. O segundo entrega a gestão que professor e admin possuem no produto. O terceiro usa o mesmo fluxo na stack efêmera da AWS, sem criar uma biblioteca paralela.

---



### Ficou de fora (e por quê)

Listem pelo menos 2 features que a equipe considerou e decidiu **não**
priorizar agora. Uma linha de justificativa basta.


| Feature descartada                                            | Por que não entrou entre as 3                                                                                                                            |
| ------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Certificado de conclusão do curso                             | Só faz sentido depois que o curso aconteceu; não desbloqueia a aula magna nem o pico de T−0. Credencial importante, mas dá para emitir depois.           |
| Catálogo, agenda e acesso autenticado (cursos/aulas + papéis) | É a base operacional, mas é commodity: o primeiro ciclo vive com 1 curso, 1 aula e logins seed. O risco e o diferencial estão no live, no chat e no VOD. |


---

