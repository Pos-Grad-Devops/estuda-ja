import { useCallback, useEffect, useState, type FormEvent } from 'react'
import { api } from '../api/client'
import type { Aluno } from '../api/client'
import { Button } from '../components/ui/Button'
import { EmptyState } from '../components/ui/EmptyState'
import { Modal } from '../components/ui/Modal'
import { PageHeader } from '../components/ui/PageHeader'

export function AlunosPage() {
  const [items, setItems] = useState<Aluno[]>([])
  const [nome, setNome] = useState('')
  const [email, setEmail] = useState('')
  const [editing, setEditing] = useState<Aluno | null>(null)
  const [open, setOpen] = useState(false)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)

  const load = useCallback(async () => {
    setLoading(true)
    try {
      setItems(await api.alunos.list())
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

  function openCreate() {
    setEditing(null)
    setNome('')
    setEmail('')
    setOpen(true)
  }

  function openEdit(aluno: Aluno) {
    setEditing(aluno)
    setNome(aluno.nome)
    setEmail(aluno.email)
    setOpen(true)
  }

  async function handleSubmit(event: FormEvent) {
    event.preventDefault()
    setSaving(true)
    try {
      const data = { nome: nome.trim(), email: email.trim() }
      if (editing) {
        await api.alunos.update(editing.id, data)
      } else {
        await api.alunos.create(data)
      }
      setOpen(false)
      await load()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Erro ao salvar')
    } finally {
      setSaving(false)
    }
  }

  async function handleDelete(aluno: Aluno) {
    if (!confirm(`Excluir o cadastro de “${aluno.nome}”?`)) return
    try {
      await api.alunos.remove(aluno.id)
      await load()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Erro ao excluir')
    }
  }

  return (
    <section>
      <PageHeader
        eyebrow="Gestão"
        title="Cadastro de alunos"
        description="Registro administrativo de alunos (nome e e-mail). Isso não é a conta de login — usuários de acesso ficam em outra tela."
        actions={<Button onClick={openCreate}>Novo aluno</Button>}
      />

      {error && <p className="mb-4 rounded-xl bg-live/12 px-3 py-2 text-sm text-live">{error}</p>}

      {loading ? (
        <p className="text-muted">Carregando cadastro...</p>
      ) : items.length === 0 ? (
        <EmptyState
          title="Nenhum aluno cadastrado"
          description="Use esta lista para o registro interno. Quem entra na plataforma é criado em Usuários de acesso."
          action={<Button onClick={openCreate}>Cadastrar aluno</Button>}
        />
      ) : (
        <div className="overflow-x-auto rounded-2xl border border-border bg-surface">
          <table className="w-full min-w-[520px] text-left text-sm">
            <thead className="border-b border-border text-muted">
              <tr>
                <th className="px-4 py-3 font-semibold">Nome</th>
                <th className="px-4 py-3 font-semibold">E-mail</th>
                <th className="px-4 py-3 font-semibold">Ações</th>
              </tr>
            </thead>
            <tbody>
              {items.map((aluno) => (
                <tr key={aluno.id} className="border-b border-border/70 last:border-0">
                  <td className="px-4 py-3 font-semibold">{aluno.nome}</td>
                  <td className="px-4 py-3 text-muted">{aluno.email}</td>
                  <td className="px-4 py-3">
                    <div className="flex flex-wrap gap-2">
                      <Button variant="secondary" className="px-3 py-1.5 text-xs" onClick={() => openEdit(aluno)}>
                        Editar
                      </Button>
                      <Button variant="danger" className="px-3 py-1.5 text-xs" onClick={() => void handleDelete(aluno)}>
                        Excluir
                      </Button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      <Modal
        open={open}
        title={editing ? 'Editar aluno' : 'Novo aluno'}
        onClose={() => setOpen(false)}
        footer={
          <>
            <Button variant="secondary" onClick={() => setOpen(false)}>
              Cancelar
            </Button>
            <Button type="submit" form="aluno-form" disabled={saving}>
              {saving ? 'Salvando...' : 'Salvar'}
            </Button>
          </>
        }
      >
        <form id="aluno-form" onSubmit={handleSubmit}>
          <label className="field">
            Nome
            <input value={nome} onChange={(e) => setNome(e.target.value)} required />
          </label>
          <label className="field">
            E-mail
            <input type="email" value={email} onChange={(e) => setEmail(e.target.value)} required />
          </label>
        </form>
      </Modal>
    </section>
  )
}
