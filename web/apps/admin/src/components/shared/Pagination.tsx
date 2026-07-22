import { Button } from '@/components/ui/button'
import { ChevronLeft, ChevronRight, ChevronsLeft, ChevronsRight } from 'lucide-react'

interface PaginationProps {
  page: number
  total: number
  pageSize?: number
  hasNext: boolean
  loading?: boolean
  onFirst: () => void
  onPrev: () => void
  onNext: () => void
  onLast: () => void
}

export function Pagination({ page, total, pageSize = 20, hasNext, loading, onFirst, onPrev, onNext, onLast }: PaginationProps) {
  const totalPages = Math.ceil(total / pageSize)
  return (
    <div className="mt-4 flex items-center justify-between">
      <span className="text-sm text-muted-foreground">
        {loading ? '加载中...' : `共 ${total} 条 · 第 ${page}/${totalPages || 1} 页`}
      </span>
      <div className="flex gap-2">
        <Button size="sm" variant="outline" disabled={loading || page <= 1} onClick={onFirst}>
          <ChevronsLeft data-icon="inline-start" />
          首页
        </Button>
        <Button size="sm" variant="outline" disabled={loading || page <= 1} onClick={onPrev}>
          <ChevronLeft data-icon="inline-start" />
          上一页
        </Button>
        <Button size="sm" variant="outline" disabled={loading || !hasNext} onClick={onNext}>
          下一页
          <ChevronRight data-icon="inline-end" />
        </Button>
        <Button size="sm" variant="outline" disabled={loading || !hasNext} onClick={onLast}>
          尾页
          <ChevronsRight data-icon="inline-end" />
        </Button>
      </div>
    </div>
  )
}
