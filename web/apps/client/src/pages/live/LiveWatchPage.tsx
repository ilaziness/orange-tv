import { useEffect, useMemo, useState } from 'react'
import { Link, useParams } from 'react-router'
import type { ClientLiveRoomDetail } from '@orange-tv/shared'
import { formatLiveCategoryName, LiveRoomStatusLiving } from '@orange-tv/shared'
import { clientApi, errorMessage } from '@/lib/api'
import { VideoPlayer } from '@/components/Player'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Empty, EmptyDescription, EmptyHeader, EmptyTitle } from '@/components/ui/empty'
import { Skeleton } from '@/components/ui/skeleton'
import { AlertCircleIcon, ArrowLeftIcon } from 'lucide-react'
import { usePageTitle } from '@/hooks/usePageTitle'
import { usePageSeo } from '@/hooks/usePageSeo'

function parseRoomId(raw: string | undefined): number {
  const id = Number(raw)
  if (!Number.isInteger(id) || id <= 0) return 0
  return id
}

export function Component() {
  const params = useParams()
  const roomId = parseRoomId(params.id)

  const [room, setRoom] = useState<ClientLiveRoomDetail | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  const title = room?.title || '直播间'
  usePageTitle(title)
  usePageSeo({
    title,
    path: roomId ? `/live/${roomId}` : '/live',
    description: room?.notice || undefined,
    image: room?.cover || undefined,
  })

  useEffect(() => {
    if (!roomId) {
      setLoading(false)
      setError('无效的直播间')
      return
    }
    let cancelled = false
    setLoading(true)
    setError('')
    void (async () => {
      try {
        const res = await clientApi.liveRoom(roomId)
        if (cancelled) return
        setRoom(res.data || null)
      } catch (err) {
        if (cancelled) return
        setRoom(null)
        setError(errorMessage(err))
      } finally {
        if (!cancelled) setLoading(false)
      }
    })()
    return () => {
      cancelled = true
    }
  }, [roomId])

  const playSrc = useMemo(() => {
    if (!room) return ''
    return room.play_hls_url || room.play_flv_url || ''
  }, [room])

  const playFormat = useMemo(() => {
    if (!room) return undefined
    if (room.play_hls_url) return 'hls'
    if (room.play_flv_url) return 'flv'
    return undefined
  }, [room])

  const nickname = room?.streamer_nickname || room?.streamer_str_id || '主播'
  const fallback = (nickname[0] || 'L').toUpperCase()
  const isLiving = room?.status === LiveRoomStatusLiving
  const categoryLabel = room
    ? formatLiveCategoryName(room.category_name, room.sub_category_name)
    : ''

  if (loading) {
    return (
      <div className="flex flex-col gap-4">
        <Skeleton className="aspect-video w-full rounded-xl" />
        <Skeleton className="h-8 w-1/2" />
        <Skeleton className="h-4 w-1/3" />
      </div>
    )
  }

  if (error) {
    return (
      <div className="flex flex-col gap-4">
        <Button variant="ghost" size="sm" className="w-fit" nativeButton={false} render={<Link to="/live" />}>
          <ArrowLeftIcon data-icon="inline-start" />
          返回直播列表
        </Button>
        <Alert variant="destructive">
          <AlertCircleIcon />
          <AlertTitle>加载失败</AlertTitle>
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      </div>
    )
  }

  if (!room) {
    return (
      <Empty>
        <EmptyHeader>
          <EmptyTitle>直播间不存在</EmptyTitle>
          <EmptyDescription>
            <Link to="/live" className="cursor-pointer text-primary underline-offset-4 hover:underline">
              返回直播列表
            </Link>
          </EmptyDescription>
        </EmptyHeader>
      </Empty>
    )
  }

  return (
    <div className="flex flex-col gap-6">
      <Button variant="ghost" size="sm" className="w-fit" nativeButton={false} render={<Link to="/live" />}>
        <ArrowLeftIcon data-icon="inline-start" />
        返回直播列表
      </Button>

      <div className="overflow-hidden rounded-xl border border-border">
        {playSrc && isLiving ? (
          <VideoPlayer src={playSrc} format={playFormat} poster={room.cover || undefined} autoplay />
        ) : (
          <div className="flex aspect-video items-center justify-center bg-muted text-sm text-muted-foreground">
            {isLiving ? '暂无可用播放地址' : '主播暂时不在直播'}
          </div>
        )}
      </div>

      <div className="flex flex-col gap-3">
        <div className="flex flex-wrap items-center gap-2">
          <h1 className="text-xl font-bold">{room.title || '未命名直播间'}</h1>
          {isLiving ? (
            <Badge variant="destructive">直播中</Badge>
          ) : (
            <Badge variant="secondary">未开播</Badge>
          )}
          {categoryLabel ? <Badge variant="outline">{categoryLabel}</Badge> : null}
        </div>

        <div className="flex items-center gap-3">
          <Avatar>
            {room.streamer_avatar ? <AvatarImage src={room.streamer_avatar} alt={nickname} /> : null}
            <AvatarFallback>{fallback}</AvatarFallback>
          </Avatar>
          <div className="min-w-0">
            <p className="font-medium">{nickname}</p>
            {room.started_at ? (
              <p className="text-xs text-muted-foreground">开播于 {room.started_at}</p>
            ) : null}
          </div>
        </div>

        {room.notice ? (
          <p className="whitespace-pre-wrap text-sm text-muted-foreground">{room.notice}</p>
        ) : null}
      </div>
    </div>
  )
}
