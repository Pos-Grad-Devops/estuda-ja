import { useState, type FormEvent } from 'react'
import { Navigate, useLocation, useNavigate } from 'react-router-dom'
import { useAuth } from '../auth/AuthContext'
import Aurora from '../components/bits/Aurora'
import GradientText from '../components/bits/GradientText'
import SplitText from '../components/bits/SplitText'
import { Button } from '../components/ui/Button'

export function LoginPage() {
  const { user, login } = useAuth()
  const navigate = useNavigate()
  const location = useLocation()
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [saving, setSaving] = useState(false)

  if (user) {
    return <Navigate to="/" replace />
  }

  async function handleSubmit(event: FormEvent) {
    event.preventDefault()
    setError('')
    setSaving(true)
    try {
      await login(email, password)
      const redirect = (location.state as { from?: string } | null)?.from ?? '/'
      navigate(redirect, { replace: true })
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Erro ao entrar')
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="relative min-h-screen overflow-hidden bg-bg">
      <div className="absolute inset-0 opacity-80">
        <Aurora colorStops={['#5227FF', '#8b5cf6', '#22d3ee']} amplitude={1.05} blend={0.55} />
      </div>
      <div className="relative z-10 grid min-h-screen place-items-center px-4 py-10">
        <form
          onSubmit={handleSubmit}
          className="w-full max-w-md rounded-3xl border border-white/10 bg-bg-elevated/85 p-8 shadow-2xl backdrop-blur-xl"
        >
          <p className="text-xs font-semibold uppercase tracking-[0.2em] text-accent-hover">Escola ao vivo</p>
          <SplitText text="EstudaJá" className="mt-3 font-display text-4xl font-bold" tag="h1" />
          <p className="mt-3 text-sm text-muted">
            Entre para ver o catálogo, a agenda e as aulas — com o mesmo acesso do seu perfil.
          </p>
          {error && <p className="mt-4 rounded-xl bg-live/12 px-3 py-2 text-sm text-live">{error}</p>}
          <label className="field mt-6">
            E-mail
            <input type="email" value={email} onChange={(e) => setEmail(e.target.value)} required />
          </label>
          <label className="field">
            Senha
            <input type="password" value={password} onChange={(e) => setPassword(e.target.value)} required />
          </label>
          <Button type="submit" className="mt-2 w-full" disabled={saving}>
            {saving ? 'Entrando...' : 'Entrar'}
          </Button>
          <p className="mt-5 text-center text-xs text-muted">
            <GradientText>Aula magna + VOD</GradientText> · pico no horário certo
          </p>
        </form>
      </div>
    </div>
  )
}
