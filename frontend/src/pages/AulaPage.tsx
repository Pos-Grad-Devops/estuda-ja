import { useCallback, useEffect, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { api } from '../api/client'
import type { Aula, Curso } from '../api/client'
import { canManageAulas } from '../auth/auth'
import { useAuth } from '../auth/AuthContext'
import FadeContent from '../components/bits/FadeContent'
import { AulaModal } from '../components/forms/AulaModal'
import { Button } from '../components/ui/Button'
import { PageHeader } from '../components/ui/PageHeader'
import { StatusPill } from '../components/ui/StatusPill'
import { formatDateBR } from '../utils/date'

export function AulaPage() {
  const { id } = useParams()
  const navigate = useNavigate()
  const { user } = useAuth()
  const [aula, setAula] = useState<Aula | null>(null)
  const [cursos, setCursos] = useState<Curso[]>([])
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(true)
  const [modalOpen, setModalOpen] = useState(false)

  const aulaId = Number(id)

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

  useEffect(() => {
    void load()
  }, [load])

  async function handleSave(data: Pick<Aula, 'curso_id' | 'titulo' | 'descricao' | 'agendada_em' | 'status'>) {
    await api.aulas.update(aulaId, data)
    await load()
  }

  async function handleDelete() {
    if (!aula || !confirm(`Excluir a aula “${aula.titulo}”?`)) return
    await api.aulas.remove(aula.id)
    navigate(aula.curso_id ? `/cursos/${aula.curso_id}` : '/aulas', { replace: true })
  }

  if (!user) return null

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
              canManageAulas(user.role) ? (
                <>
                  <Button onClick={() => setModalOpen(true)}>Editar</Button>
                  <Button variant="danger" onClick={() => void handleDelete()}>
                    Excluir
                  </Button>
                </>
              ) : undefined
            }
          />
          <div className="grid gap-4 sm:grid-cols-2">
            <div className="rounded-2xl border border-border bg-surface p-5">
              <p className="text-sm text-muted">Status</p>
              <div className="mt-3">
                <StatusPill status={aula.status} />
              </div>
            </div>
            <div className="rounded-2xl border border-border bg-surface p-5">
              <p className="text-sm text-muted">Quando</p>
              <p className="mt-3 font-display text-xl font-semibold">{formatDateBR(aula.agendada_em)}</p>
            </div>
          </div>
          {aula.status === 'ao_vivo' && (
            <p className="mt-6 rounded-2xl border border-live/30 bg-live/10 px-4 py-3 text-sm">
              Esta aula está marcada como ao vivo. O player de transmissão ainda não faz parte deste MVP.
            </p>
          )}
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
