import type { CollectLog } from '@orange-tv/shared'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'

interface CollectLogTableProps {
  logs: CollectLog[]
}

function LogStatusBadge({ status }: { status: number }) {
  if (status === 1) {
    return <Badge>已完成</Badge>
  }
  if (status === 2) {
    return <Badge variant="secondary">采集中</Badge>
  }
  if (status === 3) {
    return <Badge variant="destructive">失败</Badge>
  }
  return <Badge variant="outline">{String(status)}</Badge>
}

export function CollectLogTable({ logs }: CollectLogTableProps) {
  return (
    <div className="rounded-md border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead className="w-16">ID</TableHead>
            <TableHead className="w-32">采集源</TableHead>
            <TableHead className="w-20">状态</TableHead>
            <TableHead className="w-24">采集数量</TableHead>
            <TableHead className="w-24">耗时(秒)</TableHead>
            <TableHead>时间</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {logs.map((l) => (
            <TableRow key={l.id}>
              <TableCell>{l.id}</TableCell>
              <TableCell>{l.source_name || `#${l.source_id}`}</TableCell>
              <TableCell>
                <LogStatusBadge status={l.status} />
              </TableCell>
              <TableCell>{l.collect_count}</TableCell>
              <TableCell>{l.duration_sec}</TableCell>
              <TableCell className="text-xs text-muted-foreground">{l.created_at || '-'}</TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  )
}
