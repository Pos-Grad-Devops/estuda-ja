import { useCallback, useEffect, useState } from 'react'
import type { FormEvent } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { api } from '../api/client'
import type { Aula, AulaLive, AulaVod, Curso, LiveIngest, LivePlayback } from '../api/client'
import { canManageAulas } from '../auth/auth'
import { useAuth } from '../auth/AuthContext'
import { AulaChatPanel } from '../components/AulaChatPanel'
import FadeContent from '../components/bits/FadeContent'
import { AulaModal } from '../components/forms/AulaModal'
import { LivePlayer } from '../components/LivePlayer'
import { Button } from '../components/ui/Button'
import { EmptyState } from '../components/ui/EmptyState'
import { PageHeader } from '../components/ui/PageHeader'
import { StatusPill } from '../components/ui/StatusPill'
import { formatDateBR } from '../utils/date'

const VOD_MAX_BYTES = 52_428_800

function formatSizeMB(bytes: number): string {
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

function hasAgendadaEm(value: string | undefined | null): boolean {
  return Boolean(value && value.trim() !== '')
}

export function AulaPage() {
  const { id } = useParams()
  const navigate = useNavigate()
  const { user } = useAuth()
  const [aula, setAula] = useState<Aula | null>(null)
  const [cursos, setCursos] = useState<Curso[]>([])
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(true)
  const [modalOpen, setModalOpen] = useState(false)

  const [vod, setVod] = useState<AulaVod | null>(null)
  const [playbackUrl, setPlaybackUrl] = useState<string | null>(null)
  const [vodLoading, setVodLoading] = useState(false)
  const [vodError, setVodError] = useState('')
  const [vodBusy, setVodBusy] = useState(false)

  const [live, setLive] = useState<AulaLive | null>(null)
  const [livePlayback, setLivePlayback] = useState<LivePlayback | null>(null)
  const [liveIngest, setLiveIngest] = useState<LiveIngest | null>(null)
  const [liveLoading, setLiveLoading] = useState(false)
  const [liveError, setLiveError] = useState('')
  const [liveBusy, setLiveBusy] = useState(false)

  const aulaId = Number(id)
  const canWrite = Boolean(user && canManageAulas(user.role))

  const load = useCallback(async () => {
    if (!Number.isFinite(aulaId) || aulaId <= 0) {
      setError('Aula inválida')
      setLoading(false)
      return
    }
    setLoading(true)
    try {
      const [detail, cursosList] = await Promise.all([api.aulas.get(aulaId), api.cursos.list()])
      setAula(detail)
      setCursos(cursosList)
      setError('')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Erro ao carregar')
      setAula(null)
    } finally {
      setLoading(false)
    }
  }, [aulaId])

  const loadVod = useCallback(async (idAula: number) => {
    setVodLoading(true)
    setVodError('')
    setVod(null)
    setPlaybackUrl(null)
    try {
      const meta = await api.vod.get(idAula)
      setVod(meta)
      if (meta) {
        const playback = await api.vod.getPlayback(idAula)
        setPlaybackUrl(playback.playback_url)
      }
    } catch (err) {
      setVodError(err instanceof Error ? err.message : 'Erro ao carregar gravação')
    } finally {
      setVodLoading(false)
    }
  }, [])

  const loadLive = useCallback(
    async (idAula: number) => {
      setLiveLoading(true)
      setLiveError('')
      setLive(null)
      setLivePlayback(null)
      setLiveIngest(null)
      try {
        const statusLive = await api.live.get(idAula)
        setLive(statusLive)

        if (statusLive.status === 'ao_vivo') {
          try {
            const playback = await api.live.getPlayback(idAula)
            setLivePlayback(playback)
          } catch (err) {
            setLiveError(err instanceof Error ? err.message : 'Erro ao carregar playback ao vivo')
          }

          if (canWrite) {
            try {
              const ingest = await api.live.getIngest(idAula)
              setLiveIngest(ingest)
            } catch (err) {
              setLiveError(
                err instanceof Error ? err.message : 'Erro ao carregar credenciais de ingestão',
              )
            }
          }
        }
      } catch (err) {
        setLiveError(err instanceof Error ? err.message : 'Erro ao carregar transmissão ao vivo')
      } finally {
        setLiveLoading(false)
      }
    },
    [canWrite],
  )

  useEffect(() => {
    void load()
  }, [load])

  useEffect(() => {
    if (!Number.isFinite(aulaId) || aulaId <= 0) return
    void loadVod(aulaId)
    void loadLive(aulaId)
  }, [aulaId, loadVod, loadLive])

  async function handleSave(
    data: Pick<Aula, 'curso_id' | 'titulo' | 'descricao' | 'agendada_em' | 'status'>,
  ) {
    await api.aulas.update(aulaId, data)
    await load()
  }

  async function handleDelete() {
    if (!aula || !confirm(`Excluir a aula “${aula.titulo}”?`)) return
    await api.aulas.remove(aula.id)
    navigate(aula.curso_id ? `/cursos/${aula.curso_id}` : '/aulas', { replace: true })
  }

  async function handleVodUpload(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    const form = event.currentTarget
    const input = form.elements.namedItem('vodFile') as HTMLInputElement | null
    const file = input?.files?.[0]
    if (!file) {
      setVodError('Selecione um arquivo MP4')
      return
    }
    if (file.size > VOD_MAX_BYTES) {
      setVodError('Arquivo muito grande. O limite é 50 MB')
      return
    }
    const isMp4 = file.type === 'video/mp4' || file.name.toLowerCase().endsWith('.mp4')
    if (!isMp4) {
      setVodError('Formato inválido. Envie apenas arquivos MP4')
      return
    }

    setVodBusy(true)
    setVodError('')
    try {
      await api.vod.upload(aulaId, file)
      form.reset()
      await loadVod(aulaId)
    } catch (err) {
      setVodError(err instanceof Error ? err.message : 'Erro ao enviar gravação')
    } finally {
      setVodBusy(false)
    }
  }

  async function handleVodRemove() {
    if (!confirm('Remover a gravação desta aula?')) return
    setVodBusy(true)
    setVodError('')
    try {
      await api.vod.remove(aulaId)
      await loadVod(aulaId)
    } catch (err) {
      setVodError(err instanceof Error ? err.message : 'Erro ao remover gravação')
    } finally {
      setVodBusy(false)
    }
  }

  async function handleLiveSchedule() {
    setLiveBusy(true)
    setLiveError('')
    try {
      await api.live.schedule(aulaId)
      await loadLive(aulaId)
    } catch (err) {
      setLiveError(err instanceof Error ? err.message : 'Erro ao agendar transmissão')
    } finally {
      setLiveBusy(false)
    }
  }

  async function handleLiveCancel() {
    if (!confirm('Cancelar a transmissão agendada desta aula?')) return
    setLiveBusy(true)
    setLiveError('')
    try {
      await api.live.cancel(aulaId)
      await loadLive(aulaId)
    } catch (err) {
      setLiveError(err instanceof Error ? err.message : 'Erro ao cancelar agendamento')
    } finally {
      setLiveBusy(false)
    }
  }

  async function handleLiveStart() {
    setLiveBusy(true)
    setLiveError('')
    try {
      await api.live.start(aulaId)
      await loadLive(aulaId)
    } catch (err) {
      setLiveError(err instanceof Error ? err.message : 'Erro ao iniciar transmissão')
    } finally {
      setLiveBusy(false)
    }
  }

  async function handleLiveStop() {
    if (!confirm('Encerrar a transmissão ao vivo desta aula?')) return
    setLiveBusy(true)
    setLiveError('')
    try {
      await api.live.stop(aulaId)
      await loadLive(aulaId)
    } catch (err) {
      setLiveError(err instanceof Error ? err.message : 'Erro ao encerrar transmissão')
    } finally {
      setLiveBusy(false)
    }
  }

  if (!user) return null

  const liveIsActive = live?.status === 'ao_vivo'
  const liveIsScheduled = live?.status === 'agendada'
  const liveIsEnded = live?.status === 'encerrada'
  const liveIsInactive = !live || live.status === 'inativa'
  const missingSchedule = aula != null && !hasAgendadaEm(aula.agendada_em)
  const hasPublishedVod = vod != null && vod.status === 'publicado' && playbackUrl != null
  const canScheduleLive = canWrite && (liveIsInactive || liveIsEnded) && !missingSchedule
  const canStartLive = canWrite && !liveIsActive
  const canCancelScheduled = canWrite && liveIsScheduled

  return (
    <section>
      <div className="mb-4 flex flex-wrap gap-3 text-sm font-semibold">
        <Link to="/aulas" className="text-accent-hover hover:underline">
          ← Agenda
        </Link>
        {aula && (
          <Link to={`/cursos/${aula.curso_id}`} className="text-muted hover:text-ink">
            Curso
          </Link>
        )}
      </div>

      {error && <p className="mb-4 rounded-xl bg-live/12 px-3 py-2 text-sm text-live">{error}</p>}
      {loading && <p className="text-muted">Carregando aula...</p>}

      {aula && (
        <FadeContent>
          <PageHeader
            eyebrow={aula.curso?.titulo ?? `Curso #${aula.curso_id}`}
            title={aula.titulo}
            description={aula.descricao || 'Esta aula ainda não tem descrição.'}
            actions={
              canWrite ? (
                <>
                  <Button onClick={() => setModalOpen(true)}>Editar</Button>
                  <Button variant="danger" onClick={() => void handleDelete()}>
                    Excluir
                  </Button>
                </>
              ) : undefined
            }
          />
          <div className="mb-8 grid gap-4 sm:grid-cols-2">
            <div className="rounded-2xl border border-border bg-surface p-5">
              <p className="text-sm text-muted">Status da aula</p>
              <div className="mt-3">
                <StatusPill status={aula.status} />
              </div>
            </div>
            <div className="rounded-2xl border border-border bg-surface p-5">
              <p className="text-sm text-muted">Quando</p>
              <p className="mt-3 font-display text-xl font-semibold">
                {formatDateBR(aula.agendada_em)}
              </p>
            </div>
          </div>

          {/* —— Transmissão ao vivo (api.live / LiveStatus) —— */}
          <section id="live-block" className="mb-8">
            <PageHeader
              eyebrow="Ao vivo"
              title="Transmissão ao vivo"
              description="Status e player vêm da API de live — independente do status CRUD da aula."
            />

            {liveError && (
              <p className="mb-4 rounded-xl bg-live/12 px-3 py-2 text-sm text-live">{liveError}</p>
            )}

            {liveLoading ? (
              <EmptyState
                title="Carregando transmissão"
                description="Consultando o status ao vivo desta aula…"
              />
            ) : live ? (
              <div className="space-y-4">
                <div className="flex flex-wrap items-center gap-3 rounded-2xl border border-border bg-surface p-4">
                  <StatusPill status={live.status} />
                  {live.modo === 'stub' && (
                    <span className="text-sm text-muted">modo stub (local)</span>
                  )}
                  {live.iniciada_em && (
                    <span className="text-sm text-muted">
                      Iniciada em {formatDateBR(live.iniciada_em)}
                    </span>
                  )}
                  {liveIsEnded && live.encerrada_em && (
                    <span className="text-sm text-muted">
                      Encerrada em {formatDateBR(live.encerrada_em)}
                    </span>
                  )}
                </div>

                {liveIsScheduled && hasAgendadaEm(aula.agendada_em) && (
                  <p className="text-sm text-muted">
                    Horário previsto: {formatDateBR(aula.agendada_em)}. O player ao vivo só aparece
                    quando a transmissão iniciar.
                  </p>
                )}

                {liveIsActive && livePlayback ? (
                  <LivePlayer
                    modo={livePlayback.modo}
                    playbackUrl={livePlayback.playback_url}
                    mensagem={livePlayback.mensagem}
                  />
                ) : liveIsScheduled ? (
                  <EmptyState
                    title="Transmissão agendada"
                    description="Ainda não há sinal ao vivo nesta aula."
                  />
                ) : liveIsEnded ? (
                  <EmptyState
                    title="Transmissão encerrada"
                    description={
                      hasPublishedVod
                        ? 'Há uma gravação disponível no bloco abaixo (VOD — independente da live).'
                        : 'Não há gravação publicada para esta aula.'
                    }
                    action={
                      hasPublishedVod ? (
                        <a href="#gravacao-block" className="text-sm font-semibold text-accent-hover hover:underline">
                          Ir para gravação
                        </a>
                      ) : undefined
                    }
                  />
                ) : (
                  <EmptyState
                    title="Sem transmissão ao vivo"
                    description="Não há transmissão ao vivo nesta aula. A gravação (VOD), se existir, aparece no bloco abaixo."
                  />
                )}

                {canWrite && (
                  <div className="rounded-2xl border border-border bg-surface p-5">
                    {missingSchedule && (
                      <p className="mb-4 rounded-xl border border-accent/30 bg-accent/10 px-3 py-2 text-sm">
                        Recomendado: defina o horário agendado da aula antes de transmitir. Você
                        ainda pode iniciar a live sem horário. Agendar (estado “agendada”) exige
                        horário.
                      </p>
                    )}
                    <div className="flex flex-wrap gap-3">
                      {canScheduleLive && (
                        <Button
                          variant="secondary"
                          disabled={liveBusy || liveLoading}
                          onClick={() => void handleLiveSchedule()}
                        >
                          {liveBusy ? 'Agendando...' : 'Agendar transmissão'}
                        </Button>
                      )}
                      {canCancelScheduled && (
                        <Button
                          variant="secondary"
                          disabled={liveBusy || liveLoading}
                          onClick={() => void handleLiveCancel()}
                        >
                          {liveBusy ? 'Cancelando...' : 'Cancelar agendamento'}
                        </Button>
                      )}
                      {canStartLive && (
                        <Button
                          disabled={liveBusy || liveLoading}
                          onClick={() => void handleLiveStart()}
                        >
                          {liveBusy ? 'Iniciando...' : 'Iniciar transmissão'}
                        </Button>
                      )}
                      {liveIsActive && (
                        <Button
                          variant="danger"
                          disabled={liveBusy || liveLoading}
                          onClick={() => void handleLiveStop()}
                        >
                          {liveBusy ? 'Encerrando...' : 'Encerrar transmissão'}
                        </Button>
                      )}
                    </div>

                    {liveIsActive && liveIngest && (
                      <div className="mt-5 rounded-xl border border-border bg-bg/50 p-4">
                        <h3 className="font-display text-base font-semibold">Ingestão (OBS)</h3>
                        {liveIngest.modo === 'stub' || !liveIngest.stream_key ? (
                          <p className="mt-2 text-sm text-muted">
                            {liveIngest.mensagem ??
                              'Ingestão real só na sessão AWS. No ambiente local não há servidor nem stream key.'}
                          </p>
                        ) : (
                          <div className="mt-3 space-y-3">
                            <label className="field mb-0">
                              Servidor
                              <code className="mt-1 block break-all rounded-xl border border-border bg-surface px-3 py-2 text-sm font-normal">
                                {liveIngest.ingest_server}
                              </code>
                            </label>
                            <label className="field mb-0">
                              Stream key
                              <code className="mt-1 block break-all rounded-xl border border-border bg-surface px-3 py-2 text-sm font-normal">
                                {liveIngest.stream_key}
                              </code>
                            </label>
                            {liveIngest.observacao && (
                              <p className="text-sm text-muted">{liveIngest.observacao}</p>
                            )}
                          </div>
                        )}
                      </div>
                    )}
                  </div>
                )}
              </div>
            ) : (
              <EmptyState
                title="Status indisponível"
                description="Não foi possível obter o status da transmissão."
              />
            )}
          </section>

          {/* —— Gravação VOD —— */}
          <section id="gravacao-block" className="mb-8">
            <PageHeader
              eyebrow="VOD"
              title="Gravação"
              description="Assista a gravação publicada desta aula. Sem download do arquivo."
            />

            {vodError && (
              <p className="mb-4 rounded-xl bg-live/12 px-3 py-2 text-sm text-live">{vodError}</p>
            )}

            {vodLoading ? (
              <EmptyState
                title="Carregando gravação"
                description="Buscando metadados e playback desta aula…"
              />
            ) : vod && playbackUrl ? (
              <div className="overflow-hidden rounded-2xl border border-border bg-surface">
                <p className="border-b border-border px-4 py-2 text-sm text-muted">
                  Atualizada em {formatDateBR(vod.updated_at)} · {formatSizeMB(vod.size_bytes)}
                </p>
                <video
                  className="aspect-video w-full bg-bg"
                  controls
                  controlsList="nodownload"
                  preload="metadata"
                  src={playbackUrl}
                >
                  Seu navegador não suporta reprodução de vídeo.
                </video>
              </div>
            ) : (
              <EmptyState
                title="Nenhuma gravação"
                description="Nenhuma gravação disponível para esta aula."
                action={
                  canWrite ? (
                    <span className="text-sm text-muted">Publique um MP4 abaixo (máx. 50 MB).</span>
                  ) : undefined
                }
              />
            )}

            {canWrite && (
              <form
                className="mt-4 rounded-2xl border border-border bg-surface p-5"
                onSubmit={handleVodUpload}
              >
                <label className="field">
                  Enviar ou substituir MP4 (máx. 50 MB)
                  <input
                    type="file"
                    name="vodFile"
                    accept="video/mp4,.mp4"
                    disabled={vodBusy}
                  />
                </label>
                <div className="flex flex-wrap gap-3">
                  <Button type="submit" disabled={vodBusy}>
                    {vodBusy ? 'Enviando...' : vod ? 'Substituir gravação' : 'Publicar gravação'}
                  </Button>
                  {vod && (
                    <Button
                      type="button"
                      variant="danger"
                      disabled={vodBusy}
                      onClick={() => void handleVodRemove()}
                    >
                      Remover gravação
                    </Button>
                  )}
                </div>
              </form>
            )}
          </section>

          {/* —— Chat —— */}
          <section id="chat-block" className="mb-4">
            <PageHeader
              eyebrow="Chat"
              title="Chat da aula"
              description="Mensagens só desta conexão — ao reabrir a ficha o painel começa vazio."
            />
            <AulaChatPanel aulaId={aula.id} />
          </section>
        </FadeContent>
      )}

      <AulaModal
        open={modalOpen}
        aula={aula}
        cursos={cursos}
        onClose={() => setModalOpen(false)}
        onSubmit={handleSave}
      />
    </section>
  )
}
