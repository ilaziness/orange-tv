import { StatusBadge } from '@/components/shared'
import type { VideoListItem } from '@orange-tv/shared'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import { Pencil, Trash2 } from 'lucide-react'
import { Link } from 'react-router'

interface VideoTableProps {
  items: VideoListItem[]
  selected: Set<number>
  onToggleSelect: (id: number) => void
  onSelectAll: () => void
  onDelete: (id: number) => void
}

export function VideoTable({ items, selected, onToggleSelect, onSelectAll, onDelete }: VideoTableProps) {
  return (
    <div className="rounded-md border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead className="w-12">
              <Checkbox
                checked={selected.size === items.length && items.length > 0}
                onCheckedChange={onSelectAll}
              />
            </TableHead>
            <TableHead className="w-16">ID</TableHead>
            <TableHead>标题</TableHead>
            <TableHead className="w-20">年份</TableHead>
            <TableHead className="w-20">评分</TableHead>
            <TableHead className="w-20">状态</TableHead>
            <TableHead className="w-32">操作</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {items.map((item) => (
            <TableRow key={item.id}>
              <TableCell>
                <Checkbox
                  checked={selected.has(item.id)}
                  onCheckedChange={() => onToggleSelect(item.id)}
                />
              </TableCell>
              <TableCell>{item.id}</TableCell>
              <TableCell className="font-medium">{item.title}</TableCell>
              <TableCell>{item.year || '-'}</TableCell>
              <TableCell>{item.rating}</TableCell>
              <TableCell>
                <StatusBadge status={item.publish_status ?? 0} activeText="上架" inactiveText="下架" />
              </TableCell>
              <TableCell>
                <div className="flex gap-1">
                  <Button size="sm" variant="ghost" render={<Link to={`/content/videos/${item.id}`} />}>
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
