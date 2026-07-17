import { useCallback, useEffect, useRef, useState } from 'react'
import { adminApi, errorMessage } from '@/lib/api'
import { PageContainer, StatusBadge, ConfirmDialog } from '@/components/shared'
import { Link } from 'react-router'
import type { VideoListItem } from '@orange-tv/shared'
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Checkbox } from '@/components/ui/checkbox'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Plus, Search, Pencil, Trash2, ChevronLeft, ChevronRight } from 'lucide-react'
import { toast } from 'sonner'

export default function VideosPage() {
  const [items, setItems] = useState<VideoListItem[]>([])
  const [keyword, setKeyword] = useState('')
  const [error, setError] = useState('')
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)
  const [selected, setSelected] = useState<Set<number>>(new Set())
  const [batchAction, setBatchAction] = useState<{ type: 'publish' | 'unpublish' | 'delete'; status?: number } | null>(null)
  const [deleteId, setDeleteId] = useState<number | null>(null)
  const keywordRef = useRef(keyword)
  const pageRef = useRef(page)

  useEffect(() => { keywordRef.current = keyword }, [keyword])
  useEffect(() => { pageRef.current = page }, [page])

  const load = useCallback(async (p = pageRef.current) => {
    setError('')
    try {
      const res = await adminApi.listVideos({ keyword: keywordRef.current, page: p, page_size: 20 })
      setItems(res.data.list || [])
      setTotal(res.data.total || 0)
      setPage(res.data.page || p)
      setSelected(new Set())
    } catch (err) {
      setError(errorMessage(err))
    }
  }, [])

  useEffect(() => { void load(1) }, [load])

  function toggleSelect(id: number) {
    setSelected((prev) => {
      const next = new Set(prev)
      if (next.has(id)) next.delete(id)
      else next.add(id)
      return next
    })
  }

  function toggleSelectAll() {
    setSelected((prev) => {
      if (prev.size === items.length) return new Set()
      return new Set(items.map((i) => i.id))
    })
  }

  async function confirmBatch() {
    if (!batchAction || selected.size === 0) return
    try {
      if (batchAction.type === 'delete') {
        await adminApi.batchDeleteVideos(Array.from(selected))
        toast.success(`已删除 ${selected.size} 条影视`)
      } else if (batchAction.status !== undefined) {
        await adminApi.batchUpdatePublishStatus(Array.from(selected), batchAction.status)
        toast.success(`已${batchAction.status === 1 ? '上架' : '下架'} ${selected.size} 条影视`)
      }
      await load(page)
    } catch (err) {
      toast.error(errorMessage(err))
    } finally {
      setBatchAction(null)
    }
  }

  async function confirmDelete() {
    if (deleteId === null) return
    try {
      await adminApi.deleteVideo(deleteId)
      toast.success('影视已删除')
      await load(page)
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
            <CardTitle>影视管理</CardTitle>
            <Button size="sm" render={<Link to="/content/videos/new" />}>
              <Plus data-icon="inline-start" />
              新增影视
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          {error && (
            <Alert variant="destructive" className="mb-4">
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}
          <div className="mb-4 flex gap-2">
            <Input
              placeholder="关键词搜索"
              value={keyword}
              onChange={(e) => setKeyword(e.target.value)}
              className="max-w-xs"
              onKeyDown={(e) => { if (e.key === 'Enter') void load(1) }}
            />
            <Button variant="outline" size="sm" onClick={() => void load(1)}>
              <Search data-icon="inline-start" />
              搜索
            </Button>
          </div>

          {selected.size > 0 && (
            <div className="mb-4 flex items-center gap-2 rounded-md border bg-muted/50 p-2">
              <span className="text-sm text-muted-foreground">已选 {selected.size} 项</span>
              <Button size="sm" variant="outline" onClick={() => setBatchAction({ type: 'publish', status: 1 })}>
                批量上架
              </Button>
              <Button size="sm" variant="outline" onClick={() => setBatchAction({ type: 'unpublish', status: 0 })}>
                批量下架
              </Button>
              <Button size="sm" variant="destructive" onClick={() => setBatchAction({ type: 'delete' })}>
                批量删除
              </Button>
            </div>
          )}

          <div className="rounded-md border">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="w-12">
                    <Checkbox
                      checked={selected.size === items.length && items.length > 0}
                      onCheckedChange={toggleSelectAll}
                    />
                  </TableHead>
                  <TableHead className="w-16">ID</TableHead>
                  <TableHead>标题</TableHead>
                  <TableHead className="w-20">年份</TableHead>
                  <TableHead className="w-20">评分</TableHead>
                  <TableHead className="w-20">状态</TableHead>
                  <TableHead className="w-32">操作</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {items.map((item) => (
                  <TableRow key={item.id}>
                    <TableCell>
                      <Checkbox
                        checked={selected.has(item.id)}
                        onCheckedChange={() => toggleSelect(item.id)}
                      />
                    </TableCell>
                    <TableCell>{item.id}</TableCell>
                    <TableCell className="font-medium">{item.title}</TableCell>
                    <TableCell>{item.year || '-'}</TableCell>
                    <TableCell>{item.rating}</TableCell>
                    <TableCell>
                      <StatusBadge status={item.publish_status ?? 0} activeText="上架" inactiveText="下架" />
                    </TableCell>
                    <TableCell>
                      <div className="flex gap-1">
                        <Button size="sm" variant="ghost" render={<Link to={`/content/videos/${item.id}`} />}>
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

          <div className="mt-4 flex items-center justify-between">
            <span className="text-sm text-muted-foreground">共 {total} 条 · 第 {page} 页</span>
            <div className="flex gap-2">
              <Button size="sm" variant="outline" disabled={page <= 1} onClick={() => void load(page - 1)}>
                <ChevronLeft data-icon="inline-start" />
                上一页
              </Button>
              <Button size="sm" variant="outline" disabled={items.length < 20} onClick={() => void load(page + 1)}>
                下一页
                <ChevronRight data-icon="inline-end" />
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>

      <ConfirmDialog
        open={batchAction !== null}
        onOpenChange={(open) => { if (!open) setBatchAction(null) }}
        title="批量操作确认"
        description={
          batchAction?.type === 'delete'
            ? `确认批量删除 ${selected.size} 条影视？此操作不可撤销。`
            : `确认批量${batchAction?.status === 1 ? '上架' : '下架'} ${selected.size} 条影视？`
        }
        destructive={batchAction?.type === 'delete'}
        onConfirm={confirmBatch}
      />

      <ConfirmDialog
        open={deleteId !== null}
        onOpenChange={(open) => { if (!open) setDeleteId(null) }}
        title="删除影视"
        description="确认删除该影视？此操作不可撤销。"
        destructive
        onConfirm={confirmDelete}
      />
    </PageContainer>
  )
}
