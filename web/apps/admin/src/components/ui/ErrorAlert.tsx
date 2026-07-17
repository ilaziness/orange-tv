import type { ReactNode } from 'react'

export function ErrorAlert({ children }: { children: ReactNode }) {
  if (!children) return null
  return <p className="error">{children}</p>
}
