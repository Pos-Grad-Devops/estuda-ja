import type { Aula } from '../../api/client'

const labels: Record<Aula['status'], string> = {
  ao_vivo: 'Ao vivo',
  agendada: 'Agendada',
  encerrada: 'Encerrada',
}

const styles: Record<Aula['status'], string> = {
  ao_vivo: 'bg-live/15 text-live border-live/40',
  agendada: 'bg-accent/15 text-accent-hover border-accent/35',
  encerrada: 'bg-white/6 text-muted border-white/10',
}

export function StatusPill({ status }: { status: Aula['status'] }) {
  return (
    <span
      className={`inline-flex items-center gap-1.5 rounded-full border px-2.5 py-1 text-xs font-semibold ${styles[status]}`}
    >
      {status === 'ao_vivo' && <span className="size-1.5 animate-pulse rounded-full bg-live" />}
      {labels[status]}
    </span>
  )
}
