import { useEffect, useMemo, useState } from 'react'
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
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Field, FieldGroup, FieldLabel } from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Spinner } from '@/components/ui/spinner'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Skeleton } from '@/components/ui/skeleton'
import { Empty } from '@/components/ui/empty'
import { Edit, Plus, RefreshCw, Trash2 } from 'lucide-react'
import { toast } from 'sonner'

type CategoryForm = {
  name: string
  parentId: string
  sortOrder: string
  status: string
}

const emptyForm: CategoryForm = { name: '', parentId: '0', sortOrder: '0', status: '1' }
const statusOptions = [
  { value: '1', label: '启用' },
  { value: '0', label: '禁用' },
]

function categoryAndDescendantIDs(categories: Category[], id: number): Set<number> {
  const excluded = new Set([id])

  function collect(items: Category[]) {
    for (const item of items) {
      excluded.add(item.id)
      if (item.children) collect(item.children)
    }
  }

  function find(items: Category[]): boolean {
    for (const item of items) {
      if (item.id === id) {
        if (item.children) collect(item.children)
        return true
      }
      if (item.children && find(item.children)) return true
    }
    return false
  }

  find(categories)
  return excluded
}

export default function CategoriesPage() {
  const [tree, setTree] = useState<Category[]>([])
  const [form, setForm] = useState<CategoryForm>(emptyForm)
  const [editing, setEditing] = useState<Category | null>(null)
  const [dialogOpen, setDialogOpen] = useState(false)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [updatingId, setUpdatingId] = useState<number | null>(null)
  const [deleting, setDeleting] = useState(false)
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
  const excludedParentIDs = useMemo(
    () => editing ? categoryAndDescendantIDs(tree, editing.id) : new Set<number>(),
    [editing, tree],
  )
  const parentOptions = useMemo(() => [
    { value: '0', label: '无父级' },
    ...flat
      .filter((item) => !excludedParentIDs.has(item.id))
      .map((item) => ({ value: String(item.id), label: `${'—'.repeat(item.depth)} ${item.name}` })),
  ], [excludedParentIDs, flat])

  function openCreate() {
    setError('')
    setEditing(null)
    setForm(emptyForm)
    setDialogOpen(true)
  }

  function openEdit(item: Category) {
    setError('')
    setEditing(item)
    setForm({
      name: item.name,
      parentId: String(item.parent_id),
      sortOrder: String(item.sort_order),
      status: String(item.status),
    })
    setDialogOpen(true)
  }

  function closeDialog(open: boolean) {
    if (submitting) return
    setDialogOpen(open)
    if (!open) setEditing(null)
  }

  async function onSubmit(e: React.SyntheticEvent<HTMLFormElement>) {
    e.preventDefault()
    const name = form.name.trim()
    const parentID = Number(form.parentId)
    const sortOrder = Number(form.sortOrder)
    const status = Number(form.status)

    if (!name || name.length > 100) {
      toast.error('分类名称长度应为 1 至 100 个字符')
      return
    }
    if (!Number.isSafeInteger(parentID) || parentID < 0) {
      toast.error('请选择有效的父级分类')
      return
    }
    if (!Number.isSafeInteger(sortOrder) || sortOrder < 0 || sortOrder > 4294967295) {
      toast.error('排序必须为 0 至 4294967295 的整数')
      return
    }
    if (status !== 0 && status !== 1) {
      toast.error('请选择有效状态')
      return
    }

    setSubmitting(true)
    setError('')
    try {
      const payload = { name, parent_id: parentID, sort_order: sortOrder, status }
      if (editing) {
        await adminApi.updateCategory(editing.id, payload)
        toast.success('分类已更新')
      } else {
        await adminApi.createCategory(payload)
        toast.success('分类已创建')
      }
      setDialogOpen(false)
      setEditing(null)
      setForm(emptyForm)
      await load()
    } catch (err) {
      setError(errorMessage(err))
    } finally {
      setSubmitting(false)
    }
  }

  async function confirmDelete() {
    if (deleteId === null) return
    setDeleting(true)
    try {
      await adminApi.deleteCategory(deleteId)
      toast.success('分类已删除')
      await load()
    } catch (err) {
      toast.error(errorMessage(err))
    } finally {
      setDeleting(false)
      setDeleteId(null)
    }
  }

  async function toggleStatus(item: Category) {
    setUpdatingId(item.id)
    try {
      await adminApi.updateCategory(item.id, { status: item.status === 1 ? 0 : 1 })
      toast.success(item.status === 1 ? '已禁用' : '已启用')
      await load()
    } catch (err) {
      toast.error(errorMessage(err))
    } finally {
      setUpdatingId(null)
    }
  }

  return (
    <PageContainer>
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle>分类管理</CardTitle>
            <div className="flex gap-2">
              <Button type="button" size="sm" onClick={openCreate} disabled={loading}>
                <Plus data-icon="inline-start" />
                新增分类
              </Button>
              <Button type="button" variant="outline" size="sm" onClick={() => void load()} disabled={loading}>
                {loading ? <Spinner data-icon="inline-start" /> : <RefreshCw data-icon="inline-start" />}
                刷新
              </Button>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          {error && (
            <Alert variant="destructive" className="mb-4">
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}

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
                      <Button variant="outline" size="sm" onClick={() => openEdit(item)} disabled={updatingId === item.id}>
                        <Edit data-icon="inline-start" />
                        编辑
                      </Button>
                      <Button variant="outline" size="sm" onClick={() => void toggleStatus(item)} disabled={updatingId === item.id}>
                        {updatingId === item.id && <Spinner data-icon="inline-start" />}
                        {item.status === 1 ? '禁用' : '启用'}
                      </Button>
                      <Button variant="destructive" size="sm" onClick={() => setDeleteId(item.id)} disabled={updatingId === item.id}>
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

      <Dialog open={dialogOpen} onOpenChange={closeDialog}>
        <DialogContent className="sm:max-w-md" showCloseButton={!submitting}>
          <DialogHeader>
            <DialogTitle>{editing ? '编辑分类信息' : '新增分类信息'}</DialogTitle>
          </DialogHeader>
          <form onSubmit={onSubmit} className="flex flex-col gap-5">
            {error && (
              <Alert variant="destructive">
                <AlertDescription>{error}</AlertDescription>
              </Alert>
            )}
            <FieldGroup>
              <Field>
                <FieldLabel htmlFor="category-name">分类名称</FieldLabel>
                <Input id="category-name" value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} maxLength={100} required disabled={submitting} />
              </Field>
              <Field>
                <FieldLabel htmlFor="category-parent">父级分类</FieldLabel>
                <Select items={parentOptions} value={form.parentId} onValueChange={(value) => setForm({ ...form, parentId: value ?? '0' })} disabled={submitting}>
                  <SelectTrigger id="category-parent">
                    <SelectValue placeholder="请选择父级分类" />
                  </SelectTrigger>
                  <SelectContent>
                    {parentOptions.map((option) => <SelectItem key={option.value} value={option.value}>{option.label}</SelectItem>)}
                  </SelectContent>
                </Select>
              </Field>
              <Field>
                <FieldLabel htmlFor="category-sort-order">排序</FieldLabel>
                <Input id="category-sort-order" type="number" min="0" max="4294967295" step="1" value={form.sortOrder} onChange={(e) => setForm({ ...form, sortOrder: e.target.value })} required disabled={submitting} />
              </Field>
              <Field>
                <FieldLabel htmlFor="category-status">状态</FieldLabel>
                <Select items={statusOptions} value={form.status} onValueChange={(value) => setForm({ ...form, status: value ?? '1' })} disabled={submitting}>
                  <SelectTrigger id="category-status">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {statusOptions.map((option) => <SelectItem key={option.value} value={option.value}>{option.label}</SelectItem>)}
                  </SelectContent>
                </Select>
              </Field>
            </FieldGroup>
            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => closeDialog(false)} disabled={submitting}>取消</Button>
              <Button type="submit" disabled={submitting}>
                {submitting && <Spinner data-icon="inline-start" />}
                {submitting ? '保存中...' : '保存'}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      <ConfirmDialog
        open={deleteId !== null}
        onOpenChange={(open) => { if (!open && !deleting) setDeleteId(null) }}
        title="删除分类"
        description="确认删除该分类？此操作不可撤销。"
        destructive
        loading={deleting}
        onConfirm={confirmDelete}
      />
    </PageContainer>
  )
}
