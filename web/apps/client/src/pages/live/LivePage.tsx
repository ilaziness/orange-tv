import { useEffect, useState } from 'react'
import { Link } from 'react-router'
import type { LiveChannel } from '@orange-tv/shared'
import { clientApi, errorMessage } from '@/lib/api'
import { Card } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Empty, EmptyDescription, EmptyHeader, EmptyTitle } from '@/components/ui/empty'
import { Skeleton } from '@/components/ui/skeleton'
import { AlertCircleIcon } from 'lucide-react'

export default function LivePage() {
  const [channels, setChannels] = useState<LiveChannel[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    void (async () => {
      setLoading(true)
      try {
        const res = await clientApi.live()
        setChannels(res.data.list || [])
      } catch (err) {
        setError(errorMessage(err))
      } finally {
        setLoading(false)
      }
    })()
  }, [])

  return (
    <div className="flex flex-col gap-6">
      <h2 className="text-lg font-semibold">直播频道</h2>

      {error ? (
        <Alert variant="destructive">
          <AlertCircleIcon />
          <AlertTitle>加载失败</AlertTitle>
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      ) : null}

      {loading ? (
        <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6">
          {Array.from({ length: 12 }).map((_, i) => (
            <Card key={i} className="overflow-hidden">
              <Skeleton className="aspect-[2/3] w-full rounded-none" />
              <div className="flex flex-col gap-2 p-3">
                <Skeleton className="h-4 w-3/4" />
                <Skeleton className="h-3 w-1/2" />
              </div>
            </Card>
          ))}
        </div>
      ) : !channels.length ? (
        <Empty>
          <EmptyHeader>
            <EmptyTitle>暂无直播频道</EmptyTitle>
            <EmptyDescription>暂无可用的直播频道</EmptyDescription>
          </EmptyHeader>
        </Empty>
      ) : (
        <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6">
          {channels.map((ch) => (
            <Link key={ch.id} to={`/play/live/${ch.id}`}>
              <Card className="overflow-hidden transition-all hover:ring-primary/40">
                <div
                  className="relative aspect-[2/3] w-full bg-cover bg-center"
                  style={ch.logo ? { backgroundImage: `url(${ch.logo})` } : undefined}
                >
                  <div className="absolute top-2 left-2">
                    <Badge variant="destructive">
                      LIVE
                    </Badge>
                  </div>
                </div>
                <div className="flex flex-col gap-1 p-3">
                  <h3 className="truncate text-sm font-medium">{ch.name}</h3>
                  <p className="text-xs text-muted-foreground">{ch.description || '暂无描述'}</p>
                </div>
              </Card>
            </Link>
          ))}
        </div>
      )}
    </div>
  )
}
