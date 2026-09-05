import { useEffect, useState } from 'react'
import type { FormEvent } from 'react'
import { api } from '../api/client'
import type {
  CertificadoResumo,
  CertificadoStatusResponse,
  CertificadosListResponse,
} from '../api/client'
import { canManageCertificados } from '../auth/auth'
import { useAuth } from '../auth/AuthContext'

export type CursoCertificadoPanelProps = {
  cursoId: number
}

export function CursoCertificadoPanel({ cursoId }: CursoCertificadoPanelProps) {
  const { user } = useAuth()
  const isAdmin = user != null && canManageCertificados(user.role)
  const isAluno = user?.role === 'aluno'

  // Professor: sem status nem gestão no P1
  if (!user || (!isAluno && !isAdmin)) {
    return null
  }

  return (
    <div className="certificado-panel">
      <h3>Certificado</h3>
      {isAluno && <AlunoCertificadoBlock cursoId={cursoId} />}
      {isAdmin && <AdminCertificadoBlock cursoId={cursoId} />}
    </div>
  )
}

function AlunoCertificadoBlock({ cursoId }: { cursoId: number }) {
  const [status, setStatus] = useState<CertificadoStatusResponse | null>(null)
  const [error, setError] = useState('')
  const [info, setInfo] = useState('')
  const [loading, setLoading] = useState(true)
  const [downloading, setDownloading] = useState(false)

  async function load() {
    setLoading(true)
    try {
      setStatus(await api.certificados.getStatus(cursoId))
      setError('')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Erro ao carregar status')
      setStatus(null)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    void load()
  }, [cursoId])

  async function handleDownload() {
    setDownloading(true)
    setInfo('')
    setError('')
    try {
      await api.certificados.downloadPdf(cursoId)
      setInfo('Download do certificado iniciado.')
      await load()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Erro ao baixar certificado')
    } finally {
      setDownloading(false)
    }
  }

  if (loading) {
    return <p className="muted">Carregando status do certificado…</p>
  }

  const cert = status?.certificado
  const elegivel = status?.elegivel === true
  const podeBaixar =
    elegivel || (cert != null && cert.status === 'valido')

  return (
    <div>
      {error && <p className="error">{error}</p>}
      {info && <p className="muted">{info}</p>}

      {!elegivel && !cert && (
        <p className="muted">Você não está elegível para o certificado deste curso.</p>
      )}

      {!elegivel && cert?.status === 'invalidado' && (
        <p className="muted">
          Certificado invalidado (emitido em {cert.emitido_em}). Aguarde reabilitação pelo
          administrador.
        </p>
      )}

      {elegivel && !cert && (
        <p>Você está elegível. Solicite o certificado para emitir e baixar o PDF.</p>
      )}

      {cert?.status === 'valido' && (
        <p>
          Certificado válido · emitido em {cert.emitido_em}
          {cert.aluno_nome ? ` · ${cert.aluno_nome}` : ''}
        </p>
      )}

      <div className="actions">
        <button type="button" onClick={() => void handleDownload()} disabled={downloading || !podeBaixar}>
          {downloading
            ? 'Baixando…'
            : cert?.status === 'valido'
              ? 'Baixar PDF'
              : 'Solicitar / baixar certificado'}
        </button>
      </div>
    </div>
  )
}

function AdminCertificadoBlock({ cursoId }: { cursoId: number }) {
  const [lista, setLista] = useState<CertificadosListResponse | null>(null)
  const [userIdInput, setUserIdInput] = useState('')
  const [consultaUserId, setConsultaUserId] = useState('')
  const [consulta, setConsulta] = useState<CertificadoStatusResponse | null>(null)
  const [error, setError] = useState('')
  const [info, setInfo] = useState('')
  const [loading, setLoading] = useState(true)
  const [busy, setBusy] = useState(false)

  async function loadLista() {
    setLoading(true)
    try {
      setLista(await api.certificados.list(cursoId))
      setError('')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Erro ao listar certificados')
      setLista(null)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    void loadLista()
  }, [cursoId])

  async function handleMarcar(event: FormEvent) {
    event.preventDefault()
    const userId = Number(userIdInput)
    if (!Number.isInteger(userId) || userId <= 0) {
      setError('Informe um user_id válido de aluno')
      return
    }
    setBusy(true)
    setInfo('')
    setError('')
    try {
      await api.certificados.marcarElegibilidade(cursoId, userId)
      setInfo(`Elegibilidade marcada/reabilitada para user_id ${userId} (sem emitir PDF).`)
      setUserIdInput('')
      await loadLista()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Erro ao marcar elegibilidade')
    } finally {
      setBusy(false)
    }
  }

  async function handleConsultar(event: FormEvent) {
    event.preventDefault()
    const userId = Number(consultaUserId)
    if (!Number.isInteger(userId) || userId <= 0) {
      setError('Informe um user_id válido para consulta')
      return
    }
    setBusy(true)
    setError('')
    setInfo('')
    try {
      setConsulta(await api.certificados.getStatus(cursoId, userId))
    } catch (err) {
      setConsulta(null)
      setError(err instanceof Error ? err.message : 'Erro ao consultar')
    } finally {
      setBusy(false)
    }
  }

  async function handleInvalidar(cert: CertificadoResumo) {
    if (!confirm(`Invalidar certificado #${cert.id}?`)) return
    setBusy(true)
    setError('')
    setInfo('')
    try {
      await api.certificados.invalidar(cursoId, cert.id)
      setInfo(`Certificado #${cert.id} invalidado; elegibilidade removida.`)
      setConsulta(null)
      await loadLista()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Erro ao invalidar')
    } finally {
      setBusy(false)
    }
  }

  return (
    <div>
      {error && <p className="error">{error}</p>}
      {info && <p className="muted">{info}</p>}

      <form onSubmit={handleMarcar}>
        <label>
          Marcar / reabilitar elegibilidade (user_id do aluno)
          <input
            type="number"
            min={1}
            value={userIdInput}
            onChange={(e) => setUserIdInput(e.target.value)}
            placeholder="ex.: 3"
            required
          />
        </label>
        <div className="actions">
          <button type="submit" disabled={busy}>
            Marcar elegibilidade
          </button>
        </div>
      </form>

      <form onSubmit={handleConsultar}>
        <label>
          Consultar emissão (user_id)
          <input
            type="number"
            min={1}
            value={consultaUserId}
            onChange={(e) => setConsultaUserId(e.target.value)}
            placeholder="ex.: 3"
            required
          />
        </label>
        <div className="actions">
          <button type="submit" className="secondary" disabled={busy}>
            Consultar
          </button>
        </div>
      </form>

      {consulta && (
        <p className="muted">
          user_id {consulta.user_id}: elegível={consulta.elegivel ? 'sim' : 'não'}
          {consulta.certificado
            ? ` · cert #${consulta.certificado.id} (${consulta.certificado.status}, ${consulta.certificado.emitido_em})`
            : ' · sem certificado'}
        </p>
      )}

      <h4>Emissões do curso</h4>
      {loading ? (
        <p className="muted">Carregando…</p>
      ) : !lista ? (
        <p className="muted">Sem dados.</p>
      ) : (
        <>
          <p className="muted">
            Elegíveis:{' '}
            {lista.elegibilidades.length === 0
              ? 'nenhum'
              : lista.elegibilidades
                  .map((e) => `${e.user_nome} (id ${e.user_id})`)
                  .join(', ')}
          </p>
          {lista.certificados.length === 0 ? (
            <p className="muted">Nenhum certificado emitido.</p>
          ) : (
            <table>
              <thead>
                <tr>
                  <th>ID</th>
                  <th>Aluno</th>
                  <th>Status</th>
                  <th>Emitido em</th>
                  <th></th>
                </tr>
              </thead>
              <tbody>
                {lista.certificados.map((c) => (
                  <tr key={c.id}>
                    <td>{c.id}</td>
                    <td>
                      {c.aluno_nome}
                      {c.user_id != null ? ` (#${c.user_id})` : ''}
                    </td>
                    <td>{c.status}</td>
                    <td>{c.emitido_em}</td>
                    <td className="actions">
                      {c.status === 'valido' && (
                        <button
                          type="button"
                          className="danger"
                          disabled={busy}
                          onClick={() => void handleInvalidar(c)}
                        >
                          Invalidar
                        </button>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </>
      )}
    </div>
  )
}
