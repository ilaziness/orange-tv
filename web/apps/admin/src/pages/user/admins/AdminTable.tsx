import { StatusBadge } from '@/components/shared'
import type { AdminItem } from '@orange-tv/shared'
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
import { KeyRound, Trash2, Pencil } from 'lucide-react'

interface AdminTableProps {
  list: AdminItem[]
  loading: boolean
  onEdit: (item: AdminItem) => void
  onReset: (id: number) => void
  onDelete: (id: number) => void
}

export function AdminTable({ list, loading, onEdit, onReset, onDelete }: AdminTableProps) {
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
            <TableHead>用户名</TableHead>
            <TableHead>邮箱</TableHead>
            <TableHead className="w-28">用户组</TableHead>
            <TableHead className="w-20">状态</TableHead>
            <TableHead className="w-40">最后登录</TableHead>
            <TableHead className="w-48">操作</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {list.map((a) => (
            <TableRow key={a.id}>
              <TableCell>{a.id}</TableCell>
              <TableCell className="font-medium">{a.username}</TableCell>
              <TableCell>{a.email || '-'}</TableCell>
              <TableCell>{a.group_name}</TableCell>
              <TableCell>
                <StatusBadge status={a.status} />
              </TableCell>
              <TableCell className="text-xs text-muted-foreground">
                {a.last_login_at || '-'}
              </TableCell>
              <TableCell>
                <div className="flex gap-1">
                  <Button size="sm" variant="ghost" onClick={() => onEdit(a)}>
                    <Pencil data-icon="inline-start" />
                    编辑
                  </Button>
                  <Button size="sm" variant="ghost" onClick={() => onReset(a.id)}>
                    <KeyRound data-icon="inline-start" />
                    重置密码
                  </Button>
                  <Button size="sm" variant="ghost" onClick={() => onDelete(a.id)}>
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
