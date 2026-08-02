import { useState } from 'react'
import { Link } from 'react-router'
import type { CommentItem } from '@orange-tv/shared'
import { clientApi, errorMessage, getToken } from '@/lib/api'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { Button } from '@/components/ui/button'
import { Textarea } from '@/components/ui/textarea'
import { Field, FieldGroup, FieldLabel } from '@/components/ui/field'
import { Empty, EmptyDescription, EmptyHeader, EmptyTitle } from '@/components/ui/empty'
import { toast } from 'sonner'

type CommentSectionProps = {
  videoId: number
  comments: CommentItem[]
  onRefresh: () => void
}

export function CommentSection({ videoId, comments, onRefresh }: CommentSectionProps) {
  const [text, setText] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const isLoggedIn = !!getToken()

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
      onRefresh()
      toast.success('评论发表成功')
    } catch (err) {
      toast.error(errorMessage(err))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>评论 ({comments.length})</CardTitle>
      </CardHeader>
      <CardContent className="flex flex-col gap-4">
        {isLoggedIn ? (
          <form onSubmit={handleSubmit}>
            <FieldGroup>
              <Field>
                <FieldLabel htmlFor="comment">发表评论</FieldLabel>
                <Textarea
                  id="comment"
                  placeholder="写下你的评论（最多 200 字）"
                  maxLength={200}
                  value={text}
                  onChange={(e) => setText(e.target.value)}
                  className="min-h-20"
                />
              </Field>
              <div>
                <Button type="submit" disabled={submitting || !text.trim()}>
                  {submitting ? '发表中...' : '发表'}
                </Button>
              </div>
            </FieldGroup>
          </form>
        ) : (
          <p className="text-sm text-muted-foreground">
            <Link to="/login" className="text-primary underline-offset-4 hover:underline">登录</Link>
            {' '}后发表评论
          </p>
        )}

        <div className="flex flex-col gap-4">
          {comments.length === 0 ? (
            <Empty>
              <EmptyHeader>
                <EmptyTitle>暂无评论</EmptyTitle>
                <EmptyDescription>快来发表第一条评论吧</EmptyDescription>
              </EmptyHeader>
            </Empty>
          ) : (
            comments.map((c) => (
              <div key={c.id} className="flex gap-3">
                <Avatar size="sm">
                  {c.avatar ? <AvatarImage src={c.avatar} /> : null}
                  <AvatarFallback>{c.username?.[0]?.toUpperCase() || 'U'}</AvatarFallback>
                </Avatar>
                <div className="flex flex-1 flex-col gap-1">
                  <div className="flex items-center gap-2">
                    <span className="text-sm font-medium">{c.username}</span>
                    <span className="text-xs text-muted-foreground">
                      {c.created_at}
                    </span>
                  </div>
                  <p className="text-sm">{c.content}</p>
                </div>
              </div>
            ))
          )}
        </div>
      </CardContent>
    </Card>
  )
}
