import { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router'
import type { ClientVideoDetail, PlayEpisodeResponse, VideoDetailSourceGroup } from '@orange-tv/shared'
import { clientApi, errorMessage } from '@/lib/api'
import { useSiteStore } from '@/store/site'
import { VideoPlayer } from '@/components/Player'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Empty, EmptyDescription, EmptyHeader, EmptyTitle } from '@/components/ui/empty'
import { Skeleton } from '@/components/ui/skeleton'
import { AlertCircleIcon } from 'lucide-react'

export default function PlayPage() {
  const { id, sourceId, episodeNumber } = useParams()
  const navigate = useNavigate()
  const ad = useSiteStore((s) => s.ad)
  const [detail, setDetail] = useState<ClientVideoDetail | null>(null)
  const [episode, setEpisode] = useState<PlayEpisodeResponse | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  const sourceIdNum = Number(sourceId || 0)
  const epNum = Number(episodeNumber || 0)

  useEffect(() => {
    if (!id || !sourceId || !episodeNumber) return
    void (async () => {
      setLoading(true)
      setError('')
      try {
        const [detailRes, epRes] = await Promise.all([
          clientApi.video(Number(id)),
          clientApi.playEpisode(Number(id), Number(sourceId), Number(episodeNumber)),
        ])
        setDetail(detailRes.data || null)
        setEpisode(epRes.data || null)
      } catch (err) {
        setError(errorMessage(err))
      } finally {
        setLoading(false)
      }
    })()
  }, [id, sourceId, episodeNumber])

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

  if (!episode) {
    return (
      <Empty>
        <EmptyHeader>
          <EmptyTitle>剧集不存在</EmptyTitle>
          <EmptyDescription>该剧集可能已下架</EmptyDescription>
        </EmptyHeader>
      </Empty>
    )
  }

  const sourceGroup: VideoDetailSourceGroup | undefined = detail.sources?.find(
    (s) => s.id === sourceIdNum,
  )

  return (
    <div className="flex flex-col gap-6">
      <div className="flex flex-col gap-2">
        <p className="text-sm text-muted-foreground">
          正在播放：{detail.title} - 第{epNum}集
        </p>
        <div className="overflow-hidden rounded-xl border">
          <VideoPlayer
            src={episode.url}
            format={episode.format}
            adConfig={ad.enabled ? ad : null}
            playlist={sourceGroup?.episodes}
            currentEpisode={epNum}
            onEpisodeChange={(ep) => navigate(`/play/${id}/${sourceIdNum}/${ep}`)}
          />
        </div>
      </div>

      <div className="flex flex-col gap-3">
        <h1 className="text-xl font-bold">{detail.title}</h1>

        {detail.sources && detail.sources.length > 0 ? (
          <div className="flex flex-col gap-4">
            {detail.sources.map((source) => (
              <Card key={source.id}>
                <CardHeader>
                  <CardTitle>{source.name}</CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="flex flex-wrap gap-2">
                    {source.episodes.map((ep) => (
                      <Button
                        key={ep.episode}
                        variant={source.id === sourceIdNum && ep.episode === epNum ? 'default' : 'outline'}
                        size="sm"
                        onClick={() => navigate(`/play/${id}/${source.id}/${ep.episode}`)}
                      >
                        {ep.title || `第${ep.episode}集`}
                      </Button>
                    ))}
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        ) : null}
      </div>
    </div>
  )
}
