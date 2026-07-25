import { useCallback, useEffect, useRef, useState } from 'react'
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
import type { VideoListItem } from '@orange-tv/shared'

const PAGE_SIZE = 10

interface VideoPickerDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onSelect: (video: { id: number; title: string }) => void
}

export function VideoPickerDialog({ open, onOpenChange, onSelect }: VideoPickerDialogProps) {
  const [items, setItems] = useState<VideoListItem[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [keyword, setKeyword] = useState('')
  const [loading, setLoading] = useState(false)
  const keywordRef = useRef(keyword)
  const pageRef = useRef(page)

  useEffect(() => { keywordRef.current = keyword }, [keyword])
  useEffect(() => { pageRef.current = page }, [page])

  const load = useCallback(async (p = pageRef.current, k = keywordRef.current) => {
    setLoading(true)
    try {
      const res = await adminApi.listVideos({ keyword: k, page: p, page_size: PAGE_SIZE })
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
      void load(1, '')
    }
  }, [open, load])

  function handleSelect(video: VideoListItem) {
    onSelect({ id: video.id, title: video.title })
    onOpenChange(false)
  }

  const hasNext = page * PAGE_SIZE < total

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle>选择影视</DialogTitle>
          <DialogDescription className="sr-only">
            搜索并选择关联的影视
          </DialogDescription>
        </DialogHeader>

        <div className="mb-4 flex gap-2">
          <Input
            placeholder="搜索影视标题"
            value={keyword}
            onChange={(e) => setKeyword(e.target.value)}
            className="max-w-xs"
            onKeyDown={(e) => { if (e.key === 'Enter') void load(1) }}
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
                  <TableHead>标题</TableHead>
                  <TableHead className="w-24">操作</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {items.map((v) => (
                  <TableRow key={v.id}>
                    <TableCell>{v.id}</TableCell>
                    <TableCell className="font-medium">{v.title}</TableCell>
                    <TableCell>
                      <Button size="sm" variant="outline" onClick={() => handleSelect(v)}>
                        选择
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
              <EmptyTitle>暂无数据</EmptyTitle>
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
            loading={loading}
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
