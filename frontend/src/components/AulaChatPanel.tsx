import { useEffect, useRef, useState } from 'react'
import type { FormEvent } from 'react'
import { buildAulaChatWsUrl } from '../api/client'
import { getToken } from '../auth/auth'
import { Button } from './ui/Button'

const MAX_TEXTO = 500
const PING_INTERVAL_MS = 3 * 60 * 1000

export type ChatMessage = {
  id: string
  aula_id: number
  autor: { id: number; nome: string }
  texto: string
  enviado_em: string
}

type ConnState = 'conectando' | 'conectado' | 'desconectado' | 'erro'

type ServerFrame =
  | { type: 'chat.message'; id: string; aula_id: number; autor: { id: number; nome: string }; texto: string; enviado_em: string }
  | { type: 'chat.error'; error: string }
  | { type: 'chat.pong' }
  | { type: string }

export type AulaChatPanelProps = {
  aulaId: number
}

function countRunes(text: string): number {
  return Array.from(text).length
}

export function AulaChatPanel({ aulaId }: AulaChatPanelProps) {
  const [messages, setMessages] = useState<ChatMessage[]>([])
  const [texto, setTexto] = useState('')
  const [connState, setConnState] = useState<ConnState>('conectando')
  const [statusMsg, setStatusMsg] = useState('Conectando ao chat…')
  const [formError, setFormError] = useState('')
  const listRef = useRef<HTMLUListElement | null>(null)
  const wsRef = useRef<WebSocket | null>(null)

  useEffect(() => {
    setMessages([])
    setTexto('')
    setFormError('')
    setConnState('conectando')
    setStatusMsg('Conectando ao chat…')

    const token = getToken()
    if (!token) {
      setConnState('erro')
      setStatusMsg('não autenticado')
      return
    }

    const url = buildAulaChatWsUrl(aulaId, token)
    const ws = new WebSocket(url)
    wsRef.current = ws
    let closedByCleanup = false
    let pingTimer: ReturnType<typeof setInterval> | null = null

    ws.onopen = () => {
      if (closedByCleanup) return
      setConnState('conectado')
      setStatusMsg('')
      pingTimer = setInterval(() => {
        if (ws.readyState === WebSocket.OPEN) {
          ws.send(JSON.stringify({ type: 'chat.ping' }))
        }
      }, PING_INTERVAL_MS)
    }

    ws.onmessage = (event) => {
      if (closedByCleanup) return
      let frame: ServerFrame
      try {
        frame = JSON.parse(String(event.data)) as ServerFrame
      } catch {
        setFormError('falha ao receber mensagem')
        return
      }

      if (frame.type === 'chat.message' && 'id' in frame && 'autor' in frame) {
        const msg = frame as Extract<ServerFrame, { type: 'chat.message' }>
        setMessages((prev) => [
          ...prev,
          {
            id: msg.id,
            aula_id: msg.aula_id,
            autor: msg.autor,
            texto: msg.texto,
            enviado_em: msg.enviado_em,
          },
        ])
        return
      }

      if (frame.type === 'chat.error' && 'error' in frame) {
        setFormError((frame as { error: string }).error || 'falha ao enviar mensagem')
        return
      }
    }

    ws.onerror = () => {
      if (closedByCleanup) return
      setConnState('erro')
      setStatusMsg('falha na conexão do chat')
    }

    ws.onclose = () => {
      if (closedByCleanup) return
      setConnState('desconectado')
      setStatusMsg('conexão encerrada')
      if (pingTimer) {
        clearInterval(pingTimer)
        pingTimer = null
      }
    }

    return () => {
      closedByCleanup = true
      if (pingTimer) clearInterval(pingTimer)
      wsRef.current = null
      if (ws.readyState === WebSocket.OPEN || ws.readyState === WebSocket.CONNECTING) {
        ws.close()
      }
      setMessages([])
    }
  }, [aulaId])

  useEffect(() => {
    const el = listRef.current
    if (el) {
      el.scrollTop = el.scrollHeight
    }
  }, [messages])

  function handleSubmit(event: FormEvent) {
    event.preventDefault()
    setFormError('')

    const trimmed = texto.trim()
    if (!trimmed) {
      setFormError('mensagem vazia')
      return
    }
    if (countRunes(trimmed) > MAX_TEXTO) {
      setFormError('mensagem excede 500 caracteres')
      return
    }

    const ws = wsRef.current
    if (!ws || ws.readyState !== WebSocket.OPEN) {
      setFormError('chat desconectado')
      return
    }

    ws.send(JSON.stringify({ type: 'chat.send', texto: trimmed }))
    setTexto('')
  }

  const runeCount = countRunes(texto.trim())
  const canSend = connState === 'conectado'

  return (
    <div className="rounded-2xl border border-border bg-surface p-4 sm:p-5">
      {statusMsg && (
        <p
          className={`mb-3 text-sm ${
            connState === 'erro' || connState === 'desconectado' ? 'text-live' : 'text-muted'
          }`}
        >
          {statusMsg}
        </p>
      )}

      <ul
        className="mb-4 max-h-72 space-y-3 overflow-y-auto rounded-xl border border-border bg-bg/40 px-3 py-3"
        ref={listRef}
        aria-live="polite"
      >
        {messages.length === 0 ? (
          <li className="py-6 text-center text-sm text-muted">
            Nenhuma mensagem nesta sessão. O histórico não é guardado.
          </li>
        ) : (
          messages.map((msg) => (
            <li key={msg.id} className="rounded-xl border border-border/80 bg-bg-elevated px-3 py-2">
              <p className="text-xs text-muted">
                <strong className="text-ink">{msg.autor.nome}</strong>
                {' · '}
                {msg.enviado_em}
              </p>
              <p className="mt-1 text-sm">{msg.texto}</p>
            </li>
          ))
        )}
      </ul>

      {formError && (
        <p className="mb-3 rounded-xl bg-live/12 px-3 py-2 text-sm text-live">{formError}</p>
      )}

      <form onSubmit={handleSubmit}>
        <label className="field mb-3">
          Mensagem
          <textarea
            value={texto}
            onChange={(e) => {
              setTexto(e.target.value)
              if (formError) setFormError('')
            }}
            rows={2}
            maxLength={MAX_TEXTO * 2}
            placeholder="Escreva uma mensagem (máx. 500 caracteres)"
            disabled={!canSend}
          />
        </label>
        <div className="flex flex-wrap items-center justify-between gap-3">
          <span className="text-xs text-muted">
            {runeCount}/{MAX_TEXTO}
          </span>
          <Button type="submit" disabled={!canSend}>
            Enviar
          </Button>
        </div>
      </form>
    </div>
  )
}
