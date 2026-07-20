import { useCallback, useEffect, useRef, useState } from 'react'
import type * as React from 'react'
import { adminApi, errorMessage } from '@/lib/api'
import { PageContainer, StatusBadge, ConfirmDialog, PromptDialog } from '@/components/shared'
import type { AdminItem, UserGroupItem } from '@orange-tv/shared'
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
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Plus, Search, Trash2, KeyRound } from 'lucide-react'
import { toast } from 'sonner'

const emptyForm = { username: '', password: '', email: '', group_id: '1', status: '1' }

export default function AdminsPage() {
  const [list, setList] = useState<AdminItem[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [keyword, setKeyword] = useState('')
  const [error, setError] = useState('')
  const [showCreate, setShowCreate] = useState(false)
  const [form, setForm] = useState(emptyForm)
  const [groups, setGroups] = useState<UserGroupItem[]>([])
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
      const [res, gRes] = await Promise.all([
        adminApi.listAdmins({ page: p, page_size: 20, keyword: k || undefined }),
        adminApi.listGroups({ page_size: 100 }),
      ])
      setList(res.data.list || [])
      setTotal(res.data.total)
      setGroups(gRes.data.list || [])
    } catch (err) {
      setError(errorMessage(err))
    }
  }, [])

  useEffect(() => { void load() }, [page, queryKey, load])

  async function onCreate(e: React.SyntheticEvent<HTMLFormElement>) {
    e.preventDefault()
    setError('')
    try {
      await adminApi.createAdmin({
        username: form.username,
        password: form.password,
        email: form.email,
        group_id: Number(form.group_id),
        status: Number(form.status),
      })
      toast.success('管理员已创建')
      setShowCreate(false)
      setForm(emptyForm)
      await load()
    } catch (err) {
      setError(errorMessage(err))
    }
  }

  async function confirmResetPwd(pwd: string) {
    if (resetId === null) return
    if (!pwd || pwd.length < 6) {
      toast.error('密码至少 6 位')
      return
    }
    try {
      await adminApi.resetAdminPassword(resetId, pwd)
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
      await adminApi.deleteAdmin(deleteId)
      toast.success('管理员已删除')
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
          <div className="flex items-center justify-between">
            <CardTitle>管理员管理</CardTitle>
            <Button size="sm" onClick={() => setShowCreate(!showCreate)}>
              <Plus data-icon="inline-start" />
              新增管理员
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          {error && (
            <Alert variant="destructive" className="mb-4">
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}
          {showCreate && (
            <form onSubmit={onCreate} className="mb-4 flex flex-wrap items-end gap-2 rounded-lg border p-4">
              <Input
                placeholder="用户名"
                value={form.username}
                onChange={(e) => setForm({ ...form, username: e.target.value })}
                required
                minLength={3}
                className="max-w-xs"
              />
              <Input
                type="password"
                placeholder="密码"
                value={form.password}
                onChange={(e) => setForm({ ...form, password: e.target.value })}
                required
                minLength={6}
                className="max-w-xs"
              />
              <Input
                placeholder="邮箱"
                value={form.email}
                onChange={(e) => setForm({ ...form, email: e.target.value })}
                className="max-w-xs"
              />
              <Select value={form.group_id} onValueChange={(v) => setForm({ ...form, group_id: v ?? '1' })}>
                <SelectTrigger className="w-32">
                  <SelectValue placeholder="用户组" />
                </SelectTrigger>
                <SelectContent>
                  {groups.map((g) => (
                    <SelectItem key={g.id} value={String(g.id)}>{g.name}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
              <Select value={form.status} onValueChange={(v) => setForm({ ...form, status: v ?? '1' })}>
                <SelectTrigger className="w-28">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="1">启用</SelectItem>
                  <SelectItem value="0">禁用</SelectItem>
                </SelectContent>
              </Select>
              <Button type="submit" size="sm">保存</Button>
            </form>
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
                        <Button size="sm" variant="ghost" onClick={() => setResetId(a.id)}>
                          <KeyRound data-icon="inline-start" />
                          重置密码
                        </Button>
                        <Button size="sm" variant="ghost" onClick={() => setDeleteId(a.id)}>
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
        title="删除管理员"
        description="确认删除该管理员？此操作不可撤销。"
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
