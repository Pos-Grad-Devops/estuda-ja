import { useEffect, useRef, useState } from 'react'
import type { LiveModo } from '../api/client'

const IVS_PLAYER_CDN = 'https://player.live-video.net/1.55.0/amazon-ivs-player.min.js'

type IVSPlayerInstance = {
  attachHTMLVideoElement: (el: HTMLVideoElement) => void
  load: (url: string) => void
  play: () => void
  delete: () => void
  addEventListener: (event: string, fn: (payload?: unknown) => void) => void
  removeEventListener: (event: string, fn: (payload?: unknown) => void) => void
}

type IVSPlayerGlobal = {
  isPlayerSupported: boolean
  create: () => IVSPlayerInstance
  PlayerState?: { IDLE?: string; READY?: string; BUFFERING?: string; PLAYING?: string; ENDED?: string }
  PlayerEventType?: { ERROR?: string }
}

declare global {
  interface Window {
    IVSPlayer?: IVSPlayerGlobal
  }
}

function loadIvsScript(): Promise<IVSPlayerGlobal> {
  if (window.IVSPlayer) {
    return Promise.resolve(window.IVSPlayer)
  }
  const existing = document.querySelector<HTMLScriptElement>(`script[src="${IVS_PLAYER_CDN}"]`)
  if (existing) {
    return new Promise((resolve, reject) => {
      existing.addEventListener('load', () => {
        if (window.IVSPlayer) resolve(window.IVSPlayer)
        else reject(new Error('IVS Player não carregou'))
      })
      existing.addEventListener('error', () => reject(new Error('Falha ao carregar o player IVS')))
    })
  }
  return new Promise((resolve, reject) => {
    const script = document.createElement('script')
    script.src = IVS_PLAYER_CDN
    script.async = true
    script.onload = () => {
      if (window.IVSPlayer) resolve(window.IVSPlayer)
      else reject(new Error('IVS Player não carregou'))
    }
    script.onerror = () => reject(new Error('Falha ao carregar o player IVS'))
    document.head.appendChild(script)
  })
}

export type LivePlayerProps = {
  modo: LiveModo
  playbackUrl: string | null
  mensagem?: string
}

/**
 * Player da transmissão ao vivo. No stub / sem URL: mensagem PT (sem &lt;video&gt; vazio).
 * Com modo ivs + URL: Amazon IVS Player Web SDK (CDN) + &lt;video&gt; âncora.
 * Visual: tokens da escola (superfície, cantos, borda, cor live).
 */
export function LivePlayer({ modo, playbackUrl, mensagem }: LivePlayerProps) {
  const videoRef = useRef<HTMLVideoElement | null>(null)
  const playerRef = useRef<IVSPlayerInstance | null>(null)
  const [statusMsg, setStatusMsg] = useState<string | null>(null)
  const [fatal, setFatal] = useState<string | null>(null)

  const canPlayIvs = modo === 'ivs' && Boolean(playbackUrl)

  useEffect(() => {
    if (!canPlayIvs || !playbackUrl) {
      return
    }

    let cancelled = false
    setFatal(null)
    setStatusMsg('Conectando à transmissão…')

    void (async () => {
      try {
        const IVSPlayer = await loadIvsScript()
        if (cancelled) return
        if (!IVSPlayer.isPlayerSupported) {
          setFatal('Seu navegador não suporta o player de transmissão ao vivo.')
          setStatusMsg(null)
          return
        }
        const video = videoRef.current
        if (!video) return

        const player = IVSPlayer.create()
        playerRef.current = player
        player.attachHTMLVideoElement(video)

        const onPlaying = () => {
          if (!cancelled) setStatusMsg(null)
        }
        const onError = () => {
          if (!cancelled) {
            setStatusMsg(
              'Transmissão ao vivo ativa, mas ainda não há sinal de vídeo. Aguarde o professor iniciar o encoder (OBS).',
            )
          }
        }
        const playingState = IVSPlayer.PlayerState?.PLAYING ?? 'Playing'
        const errorEvent = IVSPlayer.PlayerEventType?.ERROR ?? 'Error'
        player.addEventListener(playingState, onPlaying)
        player.addEventListener(errorEvent, onError)

        player.load(playbackUrl)
        player.play()
        setStatusMsg(
          'Aguardando sinal de vídeo… Se o professor ainda não enviou pelo OBS, a imagem aparecerá em breve.',
        )
      } catch (err) {
        if (!cancelled) {
          setFatal(err instanceof Error ? err.message : 'Não foi possível iniciar o player')
          setStatusMsg(null)
        }
      }
    })()

    return () => {
      cancelled = true
      const player = playerRef.current
      playerRef.current = null
      if (player) {
        try {
          player.delete()
        } catch {
          /* ignore */
        }
      }
    }
  }, [canPlayIvs, playbackUrl])

  if (!canPlayIvs) {
    return (
      <div
        className="rounded-2xl border border-live/30 bg-live/10 px-4 py-6 text-sm text-ink"
        role="status"
      >
        <p>
          {mensagem ??
            (modo === 'stub'
              ? 'Ambiente local: não há sinal de vídeo real. A transmissão está marcada como ao vivo apenas para testes.'
              : 'Não há URL de reprodução disponível para esta transmissão.')}
        </p>
      </div>
    )
  }

  return (
    <div className="overflow-hidden rounded-2xl border border-border bg-surface">
      {fatal && (
        <p className="border-b border-live/30 bg-live/12 px-4 py-2 text-sm text-live">{fatal}</p>
      )}
      {statusMsg && !fatal && (
        <p className="border-b border-border px-4 py-2 text-sm text-muted">{statusMsg}</p>
      )}
      <video
        ref={videoRef}
        className="aspect-video w-full bg-bg"
        playsInline
        controls
        controlsList="nodownload"
      />
    </div>
  )
}
