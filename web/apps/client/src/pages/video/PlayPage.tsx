import { useCallback, useEffect, useRef, useState } from 'react'
import { useParams, useNavigate, Link, useLoaderData } from 'react-router'
import type { ClientVideoDetail, PlayEpisodeResponse, VideoDetailSourceGroup } from '@orange-tv/shared'
import { clientApi, errorMessage } from '@/lib/api'
import { useAuth } from '@/hooks/useAuth'
import { useSettingsStore } from '@/store/settings'
import { VideoPlayer } from '@/components/Player'
import { saveHistory } from '@/lib/playbackHistory'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Empty, EmptyDescription, EmptyHeader, EmptyTitle } from '@/components/ui/empty'
import { AlertCircleIcon } from 'lucide-react'
import { FavoriteButton, RatingStars } from '@/components/common'

type PlayLoaderData = {
  detail: ClientVideoDetail | null
  episode: PlayEpisodeResponse | null
  error: string
}

export async function loader({ params }: { params: Record<string, string | undefined> }): Promise<PlayLoaderData> {
  const { id, sourceId, episodeId } = params
  if (!id || !sourceId || !episodeId) {
    return { detail: null, episode: null, error: '' }
  }
  try {
    const [detailRes, epRes] = await Promise.all([
      clientApi.video(Number(id)),
      clientApi.playEpisode(Number(id), Number(sourceId), Number(episodeId)),
    ])
    return {
      detail: detailRes.data || null,
      episode: epRes.data || null,
      error: '',
    }
  } catch (err) {
    return { detail: null, episode: null, error: errorMessage(err) }
  }
}

export function Component() {
  const { id, sourceId, episodeId } = useParams()
  const navigate = useNavigate()
  const ad = useSettingsStore((s) => s.ad)
  const { profile } = useAuth()
  const data = useLoaderData<PlayLoaderData>()
  const { detail, episode, error } = data
  const [resumeAt, setResumeAt] = useState<number | undefined>(undefined)
  const remoteSyncInFlight = useRef(false)

  const sourceIdNum = Number(sourceId || 0)
  const epIdNum = Number(episodeId || 0)
  const videoIdNum = Number(id || 0)

  const handleProgress = useCallback(
    (currentTime: number, duration: number) => {
      if (!detail?.title) return
      saveHistory({
        videoId: videoIdNum,
        sourceId: sourceIdNum,
        episodeId: epIdNum,
        progress: currentTime,
        title: detail.title,
        updatedAt: Date.now(),
      })
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
    },
    [detail?.title, videoIdNum, sourceIdNum, epIdNum, profile],
  )

  useEffect(() => {
    if (!id) return
    setResumeAt(undefined)
    if (!profile) return
    clientApi
      .getHistory(Number(id))
      .then((res) => {
        if (res.data && res.data.progress > 0) {
          setResumeAt(res.data.progress)
        }
      })
      .catch(() => undefined)
  }, [id, profile])

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
  const playlist = sourceGroup?.episodes?.map((ep) => ({
    episodeId: ep.id,
    title: ep.title || `第${ep.episode}集`,
  }))
  const currentEpNumber = detail.sources
    ?.flatMap((s) => s.episodes)
    .find((e) => e.id === epIdNum)?.episode

  return (
    <div className="flex flex-col gap-6">
      <div className="flex flex-col gap-2">
        <p className="text-sm text-muted-foreground">
          正在播放：{detail.title}
          {currentEpNumber ? ` - 第${currentEpNumber}集` : ''}
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
            onEpisodeChange={(epId) =>
              navigate(`/play/${id}/${sourceIdNum}/${epId}`)
            }
            onProgress={handleProgress}
          />
        </div>
      </div>

      <div className="flex flex-col gap-3">
        <div className="flex items-center gap-3">
          <Link
            to={`/video/${id}`}
            className="text-xl font-bold hover:text-primary transition-colors"
          >
            {detail.title}
          </Link>
          <FavoriteButton videoId={Number(id)} />
        </div>
        <RatingStars
          videoId={Number(id)}
          rating={detail.rating}
          ratingCount={detail.rating_count}
        />

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
                        variant={
                          source.id === sourceIdNum && ep.id === epIdNum
                            ? 'default'
                            : 'outline'
                        }
                        size="sm"
                        onClick={() =>
                          navigate(`/play/${id}/${source.id}/${ep.id}`)
                        }
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
