import { useEffect, useState } from 'react'
import type * as React from 'react'
import { adminApi, errorMessage } from '@/lib/api'
import { PageContainer, ConfirmDialog } from '@/components/shared'
import type { UserGroupItem } from '@orange-tv/shared'
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
import { Plus, Trash2 } from 'lucide-react'
import { toast } from 'sonner'

const emptyForm = { name: '', permissions: '', description: '' }

export default function UserGroupsPage() {
  const [list, setList] = useState<UserGroupItem[]>([])
  const [total, setTotal] = useState(0)
  const [error, setError] = useState('')
  const [showCreate, setShowCreate] = useState(false)
  const [form, setForm] = useState(emptyForm)
  const [deleteId, setDeleteId] = useState<number | null>(null)

  async function load() {
    setError('')
    try {
      const res = await adminApi.listGroups({ page_size: 100 })
      setList(res.data.list || [])
      setTotal(res.data.total)
    } catch (err) {
      setError(errorMessage(err))
    }
  }

  useEffect(() => { void load() }, [])

  async function onCreate(e: React.SyntheticEvent<HTMLFormElement>) {
    e.preventDefault()
    setError('')
    try {
      await adminApi.createGroup(form)
      toast.success('用户组已创建')
      setShowCreate(false)
      setForm(emptyForm)
      await load()
    } catch (err) {
      setError(errorMessage(err))
    }
  }

  async function confirmDelete() {
    if (deleteId === null) return
    try {
      await adminApi.deleteGroup(deleteId)
      toast.success('用户组已删除')
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
            <CardTitle>用户组管理</CardTitle>
            <Button size="sm" onClick={() => setShowCreate(!showCreate)}>
              <Plus data-icon="inline-start" />
              新增用户组
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
                placeholder="名称"
                value={form.name}
                onChange={(e) => setForm({ ...form, name: e.target.value })}
                required
                className="max-w-xs"
              />
              <Input
                placeholder="权限（JSON）"
                value={form.permissions}
                onChange={(e) => setForm({ ...form, permissions: e.target.value })}
                className="max-w-xs"
              />
              <Input
                placeholder="描述"
                value={form.description}
                onChange={(e) => setForm({ ...form, description: e.target.value })}
                className="max-w-xs"
              />
              <Button type="submit" size="sm">保存</Button>
            </form>
          )}
          <p className="mb-2 text-sm text-muted-foreground">共 {total} 条</p>
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
                        <Button size="sm" variant="ghost" onClick={() => setDeleteId(g.id)}>
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
        </CardContent>
      </Card>

      <ConfirmDialog
        open={deleteId !== null}
        onOpenChange={(open) => { if (!open) setDeleteId(null) }}
        title="删除用户组"
        description="确认删除该用户组？此操作不可撤销。"
        destructive
        onConfirm={confirmDelete}
      />
    </PageContainer>
  )
}
