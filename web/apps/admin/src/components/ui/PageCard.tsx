import type { ReactNode } from 'react'

type PageCardProps = {
  children: ReactNode
  className?: string
}

export function PageCard({ children, className }: PageCardProps) {
  return <div className={`page-card${className ? ` ${className}` : ''}`}>{children}</div>
}
