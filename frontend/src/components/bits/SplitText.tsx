import { gsap } from 'gsap'
import { useEffect, useRef } from 'react'

type SplitTextProps = {
  text: string
  className?: string
  delay?: number
  duration?: number
  tag?: 'h1' | 'h2' | 'p' | 'span'
}

export default function SplitText({
  text,
  className = '',
  delay = 40,
  duration = 0.55,
  tag: Tag = 'h1',
}: SplitTextProps) {
  const ref = useRef<HTMLElement>(null)

  useEffect(() => {
    const el = ref.current
    if (!el) return
    const chars = el.querySelectorAll('.split-char')
    const tween = gsap.fromTo(
      chars,
      { opacity: 0, y: 28 },
      { opacity: 1, y: 0, duration, ease: 'power3.out', stagger: delay / 1000 },
    )
    return () => {
      tween.kill()
    }
  }, [text, delay, duration])

  return (
    <Tag ref={ref as never} className={className} aria-label={text}>
      {text.split('').map((char, index) => (
        <span key={`${char}-${index}`} className="split-char" aria-hidden="true">
          {char === ' ' ? '\u00A0' : char}
        </span>
      ))}
    </Tag>
  )
}
