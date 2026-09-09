import { useEffect, useState, type FormEvent } from 'react'
import type { Curso } from '../../api/client'
import { Button } from '../ui/Button'
import { Modal } from '../ui/Modal'

type CursoModalProps = {
  open: boolean
  curso?: Curso | null
  onClose: () => void
  onSubmit: (data: Pick<Curso, 'titulo' | 'descricao'>) => Promise<void>
}

export function CursoModal({ open, curso, onClose, onSubmit }: CursoModalProps) {
  const [titulo, setTitulo] = useState('')
  const [descricao, setDescricao] = useState('')
  const [error, setError] = useState('')
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    if (!open) return
    setTitulo(curso?.titulo ?? '')
    setDescricao(curso?.descricao ?? '')
    setError('')
  }, [curso, open])

  async function handleSubmit(event: FormEvent) {
    event.preventDefault()
    setSaving(true)
    setError('')
    try {
      await onSubmit({ titulo: titulo.trim(), descricao })
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
      title={curso ? 'Editar curso' : 'Novo curso'}
      onClose={onClose}
      footer={
        <>
          <Button variant="secondary" onClick={onClose}>
            Cancelar
          </Button>
          <Button type="submit" form="curso-form" disabled={saving}>
            {saving ? 'Salvando...' : 'Salvar'}
          </Button>
        </>
      }
    >
      <form id="curso-form" onSubmit={handleSubmit}>
        {error && <p className="mb-3 rounded-xl bg-live/10 px-3 py-2 text-sm text-live">{error}</p>}
        <label className="field">
          Título
          <input value={titulo} onChange={(e) => setTitulo(e.target.value)} required />
        </label>
        <label className="field">
          Descrição
          <textarea value={descricao} onChange={(e) => setDescricao(e.target.value)} rows={4} />
        </label>
      </form>
    </Modal>
  )
}
