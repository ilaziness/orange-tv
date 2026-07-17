import { useEffect, useState } from 'react'
import type * as React from 'react'
import { adminApi, errorMessage } from '@/lib/api'
import { PageContainer, StatusBadge, ConfirmDialog } from '@/components/shared'
import type { LiveChannel } from '@orange-tv/shared'
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
import { Empty } from '@/components/ui/empty'
import { RefreshCw, Pencil, Trash2 } from 'lucide-react'
import { toast } from 'sonner'

const emptyForm = { name: '', category: '', stream_url: '', logo: '', description: '', sort_order: 0, status: 1 }

export default function LivePage() {
  const [list, setList] = useState<LiveChannel[]>([])
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const [form, setForm] = useState(emptyForm)
  const [editId, setEditId] = useState(0)
  const [deleteId, setDeleteId] = useState<number | null>(null)

  async function load() {
    setLoading(true)
    setError('')
    try {
      const res = await adminApi.listLive({ page: 1, page_size: 100 })
      setList(res.data.list || [])
    } catch (err) {
      setError(errorMessage(err))
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { void load() }, [])

  async function onSubmit(e: React.SyntheticEvent<HTMLFormElement>) {
    e.preventDefault()
    setError('')
    try {
      if (editId) {
        await adminApi.updateLive(editId, form)
        toast.success('直播频道已更新')
      } else {
        await adminApi.createLive(form)
        toast.success('直播频道已创建')
      }
      setForm(emptyForm)
      setEditId(0)
      await load()
    } catch (err) {
      setError(errorMessage(err))
    }
  }

  async function confirmDelete() {
    if (deleteId === null) return
    try {
      await adminApi.deleteLive(deleteId)
      toast.success('直播频道已删除')
      await load()
    } catch (err) {
      toast.error(errorMessage(err))
    } finally {
      setDeleteId(null)
    }
  }

  function startEdit(item: LiveChannel) {
    setEditId(item.id)
    setForm({
      name: item.name,
      category: item.category || '',
      stream_url: item.stream_url,
      logo: item.logo || '',
      description: item.description || '',
      sort_order: item.sort_order || 0,
      status: item.status ?? 1,
    })
  }

  return (
    <PageContainer>
      <Card>
        <CardHeader>
          <CardTitle>直播管理</CardTitle>
        </CardHeader>
        <CardContent>
          {error && (
            <Alert variant="destructive" className="mb-4">
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}
          <form onSubmit={onSubmit} className="mb-6 flex flex-col gap-4">
            <div className="flex flex-wrap gap-2">
              <Input
                placeholder="频道名称"
                value={form.name}
                onChange={(e) => setForm({ ...form, name: e.target.value })}
                required
                className="max-w-xs"
              />
              <Input
                placeholder="分类"
                value={form.category}
                onChange={(e) => setForm({ ...form, category: e.target.value })}
                className="max-w-xs"
              />
              <Input
                placeholder="直播流地址"
                value={form.stream_url}
                onChange={(e) => setForm({ ...form, stream_url: e.target.value })}
                required
                className="min-w-[280px] flex-1"
              />
            </div>
            <div className="flex flex-wrap items-center gap-2">
              <Input
                placeholder="Logo URL"
                value={form.logo}
                onChange={(e) => setForm({ ...form, logo: e.target.value })}
                className="max-w-xs"
              />
              <Input
                placeholder="简介"
                value={form.description}
                onChange={(e) => setForm({ ...form, description: e.target.value })}
                className="max-w-xs"
              />
              <Input
                type="number"
                placeholder="排序"
                value={form.sort_order}
                onChange={(e) => setForm({ ...form, sort_order: Number(e.target.value) })}
                className="w-24"
              />
              <Select value={String(form.status)} onValueChange={(v) => setForm({ ...form, status: Number(v ?? '1') })}>
                <SelectTrigger className="w-28">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="1">启用</SelectItem>
                  <SelectItem value="0">禁用</SelectItem>
                </SelectContent>
              </Select>
              <Button type="submit" size="sm">
                {editId ? '保存修改' : '新增频道'}
              </Button>
              {editId ? (
                <Button type="button" variant="outline" size="sm" onClick={() => { setEditId(0); setForm(emptyForm) }}>
                  取消
                </Button>
              ) : null}
              <Button type="button" variant="outline" size="sm" onClick={() => void load()} disabled={loading}>
                <RefreshCw data-icon="inline-start" />
                刷新
              </Button>
            </div>
          </form>

          <div className="rounded-md border">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="w-16">ID</TableHead>
                  <TableHead>名称</TableHead>
                  <TableHead className="w-24">分类</TableHead>
                  <TableHead className="w-20">状态</TableHead>
                  <TableHead className="w-16">排序</TableHead>
                  <TableHead className="w-32">操作</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {list.map((item) => (
                  <TableRow key={item.id}>
                    <TableCell>{item.id}</TableCell>
                    <TableCell>
                      <div className="font-medium">{item.name}</div>
                      <div className="max-w-[360px] truncate text-xs text-muted-foreground">{item.stream_url}</div>
                    </TableCell>
                    <TableCell>{item.category || '-'}</TableCell>
                    <TableCell><StatusBadge status={item.status ?? 1} /></TableCell>
                    <TableCell>{item.sort_order}</TableCell>
                    <TableCell>
                      <div className="flex gap-1">
                        <Button size="sm" variant="ghost" onClick={() => startEdit(item)}>
                          <Pencil data-icon="inline-start" />
                          编辑
                        </Button>
                        <Button size="sm" variant="ghost" onClick={() => setDeleteId(item.id)}>
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
          {list.length === 0 && (
            <Empty className="py-8 text-center text-sm text-muted-foreground">暂无直播频道</Empty>
          )}
        </CardContent>
      </Card>

      <ConfirmDialog
        open={deleteId !== null}
        onOpenChange={(open) => { if (!open) setDeleteId(null) }}
        title="删除直播频道"
        description="确认删除该直播频道？此操作不可撤销。"
        destructive
        onConfirm={confirmDelete}
      />
    </PageContainer>
  )
}

