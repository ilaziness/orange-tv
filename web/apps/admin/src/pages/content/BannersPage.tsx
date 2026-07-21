import { useEffect, useState } from 'react'
import type * as React from 'react'
import { adminApi, errorMessage } from '@/lib/api'
import { PageContainer, StatusBadge, ConfirmDialog } from '@/components/shared'
import type { BannerItem } from '@orange-tv/shared'
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
import { Plus, Trash2 } from 'lucide-react'
import { toast } from 'sonner'

const emptyForm = { title: '', cover: '', link: '', video_id: 0, sort: 0, status: 1 }

export default function BannersPage() {
  const [list, setList] = useState<BannerItem[]>([])
  const [total, setTotal] = useState(0)
  const [error, setError] = useState('')
  const [showCreate, setShowCreate] = useState(false)
  const [form, setForm] = useState(emptyForm)
  const [deleteId, setDeleteId] = useState<number | null>(null)

  async function load() {
    setError('')
    try {
      const res = await adminApi.listBanners({ page_size: 100 })
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
      await adminApi.createBanner(form)
      toast.success('Banner 已创建')
      setShowCreate(false)
      setForm(emptyForm)
      await load()
    } catch (err) {
      setError(errorMessage(err))
    }
  }

  async function onToggle(b: BannerItem) {
    try {
      await adminApi.updateBanner(b.id, { status: b.status === 1 ? 0 : 1 })
      toast.success(b.status === 1 ? '已禁用' : '已启用')
      await load()
    } catch (err) {
      toast.error(errorMessage(err))
    }
  }

  async function confirmDelete() {
    if (deleteId === null) return
    try {
      await adminApi.deleteBanner(deleteId)
      toast.success('Banner 已删除')
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
            <CardTitle>Banner 管理</CardTitle>
            <Button size="sm" onClick={() => setShowCreate(!showCreate)}>
              <Plus data-icon="inline-start" />
              新增 Banner
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
                placeholder="标题"
                value={form.title}
                onChange={(e) => setForm({ ...form, title: e.target.value })}
                required
                className="max-w-xs"
              />
              <Input
                placeholder="封面URL"
                value={form.cover}
                onChange={(e) => setForm({ ...form, cover: e.target.value })}
                className="max-w-xs"
              />
              <Input
                placeholder="链接"
                value={form.link}
                onChange={(e) => setForm({ ...form, link: e.target.value })}
                className="max-w-xs"
              />
              <Input
                type="number"
                placeholder="影视ID"
                value={form.video_id || ''}
                onChange={(e) => setForm({ ...form, video_id: Number(e.target.value) })}
                className="w-28"
              />
              <Input
                type="number"
                placeholder="排序"
                value={form.sort}
                onChange={(e) => setForm({ ...form, sort: Number(e.target.value) })}
                className="w-24"
              />
              <Select items={[{ value: '1', label: '启用' }, { value: '0', label: '禁用' }]} value={String(form.status)} onValueChange={(v) => setForm({ ...form, status: Number(v ?? '1') })}>
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
          <p className="mb-2 text-sm text-muted-foreground">共 {total} 条</p>
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
                        <img src={b.cover} alt="" className="size-[60x34] rounded object-cover" style={{ width: 60, height: 34 }} />
                      ) : '-'}
                    </TableCell>
                    <TableCell>{b.sort}</TableCell>
                    <TableCell><StatusBadge status={b.status} /></TableCell>
                    <TableCell>
                      <div className="flex gap-1">
                        <Button size="sm" variant="ghost" onClick={() => void onToggle(b)}>
                          {b.status === 1 ? '禁用' : '启用'}
                        </Button>
                        <Button size="sm" variant="ghost" onClick={() => setDeleteId(b.id)}>
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
        title="删除 Banner"
        description="确认删除该 Banner？此操作不可撤销。"
        destructive
        onConfirm={confirmDelete}
      />
    </PageContainer>
  )
}
