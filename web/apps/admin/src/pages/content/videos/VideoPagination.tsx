import { Button } from '@/components/ui/button'
import { ChevronLeft, ChevronRight } from 'lucide-react'

interface VideoPaginationProps {
  page: number
  total: number
  hasNext: boolean
  onPrev: () => void
  onNext: () => void
}

export function VideoPagination({ page, total, hasNext, onPrev, onNext }: VideoPaginationProps) {
  return (
    <div className="mt-4 flex items-center justify-between">
      <span className="text-sm text-muted-foreground">共 {total} 条 · 第 {page} 页</span>
      <div className="flex gap-2">
        <Button size="sm" variant="outline" disabled={page <= 1} onClick={onPrev}>
          <ChevronLeft data-icon="inline-start" />
          上一页
        </Button>
        <Button size="sm" variant="outline" disabled={!hasNext} onClick={onNext}>
          下一页
          <ChevronRight data-icon="inline-end" />
        </Button>
      </div>
    </div>
  )
}
