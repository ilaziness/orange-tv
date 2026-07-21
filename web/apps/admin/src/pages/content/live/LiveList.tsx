import { StatusBadge } from '@/components/shared'
import type { LiveChannel } from '@orange-tv/shared'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Button } from '@/components/ui/button'
import { Pencil, Trash2 } from 'lucide-react'

interface LiveListProps {
  list: LiveChannel[]
  onEdit: (item: LiveChannel) => void
  onDelete: (id: number) => void
}

export function LiveList({ list, onEdit, onDelete }: LiveListProps) {
  return (
    <div className="rounded-md border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead className="w-16">ID</TableHead>
            <TableHead>名称</TableHead>
            <TableHead className="w-24">分类</TableHead>
            <TableHead className="w-20">状态</TableHead>
            <TableHead className="w-16">排序</TableHead>
            <TableHead className="w-32">操作</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {list.map((item) => (
            <TableRow key={item.id}>
              <TableCell>{item.id}</TableCell>
              <TableCell>
                <div className="font-medium">{item.name}</div>
                <div className="max-w-[360px] truncate text-xs text-muted-foreground">{item.stream_url}</div>
              </TableCell>
              <TableCell>{item.category || '-'}</TableCell>
              <TableCell><StatusBadge status={item.status ?? 1} /></TableCell>
              <TableCell>{item.sort_order}</TableCell>
              <TableCell>
                <div className="flex gap-1">
                  <Button size="sm" variant="ghost" onClick={() => onEdit(item)}>
                    <Pencil data-icon="inline-start" />
                    编辑
                  </Button>
                  <Button size="sm" variant="ghost" onClick={() => onDelete(item.id)}>
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
