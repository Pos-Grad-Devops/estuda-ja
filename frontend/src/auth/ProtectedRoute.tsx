import { Navigate, useLocation } from 'react-router-dom'
import { useAuth } from './AuthContext'
import type { Role } from './auth'

type Props = {
  children: React.ReactNode
  roles?: Role[]
}

export function ProtectedRoute({ children, roles }: Props) {
  const { user, loading } = useAuth()
  const location = useLocation()

  if (loading) {
    return <p className="loading-page">Carregando a escola...</p>
  }

  if (!user) {
    return <Navigate to="/login" replace state={{ from: location.pathname }} />
  }

  if (roles && !roles.includes(user.role)) {
    return <Navigate to="/" replace />
  }

  return children
}
