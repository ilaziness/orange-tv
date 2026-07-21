import { StatusBadge } from '@/components/shared'
import type { CollectSource } from '@orange-tv/shared'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Button } from '@/components/ui/button'
import { Play, Square, Trash2 } from 'lucide-react'

interface CollectSourceListProps {
  sources: CollectSource[]
  selectedId: number
  onEdit: (item: CollectSource) => void
  onStart: (id: number) => void
  onStop: (id: number) => void
  onDelete: (id: number) => void
}

export function CollectSourceList({
  sources,
  selectedId,
  onEdit,
  onStart,
  onStop,
  onDelete,
}: CollectSourceListProps) {
  return (
    <div className="rounded-md border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead className="w-16">ID</TableHead>
            <TableHead>名称</TableHead>
            <TableHead className="w-24">类型</TableHead>
            <TableHead className="w-20">状态</TableHead>
            <TableHead className="w-40">最后采集</TableHead>
            <TableHead className="w-48">操作</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {sources.map((s) => (
            <TableRow key={s.id} data-state={s.id === selectedId ? 'selected' : undefined}>
              <TableCell>{s.id}</TableCell>
              <TableCell>
                <div className="font-medium">{s.name}</div>
                <div className="max-w-[320px] truncate text-xs text-muted-foreground">{s.collect_url}</div>
              </TableCell>
              <TableCell>{s.type === 2 ? '苹果CMS' : '默认'}</TableCell>
              <TableCell><StatusBadge status={s.status} /></TableCell>
              <TableCell className="text-xs text-muted-foreground">{s.last_collect_at || '-'}</TableCell>
              <TableCell>
                <div className="flex flex-wrap gap-1">
                  <Button size="sm" variant="ghost" onClick={() => onEdit(s)}>
                    编辑/映射
                  </Button>
                  <Button size="sm" variant="ghost" onClick={() => void onStart(s.id)}>
                    <Play data-icon="inline-start" />
                    开始
                  </Button>
                  <Button size="sm" variant="ghost" onClick={() => void onStop(s.id)}>
                    <Square data-icon="inline-start" />
                    停止
                  </Button>
                  <Button size="sm" variant="ghost" onClick={() => onDelete(s.id)}>
                    <Trash2 data-icon="inline-start" />
                    删除
                  </Button>
                </div>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  )
}
