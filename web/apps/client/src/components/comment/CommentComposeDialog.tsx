import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from '@/components/ui/dialog'
import { CommentComposer } from './CommentComposer'

type CommentComposeDialogProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  videoId: number
  onSuccess?: () => void
}

export function CommentComposeDialog({
  open,
  onOpenChange,
  videoId,
  onSuccess,
}: CommentComposeDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>发表评论</DialogTitle>
          <DialogDescription>分享你的观影感受（最多 200 字）</DialogDescription>
        </DialogHeader>
        <CommentComposer
          videoId={videoId}
          idPrefix="comment-dialog"
          showLabel={false}
          onSuccess={() => {
            onOpenChange(false)
            onSuccess?.()
          }}
        />
      </DialogContent>
    </Dialog>
  )
}
