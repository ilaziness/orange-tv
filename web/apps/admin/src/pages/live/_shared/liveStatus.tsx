import { Badge } from '@/components/ui/badge'
import { applyStatusLabel, roomStatusLabel } from './liveLabels'

export function ApplyStatusBadge({ status }: { status: number }) {
  const variant = status === 0 ? 'outline' : status === 1 ? 'default' : 'secondary'
  return <Badge variant={variant}>{applyStatusLabel(status)}</Badge>
}

export function RoomStatusBadge({ status }: { status: number }) {
  const variant = status === 1 ? 'default' : status === 2 ? 'destructive' : 'secondary'
  return <Badge variant={variant}>{roomStatusLabel(status)}</Badge>
}
