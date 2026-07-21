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
import { Trash2 } from 'lucide-react'

interface GroupListProps {
  list: UserGroupItem[]
  onDelete: (id: number) => void
}

export function GroupList({ list, onDelete }: GroupListProps) {
  return (
    <div className="rounded-md border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead className="w-16">ID</TableHead>
            <TableHead>名称</TableHead>
            <TableHead>描述</TableHead>
            <TableHead className="w-24">操作</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {list.map((g) => (
            <TableRow key={g.id}>
              <TableCell>{g.id}</TableCell>
              <TableCell className="font-medium">{g.name}</TableCell>
              <TableCell>{g.description || '-'}</TableCell>
              <TableCell>
                {g.name !== 'super_admin' ? (
                  <Button size="sm" variant="ghost" onClick={() => onDelete(g.id)}>
                    <Trash2 data-icon="inline-start" />
                    删除
                  </Button>
                ) : (
                  <span className="text-sm text-muted-foreground">系统组</span>
                )}
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  )
}
