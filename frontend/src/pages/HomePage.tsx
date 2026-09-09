import { useCallback, useEffect, useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { api } from '../api/client'
import type { Aula, Curso } from '../api/client'
import { canManageCursos } from '../auth/auth'
import { useAuth } from '../auth/AuthContext'
import FadeContent from '../components/bits/FadeContent'
import CountUp from '../components/bits/CountUp'
import { CourseCard } from '../components/course/CourseCard'
import { CursoModal } from '../components/forms/CursoModal'
import { Button } from '../components/ui/Button'
import { EmptyState } from '../components/ui/EmptyState'
import { PageHeader } from '../components/ui/PageHeader'
import { StatusPill } from '../components/ui/StatusPill'
import { formatDateBR } from '../utils/date'

export function HomePage() {
  const { user } = useAuth()
  const [cursos, setCursos] = useState<Curso[]>([])
  const [aulas, setAulas] = useState<Aula[]>([])
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(true)
  const [modalOpen, setModalOpen] = useState(false)
  const [editing, setEditing] = useState<Curso | null>(null)

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const [cursosList, aulasList] = await Promise.all([api.cursos.list(), api.aulas.list()])
      setCursos(cursosList)
      setAulas(aulasList)
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

  const liveAulas = useMemo(() => aulas.filter((aula) => aula.status === 'ao_vivo'), [aulas])
  const nextLive = liveAulas[0]

  async function handleSave(data: Pick<Curso, 'titulo' | 'descricao'>) {
    if (editing) {
      await api.cursos.update(editing.id, data)
    } else {
      await api.cursos.create(data)
    }
    await load()
  }

  async function handleDelete(curso: Curso) {
    if (!confirm(`Excluir o curso “${curso.titulo}”?`)) return
    try {
      await api.cursos.remove(curso.id)
      await load()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Erro ao excluir')
    }
  }

  if (!user) return null

  return (
    <section>
      <PageHeader
        eyebrow={`Olá, ${user.nome.split(' ')[0]}`}
        title="Sua escola"
        description="Catálogo de cursos e o que está ao vivo agora. Abra um curso para ver a trilha de aulas."
        actions={
          canManageCursos(user.role) ? (
            <Button
              onClick={() => {
                setEditing(null)
                setModalOpen(true)
              }}
            >
              Novo curso
            </Button>
          ) : undefined
        }
      />

      {error && <p className="mb-4 rounded-xl bg-live/12 px-3 py-2 text-sm text-live">{error}</p>}

      <div className="mb-8 grid gap-4 sm:grid-cols-2">
        <FadeContent className="rounded-2xl border border-border bg-surface p-5">
          <p className="text-sm text-muted">Cursos</p>
          <p className="mt-2 font-display text-4xl font-semibold">
            {loading ? '—' : <CountUp to={cursos.length} />}
          </p>
        </FadeContent>
        <FadeContent delay={80} className="rounded-2xl border border-border bg-surface p-5">
          <p className="text-sm text-muted">Aulas ao vivo agora</p>
          <p className="mt-2 font-display text-4xl font-semibold text-live">
            {loading ? '—' : <CountUp to={liveAulas.length} />}
          </p>
        </FadeContent>
      </div>

      {nextLive && (
        <FadeContent className="mb-8 overflow-hidden rounded-2xl border border-live/30 bg-live/8 p-5">
          <div className="flex flex-wrap items-center justify-between gap-3">
            <div>
              <StatusPill status="ao_vivo" />
              <h2 className="mt-3 font-display text-2xl font-semibold">{nextLive.titulo}</h2>
              <p className="mt-1 text-sm text-muted">
                {nextLive.curso?.titulo ?? `Curso #${nextLive.curso_id}`} · {formatDateBR(nextLive.agendada_em)}
              </p>
            </div>
            <Link to={`/aulas/${nextLive.id}`}>
              <Button>Entrar na aula</Button>
            </Link>
          </div>
        </FadeContent>
      )}

      {loading ? (
        <p className="text-muted">Carregando catálogo...</p>
      ) : cursos.length === 0 ? (
        <EmptyState
          title="Nenhum curso ainda"
          description="Quando um administrador publicar um curso, ele aparece aqui como no catálogo de uma escola."
          action={
            canManageCursos(user.role) ? (
              <Button
                onClick={() => {
                  setEditing(null)
                  setModalOpen(true)
                }}
              >
                Criar o primeiro curso
              </Button>
            ) : undefined
          }
        />
      ) : (
        <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
          {cursos.map((curso) => (
            <CourseCard
              key={curso.id}
              curso={curso}
              role={user.role}
              onEdit={(item) => {
                setEditing(item)
                setModalOpen(true)
              }}
              onDelete={handleDelete}
            />
          ))}
        </div>
      )}

      <CursoModal
        open={modalOpen}
        curso={editing}
        onClose={() => {
          setModalOpen(false)
          setEditing(null)
        }}
        onSubmit={handleSave}
      />
    </section>
  )
}
