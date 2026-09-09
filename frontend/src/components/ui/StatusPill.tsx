import type { Aula, LiveStatus } from '../../api/client'

/** Status CRUD da aula ou estado de transmissão (`api.live`). */
export type PillStatus = Aula['status'] | LiveStatus

const labels: Record<PillStatus, string> = {
  ao_vivo: 'Ao vivo',
  agendada: 'Agendada',
  encerrada: 'Encerrada',
  inativa: 'Inativa',
}

const styles: Record<PillStatus, string> = {
  ao_vivo: 'bg-live/15 text-live border-live/40',
  agendada: 'bg-accent/15 text-accent-hover border-accent/35',
  encerrada: 'bg-white/6 text-muted border-white/10',
  inativa: 'bg-white/4 text-muted border-white/8',
}

export function StatusPill({ status }: { status: PillStatus }) {
  return (
    <span
      className={`inline-flex items-center gap-1.5 rounded-full border px-2.5 py-1 text-xs font-semibold ${styles[status]}`}
    >
      {status === 'ao_vivo' && <span className="size-1.5 animate-pulse rounded-full bg-live" />}
      {labels[status]}
    </span>
  )
}
