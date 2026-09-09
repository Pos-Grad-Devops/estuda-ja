import { useCallback, useEffect, useMemo, useState } from 'react'
import { api } from '../api/client'
import type { Aula, Curso } from '../api/client'
import { canManageAulas } from '../auth/auth'
import { useAuth } from '../auth/AuthContext'
import FadeContent from '../components/bits/FadeContent'
import { LessonRow } from '../components/course/LessonRow'
import { AulaModal } from '../components/forms/AulaModal'
import { Button } from '../components/ui/Button'
import { EmptyState } from '../components/ui/EmptyState'
import { PageHeader } from '../components/ui/PageHeader'

type StatusFilter = 'todas' | Aula['status']

export function AgendaPage() {
  const { user } = useAuth()
  const [aulas, setAulas] = useState<Aula[]>([])
  const [cursos, setCursos] = useState<Curso[]>([])
  const [status, setStatus] = useState<StatusFilter>('todas')
  const [cursoId, setCursoId] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(true)
  const [modalOpen, setModalOpen] = useState(false)

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const [aulasList, cursosList] = await Promise.all([api.aulas.list(), api.cursos.list()])
      setAulas(aulasList)
      setCursos(cursosList)
      setError('')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Erro ao carregar')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void load()
  }, [load])

  const filtered = useMemo(() => {
    return aulas.filter((aula) => {
      const statusOk = status === 'todas' || aula.status === status
      const cursoOk = !cursoId || String(aula.curso_id) === cursoId
      return statusOk && cursoOk
    })
  }, [aulas, cursoId, status])

  async function handleSave(data: Pick<Aula, 'curso_id' | 'titulo' | 'descricao' | 'agendada_em' | 'status'>) {
    await api.aulas.create(data)
    await load()
  }

  if (!user) return null

  return (
    <section>
      <PageHeader
        eyebrow="Todas as aulas"
        title="Agenda"
        description="Lista de aulas da escola. O filtro de status usa o metadado CRUD da aula — a transmissão ao vivo (canal) aparece na ficha da aula via API live."
        actions={
          canManageAulas(user.role) ? <Button onClick={() => setModalOpen(true)}>Nova aula</Button> : undefined
        }
      />

      {error && <p className="mb-4 rounded-xl bg-live/12 px-3 py-2 text-sm text-live">{error}</p>}

      <div className="mb-6 flex flex-wrap gap-3">
        <label className="field mb-0 min-w-40 flex-1">
          Status (aula)
          <select value={status} onChange={(e) => setStatus(e.target.value as StatusFilter)}>
            <option value="todas">Todas</option>
            <option value="ao_vivo">Ao vivo (metadado)</option>
            <option value="agendada">Agendada</option>
            <option value="encerrada">Encerrada</option>
          </select>
        </label>
        <label className="field mb-0 min-w-40 flex-1">
          Curso
          <select value={cursoId} onChange={(e) => setCursoId(e.target.value)}>
            <option value="">Todos</option>
            {cursos.map((curso) => (
              <option key={curso.id} value={curso.id}>
                {curso.titulo}
              </option>
            ))}
          </select>
        </label>
      </div>

      {loading ? (
        <p className="text-muted">Carregando agenda...</p>
      ) : filtered.length === 0 ? (
        <EmptyState
          title="Nenhuma aula neste recorte"
          description="Ajuste os filtros ou publique uma nova aula para preencher a agenda."
          action={
            canManageAulas(user.role) ? <Button onClick={() => setModalOpen(true)}>Nova aula</Button> : undefined
          }
        />
      ) : (
        <FadeContent className="divide-y divide-border rounded-2xl border border-border bg-surface px-2 py-1">
          {filtered.map((aula) => (
            <LessonRow key={aula.id} aula={aula} showCourse />
          ))}
        </FadeContent>
      )}

      <AulaModal
        open={modalOpen}
        cursos={cursos}
        defaultCursoId={cursoId ? Number(cursoId) : undefined}
        onClose={() => setModalOpen(false)}
        onSubmit={handleSave}
      />
    </section>
  )
}