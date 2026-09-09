import { useRef, type MouseEventHandler, type ReactNode } from 'react'

type SpotlightCardProps = {
  children: ReactNode
  className?: string
  spotlightColor?: `rgba(${number}, ${number}, ${number}, ${number})`
}

export default function SpotlightCard({
  children,
  className = '',
  spotlightColor = 'rgba(139, 92, 246, 0.28)',
}: SpotlightCardProps) {
  const divRef = useRef<HTMLDivElement>(null)

  const handleMouseMove: MouseEventHandler<HTMLDivElement> = (event) => {
    if (!divRef.current) return
    const rect = divRef.current.getBoundingClientRect()
    divRef.current.style.setProperty('--mouse-x', `${event.clientX - rect.left}px`)
    divRef.current.style.setProperty('--mouse-y', `${event.clientY - rect.top}px`)
    divRef.current.style.setProperty('--spotlight-color', spotlightColor)
  }

  return (
    <div ref={divRef} onMouseMove={handleMouseMove} className={`card-spotlight ${className}`}>
      {children}
    </div>
  )
}
