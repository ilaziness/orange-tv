import { useEffect, useState } from 'react'
import type * as React from 'react'
import { adminApi, errorMessage } from '@/lib/api'
import { flattenCategories } from '@/lib/categories'
import { PageContainer, StatusBadge, ConfirmDialog } from '@/components/shared'
import type { Category } from '@orange-tv/shared'
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Button } from '@/components/ui/button'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Skeleton } from '@/components/ui/skeleton'
import { Empty } from '@/components/ui/empty'
import { RefreshCw, Plus, Trash2 } from 'lucide-react'
import { toast } from 'sonner'

export default function CategoriesPage() {
  const [tree, setTree] = useState<Category[]>([])
  const [name, setName] = useState('')
  const [parentId, setParentId] = useState('0')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const [deleteId, setDeleteId] = useState<number | null>(null)

  async function load() {
    setLoading(true)
    setError('')
    try {
      const res = await adminApi.listCategories()
      setTree(res.data || [])
    } catch (err) {
      setError(errorMessage(err))
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { void load() }, [])

  const flat = flattenCategories(tree)

  async function onCreate(e: React.SyntheticEvent<HTMLFormElement>) {
    e.preventDefault()
    setError('')
    try {
      await adminApi.createCategory({ name, parent_id: Number(parentId), status: 1 })
      setName('')
      toast.success('分类已创建')
      await load()
    } catch (err) {
      setError(errorMessage(err))
    }
  }

  async function confirmDelete() {
    if (deleteId === null) return
    try {
      await adminApi.deleteCategory(deleteId)
      toast.success('分类已删除')
      await load()
    } catch (err) {
      toast.error(errorMessage(err))
    } finally {
      setDeleteId(null)
    }
  }

  async function toggleStatus(item: Category) {
    try {
      await adminApi.updateCategory(item.id, { status: item.status === 1 ? 0 : 1 })
      toast.success(item.status === 1 ? '已禁用' : '已启用')
      await load()
    } catch (err) {
      toast.error(errorMessage(err))
    }
  }

  return (
    <PageContainer>
      <Card>
        <CardHeader>
          <CardTitle>分类管理</CardTitle>
        </CardHeader>
        <CardContent>
          {error && (
            <Alert variant="destructive" className="mb-4">
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}
          <form onSubmit={onCreate} className="mb-6 flex flex-wrap items-end gap-2">
            <Input
              placeholder="分类名称"
              value={name}
              onChange={(e) => setName(e.target.value)}
              required
              className="w-48"
            />
            <Select value={parentId} onValueChange={(v) => setParentId(v ?? '0')}>
              <SelectTrigger className="w-48">
                <SelectValue placeholder="选择父级" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="0">无父级</SelectItem>
                {flat.map((c) => (
                  <SelectItem key={c.id} value={String(c.id)}>
                    {'—'.repeat(c.depth)} {c.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <Button type="submit" size="sm">
              <Plus data-icon="inline-start" />
              新增分类
            </Button>
            <Button type="button" variant="outline" size="sm" onClick={() => void load()} disabled={loading}>
              <RefreshCw data-icon="inline-start" />
              刷新
            </Button>
          </form>

          {loading ? (
            <div className="flex flex-col gap-2">
              {Array.from({ length: 5 }).map((_, i) => (
                <Skeleton key={i} className="h-12 w-full" />
              ))}
            </div>
          ) : flat.length === 0 ? (
            <Empty className="py-8 text-center text-sm text-muted-foreground">暂无分类</Empty>
          ) : (
            <ScrollArea className="h-[600px] pr-4">
              <div className="flex flex-col gap-1">
                {flat.map((item) => (
                  <div
                    key={item.id}
                    className="flex items-center justify-between rounded-md border p-3"
                    style={{ marginLeft: item.depth * 20 }}
                  >
                    <div>
                      <span className="font-medium">{item.name}</span>
                      <span className="ml-2 text-xs text-muted-foreground">
                        ID {item.id} · 排序 {item.sort_order}
                      </span>
                    </div>
                    <div className="flex items-center gap-2">
                      <StatusBadge status={item.status} />
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => void toggleStatus(item)}
                      >
                        {item.status === 1 ? '禁用' : '启用'}
                      </Button>
                      <Button
                        variant="destructive"
                        size="sm"
                        onClick={() => setDeleteId(item.id)}
                      >
                        <Trash2 data-icon="inline-start" />
                        删除
                      </Button>
                    </div>
                  </div>
                ))}
              </div>
            </ScrollArea>
          )}
        </CardContent>
      </Card>

      <ConfirmDialog
        open={deleteId !== null}
        onOpenChange={(open) => { if (!open) setDeleteId(null) }}
        title="删除分类"
        description="确认删除该分类？此操作不可撤销。"
        destructive
        onConfirm={confirmDelete}
      />
    </PageContainer>
  )
}
