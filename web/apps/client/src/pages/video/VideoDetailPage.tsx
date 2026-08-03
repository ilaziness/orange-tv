import { useState } from 'react'
import { useNavigate, useParams, useLoaderData } from 'react-router'
import type { CommentItem, ClientVideoDetail, ClientVideoListItem } from '@orange-tv/shared'
import { clientApi, errorMessage } from '@/lib/api'
import { VideoGrid, FavoriteButton, RatingStars } from '@/components/common'
import { CommentSection } from '@/components/CommentSection'
import { useSettings } from '@/hooks/useSettings'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Empty, EmptyDescription, EmptyHeader, EmptyTitle } from '@/components/ui/empty'
import { AlertCircleIcon, FilmIcon } from 'lucide-react'
import { usePageTitle } from '@/hooks/usePageTitle'

type VideoDetailLoaderData = {
  detail: ClientVideoDetail | null
  related: ClientVideoListItem[]
  comments: CommentItem[]
  error: string
}

export async function loader({ params }: { params: Record<string, string | undefined> }): Promise<VideoDetailLoaderData> {
  const id = params.id
  if (!id) {
    return { detail: null, related: [], comments: [], error: '' }
  }
  try {
    const [res, rel, com] = await Promise.all([
      clientApi.video(Number(id)),
      clientApi.related(Number(id), 8),
      clientApi.listComments(Number(id), 1),
    ])
    return {
      detail: res.data || null,
      related: rel.data || [],
      comments: com.data.list || [],
      error: '',
    }
  } catch (err) {
    return { detail: null, related: [], comments: [], error: errorMessage(err) }
  }
}

export function Component() {
  const { id } = useParams()
  const data = useLoaderData<VideoDetailLoaderData>()
  const navigate = useNavigate()
  const { detail, related, comments: initialComments, error } = data
  const [comments, setComments] = useState<CommentItem[]>(initialComments)
  const [posterError, setPosterError] = useState(false)
  const { feature } = useSettings()

  usePageTitle(detail ? detail.title : '影视详情')

  const loadComments = () => {
    if (!id) return
    void clientApi
      .listComments(Number(id), 1)
      .then((res) => setComments(res.data.list || []))
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
  const tagsText = detail.tags?.length
    ? detail.tags.map((t) => t.name).join(' / ')
    : ''

  return (
    <div className="flex flex-col gap-8">
      <div className="flex flex-col gap-6 md:flex-row">
        <div className="relative flex aspect-[2/3] w-40 shrink-0 items-center justify-center overflow-hidden rounded-xl bg-muted shadow-lg md:w-56">
          {detail.cover && !posterError ? (
            <img
              src={detail.cover}
              alt={detail.title}
              className="absolute inset-0 h-full w-full object-cover"
              onError={() => setPosterError(true)}
            />
          ) : (
            <FilmIcon className="size-16 text-muted-foreground/40" />
          )}
        </div>
        <div className="flex max-w-4xl flex-1 flex-col gap-3">
          <div className="flex items-center gap-3">
            <h1 className="text-2xl font-bold tracking-tight">
              {detail.title}
            </h1>
            <FavoriteButton videoId={Number(id)} />
          </div>
          <div className="flex flex-wrap items-center gap-2">
            {detail.year ? (
              <Badge variant="secondary">{detail.year}</Badge>
            ) : null}
            {detail.region ? (
              <Badge variant="secondary">{detail.region}</Badge>
            ) : null}
            {detail.language ? (
              <Badge variant="secondary">{detail.language}</Badge>
            ) : null}
          </div>
          <RatingStars
            videoId={Number(id)}
            rating={detail.rating}
            ratingCount={detail.rating_count}
          />
          <div className="flex flex-col gap-1 text-sm">
            {detail.subtitle ? (
              <p>
                <span className="text-muted-foreground">别名: </span>
                {detail.subtitle}
              </p>
            ) : null}
            {tagsText ? (
              <p>
                <span className="text-muted-foreground">类型: </span>
                {tagsText}
              </p>
            ) : null}
            <p>
              <span className="text-muted-foreground">导演: </span>
              {directorText}
            </p>
            <p>
              <span className="text-muted-foreground">主演: </span>
              {actorText}
            </p>
          </div>
          <p className="text-sm leading-relaxed">
            <span className="text-muted-foreground">剧情简介：</span>
            {detail.description || '暂无简介'}
          </p>
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
                        onClick={() =>
                          navigate(`/play/${id}/${source.id}/${ep.id}`)
                        }
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

      {feature.comment_enabled ? (
        <CommentSection
          videoId={Number(id)}
          comments={comments}
          onRefresh={loadComments}
        />
      ) : null}

      {related.length ? (
        <section className="flex flex-col gap-4">
          <h2 className="text-lg font-semibold">相关推荐</h2>
          <VideoGrid items={related} />
        </section>
      ) : null}
    </div>
  )
}
