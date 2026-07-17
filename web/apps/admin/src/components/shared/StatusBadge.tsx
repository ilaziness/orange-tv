import { Badge } from '@/components/ui/badge'

interface StatusBadgeProps {
  status: number
  activeText?: string
  inactiveText?: string
}

export function StatusBadge({
  status,
  activeText = '启用',
  inactiveText = '禁用',
}: StatusBadgeProps) {
  return (
    <Badge variant={status === 1 ? 'default' : 'secondary'}>
      {status === 1 ? activeText : inactiveText}
    </Badge>
  )
}
