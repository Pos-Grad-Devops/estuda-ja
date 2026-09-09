import { Link } from 'react-router-dom'
import type { Aula } from '../../api/client'
import { formatDateBR } from '../../utils/date'
import { StatusPill } from '../ui/StatusPill'

type LessonRowProps = {
  aula: Aula
  showCourse?: boolean
}

export function LessonRow({ aula, showCourse = false }: LessonRowProps) {
  return (
    <Link
      to={`/aulas/${aula.id}`}
      className="flex items-start justify-between gap-4 rounded-xl border border-transparent px-3 py-3 transition hover:border-border hover:bg-white/4"
    >
      <div className="min-w-0">
        <p className="font-semibold">{aula.titulo}</p>
        <p className="mt-1 text-sm text-muted">
          {showCourse && aula.curso ? `${aula.curso.titulo} · ` : ''}
          {formatDateBR(aula.agendada_em)}
        </p>
      </div>
      <StatusPill status={aula.status} />
    </Link>
  )
}