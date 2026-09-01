import { Link, Navigate, Route, Routes } from 'react-router-dom'
import {
  canManageAlunos,
  canManageAulas,
  canManageCursos,
  canManageUsers,
  roleLabel,
} from './auth/auth'
import { useAuth } from './auth/AuthContext'
import { ProtectedRoute } from './auth/ProtectedRoute'
import { AlunosPage } from './pages/AlunosPage'
import { AulasPage } from './pages/AulasPage'
import { CursosPage } from './pages/CursosPage'
import { LoginPage } from './pages/LoginPage'
import { UsuariosPage } from './pages/UsuariosPage'

function AppLayout() {
  const { user, logout } = useAuth()
  if (!user) return null

  return (
    <div className="layout">
      <header>
        <div>
          <strong>EstudaJá</strong>
          <span>
            {user.nome} · {roleLabel(user.role)}
          </span>
        </div>
        <nav>
          <Link to="/cursos">Cursos</Link>
          <Link to="/aulas">Aulas</Link>
          {canManageAlunos(user.role) && <Link to="/alunos">Alunos</Link>}
          {canManageUsers(user.role) && <Link to="/usuarios">Usuários</Link>}
          <button type="button" className="secondary" onClick={logout}>
            Sair
          </button>
        </nav>
      </header>
      <main>
        <Routes>
          <Route path="/" element={<Navigate to="/cursos" replace />} />
          <Route
            path="/cursos"
            element={
              <CursosPage
                title="Cursos"
                emptyMessage="Nenhum curso cadastrado."
                canWrite={canManageCursos(user.role)}
              />
            }
          />
          <Route
            path="/aulas"
            element={<AulasPage canWrite={canManageAulas(user.role)} />}
          />
          <Route
            path="/alunos"
            element={
              <ProtectedRoute roles={['admin']}>
                <AlunosPage canWrite={canManageAlunos(user.role)} />
              </ProtectedRoute>
            }
          />
          <Route
            path="/usuarios"
            element={
              <ProtectedRoute roles={['admin']}>
                <UsuariosPage />
              </ProtectedRoute>
            }
          />
        </Routes>
      </main>
    </div>
  )
}

export default function App() {
  const { loading } = useAuth()

  if (loading) {
    return <p className="loading-page">Carregando...</p>
  }

  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route
        path="/*"
        element={
          <ProtectedRoute>
            <AppLayout />
          </ProtectedRoute>
        }
      />
    </Routes>
  )
}
