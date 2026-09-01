import { useEffect, useState } from 'react'
import type { FormEvent } from 'react'
import { api } from '../api/client'
import type { Role, User } from '../auth/auth'
import { roleLabel } from '../auth/auth'

export function UsuariosPage() {
  const [items, setItems] = useState<User[]>([])
  const [nome, setNome] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [role, setRole] = useState<Role>('aluno')
  const [editingId, setEditingId] = useState<number | null>(null)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(true)

  async function load() {
    setLoading(true)
    try {
      setItems(await api.users.list())
      setError('')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Erro ao carregar')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    void load()
  }, [])

  function resetForm() {
    setNome('')
    setEmail('')
    setPassword('')
    setRole('aluno')
    setEditingId(null)
  }

  async function handleSubmit(event: FormEvent) {
    event.preventDefault()
    try {
      if (editingId) {
        await api.users.update(editingId, {
          nome,
          email,
          role,
          password: password || undefined,
        })
      } else {
        await api.users.create({ nome, email, password, role })
      }
      resetForm()
      await load()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Erro ao salvar')
    }
  }

  function handleEdit(user: User) {
    setEditingId(user.id)
    setNome(user.nome)
    setEmail(user.email)
    setPassword('')
    setRole(user.role)
  }

  async function handleDelete(id: number) {
    if (!confirm('Excluir este usuário?')) return
    try {
      await api.users.remove(id)
      await load()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Erro ao excluir')
    }
  }

  return (
    <section>
      <h1>Usuários</h1>
      {error && <p className="error">{error}</p>}

      <form className="card" onSubmit={handleSubmit}>
        <h2>{editingId ? 'Editar usuário' : 'Novo usuário'}</h2>
        <label>
          Nome
          <input value={nome} onChange={(e) => setNome(e.target.value)} required />
        </label>
        <label>
          E-mail
          <input type="email" value={email} onChange={(e) => setEmail(e.target.value)} required />
        </label>
        <label>
          Senha {editingId && '(deixe vazio para manter)'}
          <input
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required={!editingId}
          />
        </label>
        <label>
          Perfil
          <select value={role} onChange={(e) => setRole(e.target.value as Role)}>
            <option value="admin">Administrador</option>
            <option value="professor">Professor</option>
            <option value="aluno">Aluno</option>
          </select>
        </label>
        <div className="actions">
          <button type="submit">{editingId ? 'Salvar' : 'Cadastrar'}</button>
          {editingId && (
            <button type="button" className="secondary" onClick={resetForm}>
              Cancelar
            </button>
          )}
        </div>
      </form>

      <div className="card">
        <h2>Lista</h2>
        {loading ? (
          <p>Carregando...</p>
        ) : (
          <table>
            <thead>
              <tr>
                <th>ID</th>
                <th>Nome</th>
                <th>E-mail</th>
                <th>Perfil</th>
                <th>Ações</th>
              </tr>
            </thead>
            <tbody>
              {items.map((user) => (
                <tr key={user.id}>
                  <td>{user.id}</td>
                  <td>{user.nome}</td>
                  <td>{user.email}</td>
                  <td>{roleLabel(user.role)}</td>
                  <td className="actions">
                    <button type="button" className="secondary" onClick={() => handleEdit(user)}>
                      Editar
                    </button>
                    <button type="button" className="danger" onClick={() => handleDelete(user.id)}>
                      Excluir
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </section>
  )
}
