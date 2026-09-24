import { Link } from 'react-router'
import { buttonVariants } from '@/components/ui/button'
import { cn } from '@/lib/utils'

type InlineLinkProps = {
  to: string
  children: React.ReactNode
  className?: string
}

/** Router link styled with Button `link` variant (avoids Button+render role="button"). */
export function InlineLink({ to, children, className }: InlineLinkProps) {
  return (
    <Link
      to={to}
      className={cn(
        buttonVariants({ variant: 'link' }),
        'inline h-auto cursor-pointer p-0 whitespace-normal wrap-break-word',
        className,
      )}
    >
      {children}
    </Link>
  )
}
