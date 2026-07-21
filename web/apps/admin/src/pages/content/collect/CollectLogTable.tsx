import type { CollectLog } from '@orange-tv/shared'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'

interface CollectLogTableProps {
  logs: CollectLog[]
}

export function CollectLogTable({ logs }: CollectLogTableProps) {
  return (
    <div className="rounded-md border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead className="w-16">ID</TableHead>
            <TableHead className="w-20">源</TableHead>
            <TableHead className="w-20">状态</TableHead>
            <TableHead className="w-20">总数</TableHead>
            <TableHead className="w-20">成功</TableHead>
            <TableHead className="w-20">失败</TableHead>
            <TableHead className="w-24">耗时ms</TableHead>
            <TableHead>时间</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {logs.map((l) => (
            <TableRow key={l.id}>
              <TableCell>{l.id}</TableCell>
              <TableCell>{l.source_id}</TableCell>
              <TableCell>{l.status}</TableCell>
              <TableCell>{l.total_count}</TableCell>
              <TableCell>{l.success_count}</TableCell>
              <TableCell>{l.failed_count}</TableCell>
              <TableCell>{l.duration_ms}</TableCell>
              <TableCell className="text-xs text-muted-foreground">{l.created_at || '-'}</TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  )
}
