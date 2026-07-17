import { useEffect, useState } from 'react'
import { useParams } from 'react-router'
import type { LiveChannel } from '@orange-tv/shared'
import { clientApi, errorMessage } from '@/lib/api'
import { VideoPlayer } from '@/components/Player'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Empty, EmptyDescription, EmptyHeader, EmptyTitle } from '@/components/ui/empty'
import { Skeleton } from '@/components/ui/skeleton'
import { AlertCircleIcon } from 'lucide-react'

export default function LivePlayPage() {
  const { id } = useParams()
  const [channel, setChannel] = useState<LiveChannel | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    if (!id) return
    void (async () => {
      setLoading(true)
      try {
        const res = await clientApi.liveChannelDetail(Number(id))
        setChannel(res.data || null)
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

  if (!channel) {
    return (
      <Empty>
        <EmptyHeader>
          <EmptyTitle>频道不存在</EmptyTitle>
          <EmptyDescription>该直播频道可能已下架</EmptyDescription>
        </EmptyHeader>
      </Empty>
    )
  }

  return (
    <div className="flex flex-col gap-6">
      <div className="overflow-hidden rounded-xl border border-border">
        <VideoPlayer src={channel.stream_url} />
      </div>
      <div className="flex flex-col gap-2">
        <h1 className="text-xl font-bold">{channel.name}</h1>
        <p className="text-sm text-muted-foreground">{channel.description || '暂无描述'}</p>
      </div>
    </div>
  )
}
