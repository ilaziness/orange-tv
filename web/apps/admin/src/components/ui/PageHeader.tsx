import type { ReactNode } from 'react'

type PageHeaderProps = {
  title: ReactNode
  children?: ReactNode
}

export function PageHeader({ title, children }: PageHeaderProps) {
  return (
    <div className="page-header">
      <h1>{title}</h1>
      {children}
    </div>
  )
}
