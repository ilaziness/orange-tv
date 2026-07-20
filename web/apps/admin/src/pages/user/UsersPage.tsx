import { useCallback, useEffect, useRef, useState } from 'react'
import { adminApi, errorMessage } from '@/lib/api'
import { PageContainer, StatusBadge, ConfirmDialog, PromptDialog } from '@/components/shared'
import type { UserItem } from '@orange-tv/shared'
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Search, Trash2, KeyRound } from 'lucide-react'
import { toast } from 'sonner'

export default function UsersPage() {
  const [list, setList] = useState<UserItem[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [keyword, setKeyword] = useState('')
  const [error, setError] = useState('')
  const [queryKey, setQueryKey] = useState(0)
  const [deleteId, setDeleteId] = useState<number | null>(null)
  const [resetId, setResetId] = useState<number | null>(null)
  const keywordRef = useRef(keyword)
  const pageRef = useRef(page)

  useEffect(() => { keywordRef.current = keyword }, [keyword])
  useEffect(() => { pageRef.current = page }, [page])

  const load = useCallback(async (p = pageRef.current, k = keywordRef.current) => {
    setError('')
    try {
      const res = await adminApi.listUsers({ page: p, page_size: 20, keyword: k || undefined })
      setList(res.data.list || [])
      setTotal(res.data.total)
    } catch (err) {
      setError(errorMessage(err))
    }
  }, [])

  useEffect(() => { void load() }, [page, queryKey, load])

  async function onToggleStatus(u: UserItem) {
    try {
      await adminApi.updateUser(u.id, { status: u.status === 1 ? 0 : 1 })
      toast.success(u.status === 1 ? '已禁用' : '已启用')
      await load()
    } catch (err) {
      toast.error(errorMessage(err))
    }
  }

  async function confirmResetPwd(pwd: string) {
    if (resetId === null) return
    if (!pwd || pwd.length < 6) {
      toast.error('密码至少 6 位')
      return
    }
    try {
      await adminApi.resetUserPassword(resetId, pwd)
      toast.success('密码已重置')
    } catch (err) {
      toast.error(errorMessage(err))
    } finally {
      setResetId(null)
    }
  }

  async function confirmDelete() {
    if (deleteId === null) return
    try {
      await adminApi.deleteUser(deleteId)
      toast.success('用户已删除')
      await load()
    } catch (err) {
      toast.error(errorMessage(err))
    } finally {
      setDeleteId(null)
    }
  }

  return (
    <PageContainer>
      <Card>
        <CardHeader>
          <CardTitle>用户管理</CardTitle>
        </CardHeader>
        <CardContent>
          {error && (
            <Alert variant="destructive" className="mb-4">
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}
          <div className="mb-4 flex gap-2">
            <Input
              placeholder="用户名/邮箱"
              value={keyword}
              onChange={(e) => setKeyword(e.target.value)}
              className="max-w-xs"
              onKeyDown={(e) => { if (e.key === 'Enter') { setPage(1); setQueryKey((q) => q + 1) } }}
            />
            <Button variant="outline" size="sm" onClick={() => { setPage(1); setQueryKey((q) => q + 1) }}>
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
                  <TableHead className="w-20">状态</TableHead>
                  <TableHead className="w-40">最后登录</TableHead>
                  <TableHead className="w-40">注册时间</TableHead>
                  <TableHead className="w-48">操作</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {list.map((u) => (
                  <TableRow key={u.id}>
                    <TableCell>{u.id}</TableCell>
                    <TableCell className="font-medium">{u.username}</TableCell>
                    <TableCell>{u.email || '-'}</TableCell>
                    <TableCell><StatusBadge status={u.status} /></TableCell>
                    <TableCell className="text-xs text-muted-foreground">{u.last_login_at || '-'}</TableCell>
                    <TableCell className="text-xs text-muted-foreground">{u.created_at || '-'}</TableCell>
                    <TableCell>
                      <div className="flex gap-1">
                        <Button size="sm" variant="ghost" onClick={() => void onToggleStatus(u)}>
                          {u.status === 1 ? '禁用' : '启用'}
                        </Button>
                        <Button size="sm" variant="ghost" onClick={() => setResetId(u.id)}>
                          <KeyRound data-icon="inline-start" />
                          重置密码
                        </Button>
                        <Button size="sm" variant="ghost" onClick={() => setDeleteId(u.id)}>
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
        </CardContent>
      </Card>

      <ConfirmDialog
        open={deleteId !== null}
        onOpenChange={(open) => { if (!open) setDeleteId(null) }}
        title="删除用户"
        description="确认删除该用户？此操作不可撤销。"
        destructive
        onConfirm={confirmDelete}
      />

      <PromptDialog
        open={resetId !== null}
        onOpenChange={(open) => { if (!open) setResetId(null) }}
        title="重置密码"
        description="输入新密码（至少 6 位）"
        label="新密码"
        placeholder="请输入新密码"
        confirmText="重置"
        onConfirm={confirmResetPwd}
      />
    </PageContainer>
  )
}
