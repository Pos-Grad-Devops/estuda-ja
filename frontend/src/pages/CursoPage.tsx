import { useCallback, useEffect, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { api } from '../api/client'
import type { Aula, Curso } from '../api/client'
import { canManageAulas, canManageCursos } from '../auth/auth'
import { useAuth } from '../auth/AuthContext'
import FadeContent from '../components/bits/FadeContent'
import { CursoCertificadoPanel } from '../components/CursoCertificadoPanel'
import { LessonRow } from '../components/course/LessonRow'
import { AulaModal } from '../components/forms/AulaModal'
import { CursoModal } from '../components/forms/CursoModal'
import { Button } from '../components/ui/Button'
import { EmptyState } from '../components/ui/EmptyState'
import { PageHeader } from '../components/ui/PageHeader'
import { courseHue, courseInitials } from '../utils/course'

export function CursoPage() {
  const { id } = useParams()
  const navigate = useNavigate()
  const { user } = useAuth()
  const [curso, setCurso] = useState<Curso | null>(null)
  const [cursos, setCursos] = useState<Curso[]>([])
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(true)
  const [cursoModal, setCursoModal] = useState(false)
  const [aulaModal, setAulaModal] = useState(false)

  const cursoId = Number(id)

  const load = useCallback(async () => {
    if (!Number.isFinite(cursoId) || cursoId <= 0) {
      setError('Curso inválido')
      setLoading(false)
      return
    }
    setLoading(true)
    try {
      const [detail, list] = await Promise.all([api.cursos.get(cursoId), api.cursos.list()])
      // Backend pode não embutir aulas[] — fallback só no front (sem novo endpoint).
      const withAulas =
        detail.aulas != null
          ? detail
          : { ...detail, aulas: await api.aulas.list(cursoId) }
      setCurso(withAulas)
      setCursos(list)
      setError('')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Erro ao carregar')
      setCurso(null)
    } finally {
      setLoading(false)
    }
  }, [cursoId])

  useEffect(() => {
    void load()
  }, [load])

  async function handleSaveCurso(data: Pick<Curso, 'titulo' | 'descricao'>) {
    await api.cursos.update(cursoId, data)
    await load()
  }

  async function handleDeleteCurso() {
    if (!curso || !confirm(`Excluir o curso “${curso.titulo}”?`)) return
    await api.cursos.remove(curso.id)
    navigate('/', { replace: true })
  }

  async function handleSaveAula(data: Pick<Aula, 'curso_id' | 'titulo' | 'descricao' | 'agendada_em' | 'status'>) {
    await api.aulas.create(data)
    await load()
  }

  if (!user) return null

  const aulas = curso?.aulas ?? []
  const hue = curso ? courseHue(curso.id) : '#7c3aed'

  return (
    <section>
      <Link to="/" className="mb-4 inline-block text-sm font-semibold text-accent-hover hover:underline">
        ← Voltar ao catálogo
      </Link>

      {error && <p className="mb-4 rounded-xl bg-live/12 px-3 py-2 text-sm text-live">{error}</p>}
      {loading && <p className="text-muted">Carregando curso...</p>}

      {curso && (
        <>
          <div
            className="mb-6 flex h-36 items-end rounded-3xl px-6 pb-5"
            style={{ background: `linear-gradient(135deg, ${hue}, #0b0714)` }}
          >
            <span className="font-display text-5xl font-bold text-white/90">{courseInitials(curso.titulo)}</span>
          </div>
          <PageHeader
            eyebrow="Curso"
            title={curso.titulo}
            description={curso.descricao || 'Este curso ainda não tem uma descrição.'}
            actions={
              <>
                {canManageAulas(user.role) && <Button onClick={() => setAulaModal(true)}>Nova aula</Button>}
                {canManageCursos(user.role) && (
                  <>
                    <Button variant="secondary" onClick={() => setCursoModal(true)}>
                      Editar
                    </Button>
                    <Button variant="danger" onClick={() => void handleDeleteCurso()}>
                      Excluir
                    </Button>
                  </>
                )}
              </>
            }
          />

          <FadeContent>
            <h2 className="mb-3 font-display text-xl font-semibold">Trilha de aulas</h2>
            {aulas.length === 0 ? (
              <EmptyState
                title="Nenhuma aula neste curso"
                description="Quando alguém publicar uma aula, ela entra nesta trilha com data e status."
                action={
                  canManageAulas(user.role) ? <Button onClick={() => setAulaModal(true)}>Criar aula</Button> : undefined
                }
              />
            ) : (
              <div className="divide-y divide-border rounded-2xl border border-border bg-surface px-2 py-1">
                {aulas.map((aula) => (
                  <LessonRow key={aula.id} aula={{ ...aula, curso }} />
                ))}
              </div>
            )}
          </FadeContent>

          <CursoCertificadoPanel cursoId={curso.id} />
        </>
      )}

      <CursoModal open={cursoModal} curso={curso} onClose={() => setCursoModal(false)} onSubmit={handleSaveCurso} />
      <AulaModal
        open={aulaModal}
        cursos={cursos}
        defaultCursoId={cursoId}
        onClose={() => setAulaModal(false)}
        onSubmit={handleSaveAula}
      />
    </section>
  )
}
