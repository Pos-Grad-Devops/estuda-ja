import { Navigate, Route, Routes } from 'react-router-dom'
import { useAuth } from './auth/AuthContext'
import { ProtectedRoute } from './auth/ProtectedRoute'
import { AppShell } from './components/layout/AppShell'
import { AgendaPage } from './pages/AgendaPage'
import { AlunosPage } from './pages/AlunosPage'
import { AulaPage } from './pages/AulaPage'
import { CursoPage } from './pages/CursoPage'
import { HomePage } from './pages/HomePage'
import { LoginPage } from './pages/LoginPage'
import { UsuariosPage } from './pages/UsuariosPage'

function AppLayout() {
  return (
    <AppShell>
      <Routes>
        <Route path="/" element={<HomePage />} />
        <Route path="/cursos" element={<Navigate to="/" replace />} />
        <Route path="/cursos/:id" element={<CursoPage />} />
        <Route path="/aulas" element={<AgendaPage />} />
        <Route path="/aulas/:id" element={<AulaPage />} />
        <Route
          path="/alunos"
          element={
            <ProtectedRoute roles={['admin']}>
              <AlunosPage />
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
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </AppShell>
  )
}

export default function App() {
  const { loading } = useAuth()

  if (loading) {
    return <p className="loading-page">Carregando a escola...</p>
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
