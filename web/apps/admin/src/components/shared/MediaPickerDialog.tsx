import { useCallback, useEffect, useRef, useState } from 'react'
import type { MediaAsset } from '@orange-tv/shared'
import { adminApi, errorMessage } from '@/lib/api'
import { Pagination } from '@/components/shared'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Skeleton } from '@/components/ui/skeleton'
import { Separator } from '@/components/ui/separator'
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyTitle,
} from '@/components/ui/empty'
import { toast } from 'sonner'

const PAGE_SIZE = 12

interface MediaPickerDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onSelect: (media: MediaAsset) => void
}

export function MediaPickerDialog({ open, onOpenChange, onSelect }: MediaPickerDialogProps) {
  const [items, setItems] = useState<MediaAsset[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [loading, setLoading] = useState(false)
  const pageRef = useRef(page)

  useEffect(() => {
    pageRef.current = page
  }, [page])

  const load = useCallback(async (p = pageRef.current) => {
    setLoading(true)
    try {
      const res = await adminApi.listMedia({
        media_type: 'image',
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
      void load(1)
    }
  }, [open, load])

  function handleSelect(media: MediaAsset) {
    onSelect(media)
    onOpenChange(false)
  }

  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE))
  const hasNext = page < totalPages

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-3xl">
        <DialogHeader>
          <DialogTitle>从媒体库选择</DialogTitle>
          <DialogDescription>选择一张图片，URL 将填入输入框（不会在此上传）</DialogDescription>
        </DialogHeader>

        {loading ? (
          <div className="grid grid-cols-3 gap-3 sm:grid-cols-4">
            {Array.from({ length: 8 }).map((_, i) => (
              <Skeleton key={i} className="aspect-square w-full rounded-lg" />
            ))}
          </div>
        ) : items.length === 0 ? (
          <Empty>
            <EmptyHeader>
              <EmptyTitle>媒体库为空</EmptyTitle>
              <EmptyDescription>请先到「媒体库」菜单上传图片</EmptyDescription>
            </EmptyHeader>
          </Empty>
        ) : (
          <div className="grid max-h-[60vh] grid-cols-3 gap-3 overflow-y-auto sm:grid-cols-4">
            {items.map((item) => (
              <button
                key={item.id}
                type="button"
                className="cursor-pointer overflow-hidden rounded-lg border text-left transition-colors hover:border-primary"
                onClick={() => handleSelect(item)}
              >
                <img
                  src={item.url}
                  alt={item.original_name || `media-${item.id}`}
                  className="aspect-square w-full bg-muted object-cover"
                  loading="lazy"
                />
                <Separator />
                <div className="truncate px-2 py-1 text-xs text-muted-foreground">
                  {item.original_name || item.url}
                </div>
              </button>
            ))}
          </div>
        )}

        {total > 0 && (
          <Pagination
            page={page}
            total={total}
            pageSize={PAGE_SIZE}
            hasNext={hasNext}
            loading={loading}
            onFirst={() => void load(1)}
            onPrev={() => void load(Math.max(1, page - 1))}
            onNext={() => void load(page + 1)}
            onLast={() => void load(totalPages)}
          />
        )}
      </DialogContent>
    </Dialog>
  )
}
