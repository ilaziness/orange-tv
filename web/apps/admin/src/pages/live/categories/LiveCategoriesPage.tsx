import { useCallback, useEffect, useRef, useState, type SyntheticEvent } from 'react'
import type { AdminLiveCategoryItem } from '@orange-tv/shared'
import { z } from 'zod'
import { adminApi, errorMessage } from '@/lib/api'
import {
  PageContainer,
  Pagination,
  ConfirmDialog,
  StatusBadge,
} from '@/components/shared'
import { Card, CardAction, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Field, FieldGroup, FieldLabel } from '@/components/ui/field'
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Skeleton } from '@/components/ui/skeleton'
import { Spinner } from '@/components/ui/spinner'
import { Empty, EmptyHeader, EmptyTitle, EmptyDescription } from '@/components/ui/empty'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Pencil, Plus, RefreshCw, Search, Trash2 } from 'lucide-react'
import { toast } from 'sonner'
import { DEFAULT_PAGE_SIZE } from '@/lib/constants'

type StatusFilter = 'all' | '0' | '1'

type CategoryForm = {
  name: string
  sort: string
  status: '0' | '1'
}

const emptyForm = (): CategoryForm => ({ name: '', sort: '0', status: '1' })

const categorySchema = z.object({
  name: z.string().trim().min(1, '名称不能为空').max(50, '名称过长'),
  sort: z
    .string()
    .regex(/^\d+$/, '排序须为非负整数')
    .transform((v) => Number(v)),
  status: z.enum(['0', '1']),
})

export default function LiveCategoriesPage() {
  const [items, setItems] = useState<AdminLiveCategoryItem[]>([])
  const [keyword, setKeyword] = useState('')
  const [status, setStatus] = useState<StatusFilter>('all')
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(false)
  const [dialogOpen, setDialogOpen] = useState(false)
  const [editing, setEditing] = useState<AdminLiveCategoryItem | null>(null)
  const [form, setForm] = useState<CategoryForm>(emptyForm)
  const [submitting, setSubmitting] = useState(false)
  const [deleteId, setDeleteId] = useState<number | null>(null)
  const [deleting, setDeleting] = useState(false)
  const [togglingId, setTogglingId] = useState<number | null>(null)

  const keywordRef = useRef(keyword)
  const statusRef = useRef(status)
  const pageRef = useRef(page)

  useEffect(() => {
    keywordRef.current = keyword
  }, [keyword])
  useEffect(() => {
    statusRef.current = status
  }, [status])
  useEffect(() => {
    pageRef.current = page
  }, [page])

  const load = useCallback(async (p = pageRef.current) => {
    setLoading(true)
    try {
      const query: Record<string, string | number | undefined> = {
        keyword: keywordRef.current || undefined,
        page: p,
        page_size: DEFAULT_PAGE_SIZE,
      }
      if (statusRef.current !== 'all') {
        query.status = Number(statusRef.current)
      }
      const res = await adminApi.listLiveCategories(query)
      setItems(res.data.list || [])
      setTotal(res.data.total || 0)
      setPage(res.data.page || p)
    } catch (err) {
      toast.error(errorMessage(err))
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void load(1)
  }, [load])

  function openCreate() {
    setEditing(null)
    setForm(emptyForm())
    setDialogOpen(true)
  }

  function openEdit(item: AdminLiveCategoryItem) {
    setEditing(item)
    setForm({
      name: item.name,
      sort: String(item.sort ?? 0),
      status: item.status === 1 ? '1' : '0',
    })
    setDialogOpen(true)
  }

  async function onSubmit(e: SyntheticEvent<HTMLFormElement>) {
    e.preventDefault()
    if (submitting) return
    const parsed = categorySchema.safeParse(form)
    if (!parsed.success) {
      toast.error(parsed.error.issues[0]?.message || '表单校验失败')
      return
    }
    setSubmitting(true)
    try {
      const body = {
        name: parsed.data.name,
        sort: parsed.data.sort,
        status: Number(parsed.data.status) as 0 | 1,
      }
      if (editing) {
        await adminApi.updateLiveCategory(editing.id, body)
        toast.success('分类已更新')
      } else {
        await adminApi.createLiveCategory(body)
        toast.success('分类已创建')
      }
      setDialogOpen(false)
      await load(editing ? page : 1)
    } catch (err) {
      toast.error(errorMessage(err))
    } finally {
      setSubmitting(false)
    }
  }

  async function confirmDelete() {
    if (deleteId === null) return
    setDeleting(true)
    try {
      await adminApi.deleteLiveCategory(deleteId)
      toast.success('分类已删除')
      setDeleteId(null)
      await load(page)
    } catch (err) {
      toast.error(errorMessage(err))
    } finally {
      setDeleting(false)
    }
  }

  async function toggleStatus(item: AdminLiveCategoryItem) {
    const next = item.status === 1 ? 0 : 1
    setTogglingId(item.id)
    try {
      await adminApi.updateLiveCategory(item.id, { status: next })
      setItems((prev) => prev.map((it) => (it.id === item.id ? { ...it, status: next } : it)))
      toast.success(next === 1 ? '已启用' : '已禁用')
    } catch (err) {
      toast.error(errorMessage(err))
    } finally {
      setTogglingId(null)
    }
  }

  const statusOptions: { value: StatusFilter; label: string }[] = [
    { value: 'all', label: '全部状态' },
    { value: '1', label: '启用' },
    { value: '0', label: '禁用' },
  ]

  return (
    <PageContainer>
      <Card>
        <CardHeader>
          <CardTitle>直播分类</CardTitle>
          <CardAction>
            <div className="flex gap-2">
              <Button size="sm" variant="outline" onClick={() => void load(page)} disabled={loading}>
                {loading ? (
                  <Spinner data-icon="inline-start" />
                ) : (
                  <RefreshCw data-icon="inline-start" />
                )}
                刷新
              </Button>
              <Button size="sm" onClick={openCreate}>
                <Plus data-icon="inline-start" />
                新增分类
              </Button>
            </div>
          </CardAction>
        </CardHeader>
        <CardContent>
          <div className="mb-4 flex flex-wrap gap-2">
            <Input
              placeholder="分类名称"
              value={keyword}
              onChange={(e) => setKeyword(e.target.value)}
              className="max-w-xs"
              onKeyDown={(e) => {
                if (e.key === 'Enter') void load(1)
              }}
              disabled={loading}
            />
            <Select value={status} onValueChange={(v) => setStatus((v ?? 'all') as StatusFilter)}>
              <SelectTrigger className="w-32" disabled={loading}>
                <SelectValue placeholder="状态">
                  {statusOptions.find((o) => o.value === status)?.label}
                </SelectValue>
              </SelectTrigger>
              <SelectContent>
                <SelectGroup>
                  {statusOptions.map((o) => (
                    <SelectItem key={o.value} value={o.value}>
                      {o.label}
                    </SelectItem>
                  ))}
                </SelectGroup>
              </SelectContent>
            </Select>
            <Button variant="outline" size="sm" onClick={() => void load(1)} disabled={loading}>
              {loading ? <Spinner data-icon="inline-start" /> : <Search data-icon="inline-start" />}
              查询
            </Button>
          </div>

          {loading && items.length === 0 ? (
            <div className="flex flex-col gap-2">
              {Array.from({ length: 5 }).map((_, i) => (
                <Skeleton key={i} className="h-12 w-full" />
              ))}
            </div>
          ) : items.length > 0 ? (
            <>
              <div className="relative rounded-md border">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead className="w-16">ID</TableHead>
                      <TableHead>名称</TableHead>
                      <TableHead className="w-20">排序</TableHead>
                      <TableHead className="w-20">状态</TableHead>
                      <TableHead className="w-36">更新时间</TableHead>
                      <TableHead className="w-52">操作</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {items.map((item) => {
                      const toggling = togglingId === item.id
                      return (
                        <TableRow key={item.id}>
                          <TableCell>{item.id}</TableCell>
                          <TableCell>{item.name}</TableCell>
                          <TableCell>{item.sort}</TableCell>
                          <TableCell>
                            <StatusBadge status={item.status} />
                          </TableCell>
                          <TableCell className="text-muted-foreground">
                            {item.updated_at || '-'}
                          </TableCell>
                          <TableCell>
                            <div className="flex gap-1">
                              <Button size="sm" variant="ghost" onClick={() => openEdit(item)}>
                                <Pencil data-icon="inline-start" />
                                编辑
                              </Button>
                              <Button
                                size="sm"
                                variant="ghost"
                                disabled={toggling}
                                onClick={() => void toggleStatus(item)}
                              >
                                {toggling ? <Spinner data-icon="inline-start" /> : null}
                                {item.status === 1 ? '禁用' : '启用'}
                              </Button>
                              <Button
                                size="sm"
                                variant="ghost"
                                onClick={() => setDeleteId(item.id)}
                              >
                                <Trash2 data-icon="inline-start" />
                                删除
                              </Button>
                            </div>
                          </TableCell>
                        </TableRow>
                      )
                    })}
                  </TableBody>
                </Table>
                {loading && (
                  <div className="absolute inset-0 flex items-center justify-center bg-background/40">
                    <Spinner className="size-5" />
                  </div>
                )}
              </div>
              <Pagination
                page={page}
                total={total}
                pageSize={DEFAULT_PAGE_SIZE}
                loading={loading}
                hasNext={page * DEFAULT_PAGE_SIZE < total}
                onFirst={() => void load(1)}
                onPrev={() => void load(page - 1)}
                onNext={() => void load(page + 1)}
                onLast={() => void load(Math.ceil(total / DEFAULT_PAGE_SIZE))}
              />
            </>
          ) : (
            <Empty className="py-8">
              <EmptyHeader>
                <EmptyTitle>暂无分类</EmptyTitle>
                <EmptyDescription>点击右上角「新增分类」添加</EmptyDescription>
              </EmptyHeader>
            </Empty>
          )}
        </CardContent>
      </Card>

      <Dialog
        open={dialogOpen}
        onOpenChange={(open) => {
          if (!submitting) setDialogOpen(open)
        }}
      >
        <DialogContent className="sm:max-w-md" showCloseButton={!submitting}>
          <DialogHeader>
            <DialogTitle>{editing ? '编辑分类' : '新增分类'}</DialogTitle>
            <DialogDescription className="sr-only">填写分类信息后保存</DialogDescription>
          </DialogHeader>
          <form onSubmit={onSubmit} className="flex flex-col gap-5">
            <FieldGroup>
              <Field data-disabled={submitting ? true : undefined}>
                <FieldLabel htmlFor="live-cat-name">
                  名称<span className="ml-0.5 text-destructive">*</span>
                </FieldLabel>
                <Input
                  id="live-cat-name"
                  value={form.name}
                  onChange={(e) => setForm((prev) => ({ ...prev, name: e.target.value }))}
                  maxLength={50}
                  required
                  disabled={submitting}
                  autoFocus
                />
              </Field>
              <Field data-disabled={submitting ? true : undefined}>
                <FieldLabel htmlFor="live-cat-sort">排序</FieldLabel>
                <Input
                  id="live-cat-sort"
                  value={form.sort}
                  onChange={(e) =>
                    setForm((prev) => ({
                      ...prev,
                      sort: e.target.value.replace(/[^0-9]/g, ''),
                    }))
                  }
                  disabled={submitting}
                />
              </Field>
              <Field data-disabled={submitting ? true : undefined}>
                <FieldLabel>状态</FieldLabel>
                <Select
                  value={form.status}
                  onValueChange={(v) =>
                    setForm((prev) => ({ ...prev, status: (v === '0' ? '0' : '1') as '0' | '1' }))
                  }
                >
                  <SelectTrigger disabled={submitting}>
                    <SelectValue>{form.status === '1' ? '启用' : '禁用'}</SelectValue>
                  </SelectTrigger>
                  <SelectContent>
                    <SelectGroup>
                      <SelectItem value="1">启用</SelectItem>
                      <SelectItem value="0">禁用</SelectItem>
                    </SelectGroup>
                  </SelectContent>
                </Select>
              </Field>
            </FieldGroup>
            <DialogFooter>
              <Button
                type="button"
                variant="outline"
                onClick={() => setDialogOpen(false)}
                disabled={submitting}
              >
                取消
              </Button>
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
        onOpenChange={(open) => {
          if (!open && !deleting) setDeleteId(null)
        }}
        title="删除分类"
        description="确认删除该分类？已在播房间不受影响，新开播不可再选。"
        destructive
        loading={deleting}
        onConfirm={() => void confirmDelete()}
      />
    </PageContainer>
  )
}
