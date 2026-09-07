import { useCallback, useEffect, useRef, useState } from 'react'
import type { MediaAsset } from '@orange-tv/shared'
import { adminApi, errorMessage } from '@/lib/api'
import { ConfirmDialog, PageContainer, Pagination } from '@/components/shared'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardAction,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyTitle,
} from '@/components/ui/empty'
import { Skeleton } from '@/components/ui/skeleton'
import { Spinner } from '@/components/ui/spinner'
import { Separator } from '@/components/ui/separator'
import { Copy, Trash2, Upload } from 'lucide-react'
import { toast } from 'sonner'

const PAGE_SIZE = 24
const MAX_IMAGE_BYTES = 5 * 1024 * 1024
const ACCEPT_TYPES = new Set(['image/jpeg', 'image/png', 'image/webp', 'image/gif'])

function formatSize(n: number): string {
  if (n < 1024) return `${n} B`
  if (n < 1024 * 1024) return `${(n / 1024).toFixed(1)} KB`
  return `${(n / (1024 * 1024)).toFixed(1)} MB`
}

export default function MediaLibraryPage() {
  const [items, setItems] = useState<MediaAsset[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [loading, setLoading] = useState(false)
  const [uploading, setUploading] = useState(false)
  const [deleteId, setDeleteId] = useState<number | null>(null)
  const [deleting, setDeleting] = useState(false)
  const fileRef = useRef<HTMLInputElement>(null)
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
    void load(1)
  }, [load])

  async function onUpload(file: File | undefined) {
    if (!file || uploading) return
    if (file.size > MAX_IMAGE_BYTES) {
      toast.error('文件过大（最大 5MB）')
      if (fileRef.current) fileRef.current.value = ''
      return
    }
    if (file.type && !ACCEPT_TYPES.has(file.type)) {
      toast.error('仅支持 jpeg/png/webp/gif')
      if (fileRef.current) fileRef.current.value = ''
      return
    }
    setUploading(true)
    try {
      await adminApi.uploadMedia(file)
      toast.success('上传成功')
      await load(1)
    } catch (err) {
      toast.error(errorMessage(err))
    } finally {
      setUploading(false)
      if (fileRef.current) fileRef.current.value = ''
    }
  }

  async function confirmDelete() {
    if (deleteId == null || deleting) return
    setDeleting(true)
    try {
      await adminApi.deleteMedia(deleteId)
      toast.success('已删除')
      setDeleteId(null)
      const nextTotal = Math.max(0, total - 1)
      const maxPage = Math.max(1, Math.ceil(nextTotal / PAGE_SIZE))
      await load(Math.min(pageRef.current, maxPage))
    } catch (err) {
      toast.error(errorMessage(err))
    } finally {
      setDeleting(false)
    }
  }

  async function copyURL(url: string) {
    try {
      await navigator.clipboard.writeText(url)
      toast.success('已复制 URL')
    } catch {
      toast.error('复制失败')
    }
  }

  const hasNext = page * PAGE_SIZE < total

  return (
    <PageContainer>
      <Card>
        <CardHeader>
          <CardTitle>媒体库</CardTitle>
          <CardDescription>
            单文件上传图片（jpeg/png/webp/gif，最大 5MB）。删除不会清理业务表中已引用的 URL。
          </CardDescription>
          <CardAction>
            <input
              ref={fileRef}
              type="file"
              accept="image/jpeg,image/png,image/webp,image/gif,.jpg,.jpeg,.png,.webp,.gif"
              className="hidden"
              onChange={(e) => void onUpload(e.target.files?.[0])}
            />
            <Button
              type="button"
              disabled={uploading}
              className="cursor-pointer"
              onClick={() => fileRef.current?.click()}
            >
              {uploading ? <Spinner data-icon="inline-start" /> : <Upload data-icon="inline-start" />}
              {uploading ? '上传中...' : '上传图片'}
            </Button>
          </CardAction>
        </CardHeader>
        <CardContent className="flex flex-col gap-4">
          {loading ? (
            <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6">
              {Array.from({ length: 12 }).map((_, i) => (
                <Skeleton key={i} className="aspect-square w-full rounded-lg" />
              ))}
            </div>
          ) : items.length === 0 ? (
            <Empty>
              <EmptyHeader>
                <EmptyTitle>暂无媒体</EmptyTitle>
                <EmptyDescription>请先配置云存储，然后上传图片</EmptyDescription>
              </EmptyHeader>
            </Empty>
          ) : (
            <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6">
              {items.map((item) => (
                <div key={item.id} className="flex flex-col overflow-hidden rounded-lg border">
                  <img
                    src={item.url}
                    alt={item.original_name || `media-${item.id}`}
                    className="aspect-square w-full bg-muted object-cover"
                    loading="lazy"
                  />
                  <Separator />
                  <div className="flex flex-col gap-2 bg-muted/40 p-2">
                    <div className="truncate text-xs text-muted-foreground">
                      {item.original_name || item.url}
                    </div>
                    <div className="text-xs text-muted-foreground">{formatSize(item.size)}</div>
                    <div className="flex gap-1">
                      <Button
                        type="button"
                        size="icon-sm"
                        variant="outline"
                        className="cursor-pointer"
                        aria-label="复制 URL"
                        onClick={() => void copyURL(item.url)}
                      >
                        <Copy />
                      </Button>
                      <Button
                        type="button"
                        size="icon-sm"
                        variant="destructive"
                        className="cursor-pointer"
                        aria-label="删除"
                        onClick={() => setDeleteId(item.id)}
                      >
                        <Trash2 />
                      </Button>
                    </div>
                  </div>
                </div>
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
              onLast={() => void load(Math.max(1, Math.ceil(total / PAGE_SIZE)))}
            />
          )}
        </CardContent>
      </Card>

      <ConfirmDialog
        open={deleteId != null}
        onOpenChange={(open) => {
          if (!open && !deleting) setDeleteId(null)
        }}
        title="删除媒体"
        description="确定删除该文件？已填入业务表单的 URL 不会自动清空。"
        confirmText="删除"
        destructive
        loading={deleting}
        onConfirm={() => void confirmDelete()}
      />
    </PageContainer>
  )
}
