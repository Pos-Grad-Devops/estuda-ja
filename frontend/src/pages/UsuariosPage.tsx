import { useCallback, useEffect, useState, type FormEvent } from 'react'
import { api } from '../api/client'
import { roleLabel, type Role, type User } from '../auth/auth'
import { Button } from '../components/ui/Button'
import { EmptyState } from '../components/ui/EmptyState'
import { Modal } from '../components/ui/Modal'
import { PageHeader } from '../components/ui/PageHeader'

export function UsuariosPage() {
  const [items, setItems] = useState<User[]>([])
  const [nome, setNome] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [role, setRole] = useState<Role>('aluno')
  const [editing, setEditing] = useState<User | null>(null)
  const [open, setOpen] = useState(false)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)

  const load = useCallback(async () => {
    setLoading(true)
    try {
      setItems(await api.users.list())
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
    setPassword('')
    setRole('aluno')
    setOpen(true)
  }

  function openEdit(user: User) {
    setEditing(user)
    setNome(user.nome)
    setEmail(user.email)
    setPassword('')
    setRole(user.role)
    setOpen(true)
  }

  async function handleSubmit(event: FormEvent) {
    event.preventDefault()
    setSaving(true)
    try {
      if (editing) {
        await api.users.update(editing.id, {
          nome: nome.trim(),
          email: email.trim(),
          role,
          password: password || undefined,
        })
      } else {
        await api.users.create({ nome: nome.trim(), email: email.trim(), password, role })
      }
      setOpen(false)
      await load()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Erro ao salvar')
    } finally {
      setSaving(false)
    }
  }

  async function handleDelete(user: User) {
    if (!confirm(`Excluir o usuário “${user.nome}”?`)) return
    try {
      await api.users.remove(user.id)
      await load()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Erro ao excluir')
    }
  }

  return (
    <section>
      <PageHeader
        eyebrow="Gestão"
        title="Usuários de acesso"
        description="Contas que entram na plataforma (admin, professor ou aluno). Diferente do cadastro administrativo de alunos."
        actions={<Button onClick={openCreate}>Novo usuário</Button>}
      />

      {error && <p className="mb-4 rounded-xl bg-live/12 px-3 py-2 text-sm text-live">{error}</p>}

      {loading ? (
        <p className="text-muted">Carregando usuários...</p>
      ) : items.length === 0 ? (
        <EmptyState
          title="Nenhum usuário"
          description="Crie contas para professor e aluno testarem o catálogo com permissões diferentes."
          action={<Button onClick={openCreate}>Criar usuário</Button>}
        />
      ) : (
        <div className="overflow-x-auto rounded-2xl border border-border bg-surface">
          <table className="w-full min-w-[640px] text-left text-sm">
            <thead className="border-b border-border text-muted">
              <tr>
                <th className="px-4 py-3 font-semibold">Nome</th>
                <th className="px-4 py-3 font-semibold">E-mail</th>
                <th className="px-4 py-3 font-semibold">Perfil</th>
                <th className="px-4 py-3 font-semibold">Ações</th>
              </tr>
            </thead>
            <tbody>
              {items.map((user) => (
                <tr key={user.id} className="border-b border-border/70 last:border-0">
                  <td className="px-4 py-3 font-semibold">{user.nome}</td>
                  <td className="px-4 py-3 text-muted">{user.email}</td>
                  <td className="px-4 py-3">{roleLabel(user.role)}</td>
                  <td className="px-4 py-3">
                    <div className="flex flex-wrap gap-2">
                      <Button variant="secondary" className="px-3 py-1.5 text-xs" onClick={() => openEdit(user)}>
                        Editar
                      </Button>
                      <Button variant="danger" className="px-3 py-1.5 text-xs" onClick={() => void handleDelete(user)}>
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
        title={editing ? 'Editar usuário' : 'Novo usuário'}
        onClose={() => setOpen(false)}
        footer={
          <>
            <Button variant="secondary" onClick={() => setOpen(false)}>
              Cancelar
            </Button>
            <Button type="submit" form="user-form" disabled={saving}>
              {saving ? 'Salvando...' : 'Salvar'}
            </Button>
          </>
        }
      >
        <form id="user-form" onSubmit={handleSubmit}>
          <label className="field">
            Nome
            <input value={nome} onChange={(e) => setNome(e.target.value)} required />
          </label>
          <label className="field">
            E-mail
            <input type="email" value={email} onChange={(e) => setEmail(e.target.value)} required />
          </label>
          <label className="field">
            Senha {editing ? '(deixe vazio para manter)' : ''}
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required={!editing}
            />
          </label>
          <label className="field">
            Perfil
            <select value={role} onChange={(e) => setRole(e.target.value as Role)}>
              <option value="admin">Administrador</option>
              <option value="professor">Professor</option>
              <option value="aluno">Aluno</option>
            </select>
          </label>
        </form>
      </Modal>
    </section>
  )
}
