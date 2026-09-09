import { useState, type ReactNode } from 'react'
import { NavLink } from 'react-router-dom'
import { canManageAlunos, canManageUsers, roleLabel } from '../../auth/auth'
import { useAuth } from '../../auth/AuthContext'
import GradientText from '../bits/GradientText'
import { Button } from '../ui/Button'

type AppShellProps = {
  children: ReactNode
}

const linkClass = ({ isActive }: { isActive: boolean }) =>
  `block rounded-xl px-3 py-2 text-sm font-semibold transition ${
    isActive ? 'bg-accent text-white' : 'text-muted hover:bg-white/6 hover:text-ink'
  }`

export function AppShell({ children }: AppShellProps) {
  const { user, logout } = useAuth()
  const [open, setOpen] = useState(false)
  if (!user) return null

  const nav = (
    <nav className="flex flex-col gap-1">
      <NavLink to="/" end className={linkClass} onClick={() => setOpen(false)}>
        Início
      </NavLink>
      <NavLink to="/aulas" end className={linkClass} onClick={() => setOpen(false)}>
        Agenda
      </NavLink>
      {canManageAlunos(user.role) && (
        <NavLink to="/alunos" className={linkClass} onClick={() => setOpen(false)}>
          Cadastro de alunos
        </NavLink>
      )}
      {canManageUsers(user.role) && (
        <NavLink to="/usuarios" className={linkClass} onClick={() => setOpen(false)}>
          Usuários de acesso
        </NavLink>
      )}
    </nav>
  )

  return (
    <div className="min-h-screen bg-bg lg:grid lg:grid-cols-[260px_1fr]">
      <aside className="hidden border-r border-border bg-bg-elevated px-5 py-6 lg:flex lg:flex-col">
        <GradientText className="mb-8 font-display text-2xl font-bold">EstudaJá</GradientText>
        {nav}
        <div className="mt-auto border-t border-border pt-4">
          <p className="text-sm font-semibold">{user.nome}</p>
          <p className="text-xs text-muted">{roleLabel(user.role)}</p>
          <Button variant="secondary" className="mt-3 w-full" onClick={logout}>
            Sair
          </Button>
        </div>
      </aside>

      <div className="min-w-0">
        <header className="flex items-center justify-between border-b border-border bg-bg-elevated px-4 py-3 lg:hidden">
          <GradientText className="font-display text-xl font-bold">EstudaJá</GradientText>
          <Button variant="secondary" onClick={() => setOpen((value) => !value)}>
            Menu
          </Button>
        </header>
        {open && (
          <div className="border-b border-border bg-bg-elevated px-4 py-4 lg:hidden">
            {nav}
            <div className="mt-4 border-t border-border pt-3">
              <p className="text-sm font-semibold">{user.nome}</p>
              <p className="text-xs text-muted">{roleLabel(user.role)}</p>
              <Button variant="secondary" className="mt-3" onClick={logout}>
                Sair
              </Button>
            </div>
          </div>
        )}
        <main className="mx-auto w-full max-w-6xl px-4 py-8 sm:px-6">{children}</main>
      </div>
    </div>
  )
}
