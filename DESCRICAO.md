# Projeto Fictício — Aula 1 (Projeto de Software com DevOps)

## 5. EstudaJá — Plataforma de Cursos ao Vivo

**Contexto:** EdTech que oferece aulas ao vivo em massa (estilo "aula magna" com milhares
de alunos simultâneos) e conteúdo gravado sob demanda.

**Escopo:** transmissão ao vivo, chat da aula, biblioteca de vídeos gravados, emissão de
certificado ao final do curso.

**Restrição especial:** a aula ao vivo tem horário fixo e não pode simplesmente "escalar
aos poucos" — o pico de acesso é instantâneo no minuto que a aula começa.

**SLA desejado:** 99.9% de disponibilidade durante o horário de transmissão ao vivo;
fora desse horário, 99% já é aceitável.

**Stack/infra sugerida:** bom cenário pra CDN/streaming e para discutir "cold start" de
infraestrutura que precisa estar pronta num instante específico, não gradualmente.

**Pico de uso:** pico extremo e instantâneo, concentrado em horários agendados.

