import { useEffect, useRef } from 'react'

type CountUpProps = {
  to: number
  from?: number
  duration?: number
  className?: string
}

export default function CountUp({ to, from = 0, duration = 1.15, className = '' }: CountUpProps) {
  const ref = useRef<HTMLSpanElement>(null)

  useEffect(() => {
    const el = ref.current
    if (!el) return
    const start = performance.now()
    let frame = 0

    const tick = (now: number) => {
      const progress = Math.min((now - start) / (duration * 1000), 1)
      const eased = 1 - (1 - progress) ** 3
      el.textContent = String(Math.round(from + (to - from) * eased))
      if (progress < 1) {
        frame = requestAnimationFrame(tick)
      }
    }

    el.textContent = String(from)
    frame = requestAnimationFrame(tick)
    return () => cancelAnimationFrame(frame)
  }, [duration, from, to])

  return <span ref={ref} className={className} />
}