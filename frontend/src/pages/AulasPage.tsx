import { useEffect, useState } from 'react'
import type { FormEvent } from 'react'
import { api } from '../api/client'
import type { Aula, AulaLive, AulaVod, Curso, LiveIngest, LivePlayback } from '../api/client'
import { LivePlayer } from '../components/LivePlayer'
import { formatDateBR, isValidDateTimeBR, toInputDateTimeBR } from '../utils/date'

const VOD_MAX_BYTES = 52_428_800

function formatSizeMB(bytes: number): string {
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

function liveStatusLabel(status: AulaLive['status']): string {
  switch (status) {
    case 'agendada':
      return 'Agendada'
    case 'ao_vivo':
      return 'Ao vivo'
    case 'encerrada':
      return 'Encerrada'
    default:
      return 'Inativa'
  }
}

function liveStatusClass(status: AulaLive['status']): string | undefined {
  switch (status) {
    case 'agendada':
      return 'live-status-agendada'
    case 'ao_vivo':
      return 'live-status-ao_vivo'
    case 'encerrada':
      return 'live-status-encerrada'
    default:
      return undefined
  }
}

function hasAgendadaEm(value: string | undefined | null): boolean {
  return Boolean(value && value.trim() !== '')
}

export function AulasPage({ canWrite = false }: { canWrite?: boolean }) {
  const [items, setItems] = useState<Aula[]>([])
  const [cursos, setCursos] = useState<Curso[]>([])
  const [cursoId, setCursoId] = useState('')
  const [titulo, setTitulo] = useState('')
  const [descricao, setDescricao] = useState('')
  const [agendadaEm, setAgendadaEm] = useState('')
  const [status, setStatus] = useState<Aula['status']>('agendada')
  const [editingId, setEditingId] = useState<number | null>(null)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(true)

  const [selectedAulaId, setSelectedAulaId] = useState<number | null>(null)
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

  async function load() {
    setLoading(true)
    try {
      const [aulas, cursosList] = await Promise.all([api.aulas.list(), api.cursos.list()])
      setItems(aulas)
      setCursos(cursosList)
      setError('')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Erro ao carregar')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    void load()
  }, [])

  async function loadVod(aulaId: number) {
    setVodLoading(true)
    setVodError('')
    setVod(null)
    setPlaybackUrl(null)
    try {
      const meta = await api.vod.get(aulaId)
      setVod(meta)
      if (meta) {
        const playback = await api.vod.getPlayback(aulaId)
        setPlaybackUrl(playback.playback_url)
      }
    } catch (err) {
      setVodError(err instanceof Error ? err.message : 'Erro ao carregar gravação')
    } finally {
      setVodLoading(false)
    }
  }

  async function loadLive(aulaId: number) {
    setLiveLoading(true)
    setLiveError('')
    setLive(null)
    setLivePlayback(null)
    setLiveIngest(null)
    try {
      const statusLive = await api.live.get(aulaId)
      setLive(statusLive)

      if (statusLive.status === 'ao_vivo') {
        try {
          const playback = await api.live.getPlayback(aulaId)
          setLivePlayback(playback)
        } catch (err) {
          setLiveError(err instanceof Error ? err.message : 'Erro ao carregar playback ao vivo')
        }

        if (canWrite) {
          try {
            const ingest = await api.live.getIngest(aulaId)
            setLiveIngest(ingest)
          } catch (err) {
            setLiveError(err instanceof Error ? err.message : 'Erro ao carregar credenciais de ingestão')
          }
        }
      }
    } catch (err) {
      setLiveError(err instanceof Error ? err.message : 'Erro ao carregar transmissão ao vivo')
    } finally {
      setLiveLoading(false)
    }
  }

  useEffect(() => {
    if (selectedAulaId == null) {
      setVod(null)
      setPlaybackUrl(null)
      setVodError('')
      setLive(null)
      setLivePlayback(null)
      setLiveIngest(null)
      setLiveError('')
      return
    }
    void loadVod(selectedAulaId)
    void loadLive(selectedAulaId)
  }, [selectedAulaId, canWrite])

  function resetForm() {
    setCursoId('')
    setTitulo('')
    setDescricao('')
    setAgendadaEm('')
    setStatus('agendada')
    setEditingId(null)
  }

  async function handleSubmit(event: FormEvent) {
    event.preventDefault()
    if (!isValidDateTimeBR(agendadaEm)) {
      setError('Data inválida. Use o formato dd/mm/yyyy ou dd/mm/yyyy HH:mm')
      return
    }
    const payload = {
      curso_id: Number(cursoId),
      titulo,
      descricao,
      agendada_em: agendadaEm.trim(),
      status,
    }
    try {
      if (editingId) {
        await api.aulas.update(editingId, payload)
      } else {
        await api.aulas.create(payload)
      }
      resetForm()
      await load()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Erro ao salvar')
    }
  }

  function handleEdit(aula: Aula) {
    setEditingId(aula.id)
    setCursoId(String(aula.curso_id))
    setTitulo(aula.titulo)
    setDescricao(aula.descricao)
    setAgendadaEm(toInputDateTimeBR(aula.agendada_em))
    setStatus(aula.status)
    setSelectedAulaId(aula.id)
  }

  async function handleDelete(id: number) {
    if (!confirm('Excluir esta aula?')) return
    try {
      await api.aulas.remove(id)
      if (selectedAulaId === id) {
        setSelectedAulaId(null)
      }
      await load()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Erro ao excluir')
    }
  }

  async function handleVodUpload(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (selectedAulaId == null) return
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
    const isMp4 =
      file.type === 'video/mp4' || file.name.toLowerCase().endsWith('.mp4')
    if (!isMp4) {
      setVodError('Formato inválido. Envie apenas arquivos MP4')
      return
    }

    setVodBusy(true)
    setVodError('')
    try {
      await api.vod.upload(selectedAulaId, file)
      form.reset()
      await loadVod(selectedAulaId)
    } catch (err) {
      setVodError(err instanceof Error ? err.message : 'Erro ao enviar gravação')
    } finally {
      setVodBusy(false)
    }
  }

  async function handleVodRemove() {
    if (selectedAulaId == null) return
    if (!confirm('Remover a gravação desta aula?')) return
    setVodBusy(true)
    setVodError('')
    try {
      await api.vod.remove(selectedAulaId)
      await loadVod(selectedAulaId)
    } catch (err) {
      setVodError(err instanceof Error ? err.message : 'Erro ao remover gravação')
    } finally {
      setVodBusy(false)
    }
  }

  async function handleLiveSchedule() {
    if (selectedAulaId == null) return
    setLiveBusy(true)
    setLiveError('')
    try {
      await api.live.schedule(selectedAulaId)
      await loadLive(selectedAulaId)
    } catch (err) {
      setLiveError(err instanceof Error ? err.message : 'Erro ao agendar transmissão')
    } finally {
      setLiveBusy(false)
    }
  }

  async function handleLiveCancel() {
    if (selectedAulaId == null) return
    if (!confirm('Cancelar a transmissão agendada desta aula?')) return
    setLiveBusy(true)
    setLiveError('')
    try {
      await api.live.cancel(selectedAulaId)
      await loadLive(selectedAulaId)
    } catch (err) {
      setLiveError(err instanceof Error ? err.message : 'Erro ao cancelar agendamento')
    } finally {
      setLiveBusy(false)
    }
  }

  async function handleLiveStart() {
    if (selectedAulaId == null) return
    setLiveBusy(true)
    setLiveError('')
    try {
      await api.live.start(selectedAulaId)
      await loadLive(selectedAulaId)
    } catch (err) {
      setLiveError(err instanceof Error ? err.message : 'Erro ao iniciar transmissão')
    } finally {
      setLiveBusy(false)
    }
  }

  async function handleLiveStop() {
    if (selectedAulaId == null) return
    if (!confirm('Encerrar a transmissão ao vivo desta aula?')) return
    setLiveBusy(true)
    setLiveError('')
    try {
      await api.live.stop(selectedAulaId)
      await loadLive(selectedAulaId)
    } catch (err) {
      setLiveError(err instanceof Error ? err.message : 'Erro ao encerrar transmissão')
    } finally {
      setLiveBusy(false)
    }
  }

  const selectedAula = items.find((a) => a.id === selectedAulaId) ?? null
  const liveIsActive = live?.status === 'ao_vivo'
  const liveIsScheduled = live?.status === 'agendada'
  const liveIsEnded = live?.status === 'encerrada'
  const liveIsInactive = !live || live.status === 'inativa'
  const missingSchedule = selectedAula != null && !hasAgendadaEm(selectedAula.agendada_em)
  const hasPublishedVod = vod != null && vod.status === 'publicado' && playbackUrl != null
  const canScheduleLive = canWrite && (liveIsInactive || liveIsEnded) && !missingSchedule
  const canStartLive = canWrite && !liveIsActive
  const canCancelScheduled = canWrite && liveIsScheduled

  return (
    <section>
      <h1>Aulas</h1>
      {error && <p className="error">{error}</p>}

      {canWrite && (
        <form className="card" onSubmit={handleSubmit}>
          <h2>{editingId ? 'Editar aula' : 'Nova aula'}</h2>
          <label>
            Curso
            <select value={cursoId} onChange={(e) => setCursoId(e.target.value)} required>
              <option value="">Selecione</option>
              {cursos.map((curso) => (
                <option key={curso.id} value={curso.id}>
                  {curso.titulo}
                </option>
              ))}
            </select>
          </label>
          <label>
            Título
            <input value={titulo} onChange={(e) => setTitulo(e.target.value)} required />
          </label>
          <label>
            Descrição
            <textarea value={descricao} onChange={(e) => setDescricao(e.target.value)} rows={3} />
          </label>
          <label>
            Agendada em
            <input
              type="text"
              value={agendadaEm}
              onChange={(e) => setAgendadaEm(e.target.value)}
              placeholder="dd/mm/yyyy HH:mm"
              required
            />
          </label>
          <label>
            Status
            <select value={status} onChange={(e) => setStatus(e.target.value as Aula['status'])}>
              <option value="agendada">Agendada</option>
              <option value="ao_vivo">Ao vivo</option>
              <option value="encerrada">Encerrada</option>
            </select>
          </label>
          <div className="actions">
            <button type="submit">{editingId ? 'Salvar' : 'Cadastrar'}</button>
            {editingId && (
              <button type="button" className="secondary" onClick={resetForm}>
                Cancelar
              </button>
            )}
          </div>
        </form>
      )}

      <div className="card">
        <h2>Lista</h2>
        {loading ? (
          <p>Carregando...</p>
        ) : items.length === 0 ? (
          <p>Nenhuma aula cadastrada.</p>
        ) : (
          <table>
            <thead>
              <tr>
                <th>ID</th>
                <th>Curso</th>
                <th>Título</th>
                <th>Agendada em</th>
                <th>Status</th>
                <th>Ficha</th>
                {canWrite && <th>Ações</th>}
              </tr>
            </thead>
            <tbody>
              {items.map((aula) => (
                <tr key={aula.id}>
                  <td>{aula.id}</td>
                  <td>{aula.curso?.titulo ?? aula.curso_id}</td>
                  <td>{aula.titulo}</td>
                  <td>{formatDateBR(aula.agendada_em)}</td>
                  <td>{aula.status}</td>
                  <td>
                    <button
                      type="button"
                      className="secondary"
                      onClick={() => setSelectedAulaId(aula.id)}
                    >
                      {selectedAulaId === aula.id ? 'Selecionada' : 'Abrir'}
                    </button>
                  </td>
                  {canWrite && (
                    <td className="actions">
                      <button type="button" className="secondary" onClick={() => handleEdit(aula)}>
                        Editar
                      </button>
                      <button type="button" className="danger" onClick={() => handleDelete(aula.id)}>
                        Excluir
                      </button>
                    </td>
                  )}
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>

      {selectedAula && (
        <>
          <div className="card" id="live-block">
            <h2>Transmissão ao vivo</h2>
            <p>
              Aula: <strong>{selectedAula.titulo}</strong>
            </p>

            {liveError && <p className="error">{liveError}</p>}

            {liveLoading ? (
              <p>Carregando transmissão...</p>
            ) : live ? (
              <>
                <p>
                  Status:{' '}
                  <span className={liveStatusClass(live.status)}>
                    {liveStatusLabel(live.status)}
                  </span>
                  {live.modo === 'stub' && <span className="muted"> · modo stub (local)</span>}
                </p>
                {liveIsScheduled && hasAgendadaEm(selectedAula.agendada_em) && (
                  <p className="muted">
                    Horário previsto: {formatDateBR(selectedAula.agendada_em)}. O player ao vivo
                    só aparece quando a transmissão iniciar.
                  </p>
                )}
                {live.iniciada_em && (
                  <p className="muted">Iniciada em {formatDateBR(live.iniciada_em)}</p>
                )}
                {liveIsEnded && live.encerrada_em && (
                  <p className="muted">Encerrada em {formatDateBR(live.encerrada_em)}</p>
                )}

                {liveIsActive && livePlayback ? (
                  <LivePlayer
                    modo={livePlayback.modo}
                    playbackUrl={livePlayback.playback_url}
                    mensagem={livePlayback.mensagem}
                  />
                ) : liveIsScheduled ? (
                  <p className="muted">
                    Transmissão agendada — ainda não há sinal ao vivo nesta aula.
                  </p>
                ) : liveIsEnded ? (
                  <p className="muted">
                    Transmissão encerrada.
                    {hasPublishedVod ? (
                      <>
                        {' '}
                        Há uma{' '}
                        <a href="#gravacao-block">gravação disponível</a> no bloco abaixo (VOD —
                        independente da live).
                      </>
                    ) : (
                      <> Não há gravação publicada para esta aula.</>
                    )}
                  </p>
                ) : (
                  <p className="muted">
                    Não há transmissão ao vivo nesta aula. A gravação (VOD), se existir, aparece no
                    bloco abaixo.
                  </p>
                )}
              </>
            ) : (
              <p className="muted">Não foi possível obter o status da transmissão.</p>
            )}

            {canWrite && (
              <div className="live-manage">
                {missingSchedule && (
                  <p className="warning">
                    Recomendado: defina o horário agendado da aula antes de transmitir. Você ainda
                    pode iniciar a live sem horário. Agendar (estado “agendada”) exige horário.
                  </p>
                )}
                <div className="actions">
                  {canScheduleLive && (
                    <button
                      type="button"
                      className="secondary"
                      disabled={liveBusy || liveLoading}
                      onClick={() => void handleLiveSchedule()}
                    >
                      {liveBusy ? 'Agendando...' : 'Agendar transmissão'}
                    </button>
                  )}
                  {canCancelScheduled && (
                    <button
                      type="button"
                      className="secondary"
                      disabled={liveBusy || liveLoading}
                      onClick={() => void handleLiveCancel()}
                    >
                      {liveBusy ? 'Cancelando...' : 'Cancelar agendamento'}
                    </button>
                  )}
                  {canStartLive && (
                    <button
                      type="button"
                      disabled={liveBusy || liveLoading}
                      onClick={() => void handleLiveStart()}
                    >
                      {liveBusy ? 'Iniciando...' : 'Iniciar transmissão'}
                    </button>
                  )}
                  {liveIsActive && (
                    <button
                      type="button"
                      className="danger"
                      disabled={liveBusy || liveLoading}
                      onClick={() => void handleLiveStop()}
                    >
                      {liveBusy ? 'Encerrando...' : 'Encerrar transmissão'}
                    </button>
                  )}
                </div>

                {liveIsActive && liveIngest && (
                  <div className="live-ingest">
                    <h3>Ingestão (OBS)</h3>
                    {liveIngest.modo === 'stub' || !liveIngest.stream_key ? (
                      <p className="muted">
                        {liveIngest.mensagem ??
                          'Ingestão real só na sessão AWS. No ambiente local não há servidor nem stream key.'}
                      </p>
                    ) : (
                      <>
                        <label>
                          Servidor
                          <code>{liveIngest.ingest_server}</code>
                        </label>
                        <label>
                          Stream key
                          <code>{liveIngest.stream_key}</code>
                        </label>
                        {liveIngest.observacao && (
                          <p className="muted">{liveIngest.observacao}</p>
                        )}
                      </>
                    )}
                  </div>
                )}
              </div>
            )}
          </div>

          <div className="card" id="gravacao-block">
            <h2>Gravação</h2>
            <p>
              Aula: <strong>{selectedAula.titulo}</strong>
            </p>

            {vodError && <p className="error">{vodError}</p>}

            {vodLoading ? (
              <p>Carregando gravação...</p>
            ) : vod && playbackUrl ? (
              <div className="vod-block">
                <p className="muted">
                  Atualizada em {formatDateBR(vod.updated_at)} · {formatSizeMB(vod.size_bytes)}
                </p>
                <video
                  className="vod-player"
                  controls
                  controlsList="nodownload"
                  preload="metadata"
                  src={playbackUrl}
                >
                  Seu navegador não suporta reprodução de vídeo.
                </video>
              </div>
            ) : (
              <p className="muted">Nenhuma gravação disponível para esta aula.</p>
            )}

            {canWrite && (
              <div className="vod-manage">
                <form onSubmit={handleVodUpload}>
                  <label>
                    Enviar ou substituir MP4 (máx. 50 MB)
                    <input
                      type="file"
                      name="vodFile"
                      accept="video/mp4,.mp4"
                      disabled={vodBusy}
                    />
                  </label>
                  <div className="actions">
                    <button type="submit" disabled={vodBusy}>
                      {vodBusy ? 'Enviando...' : vod ? 'Substituir gravação' : 'Publicar gravação'}
                    </button>
                    {vod && (
                      <button
                        type="button"
                        className="danger"
                        disabled={vodBusy}
                        onClick={() => void handleVodRemove()}
                      >
                        Remover gravação
                      </button>
                    )}
                  </div>
                </form>
              </div>
            )}
          </div>
        </>
      )}
    </section>
  )
}
