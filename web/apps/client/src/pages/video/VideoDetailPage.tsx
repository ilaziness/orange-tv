import { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router'
import type { CommentItem, ClientVideoDetail, VideoListItem } from '@orange-tv/shared'
import { clientApi, errorMessage } from '@/lib/api'
import { VideoGrid } from '@/components/common'
import { CommentSection } from '@/components/CommentSection'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Empty, EmptyDescription, EmptyHeader, EmptyTitle } from '@/components/ui/empty'
import { Skeleton } from '@/components/ui/skeleton'
import { AlertCircleIcon, FilmIcon } from 'lucide-react'

export default function VideoDetailPage() {
  const { id } = useParams()
  const navigate = useNavigate()
  const [detail, setDetail] = useState<ClientVideoDetail | null>(null)
  const [related, setRelated] = useState<VideoListItem[]>([])
  const [comments, setComments] = useState<CommentItem[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [posterError, setPosterError] = useState(false)

  const loadComments = () => {
    if (!id) return
    void clientApi.listComments(Number(id), 1).then((res) => setComments(res.data.list || []))
  }

  useEffect(() => {
    if (!id) return
    void (async () => {
      setLoading(true)
      try {
        const [res, rel, com] = await Promise.all([
          clientApi.video(Number(id)),
          clientApi.related(Number(id), 6),
          clientApi.listComments(Number(id), 1),
        ])
        setDetail(res.data || null)
        setRelated(rel.data || [])
        setComments(com.data.list || [])
      } catch (err) {
        setError(errorMessage(err))
      } finally {
        setLoading(false)
      }
    })()
  }, [id])

  if (loading) {
    return (
      <div className="flex flex-col gap-6">
        <div className="flex gap-6">
          <Skeleton className="aspect-[2/3] w-48 shrink-0 rounded-xl" />
          <div className="flex flex-1 flex-col gap-3">
            <Skeleton className="h-8 w-2/3" />
            <Skeleton className="h-4 w-1/3" />
            <Skeleton className="h-20 w-full" />
          </div>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <Alert variant="destructive">
        <AlertCircleIcon />
        <AlertTitle>加载失败</AlertTitle>
        <AlertDescription>{error}</AlertDescription>
      </Alert>
    )
  }

  if (!detail) {
    return (
      <Empty>
        <EmptyHeader>
          <EmptyTitle>影视不存在</EmptyTitle>
          <EmptyDescription>该视频可能已下架</EmptyDescription>
        </EmptyHeader>
      </Empty>
    )
  }

  const directorText = detail.directors?.length
    ? detail.directors.map((d) => d.name).join(' / ')
    : '暂无'
  const actorText = detail.actors?.length
    ? detail.actors.map((a) => a.name).join(' / ')
    : '暂无'

  return (
    <div className="flex flex-col gap-8">
      <div className="flex flex-col gap-6 md:flex-row">
        <div className="relative flex aspect-[2/3] w-40 shrink-0 items-center justify-center overflow-hidden rounded-xl bg-muted shadow-lg md:w-56">
          {detail.poster && !posterError ? (
            <img
              src={detail.poster}
              alt={detail.title}
              className="absolute inset-0 h-full w-full object-cover"
              onError={() => setPosterError(true)}
            />
          ) : (
            <FilmIcon className="size-16 text-muted-foreground/40" />
          )}
        </div>
        <div className="flex max-w-4xl flex-1 flex-col gap-3">
          <h1 className="text-2xl font-bold tracking-tight">{detail.title}</h1>
          {detail.subtitle ? (
            <p className="text-muted-foreground">{detail.subtitle}</p>
          ) : null}
          <div className="flex flex-wrap gap-2">
            <Badge variant="secondary">评分: {detail.rating?.toFixed(1) || 'N/A'}</Badge>
            <Badge variant="secondary">{detail.year || '未知'}</Badge>
            <Badge variant="secondary">{detail.region || '未知'}</Badge>
            <Badge variant="secondary">{detail.language || '未知'}</Badge>
          </div>
          <div className="flex flex-col gap-1 text-sm">
            <p>
              <span className="text-muted-foreground">导演: </span>
              {directorText}
            </p>
            <p>
              <span className="text-muted-foreground">主演: </span>
              {actorText}
            </p>
          </div>
          <p className="text-sm leading-relaxed">{detail.description || '暂无简介'}</p>
          {detail.tags?.length ? (
            <div className="flex flex-wrap gap-2">
              {detail.tags.map((t) => (
                <Badge key={t.id} variant="outline">{t.name}</Badge>
              ))}
            </div>
          ) : null}
        </div>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>播放源</CardTitle>
        </CardHeader>
        <CardContent>
          {detail.sources && detail.sources.length > 0 ? (
            <div className="flex flex-col gap-4">
              {detail.sources.map((source) => (
                <div key={source.id} className="flex flex-col gap-2">
                  <p className="text-sm font-medium">{source.name}</p>
                  <div className="flex flex-wrap gap-2">
                    {source.episodes.map((ep) => (
                      <Button
                        key={ep.id}
                        variant="outline"
                        size="sm"
                        onClick={() => navigate(`/play/${id}/${source.id}/${ep.id}`)}
                      >
                        {ep.title || `第${ep.episode}集`}
                      </Button>
                    ))}
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <Empty>
              <EmptyHeader>
                <EmptyTitle>暂无播放源</EmptyTitle>
              </EmptyHeader>
            </Empty>
          )}
        </CardContent>
      </Card>

      <CommentSection
        videoId={Number(id)}
        comments={comments}
        onRefresh={loadComments}
      />

      {related.length ? (
        <section className="flex flex-col gap-4">
          <h2 className="text-lg font-semibold">相关推荐</h2>
          <VideoGrid items={related} />
        </section>
      ) : null}
    </div>
  )
}
