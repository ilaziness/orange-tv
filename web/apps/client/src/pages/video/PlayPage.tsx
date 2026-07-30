import { useEffect, useState, useCallback, useRef } from 'react'
import { useParams, useNavigate } from 'react-router'
import type { ClientVideoDetail, PlayEpisodeResponse, VideoDetailSourceGroup } from '@orange-tv/shared'
import { clientApi, errorMessage } from '@/lib/api'
import { useAuth } from '@/hooks/useAuth'
import { useSiteStore } from '@/store/site'
import { VideoPlayer } from '@/components/Player'
import { saveHistory } from '@/lib/playbackHistory'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Empty, EmptyDescription, EmptyHeader, EmptyTitle } from '@/components/ui/empty'
import { Skeleton } from '@/components/ui/skeleton'
import { AlertCircleIcon } from 'lucide-react'

export default function PlayPage() {
  const { id, sourceId, episodeId } = useParams()
  const navigate = useNavigate()
  const ad = useSiteStore((s) => s.ad)
  const { profile } = useAuth()
  const [detail, setDetail] = useState<ClientVideoDetail | null>(null)
  const [episode, setEpisode] = useState<PlayEpisodeResponse | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [resumeAt, setResumeAt] = useState<number | undefined>(undefined)
  const remoteSyncInFlight = useRef(false)

  const sourceIdNum = Number(sourceId || 0)
  const epIdNum = Number(episodeId || 0)
  const videoIdNum = Number(id || 0)

  const handleProgress = useCallback((currentTime: number, duration: number) => {
    if (!detail?.title) return
    // 本地写入始终执行（同步 localStorage，无并发问题）
    saveHistory({
      videoId: videoIdNum,
      sourceId: sourceIdNum,
      episodeId: epIdNum,
      progress: currentTime,
      title: detail.title,
      updatedAt: Date.now(),
    })
    // 远端同步：已登录才发；上一个请求未返回时跳过本次，避免请求堆积
    if (!profile) return
    if (remoteSyncInFlight.current) return
    remoteSyncInFlight.current = true
    clientApi
      .upsertHistory({
        video_id: videoIdNum,
        play_source_id: sourceIdNum,
        episode_id: epIdNum,
        progress: currentTime,
        duration,
      })
      .catch((err) => {
        console.warn('sync play history to remote failed:', err)
      })
      .finally(() => {
        remoteSyncInFlight.current = false
      })
  }, [detail?.title, videoIdNum, sourceIdNum, epIdNum, profile])

  useEffect(() => {
    if (!id || !sourceId || !episodeId) return
    void (async () => {
      setLoading(true)
      setError('')
      setResumeAt(undefined)
      try {
        const [detailRes, epRes] = await Promise.all([
          clientApi.video(Number(id)),
          clientApi.playEpisode(Number(id), Number(sourceId), Number(episodeId)),
        ])
        setDetail(detailRes.data || null)
        setEpisode(epRes.data || null)
        // 已登录则并行拉取单条历史，用于恢复进度（失败不影响主流程）
        if (profile) {
          clientApi
            .getHistory(Number(id))
            .then((res) => {
              if (res.data && res.data.progress > 0) {
                setResumeAt(res.data.progress)
              }
            })
            .catch(() => undefined)
        }
      } catch (err) {
        setError(errorMessage(err))
      } finally {
        setLoading(false)
      }
    })()
  }, [id, sourceId, episodeId, profile])

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
  const playlist = sourceGroup?.episodes?.map((ep) => ({ episodeId: ep.id, title: ep.title || `第${ep.episode}集` }))
  // 从详情中查找当前剧集的集数（epIdNum 是数据库主键，不是集数）
  const currentEpNumber = detail.sources?.flatMap((s) => s.episodes).find((e) => e.id === epIdNum)?.episode

  return (
    <div className="flex flex-col gap-6">
      <div className="flex flex-col gap-2">
        <p className="text-sm text-muted-foreground">
          正在播放：{detail.title}{currentEpNumber ? ` - 第${currentEpNumber}集` : ''}
        </p>
        <div className="overflow-hidden rounded-xl border">
          <VideoPlayer
            src={episode.url}
            format={episode.format}
            videoId={videoIdNum}
            sourceId={sourceIdNum}
            episodeId={epIdNum}
            resumeAt={resumeAt}
            adConfig={ad.enabled ? ad : null}
            playlist={playlist}
            currentEpisodeId={epIdNum}
            onEpisodeChange={(epId) => navigate(`/play/${id}/${sourceIdNum}/${epId}`)}
            onProgress={handleProgress}
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
                        key={ep.id}
                        variant={source.id === sourceIdNum && ep.id === epIdNum ? 'default' : 'outline'}
                        size="sm"
                        onClick={() => navigate(`/play/${id}/${source.id}/${ep.id}`)}
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
