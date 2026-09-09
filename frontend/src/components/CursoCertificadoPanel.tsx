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
import { Button } from './ui/Button'
import { EmptyState } from './ui/EmptyState'
import { PageHeader } from './ui/PageHeader'

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
    <section id={`curso-${cursoId}-certificado`} className="mt-8">
      <PageHeader
        eyebrow="Certificado"
        title="Certificado do curso"
        description={
          isAdmin
            ? 'Marque elegibilidade e gerencie emissões. O PDF só é gerado na primeira solicitação do aluno.'
            : 'Consulte sua elegibilidade e baixe o PDF quando disponível.'
        }
      />
      {isAluno && <AlunoCertificadoBlock cursoId={cursoId} />}
      {isAdmin && <AdminCertificadoBlock cursoId={cursoId} />}
    </section>
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
    return (
      <EmptyState
        title="Carregando certificado"
        description="Consultando elegibilidade e emissão deste curso…"
      />
    )
  }

  const cert = status?.certificado
  const elegivel = status?.elegivel === true
  const podeBaixar = elegivel || (cert != null && cert.status === 'valido')

  return (
    <div className="rounded-2xl border border-border bg-surface p-5">
      {error && (
        <p className="mb-3 rounded-xl bg-live/12 px-3 py-2 text-sm text-live">{error}</p>
      )}
      {info && <p className="mb-3 text-sm text-muted">{info}</p>}

      {!elegivel && !cert && (
        <EmptyState
          title="Sem elegibilidade"
          description="Você não está elegível para o certificado deste curso."
        />
      )}

      {!elegivel && cert?.status === 'invalidado' && (
        <p className="mb-4 text-sm text-muted">
          Certificado invalidado (emitido em {cert.emitido_em}). Aguarde reabilitação pelo
          administrador.
        </p>
      )}

      {elegivel && !cert && (
        <p className="mb-4 text-sm">
          Você está elegível. Solicite o certificado para emitir e baixar o PDF.
        </p>
      )}

      {cert?.status === 'valido' && (
        <p className="mb-4 text-sm">
          Certificado válido · emitido em {cert.emitido_em}
          {cert.aluno_nome ? ` · ${cert.aluno_nome}` : ''}
        </p>
      )}

      <div className="mt-4 flex flex-wrap gap-3">
        <Button onClick={() => void handleDownload()} disabled={downloading || !podeBaixar}>
          {downloading
            ? 'Baixando…'
            : cert?.status === 'valido'
              ? 'Baixar PDF'
              : 'Solicitar / baixar certificado'}
        </Button>
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
    <div className="space-y-5">
      {error && (
        <p className="rounded-xl bg-live/12 px-3 py-2 text-sm text-live">{error}</p>
      )}
      {info && <p className="text-sm text-muted">{info}</p>}

      <form
        className="rounded-2xl border border-border bg-surface p-5"
        onSubmit={handleMarcar}
      >
        <label className="field">
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
        <Button type="submit" disabled={busy}>
          Marcar elegibilidade
        </Button>
      </form>

      <form
        className="rounded-2xl border border-border bg-surface p-5"
        onSubmit={handleConsultar}
      >
        <label className="field">
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
        <Button type="submit" variant="secondary" disabled={busy}>
          Consultar
        </Button>
      </form>

      {consulta && (
        <p className="rounded-xl border border-border bg-surface/60 px-4 py-3 text-sm text-muted">
          user_id {consulta.user_id}: elegível={consulta.elegivel ? 'sim' : 'não'}
          {consulta.certificado
            ? ` · cert #${consulta.certificado.id} (${consulta.certificado.status}, ${consulta.certificado.emitido_em})`
            : ' · sem certificado'}
        </p>
      )}

      <div className="rounded-2xl border border-border bg-surface p-5">
        <h3 className="mb-3 font-display text-lg font-semibold">Emissões do curso</h3>
        {loading ? (
          <p className="text-sm text-muted">Carregando…</p>
        ) : !lista ? (
          <EmptyState title="Sem dados" description="Não foi possível carregar elegibilidades e certificados." />
        ) : (
          <>
            <p className="mb-4 text-sm text-muted">
              Elegíveis:{' '}
              {lista.elegibilidades.length === 0
                ? 'nenhum'
                : lista.elegibilidades
                    .map((e) => `${e.user_nome} (id ${e.user_id})`)
                    .join(', ')}
            </p>
            {lista.certificados.length === 0 ? (
              <EmptyState
                title="Nenhum certificado emitido"
                description="Quando um aluno elegível solicitar o PDF, a emissão aparece aqui."
              />
            ) : (
              <div className="divide-y divide-border overflow-hidden rounded-xl border border-border">
                {lista.certificados.map((c) => (
                  <div
                    key={c.id}
                    className="flex flex-wrap items-center justify-between gap-3 px-4 py-3"
                  >
                    <div className="min-w-0">
                      <p className="font-semibold">
                        #{c.id} · {c.aluno_nome}
                        {c.user_id != null ? ` (#${c.user_id})` : ''}
                      </p>
                      <p className="mt-1 text-sm text-muted">
                        {c.status} · emitido em {c.emitido_em}
                      </p>
                    </div>
                    {c.status === 'valido' && (
                      <Button
                        variant="danger"
                        disabled={busy}
                        onClick={() => void handleInvalidar(c)}
                      >
                        Invalidar
                      </Button>
                    )}
                  </div>
                ))}
              </div>
            )}
          </>
        )}
      </div>
    </div>
  )
}
