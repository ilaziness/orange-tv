import { useEffect, useState } from 'react'
import { Link } from 'react-router'
import type { CommentItem } from '@orange-tv/shared'
import { clientApi, errorMessage, getToken } from '@/lib/api'
import { cn } from '@/lib/utils'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { Button } from '@/components/ui/button'
import { Textarea } from '@/components/ui/textarea'
import { Field, FieldGroup, FieldLabel } from '@/components/ui/field'
import { Empty, EmptyDescription, EmptyHeader, EmptyTitle } from '@/components/ui/empty'
import { Loader2, ThumbsDown, ThumbsUp } from 'lucide-react'
import { toast } from 'sonner'

type CommentSectionProps = {
  videoId: number
  comments: CommentItem[]
  onRefresh: () => void
}

type CommentNodeProps = {
  comment: CommentItem
  videoId: number
  depth?: number
}

function CommentNode({ comment, videoId, depth = 0 }: CommentNodeProps) {
  const [data, setData] = useState<CommentItem>(comment)
  const [replies, setReplies] = useState<CommentItem[]>(comment.replies ?? [])
  const [replyCount, setReplyCount] = useState(comment.reply_count ?? 0)
  const [expanded, setExpanded] = useState(false)
  const [loading, setLoading] = useState(false)
  const [repliesPage, setRepliesPage] = useState(1)
  const [hasMore, setHasMore] = useState(false)
  const [showReplyInput, setShowReplyInput] = useState(false)
  const [replyText, setReplyText] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const isLoggedIn = !!getToken()

  useEffect(() => {
    setData(comment)
    setReplyCount(comment.reply_count ?? 0)
    if (comment.replies && comment.replies.length > 0) {
      setReplies(comment.replies)
      setExpanded(true)
    }
  }, [comment])

  const handleVote = async (action: 'like' | 'dislike' | 'cancel') => {
    if (!isLoggedIn) {
      toast.error('请登录后再操作')
      return
    }
    try {
      const res = await clientApi.voteComment(data.id, action)
      setData((prev) => ({
        ...prev,
        like_count: res.data.like_count,
        dislike_count: res.data.dislike_count,
        my_vote: res.data.my_vote,
      }))
    } catch (err) {
      toast.error(errorMessage(err))
    }
  }

  const loadReplies = async (page: number) => {
    setLoading(true)
    try {
      const res = await clientApi.listReplies(data.id, page)
      const list = res.data.list || []
      setReplies((prev) => (page === 1 ? list : [...prev, ...list]))
      setReplyCount(res.data.total ?? 0)
      setRepliesPage(res.data.page)
      setHasMore(res.data.page < res.data.total_pages)
      setExpanded(true)
    } catch (err) {
      toast.error(errorMessage(err))
    } finally {
      setLoading(false)
    }
  }

  const handleReplySubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    const content = replyText.trim()
    if (!content || content.length > 200) {
      toast.error('回复内容不能为空且最多 200 字')
      return
    }
    setSubmitting(true)
    try {
      const res = await clientApi.createComment(videoId, content, data.id)
      setReplies((prev) => [res.data, ...prev])
      setReplyCount((prev) => prev + 1)
      setHasMore(replyCount > replies.length)
      setReplyText('')
      setShowReplyInput(false)
      setExpanded(true)
      toast.success('回复发表成功')
    } catch (err) {
      toast.error(errorMessage(err))
    } finally {
      setSubmitting(false)
    }
  }

  const likeActive = data.my_vote === 1
  const dislikeActive = data.my_vote === -1

  return (
    <div
      className={cn(
        'flex gap-3',
        depth > 0 && 'mt-3 border-l-2 border-muted pl-4',
      )}
    >
      <Avatar size="sm">
        {data.avatar ? <AvatarImage src={data.avatar} /> : null}
        <AvatarFallback>
          {data.username?.[0]?.toUpperCase() || 'U'}
        </AvatarFallback>
      </Avatar>
      <div className="flex flex-1 flex-col gap-1">
        <div className="flex items-center gap-2">
          <span className="text-sm font-medium">{data.username}</span>
          <span className="text-xs text-muted-foreground">{data.created_at}</span>
        </div>
        <p className="text-sm">{data.content}</p>
        <div className="mt-1 flex flex-wrap items-center gap-2">
          <Button
            type="button"
            variant={likeActive ? 'default' : 'ghost'}
            size="xs"
            onClick={() => handleVote(likeActive ? 'cancel' : 'like')}
          >
            <ThumbsUp data-icon="inline-start" />
            {data.like_count}
          </Button>
          <Button
            type="button"
            variant={dislikeActive ? 'destructive' : 'ghost'}
            size="xs"
            onClick={() => handleVote(dislikeActive ? 'cancel' : 'dislike')}
          >
            <ThumbsDown data-icon="inline-start" />
            {data.dislike_count}
          </Button>
          {isLoggedIn && (
            <Button
              type="button"
              variant="ghost"
              size="xs"
              onClick={() => setShowReplyInput((v) => !v)}
            >
              回复
            </Button>
          )}
          {replyCount > 0 && !expanded && (
            <Button
              type="button"
              variant="link"
              size="xs"
              disabled={loading}
              onClick={() => loadReplies(1)}
            >
              {loading ? (
                <>
                  <Loader2 data-icon="inline-start" className="animate-spin" />
                  加载中...
                </>
              ) : (
                `${replyCount} 条回复`
              )}
            </Button>
          )}
        </div>

        {showReplyInput && (
          <form onSubmit={handleReplySubmit} className="mt-2 flex flex-col gap-2">
            <Textarea
              value={replyText}
              onChange={(e) => setReplyText(e.target.value)}
              maxLength={200}
              placeholder="写下你的回复（最多 200 字）"
              className="min-h-16"
            />
            <div className="flex gap-2">
              <Button
                type="submit"
                size="sm"
                disabled={submitting || !replyText.trim()}
              >
                {submitting ? '发表中...' : '回复'}
              </Button>
              <Button
                type="button"
                size="sm"
                variant="ghost"
                onClick={() => setShowReplyInput(false)}
              >
                取消
              </Button>
            </div>
          </form>
        )}

        {expanded && replies.length > 0 && (
          <div className="mt-2 flex flex-col gap-3">
            {replies.map((reply) => (
              <CommentNode
                key={reply.id}
                comment={reply}
                videoId={videoId}
                depth={depth + 1}
              />
            ))}
            {hasMore && (
              <Button
                type="button"
                variant="ghost"
                size="sm"
                onClick={() => loadReplies(repliesPage + 1)}
                disabled={loading}
              >
                {loading ? '加载中...' : '加载更多'}
              </Button>
            )}
          </div>
        )}
      </div>
    </div>
  )
}

export function CommentSection({
  videoId,
  comments,
  onRefresh,
}: CommentSectionProps) {
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
            <Link
              to="/login"
              className="text-primary underline-offset-4 hover:underline"
            >
              登录
            </Link>{' '}
            后发表评论
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
              <CommentNode key={c.id} comment={c} videoId={videoId} />
            ))
          )}
        </div>
      </CardContent>
    </Card>
  )
}
