import { useCallback, useEffect, useRef, useState } from 'react'
import { Link } from 'react-router'
import { z } from 'zod'
import { errorMessage } from '@/lib/api'
import { PageContainer, ConfirmDialog, Pagination } from '@/components/shared'
import { NamedResourceDialog } from './NamedResourceDialog'
import type { NamedItem, PageData } from '@orange-tv/shared'
import { Card, CardAction, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Spinner } from '@/components/ui/spinner'
import { Skeleton } from '@/components/ui/skeleton'
import { Empty, EmptyHeader, EmptyTitle, EmptyDescription } from '@/components/ui/empty'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Search, Plus, Pencil, Film, Trash2, RefreshCw } from 'lucide-react'
import { toast } from 'sonner'
import { DEFAULT_PAGE_SIZE } from '@/lib/constants'

const nameSchema = z.string().trim().min(1, '请输入名称')

type ResourceType = 'director' | 'actor' | 'tag'

export function NamedResourcePage({
  title,
  resourceType,
  list,
  create,
  update,
  remove,
}: {
  title: string
  resourceType: ResourceType
  list: (query?: {
    keyword?: string
    page?: number
    page_size?: number
  }) => Promise<{ data: PageData<NamedItem> }>
  create: (name: string) => Promise<unknown>
  update: (id: number, name: string) => Promise<unknown>
  remove: (id: number) => Promise<unknown>
}) {
  const [items, setItems] = useState<NamedItem[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [keyword, setKeyword] = useState('')
  const [loading, setLoading] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [deleting, setDeleting] = useState(false)
  const [dialogOpen, setDialogOpen] = useState(false)
  const [editId, setEditId] = useState(0)
  const [dialogName, setDialogName] = useState('')
  const [dialogError, setDialogError] = useState('')
  const [fieldError, setFieldError] = useState('')
  const [deleteId, setDeleteId] = useState<number | null>(null)
  const keywordRef = useRef(keyword)
  const pageRef = useRef(page)
  const listRef = useRef(list)

  useEffect(() => {
    keywordRef.current = keyword
  }, [keyword])
  useEffect(() => {
    pageRef.current = page
  }, [page])
  useEffect(() => {
    listRef.current = list
  }, [list])

  const load = useCallback(async (p = pageRef.current, k = keywordRef.current) => {
    setLoading(true)
    try {
      const res = await listRef.current({ keyword: k, page: p, page_size: DEFAULT_PAGE_SIZE })
      setItems(res.data.list || [])
      setTotal(res.data.total || 0)
      setPage(p)
    } catch (err) {
      toast.error(errorMessage(err))
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void load(1, '')
  }, [load])

  function openCreate() {
    setEditId(0)
    setDialogName('')
    setDialogError('')
    setFieldError('')
    setDialogOpen(true)
  }

  function openEdit(item: NamedItem) {
    setEditId(item.id)
    setDialogName(item.name)
    setDialogError('')
    setFieldError('')
    setDialogOpen(true)
  }

  async function onSubmit(e: React.SyntheticEvent<HTMLFormElement>) {
    e.preventDefault()
    if (submitting) return
    setDialogError('')
    setFieldError('')
    const result = nameSchema.safeParse(dialogName)
    if (!result.success) {
      setFieldError(result.error.issues[0]?.message || '请输入名称')
      return
    }
    setSubmitting(true)
    try {
      if (editId) {
        await update(editId, result.data)
        toast.success('更新成功')
      } else {
        await create(result.data)
        toast.success('创建成功')
      }
      setDialogOpen(false)
      await load(editId ? page : 1)
    } catch (err) {
      setDialogError(errorMessage(err))
    } finally {
      setSubmitting(false)
    }
  }

  async function confirmDelete() {
    if (deleteId === null) return
    setDeleting(true)
    try {
      await remove(deleteId)
      toast.success('删除成功')
      await load(page)
    } catch (err) {
      toast.error(errorMessage(err))
    } finally {
      setDeleting(false)
      setDeleteId(null)
    }
  }

  const hasNext = page * DEFAULT_PAGE_SIZE < total

  return (
    <PageContainer>
      <Card>
        <CardHeader>
          <CardTitle>{title}</CardTitle>
          <CardAction>
            <div className="flex gap-2">
              <Button
                size="sm"
                variant="outline"
                onClick={() => void load(page)}
                disabled={loading}
              >
                {loading ? (
                  <Spinner data-icon="inline-start" />
                ) : (
                  <RefreshCw data-icon="inline-start" />
                )}
                刷新
              </Button>
              <Button size="sm" onClick={openCreate}>
                <Plus data-icon="inline-start" />
                新增
              </Button>
            </div>
          </CardAction>
        </CardHeader>
        <CardContent>
          <div className="mb-4 flex flex-wrap gap-2">
            <Input
              placeholder="搜索"
              value={keyword}
              onChange={(e) => setKeyword(e.target.value)}
              className="max-w-xs"
              onKeyDown={(e) => {
                if (e.key === 'Enter') void load(1)
              }}
            />
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
                {loading && (
                  <div className="absolute inset-0 z-10 flex items-center justify-center bg-background/50">
                    <Spinner />
                  </div>
                )}
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead className="w-16">ID</TableHead>
                      <TableHead>名称</TableHead>
                      <TableHead className="w-64">操作</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {items.map((item) => (
                      <TableRow key={item.id}>
                        <TableCell>{item.id}</TableCell>
                        <TableCell className="font-medium">{item.name}</TableCell>
                        <TableCell>
                          <div className="flex gap-1">
                            <Button size="sm" variant="ghost" onClick={() => openEdit(item)}>
                              <Pencil data-icon="inline-start" />
                              编辑
                            </Button>
                            <Button
                              size="sm"
                              variant="ghost"
                              render={<Link to={`/content/videos?${resourceType}_id=${item.id}`} />}
                            >
                              <Film data-icon="inline-start" />
                              查看影视
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
              <Pagination
                page={page}
                total={total}
                pageSize={DEFAULT_PAGE_SIZE}
                hasNext={hasNext}
                loading={loading}
                onFirst={() => void load(1)}
                onPrev={() => void load(page - 1)}
                onNext={() => void load(page + 1)}
                onLast={() => void load(Math.ceil(total / DEFAULT_PAGE_SIZE))}
              />
            </>
          ) : (
            <Empty className="py-8">
              <EmptyHeader>
                <EmptyTitle>暂无数据</EmptyTitle>
                <EmptyDescription>点击右上角「新增」添加第一条数据</EmptyDescription>
              </EmptyHeader>
            </Empty>
          )}
        </CardContent>
      </Card>

      <NamedResourceDialog
        open={dialogOpen}
        onOpenChange={setDialogOpen}
        title={editId ? '编辑' : '新增'}
        name={dialogName}
        onNameChange={setDialogName}
        submitting={submitting}
        error={dialogError}
        fieldError={fieldError}
        onSubmit={onSubmit}
      />

      <ConfirmDialog
        open={deleteId !== null}
        onOpenChange={(open) => {
          if (!open && !deleting) setDeleteId(null)
        }}
        title="删除确认"
        description="确认删除该项？此操作不可撤销。"
        destructive
        loading={deleting}
        onConfirm={confirmDelete}
      />
    </PageContainer>
  )
}
