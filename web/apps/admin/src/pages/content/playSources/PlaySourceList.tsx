import { StatusBadge } from '@/components/shared'
import type { PlaySource } from '@orange-tv/shared'
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
import { Spinner } from '@/components/ui/spinner'

interface PlaySourceListProps {
  items: PlaySource[]
  togglingId: number | null
  onEdit: (item: PlaySource) => void
  onToggle: (item: PlaySource) => void
  onDelete: (id: number) => void
}

export function PlaySourceList({ items, togglingId, onEdit, onToggle, onDelete }: PlaySourceListProps) {
  return (
    <div className="rounded-md border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead className="w-16">ID</TableHead>
            <TableHead>名称</TableHead>
            <TableHead className="w-16">排序</TableHead>
            <TableHead className="w-20">状态</TableHead>
            <TableHead className="w-48">操作</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {items.map((item) => {
            const toggling = togglingId === item.id
            return (
              <TableRow key={item.id}>
                <TableCell>{item.id}</TableCell>
                <TableCell className="font-medium">{item.name}</TableCell>
                <TableCell>{item.sort_order}</TableCell>
                <TableCell><StatusBadge status={item.status} /></TableCell>
                <TableCell>
                  <div className="flex gap-1">
                    <Button size="sm" variant="ghost" onClick={() => onEdit(item)}>
                      <Pencil data-icon="inline-start" />
                      编辑
                    </Button>
                    <Button
                      size="sm"
                      variant="ghost"
                      disabled={toggling}
                      onClick={() => void onToggle(item)}
                    >
                      {toggling && <Spinner data-icon="inline-start" />}
                      {item.status === 1 ? '禁用' : '启用'}
                    </Button>
                    <Button size="sm" variant="ghost" onClick={() => onDelete(item.id)}>
                      <Trash2 data-icon="inline-start" />
                      删除
                    </Button>
                  </div>
                </TableCell>
              </TableRow>
            )
          })}
        </TableBody>
      </Table>
    </div>
  )
}
