import { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router'
import type { VideoDetail } from '@orange-tv/shared'
import { clientApi, errorMessage } from '@/lib/api'
import { VideoPlayer } from '@/components/Player'
import { EpisodeList } from '@/components/EpisodeList'
import { Button } from '@/components/ui/button'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Empty, EmptyDescription, EmptyHeader, EmptyTitle } from '@/components/ui/empty'
import { Skeleton } from '@/components/ui/skeleton'
import { AlertCircleIcon } from 'lucide-react'

export default function PlayPage() {
  const { id, sourceIdx } = useParams()
  const navigate = useNavigate()
  const [detail, setDetail] = useState<VideoDetail | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const sourceIndex = Number(sourceIdx || 0)

  useEffect(() => {
    if (!id) return
    void (async () => {
      setLoading(true)
      try {
        const res = await clientApi.video(Number(id))
        setDetail(res.data || null)
      } catch (err) {
        setError(errorMessage(err))
      } finally {
        setLoading(false)
      }
    })()
  }, [id])

  if (loading) {
    return (
      <div className="flex flex-col gap-4">
        <Skeleton className="aspect-video w-full rounded-xl" />
        <Skeleton className="h-8 w-1/3" />
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

  const sourceGroup = detail.sources?.[sourceIndex]
  if (!sourceGroup) {
    return (
      <Empty>
        <EmptyHeader>
          <EmptyTitle>播放源不存在</EmptyTitle>
        </EmptyHeader>
      </Empty>
    )
  }
  const source = sourceGroup.episodes?.[0]
  if (!source) {
    return (
      <Empty>
        <EmptyHeader>
          <EmptyTitle>暂无剧集</EmptyTitle>
        </EmptyHeader>
      </Empty>
    )
  }

  return (
    <div className="flex flex-col gap-6">
      <div className="overflow-hidden rounded-xl border border-border">
        <VideoPlayer src={source.url} format={source.format} />
      </div>

      <div className="flex flex-col gap-3">
        <h1 className="text-xl font-bold">{detail.title}</h1>
        <p className="text-sm text-muted-foreground">
          {sourceGroup.name || `播放源 ${sourceIndex + 1}`}
        </p>

        <div className="flex flex-wrap gap-2">
          {detail.sources?.map((s, idx) => (
            <Button
              key={idx}
              variant={idx === sourceIndex ? 'default' : 'outline'}
              size="sm"
              onClick={() => navigate(`/play/${id}/${idx}`)}
            >
              {s.name || `播放源 ${idx + 1}`}
            </Button>
          ))}
        </div>

        {sourceGroup.episodes.length > 1 ? (
          <EpisodeList
            source={sourceGroup}
            currentEpisode={source.episode}
            onSelect={() => {}}
          />
        ) : null}
      </div>
    </div>
  )
}
