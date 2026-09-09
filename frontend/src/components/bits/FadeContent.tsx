import { useEffect, useRef, type CSSProperties, type ReactNode } from 'react'

type FadeContentProps = {
  children: ReactNode
  className?: string
  style?: CSSProperties
  duration?: number
  delay?: number
  blur?: boolean
}

export default function FadeContent({
  children,
  className = '',
  style,
  duration = 700,
  delay = 0,
  blur = false,
}: FadeContentProps) {
  const ref = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const el = ref.current
    if (!el) return
    el.style.opacity = '0'
    el.style.transform = 'translateY(16px)'
    if (blur) el.style.filter = 'blur(8px)'

    const observer = new IntersectionObserver(
      ([entry]) => {
        if (!entry.isIntersecting) return
        el.style.transition = `opacity ${duration}ms ease, transform ${duration}ms ease, filter ${duration}ms ease`
        el.style.transitionDelay = `${delay}ms`
        el.style.opacity = '1'
        el.style.transform = 'translateY(0)'
        el.style.filter = 'blur(0)'
        observer.disconnect()
      },
      { threshold: 0.12 },
    )
    observer.observe(el)
    return () => observer.disconnect()
  }, [blur, delay, duration])

  return (
    <div ref={ref} className={className} style={style}>
      {children}
    </div>
  )
}
