import { useCallback, useEffect, useRef, useState } from 'react'
import { useParams, useNavigate, Link, useLoaderData } from 'react-router'
import type {
  ClientAdItem,
  ClientVideoDetail,
  CommentItem,
  PlayEpisodeResponse,
  VideoDetailSourceGroup,
} from '@orange-tv/shared'
import { clientApi, errorMessage } from '@/lib/api'
import { useAuth } from '@/hooks/useAuth'
import { useSettings } from '@/hooks/useSettings'
import { VideoPlayer } from '@/components/Player'
import { saveHistory } from '@/lib/playbackHistory'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Empty, EmptyDescription, EmptyHeader, EmptyTitle } from '@/components/ui/empty'
import { AlertCircleIcon } from 'lucide-react'
import { FavoriteButton, RatingStars } from '@/components/common'
import { CommentSection, QuickCommentInput } from '@/components/comment'
import { PlaySourceEpisodeList } from '@/components/PlaySourceEpisodeList'
import { usePageTitle } from '@/hooks/usePageTitle'
import { usePageSeo } from '@/hooks/usePageSeo'
import { toast } from 'sonner'

type PlayLoaderData = {
  detail: ClientVideoDetail | null
  episode: PlayEpisodeResponse | null
  error: string
}

export async function loader({
  params,
}: {
  params: Record<string, string | undefined>
}): Promise<PlayLoaderData> {
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
  const [videoAds, setVideoAds] = useState<ClientAdItem[]>([])
  const { profile } = useAuth()
  const { feature } = useSettings()
  const data = useLoaderData<PlayLoaderData>()
  const { detail, episode, error } = data
  const [comments, setComments] = useState<CommentItem[] | null>(null)
  const [commentTotal, setCommentTotal] = useState(0)
  const [commentPage, setCommentPage] = useState(1)
  const [commentTotalPages, setCommentTotalPages] = useState(0)
  const [commentsLoading, setCommentsLoading] = useState(false)
  const [resumeAt, setResumeAt] = useState<number | undefined>(undefined)
  const remoteSyncInFlight = useRef(false)
  const commentsFetchSeq = useRef(0)

  const sourceIdNum = Number(sourceId || 0)
  const epIdNum = Number(episodeId || 0)
  const videoIdNum = Number(id || 0)
  const commentEnabled = feature.comment_enabled

  const loadComments = useCallback(
    (page = 1) => {
      if (!id) return
      const seq = ++commentsFetchSeq.current
      setCommentsLoading(true)
      void clientApi
        .listComments(Number(id), page)
        .then((res) => {
          if (seq !== commentsFetchSeq.current) return
          setComments(res.data.list || [])
          setCommentTotal(res.data.total || 0)
          setCommentPage(res.data.page || page)
          setCommentTotalPages(res.data.total_pages || 0)
        })
        .catch((err) => {
          if (seq !== commentsFetchSeq.current) return
          setComments((prev) => prev ?? [])
          toast.error(errorMessage(err))
        })
        .finally(() => {
          if (seq === commentsFetchSeq.current) setCommentsLoading(false)
        })
    },
    [id],
  )

  useEffect(() => {
    if (!commentEnabled || !id) {
      setComments(null)
      setCommentTotal(0)
      setCommentPage(1)
      setCommentTotalPages(0)
      setCommentsLoading(false)
      return
    }
    setComments(null)
    loadComments(1)
    return () => {
      commentsFetchSeq.current += 1
    }
  }, [commentEnabled, id, loadComments])

  // Load video loading ads once on mount
  useEffect(() => {
    let mounted = true
    clientApi
      .ads('video_loading')
      .then((res) => {
        if (mounted) setVideoAds(res.data || [])
      })
      .catch(() => {
        if (mounted) setVideoAds([])
      })
    return () => {
      mounted = false
    }
  }, [])

  const currentEpNumber = detail?.sources
    ?.flatMap((s) => s.episodes)
    ?.find((e) => e.id === epIdNum)?.episode

  usePageTitle(
    detail
      ? `${detail.title}${currentEpNumber ? ` 第${currentEpNumber}集` : ''} - 播放`
      : '视频播放',
  )
  usePageSeo({
    title: detail
      ? `${detail.title}${currentEpNumber ? ` 第${currentEpNumber}集` : ''} - 播放`
      : '视频播放',
    description: detail?.description,
    image: detail?.cover || undefined,
    path: id ? `/video/${id}` : undefined,
    noindex: true,
  })

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
          last_played_at: new Date().toISOString(),
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
            ads={videoAds}
            playlist={playlist}
            currentEpisodeId={epIdNum}
            onEpisodeChange={(epId) => navigate(`/play/${id}/${sourceIdNum}/${epId}`)}
            onProgress={handleProgress}
          />
        </div>
      </div>

      <div className="flex flex-col gap-3">
        <div className="flex flex-wrap items-center gap-3">
          <Link
            to={`/video/${id}`}
            className="shrink-0 text-xl font-bold transition-colors hover:text-primary"
          >
            {detail.title}
          </Link>
          <FavoriteButton videoId={videoIdNum} />
          {commentEnabled ? (
            <QuickCommentInput videoId={videoIdNum} onSuccess={() => loadComments(1)} />
          ) : null}
        </div>
        <RatingStars
          videoId={videoIdNum}
          rating={detail.rating}
          ratingCount={detail.rating_count}
        />

        {detail.sources && detail.sources.length > 0 ? (
          <PlaySourceEpisodeList
            sources={detail.sources}
            currentSourceId={sourceIdNum}
            currentEpisodeId={epIdNum}
            onSelectEpisode={(srcId, epId) => navigate(`/play/${id}/${srcId}/${epId}`)}
          />
        ) : null}
      </div>

      {commentEnabled && comments ? (
        <CommentSection
          videoId={videoIdNum}
          comments={comments}
          total={commentTotal}
          page={commentPage}
          totalPages={commentTotalPages}
          loading={commentsLoading}
          onRefresh={() => loadComments(1)}
          onPageChange={loadComments}
        />
      ) : null}
    </div>
  )
}
