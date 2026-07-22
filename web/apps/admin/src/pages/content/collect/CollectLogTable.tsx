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
                {l.status === 1 ? (
                  <span className="text-xs text-green-600">已完成</span>
                ) : l.status === 2 ? (
                  <span className="text-xs text-blue-600">采集中</span>
                ) : l.status === 3 ? (
                  <span className="text-xs text-red-600">失败</span>
                ) : (
                  <span className="text-xs text-muted-foreground">{l.status}</span>
                )}
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
