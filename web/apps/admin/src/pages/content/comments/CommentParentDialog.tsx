import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Spinner } from '@/components/ui/spinner'
import type { AdminCommentParentItem } from '@orange-tv/shared'

interface CommentParentDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  parents: AdminCommentParentItem[]
  loading?: boolean
}

export function CommentParentDialog({
  open,
  onOpenChange,
  parents,
  loading,
}: CommentParentDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>父级评论</DialogTitle>
          <DialogDescription>从根评论到直接父评论的回复链</DialogDescription>
        </DialogHeader>
        {loading ? (
          <div className="py-8 text-center">
            <Spinner className="mx-auto" />
            <span className="mt-2 block text-sm text-muted-foreground">加载中...</span>
          </div>
        ) : parents.length === 0 ? (
          <div className="py-8 text-center text-sm text-muted-foreground">暂无父级评论</div>
        ) : (
          <div className="max-h-[60vh] space-y-3 overflow-y-auto">
            {parents.map((p, idx) => (
              <div key={p.id} className="rounded-lg border p-3">
                <div className="mb-1 flex items-center justify-between text-sm text-muted-foreground">
                  <span>
                    #{idx + 1} 用户: {p.nickname || '-'} (ID:{p.user_id})
                  </span>
                  <span>
                    ID: {p.id} {p.parent_id ? `→ 父ID: ${p.parent_id}` : ''}
                  </span>
                </div>
                <p className="text-sm">{p.content}</p>
                <div className="mt-1 text-xs text-muted-foreground">{p.created_at || '-'}</div>
              </div>
            ))}
          </div>
        )}
      </DialogContent>
    </Dialog>
  )
}
