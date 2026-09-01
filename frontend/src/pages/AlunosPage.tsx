import { useEffect, useState } from 'react'
import type { FormEvent } from 'react'
import { api } from '../api/client'
import type { Aluno } from '../api/client'

export function AlunosPage({ canWrite = false }: { canWrite?: boolean }) {
  const [items, setItems] = useState<Aluno[]>([])
  const [nome, setNome] = useState('')
  const [email, setEmail] = useState('')
  const [editingId, setEditingId] = useState<number | null>(null)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(true)

  async function load() {
    setLoading(true)
    try {
      setItems(await api.alunos.list())
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
    setEditingId(null)
  }

  async function handleSubmit(event: FormEvent) {
    event.preventDefault()
    try {
      if (editingId) {
        await api.alunos.update(editingId, { nome, email })
      } else {
        await api.alunos.create({ nome, email })
      }
      resetForm()
      await load()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Erro ao salvar')
    }
  }

  function handleEdit(aluno: Aluno) {
    setEditingId(aluno.id)
    setNome(aluno.nome)
    setEmail(aluno.email)
  }

  async function handleDelete(id: number) {
    if (!confirm('Excluir este aluno?')) return
    try {
      await api.alunos.remove(id)
      await load()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Erro ao excluir')
    }
  }

  return (
    <section>
      <h1>Alunos</h1>
      {error && <p className="error">{error}</p>}

      {canWrite && (
        <form className="card" onSubmit={handleSubmit}>
        <h2>{editingId ? 'Editar aluno' : 'Novo aluno'}</h2>
        <label>
          Nome
          <input value={nome} onChange={(e) => setNome(e.target.value)} required />
        </label>
        <label>
          E-mail
          <input type="email" value={email} onChange={(e) => setEmail(e.target.value)} required />
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
      )}

      <div className="card">
        <h2>Lista</h2>
        {loading ? (
          <p>Carregando...</p>
        ) : items.length === 0 ? (
          <p>Nenhum aluno cadastrado.</p>
        ) : (
          <table>
            <thead>
              <tr>
                <th>ID</th>
                <th>Nome</th>
                <th>E-mail</th>
                {canWrite && <th>Ações</th>}
              </tr>
            </thead>
            <tbody>
              {items.map((aluno) => (
                <tr key={aluno.id}>
                  <td>{aluno.id}</td>
                  <td>{aluno.nome}</td>
                  <td>{aluno.email}</td>
                  {canWrite && (
                    <td className="actions">
                      <button type="button" className="secondary" onClick={() => handleEdit(aluno)}>
                        Editar
                      </button>
                      <button type="button" className="danger" onClick={() => handleDelete(aluno.id)}>
                        Excluir
                      </button>
                    </td>
                  )}
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </section>
  )
}
