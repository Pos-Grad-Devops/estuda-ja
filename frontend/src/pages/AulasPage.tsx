import { useEffect, useState } from 'react'
import type { FormEvent } from 'react'
import { api } from '../api/client'
import type { Aula, Curso } from '../api/client'
import { formatDateBR, isValidDateTimeBR, toInputDateTimeBR } from '../utils/date'

export function AulasPage({ canWrite = false }: { canWrite?: boolean }) {
  const [items, setItems] = useState<Aula[]>([])
  const [cursos, setCursos] = useState<Curso[]>([])
  const [cursoId, setCursoId] = useState('')
  const [titulo, setTitulo] = useState('')
  const [descricao, setDescricao] = useState('')
  const [agendadaEm, setAgendadaEm] = useState('')
  const [status, setStatus] = useState<Aula['status']>('agendada')
  const [editingId, setEditingId] = useState<number | null>(null)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(true)

  async function load() {
    setLoading(true)
    try {
      const [aulas, cursosList] = await Promise.all([api.aulas.list(), api.cursos.list()])
      setItems(aulas)
      setCursos(cursosList)
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
    setCursoId('')
    setTitulo('')
    setDescricao('')
    setAgendadaEm('')
    setStatus('agendada')
    setEditingId(null)
  }

  async function handleSubmit(event: FormEvent) {
    event.preventDefault()
    if (!isValidDateTimeBR(agendadaEm)) {
      setError('Data inválida. Use o formato dd/mm/yyyy ou dd/mm/yyyy HH:mm')
      return
    }
    const payload = {
      curso_id: Number(cursoId),
      titulo,
      descricao,
      agendada_em: agendadaEm.trim(),
      status,
    }
    try {
      if (editingId) {
        await api.aulas.update(editingId, payload)
      } else {
        await api.aulas.create(payload)
      }
      resetForm()
      await load()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Erro ao salvar')
    }
  }

  function handleEdit(aula: Aula) {
    setEditingId(aula.id)
    setCursoId(String(aula.curso_id))
    setTitulo(aula.titulo)
    setDescricao(aula.descricao)
    setAgendadaEm(toInputDateTimeBR(aula.agendada_em))
    setStatus(aula.status)
  }

  async function handleDelete(id: number) {
    if (!confirm('Excluir esta aula?')) return
    try {
      await api.aulas.remove(id)
      await load()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Erro ao excluir')
    }
  }

  return (
    <section>
      <h1>Aulas</h1>
      {error && <p className="error">{error}</p>}

      {canWrite && (
        <form className="card" onSubmit={handleSubmit}>
        <h2>{editingId ? 'Editar aula' : 'Nova aula'}</h2>
        <label>
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
        <label>
          Título
          <input value={titulo} onChange={(e) => setTitulo(e.target.value)} required />
        </label>
        <label>
          Descrição
          <textarea value={descricao} onChange={(e) => setDescricao(e.target.value)} rows={3} />
        </label>
        <label>
          Agendada em
          <input
            type="text"
            value={agendadaEm}
            onChange={(e) => setAgendadaEm(e.target.value)}
            placeholder="dd/mm/yyyy HH:mm"
            required
          />
        </label>
        <label>
          Status
          <select value={status} onChange={(e) => setStatus(e.target.value as Aula['status'])}>
            <option value="agendada">Agendada</option>
            <option value="ao_vivo">Ao vivo</option>
            <option value="encerrada">Encerrada</option>
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
      )}

      <div className="card">
        <h2>Lista</h2>
        {loading ? (
          <p>Carregando...</p>
        ) : items.length === 0 ? (
          <p>Nenhuma aula cadastrada.</p>
        ) : (
          <table>
            <thead>
              <tr>
                <th>ID</th>
                <th>Curso</th>
                <th>Título</th>
                <th>Agendada em</th>
                <th>Status</th>
                {canWrite && <th>Ações</th>}
              </tr>
            </thead>
            <tbody>
              {items.map((aula) => (
                <tr key={aula.id}>
                  <td>{aula.id}</td>
                  <td>{aula.curso?.titulo ?? aula.curso_id}</td>
                  <td>{aula.titulo}</td>
                  <td>{formatDateBR(aula.agendada_em)}</td>
                  <td>{aula.status}</td>
                  {canWrite && (
                    <td className="actions">
                      <button type="button" className="secondary" onClick={() => handleEdit(aula)}>
                        Editar
                      </button>
                      <button type="button" className="danger" onClick={() => handleDelete(aula.id)}>
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
