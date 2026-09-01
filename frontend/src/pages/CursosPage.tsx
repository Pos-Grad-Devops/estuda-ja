import { useEffect, useState } from 'react'
import type { FormEvent } from 'react'
import { api } from '../api/client'
import type { Curso } from '../api/client'

type Props = {
  title: string
  emptyMessage: string
  canWrite?: boolean
}

export function CursosPage({ title, emptyMessage, canWrite = false }: Props) {
  const [items, setItems] = useState<Curso[]>([])
  const [titulo, setTitulo] = useState('')
  const [descricao, setDescricao] = useState('')
  const [editingId, setEditingId] = useState<number | null>(null)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(true)

  async function load() {
    setLoading(true)
    try {
      setItems(await api.cursos.list())
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
    setTitulo('')
    setDescricao('')
    setEditingId(null)
  }

  async function handleSubmit(event: FormEvent) {
    event.preventDefault()
    try {
      if (editingId) {
        await api.cursos.update(editingId, { titulo, descricao })
      } else {
        await api.cursos.create({ titulo, descricao })
      }
      resetForm()
      await load()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Erro ao salvar')
    }
  }

  async function handleEdit(curso: Curso) {
    setEditingId(curso.id)
    setTitulo(curso.titulo)
    setDescricao(curso.descricao)
  }

  async function handleDelete(id: number) {
    if (!confirm('Excluir este curso?')) return
    try {
      await api.cursos.remove(id)
      await load()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Erro ao excluir')
    }
  }

  return (
    <section>
      <h1>{title}</h1>
      {error && <p className="error">{error}</p>}

      {canWrite && (
        <form className="card" onSubmit={handleSubmit}>
          <h2>{editingId ? 'Editar curso' : 'Novo curso'}</h2>
          <label>
            Título
            <input value={titulo} onChange={(e) => setTitulo(e.target.value)} required />
          </label>
          <label>
            Descrição
            <textarea value={descricao} onChange={(e) => setDescricao(e.target.value)} rows={3} />
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
          <p>{emptyMessage}</p>
        ) : (
          <table>
            <thead>
              <tr>
                <th>ID</th>
                <th>Título</th>
                <th>Descrição</th>
                {canWrite && <th>Ações</th>}
              </tr>
            </thead>
            <tbody>
              {items.map((curso) => (
                <tr key={curso.id}>
                  <td>{curso.id}</td>
                  <td>{curso.titulo}</td>
                  <td>{curso.descricao || '—'}</td>
                  {canWrite && (
                    <td className="actions">
                      <button type="button" className="secondary" onClick={() => handleEdit(curso)}>
                        Editar
                      </button>
                      <button type="button" className="danger" onClick={() => handleDelete(curso.id)}>
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
