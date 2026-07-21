import { useCallback, useEffect, useRef, useState } from 'react'
import { z } from 'zod'
import { errorMessage } from '@/lib/api'
import { PageContainer, ConfirmDialog } from '@/components/shared'
import type { NamedItem } from '@orange-tv/shared'
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
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
import { Search, Plus, Trash2 } from 'lucide-react'
import { toast } from 'sonner'

const nameSchema = z.string().trim().min(1, '请输入名称')

export function NamedResourcePage({
  title,
  list,
  create,
  remove,
}: {
  title: string
  list: (keyword?: string) => Promise<{ data: { list: NamedItem[] } }>
  create: (name: string) => Promise<unknown>
  remove: (id: number) => Promise<unknown>
}) {
  const [items, setItems] = useState<NamedItem[]>([])
  const [keyword, setKeyword] = useState('')
  const [name, setName] = useState('')
  const [error, setError] = useState('')
  const [deleteId, setDeleteId] = useState<number | null>(null)
  const keywordRef = useRef(keyword)
  const listRef = useRef(list)

  useEffect(() => { keywordRef.current = keyword }, [keyword])
  useEffect(() => { listRef.current = list }, [list])

  const load = useCallback(async (k = keywordRef.current) => {
    setError('')
    try {
      const res = await listRef.current(k)
      setItems(res.data.list || [])
    } catch (err) {
      setError(errorMessage(err))
    }
  }, [])

  useEffect(() => { void load('') }, [load])

  async function handleCreate() {
    setError('')
    const result = nameSchema.safeParse(name)
    if (!result.success) {
      setError(result.error.issues[0]?.message || '请输入名称')
      return
    }
    try {
      await create(result.data)
      setName('')
      toast.success('创建成功')
      await load()
    } catch (err) {
      toast.error(errorMessage(err))
    }
  }

  async function confirmDelete() {
    if (deleteId === null) return
    try {
      await remove(deleteId)
      toast.success('删除成功')
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
          <CardTitle>{title}</CardTitle>
        </CardHeader>
        <CardContent>
          {error && (
            <Alert variant="destructive" className="mb-4">
              <AlertTitle>出错了</AlertTitle>
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}
          <div className="mb-4 flex flex-wrap gap-2">
            <Input
              placeholder="搜索"
              value={keyword}
              onChange={(e) => setKeyword(e.target.value)}
              className="max-w-xs"
              onKeyDown={(e) => { if (e.key === 'Enter') void load() }}
            />
            <Button variant="outline" size="sm" onClick={() => void load()}>
              <Search data-icon="inline-start" />
              查询
            </Button>
            <Input
              placeholder="新名称"
              value={name}
              onChange={(e) => setName(e.target.value)}
              className="max-w-xs"
              onKeyDown={(e) => { if (e.key === 'Enter') void handleCreate() }}
            />
            <Button size="sm" onClick={handleCreate}>
              <Plus data-icon="inline-start" />
              新增
            </Button>
          </div>
          <div className="rounded-md border">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="w-16">ID</TableHead>
                  <TableHead>名称</TableHead>
                  <TableHead className="w-24">操作</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {items.map((item) => (
                  <TableRow key={item.id}>
                    <TableCell>{item.id}</TableCell>
                    <TableCell className="font-medium">{item.name}</TableCell>
                    <TableCell>
                      <Button size="sm" variant="ghost" onClick={() => setDeleteId(item.id)}>
                        <Trash2 data-icon="inline-start" />
                        删除
                      </Button>
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
        title="删除确认"
        description="确认删除该项？此操作不可撤销。"
        destructive
        onConfirm={confirmDelete}
      />
    </PageContainer>
  )
}
