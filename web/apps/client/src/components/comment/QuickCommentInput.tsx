import { useState } from 'react'
import { useAuth } from '@/hooks/useAuth'
import { useLoginDialogStore } from '@/store/loginDialog'
import { InputGroup, InputGroupAddon, InputGroupInput } from '@/components/ui/input-group'
import { MessageSquareIcon } from 'lucide-react'
import { CommentComposeDialog } from './CommentComposeDialog'

type QuickCommentInputProps = {
  videoId: number
  onSuccess?: () => void
}

export function QuickCommentInput({ videoId, onSuccess }: QuickCommentInputProps) {
  const [dialogOpen, setDialogOpen] = useState(false)
  const { profile } = useAuth()
  const openLoginDialog = useLoginDialogStore((s) => s.open)

  const openComposer = () => {
    if (!profile) {
      openLoginDialog()
      return
    }
    setDialogOpen(true)
  }

  return (
    <>
      <InputGroup className="h-9 w-48 cursor-pointer sm:w-56" onClick={openComposer}>
        <InputGroupAddon>
          <MessageSquareIcon />
        </InputGroupAddon>
        <InputGroupInput
          readOnly
          placeholder="说点什么…"
          className="cursor-pointer"
          onKeyDown={(e) => {
            if (e.key === 'Enter' || e.key === ' ') {
              e.preventDefault()
              openComposer()
            }
          }}
        />
      </InputGroup>
      <CommentComposeDialog
        open={dialogOpen}
        onOpenChange={setDialogOpen}
        videoId={videoId}
        onSuccess={onSuccess}
      />
    </>
  )
}
