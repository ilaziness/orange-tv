import type { UserGroupItem } from '@orange-tv/shared'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Button } from '@/components/ui/button'
import { Spinner } from '@/components/ui/spinner'
import { Trash2, Pencil } from 'lucide-react'

interface GroupListProps {
  list: UserGroupItem[]
  loading: boolean
  onEdit: (item: UserGroupItem) => void
  onDelete: (id: number) => void
}

export function GroupList({ list, loading, onEdit, onDelete }: GroupListProps) {
  return (
    <div className="relative rounded-md border">
      {loading && (
        <div className="absolute inset-0 z-10 flex items-center justify-center bg-background/50">
          <Spinner />
        </div>
      )}
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead className="w-16">ID</TableHead>
            <TableHead>名称</TableHead>
            <TableHead>描述</TableHead>
            <TableHead className="w-32">操作</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {list.map((g) => (
            <TableRow key={g.id}>
              <TableCell>{g.id}</TableCell>
              <TableCell className="font-medium">{g.name}</TableCell>
              <TableCell>{g.description || '-'}</TableCell>
              <TableCell>
                <div className="flex gap-1">
                  {g.name !== 'super_admin' && (
                    <Button size="sm" variant="ghost" onClick={() => onEdit(g)}>
                      <Pencil data-icon="inline-start" />
                      编辑
                    </Button>
                  )}
                  {g.name !== 'super_admin' ? (
                    <Button size="sm" variant="ghost" onClick={() => onDelete(g.id)}>
                      <Trash2 data-icon="inline-start" />
                      删除
                    </Button>
                  ) : (
                    <span className="text-sm text-muted-foreground">系统组</span>
                  )}
                </div>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  )
}
