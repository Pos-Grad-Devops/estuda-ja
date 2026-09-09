import { Link } from 'react-router-dom'
import type { Curso } from '../../api/client'
import { canManageCursos, type Role } from '../../auth/auth'
import { courseHue, courseInitials } from '../../utils/course'
import SpotlightCard from '../bits/SpotlightCard'
import { Button } from '../ui/Button'

type CourseCardProps = {
  curso: Curso
  role: Role
  onEdit: (curso: Curso) => void
  onDelete: (curso: Curso) => void
}

export function CourseCard({ curso, role, onEdit, onDelete }: CourseCardProps) {
  const hue = courseHue(curso.id)

  return (
    <SpotlightCard className="h-full p-0">
      <div className="relative z-10 flex h-full flex-col">
        <div
          className="flex h-28 items-end px-5 pb-4"
          style={{ background: `linear-gradient(135deg, ${hue}, #140d22)` }}
        >
          <span className="font-display text-3xl font-bold text-white/90">{courseInitials(curso.titulo)}</span>
        </div>
        <div className="flex flex-1 flex-col p-5">
          <h3 className="font-display text-lg font-semibold leading-snug">
            <Link to={`/cursos/${curso.id}`} className="hover:text-accent-hover">
              {curso.titulo}
            </Link>
          </h3>
          <p className="mt-2 line-clamp-3 flex-1 text-sm text-muted">{curso.descricao || 'Sem descrição ainda.'}</p>
          <div className="mt-4 flex flex-wrap items-center gap-2">
            <Link
              to={`/cursos/${curso.id}`}
              className="text-sm font-semibold text-accent-hover hover:underline"
            >
              Abrir curso
            </Link>
            {canManageCursos(role) && (
              <>
                <Button variant="ghost" className="px-2 py-1 text-xs" onClick={() => onEdit(curso)}>
                  Editar
                </Button>
                <Button variant="ghost" className="px-2 py-1 text-xs text-live" onClick={() => onDelete(curso)}>
                  Excluir
                </Button>
              </>
            )}
          </div>
        </div>
      </div>
    </SpotlightCard>
  )
}
