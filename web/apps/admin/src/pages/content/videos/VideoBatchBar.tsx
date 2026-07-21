import { Button } from '@/components/ui/button'

interface VideoBatchBarProps {
  count: number
  loading?: boolean
  onPublish: () => void
  onUnpublish: () => void
  onDelete: () => void
}

export function VideoBatchBar({ count, loading, onPublish, onUnpublish, onDelete }: VideoBatchBarProps) {
  if (count === 0) return null
  return (
    <div className="mb-4 flex items-center gap-2 rounded-md border bg-muted/50 p-2">
      <span className="text-sm text-muted-foreground">已选 {count} 项</span>
      <Button size="sm" variant="outline" disabled={loading} onClick={onPublish}>
        批量上架
      </Button>
      <Button size="sm" variant="outline" disabled={loading} onClick={onUnpublish}>
        批量下架
      </Button>
      <Button size="sm" variant="destructive" disabled={loading} onClick={onDelete}>
        批量删除
      </Button>
    </div>
  )
}
