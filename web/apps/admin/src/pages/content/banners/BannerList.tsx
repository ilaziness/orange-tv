import { StatusBadge } from '@/components/shared'
import type { BannerItem } from '@orange-tv/shared'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Button } from '@/components/ui/button'
import { Trash2 } from 'lucide-react'

interface BannerListProps {
  list: BannerItem[]
  onToggle: (b: BannerItem) => void
  onDelete: (id: number) => void
}

export function BannerList({ list, onToggle, onDelete }: BannerListProps) {
  return (
    <div className="rounded-md border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead className="w-16">ID</TableHead>
            <TableHead>标题</TableHead>
            <TableHead className="w-24">封面</TableHead>
            <TableHead className="w-16">排序</TableHead>
            <TableHead className="w-20">状态</TableHead>
            <TableHead className="w-32">操作</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {list.map((b) => (
            <TableRow key={b.id}>
              <TableCell>{b.id}</TableCell>
              <TableCell className="font-medium">{b.title}</TableCell>
              <TableCell>
                {b.cover ? (
                  <img src={b.cover} alt={b.title} className="rounded object-cover" style={{ width: 60, height: 34 }} />
                ) : '-'}
              </TableCell>
              <TableCell>{b.sort}</TableCell>
              <TableCell><StatusBadge status={b.status} /></TableCell>
              <TableCell>
                <div className="flex gap-1">
                  <Button size="sm" variant="ghost" onClick={() => void onToggle(b)}>
                    {b.status === 1 ? '禁用' : '启用'}
                  </Button>
                  <Button size="sm" variant="ghost" onClick={() => onDelete(b.id)}>
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
