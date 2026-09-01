export type Role = 'admin' | 'professor' | 'aluno'

export type User = {
  id: number
  nome: string
  email: string
  role: Role
  created_at: string
  updated_at: string
}

export type AuthResponse = {
  token: string
  user: User
}

const TOKEN_KEY = 'estudaja_token'
const USER_KEY = 'estudaja_user'

export function saveAuth(data: AuthResponse) {
  localStorage.setItem(TOKEN_KEY, data.token)
  localStorage.setItem(USER_KEY, JSON.stringify(data.user))
}

export function clearAuth() {
  localStorage.removeItem(TOKEN_KEY)
  localStorage.removeItem(USER_KEY)
}

export function getToken(): string | null {
  return localStorage.getItem(TOKEN_KEY)
}

export function getStoredUser(): User | null {
  const raw = localStorage.getItem(USER_KEY)
  if (!raw) return null
  try {
    return JSON.parse(raw) as User
  } catch {
    return null
  }
}

export function canManageCursos(role: Role) {
  return role === 'admin'
}

export function canManageAulas(role: Role) {
  return role === 'admin' || role === 'professor'
}

export function canManageAlunos(role: Role) {
  return role === 'admin'
}

export function canManageUsers(role: Role) {
  return role === 'admin'
}

export function roleLabel(role: Role) {
  switch (role) {
    case 'admin':
      return 'Administrador'
    case 'professor':
      return 'Professor'
    case 'aluno':
      return 'Aluno'
  }
}
