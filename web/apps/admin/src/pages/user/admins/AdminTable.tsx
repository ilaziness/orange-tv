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
import { Input } from '@/components/ui/input'
import { Search, KeyRound, Trash2 } from 'lucide-react'

interface AdminTableProps {
  list: AdminItem[]
  total: number
  keyword: string
  setKeyword: (v: string) => void
  setQueryKey: (f: (v: number) => number) => void
  onReset: (id: number) => void
  onDelete: (id: number) => void
}

export function AdminTable({
  list,
  total,
  keyword,
  setKeyword,
  setQueryKey,
  onReset,
  onDelete,
}: AdminTableProps) {
  function handleSearch() {
    setQueryKey((q) => q + 1)
  }

  return (
    <>
      <div className="mb-4 flex gap-2">
        <Input
          placeholder="用户名/邮箱"
          value={keyword}
          onChange={(e) => setKeyword(e.target.value)}
          className="max-w-xs"
          onKeyDown={(e) => { if (e.key === 'Enter') handleSearch() }}
        />
        <Button variant="outline" size="sm" onClick={handleSearch}>
          <Search data-icon="inline-start" />
          查询
        </Button>
      </div>
      <p className="mb-2 text-sm text-muted-foreground">共 {total} 条</p>
      <div className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className="w-16">ID</TableHead>
              <TableHead>用户名</TableHead>
              <TableHead>邮箱</TableHead>
              <TableHead className="w-28">用户组</TableHead>
              <TableHead className="w-20">状态</TableHead>
              <TableHead className="w-40">最后登录</TableHead>
              <TableHead className="w-40">操作</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {list.map((a) => (
              <TableRow key={a.id}>
                <TableCell>{a.id}</TableCell>
                <TableCell className="font-medium">{a.username}</TableCell>
                <TableCell>{a.email || '-'}</TableCell>
                <TableCell>{a.group_name}</TableCell>
                <TableCell><StatusBadge status={a.status} /></TableCell>
                <TableCell className="text-xs text-muted-foreground">{a.last_login_at || '-'}</TableCell>
                <TableCell>
                  <div className="flex gap-1">
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
    </>
  )
}
