import type { ReactNode } from 'react'

type GradientTextProps = {
  children: ReactNode
  className?: string
  colors?: string[]
}

export default function GradientText({
  children,
  className = '',
  colors = ['#8b5cf6', '#f0abfc', '#67e8f9', '#8b5cf6'],
}: GradientTextProps) {
  return (
    <span className={`animated-gradient-text ${className}`}>
      <span
        className="text-content"
        style={{ backgroundImage: `linear-gradient(to right, ${colors.join(', ')})` }}
      >
        {children}
      </span>
    </span>
  )
}
