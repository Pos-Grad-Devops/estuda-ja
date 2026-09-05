import { useEffect, useState } from 'react'
import type { FormEvent } from 'react'
import { api } from '../api/client'
import type { Aula, AulaVod, Curso } from '../api/client'
import { formatDateBR, isValidDateTimeBR, toInputDateTimeBR } from '../utils/date'

const VOD_MAX_BYTES = 52_428_800

function formatSizeMB(bytes: number): string {
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
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

  useEffect(() => {
    if (selectedAulaId == null) {
      setVod(null)
      setPlaybackUrl(null)
      setVodError('')
      return
    }
    void loadVod(selectedAulaId)
  }, [selectedAulaId])

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

  const selectedAula = items.find((a) => a.id === selectedAulaId) ?? null

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
        <div className="card">
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
      )}
    </section>
  )
}
