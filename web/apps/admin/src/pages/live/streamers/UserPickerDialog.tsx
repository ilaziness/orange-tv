import { useCallback, useEffect, useRef, useState } from 'react'
import type { UserItem } from '@orange-tv/shared'
import { adminApi, errorMessage } from '@/lib/api'
import { Pagination } from '@/components/shared'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Spinner } from '@/components/ui/spinner'
import { Skeleton } from '@/components/ui/skeleton'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Search } from 'lucide-react'
import { Empty, EmptyHeader, EmptyTitle, EmptyDescription } from '@/components/ui/empty'
import { toast } from 'sonner'

const PAGE_SIZE = 10

interface UserPickerDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onGranted: () => void
}

export function UserPickerDialog({ open, onOpenChange, onGranted }: UserPickerDialogProps) {
  const [items, setItems] = useState<UserItem[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [keyword, setKeyword] = useState('')
  const [loading, setLoading] = useState(false)
  const [grantingId, setGrantingId] = useState<number | null>(null)
  const keywordRef = useRef(keyword)
  const pageRef = useRef(page)

  useEffect(() => {
    keywordRef.current = keyword
  }, [keyword])
  useEffect(() => {
    pageRef.current = page
  }, [page])

  const load = useCallback(async (p = pageRef.current, k = keywordRef.current) => {
    setLoading(true)
    try {
      const res = await adminApi.listUsers({
        keyword: k || undefined,
        status: 1,
        page: p,
        page_size: PAGE_SIZE,
      })
      setItems(res.data.list || [])
      setTotal(res.data.total || 0)
      setPage(p)
    } catch (err) {
      toast.error(errorMessage(err))
      setItems([])
      setTotal(0)
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    if (open) {
      setKeyword('')
      setGrantingId(null)
      void load(1, '')
    }
  }, [open, load])

  async function handleGrant(user: UserItem) {
    if (grantingId !== null) return
    setGrantingId(user.id)
    try {
      await adminApi.grantLiveStreamer(user.id)
      toast.success(`已将「${user.nickname || user.email}」设为主播`)
      onOpenChange(false)
      onGranted()
    } catch (err) {
      toast.error(errorMessage(err))
    } finally {
      setGrantingId(null)
    }
  }

  const hasNext = page * PAGE_SIZE < total
  const busy = grantingId !== null

  return (
    <Dialog
      open={open}
      onOpenChange={(next) => {
        if (!busy) onOpenChange(next)
      }}
    >
      <DialogContent className="sm:max-w-2xl" showCloseButton={!busy}>
        <DialogHeader>
          <DialogTitle>添加主播</DialogTitle>
          <DialogDescription className="sr-only">搜索用户并设为主播</DialogDescription>
        </DialogHeader>

        <div className="mb-4 flex gap-2">
          <Input
            placeholder="昵称/邮箱"
            value={keyword}
            onChange={(e) => setKeyword(e.target.value)}
            className="max-w-xs"
            onKeyDown={(e) => {
              if (e.key === 'Enter') void load(1)
            }}
            disabled={loading || busy}
          />
          <Button variant="outline" size="sm" onClick={() => void load(1)} disabled={loading || busy}>
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
                  <TableHead>用户</TableHead>
                  <TableHead className="w-28">操作</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {items.map((u) => (
                  <TableRow key={u.id}>
                    <TableCell>{u.id}</TableCell>
                    <TableCell>
                      <div className="font-medium">{u.nickname || '-'}</div>
                      <div className="text-xs text-muted-foreground">{u.email}</div>
                    </TableCell>
                    <TableCell>
                      <Button
                        size="sm"
                        variant="outline"
                        disabled={busy}
                        onClick={() => void handleGrant(u)}
                      >
                        {grantingId === u.id ? <Spinner data-icon="inline-start" /> : null}
                        设为主播
                      </Button>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        ) : (
          <Empty className="py-8">
            <EmptyHeader>
              <EmptyTitle>暂无用户</EmptyTitle>
              <EmptyDescription>请尝试搜索其他关键词</EmptyDescription>
            </EmptyHeader>
          </Empty>
        )}

        {items.length > 0 && (
          <Pagination
            page={page}
            total={total}
            pageSize={PAGE_SIZE}
            hasNext={hasNext}
            loading={loading || busy}
            onFirst={() => void load(1)}
            onPrev={() => void load(page - 1)}
            onNext={() => void load(page + 1)}
            onLast={() => void load(Math.ceil(total / PAGE_SIZE))}
          />
        )}
      </DialogContent>
    </Dialog>
  )
}
