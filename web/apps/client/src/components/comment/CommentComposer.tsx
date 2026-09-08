import { useState } from 'react'
import { clientApi, errorMessage } from '@/lib/api'
import { Button } from '@/components/ui/button'
import { Textarea } from '@/components/ui/textarea'
import { Field, FieldGroup, FieldLabel } from '@/components/ui/field'
import { Spinner } from '@/components/ui/spinner'
import { toast } from 'sonner'

type CommentComposerProps = {
  videoId: number
  idPrefix?: string
  showLabel?: boolean
  onSuccess?: () => void
}

export function CommentComposer({
  videoId,
  idPrefix = 'comment',
  showLabel = true,
  onSuccess,
}: CommentComposerProps) {
  const [text, setText] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const fieldId = `${idPrefix}-content`

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    const content = text.trim()
    if (!content || content.length > 200) {
      toast.error('评论内容不能为空且最多 200 字')
      return
    }
    setSubmitting(true)
    try {
      await clientApi.createComment(videoId, content)
      setText('')
      toast.success('评论发表成功')
      onSuccess?.()
    } catch (err) {
      toast.error(errorMessage(err))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <form onSubmit={handleSubmit}>
      <FieldGroup>
        <Field>
          {showLabel ? <FieldLabel htmlFor={fieldId}>发表评论</FieldLabel> : null}
          <Textarea
            id={fieldId}
            placeholder="写下你的评论（最多 200 字）"
            maxLength={200}
            value={text}
            onChange={(e) => setText(e.target.value)}
            className="min-h-20"
            aria-label={showLabel ? undefined : '发表评论'}
          />
        </Field>
        <div>
          <Button type="submit" disabled={submitting || !text.trim()}>
            {submitting ? (
              <>
                <Spinner data-icon="inline-start" />
                发表中...
              </>
            ) : (
              '发表'
            )}
          </Button>
        </div>
      </FieldGroup>
    </form>
  )
}
