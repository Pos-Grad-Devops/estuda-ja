import { useEffect, useState, type FormEvent } from 'react'
import type { Aula, Curso } from '../../api/client'
import { isValidDateTimeBR } from '../../utils/date'
import { Button } from '../ui/Button'
import { Modal } from '../ui/Modal'

type AulaPayload = Pick<Aula, 'curso_id' | 'titulo' | 'descricao' | 'agendada_em' | 'status'>

type AulaModalProps = {
  open: boolean
  aula?: Aula | null
  cursos: Curso[]
  defaultCursoId?: number
  onClose: () => void
  onSubmit: (data: AulaPayload) => Promise<void>
}

export function AulaModal({ open, aula, cursos, defaultCursoId, onClose, onSubmit }: AulaModalProps) {
  const [cursoId, setCursoId] = useState('')
  const [titulo, setTitulo] = useState('')
  const [descricao, setDescricao] = useState('')
  const [agendadaEm, setAgendadaEm] = useState('')
  const [status, setStatus] = useState<Aula['status']>('agendada')
  const [error, setError] = useState('')
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    if (!open) return
    setCursoId(aula ? String(aula.curso_id) : defaultCursoId ? String(defaultCursoId) : '')
    setTitulo(aula?.titulo ?? '')
    setDescricao(aula?.descricao ?? '')
    setAgendadaEm(aula?.agendada_em ?? '')
    setStatus(aula?.status ?? 'agendada')
    setError('')
  }, [aula, defaultCursoId, open])

  async function handleSubmit(event: FormEvent) {
    event.preventDefault()
    if (!isValidDateTimeBR(agendadaEm)) {
      setError('Data inválida. Use dd/mm/yyyy ou dd/mm/yyyy HH:mm')
      return
    }
    setSaving(true)
    setError('')
    try {
      await onSubmit({
        curso_id: Number(cursoId),
        titulo: titulo.trim(),
        descricao,
        agendada_em: agendadaEm.trim(),
        status,
      })
      onClose()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Erro ao salvar')
    } finally {
      setSaving(false)
    }
  }

  return (
    <Modal
      open={open}
      title={aula ? 'Editar aula' : 'Nova aula'}
      onClose={onClose}
      footer={
        <>
          <Button variant="secondary" onClick={onClose}>
            Cancelar
          </Button>
          <Button type="submit" form="aula-form" disabled={saving}>
            {saving ? 'Salvando...' : 'Salvar'}
          </Button>
        </>
      }
    >
      <form id="aula-form" onSubmit={handleSubmit}>
        {error && <p className="mb-3 rounded-xl bg-live/10 px-3 py-2 text-sm text-live">{error}</p>}
        <label className="field">
          Curso
          <select value={cursoId} onChange={(e) => setCursoId(e.target.value)} required>
            <option value="">Selecione</option>
            {cursos.map((curso) => (
              <option key={curso.id} value={curso.id}>
                {curso.titulo}
              </option>
            ))}
          </select>
        </label>
        <label className="field">
          Título
          <input value={titulo} onChange={(e) => setTitulo(e.target.value)} required />
        </label>
        <label className="field">
          Descrição
          <textarea value={descricao} onChange={(e) => setDescricao(e.target.value)} rows={3} />
        </label>
        <label className="field">
          Agendada em
          <input
            value={agendadaEm}
            onChange={(e) => setAgendadaEm(e.target.value)}
            placeholder="dd/mm/yyyy HH:mm"
            required
          />
        </label>
        <label className="field">
          Status
          <select value={status} onChange={(e) => setStatus(e.target.value as Aula['status'])}>
            <option value="agendada">Agendada</option>
            <option value="ao_vivo">Ao vivo</option>
            <option value="encerrada">Encerrada</option>
          </select>
        </label>
      </form>
    </Modal>
  )
}