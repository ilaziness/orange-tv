import { StatusBadge } from '@/components/shared'
import type { UserItem } from '@orange-tv/shared'
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

interface UserTableProps {
  list: UserItem[]
  loading: boolean
  onEdit: (item: UserItem) => void
  onToggle: (u: UserItem) => void
  onReset: (id: number) => void
  onDelete: (id: number) => void
}

export function UserTable({ list, loading, onEdit, onToggle, onReset, onDelete }: UserTableProps) {
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
            <TableHead>昵称</TableHead>
            <TableHead>邮箱</TableHead>
            <TableHead className="w-20">状态</TableHead>
            <TableHead className="w-40">最后登录</TableHead>
            <TableHead className="w-40">注册时间</TableHead>
            <TableHead className="w-56">操作</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {list.map((u) => (
            <TableRow key={u.id}>
              <TableCell>{u.id}</TableCell>
              <TableCell className="font-medium">{u.nickname || u.email || '-'}</TableCell>
              <TableCell>{u.email || '-'}</TableCell>
              <TableCell>
                <StatusBadge status={u.status} />
              </TableCell>
              <TableCell className="text-xs text-muted-foreground">
                {u.last_login_at || '-'}
              </TableCell>
              <TableCell className="text-xs text-muted-foreground">{u.created_at || '-'}</TableCell>
              <TableCell>
                <div className="flex gap-1">
                  <Button size="sm" variant="ghost" onClick={() => onEdit(u)}>
                    <Pencil data-icon="inline-start" />
                    编辑
                  </Button>
                  <Button size="sm" variant="ghost" onClick={() => void onToggle(u)}>
                    {u.status === 1 ? '禁用' : '启用'}
                  </Button>
                  <Button size="sm" variant="ghost" onClick={() => onReset(u.id)}>
                    <KeyRound data-icon="inline-start" />
                    重置密码
                  </Button>
                  <Button size="sm" variant="ghost" onClick={() => onDelete(u.id)}>
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
