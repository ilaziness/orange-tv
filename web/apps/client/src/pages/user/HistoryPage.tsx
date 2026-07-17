import { useEffect, useState } from 'react'
import { Link } from 'react-router'
import type { HistoryItem } from '@orange-tv/shared'
import { clientApi, errorMessage } from '@/lib/api'
import { Card } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Empty, EmptyDescription, EmptyHeader, EmptyTitle } from '@/components/ui/empty'
import { Skeleton } from '@/components/ui/skeleton'
import { AlertCircleIcon } from 'lucide-react'

export default function HistoryPage() {
  const [history, setHistory] = useState<HistoryItem[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    void (async () => {
      setLoading(true)
      try {
        const res = await clientApi.listHistory()
        setHistory(res.data.list || [])
      } catch (err) {
        setError(errorMessage(err))
      } finally {
        setLoading(false)
      }
    })()
  }, [])

  return (
    <div className="flex flex-col gap-6">
      <h2 className="text-lg font-semibold">观看历史</h2>

      {error ? (
        <Alert variant="destructive">
          <AlertCircleIcon />
          <AlertTitle>加载失败</AlertTitle>
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      ) : null}

      {loading ? (
        <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6">
          {Array.from({ length: 6 }).map((_, i) => (
            <Card key={i} className="overflow-hidden">
              <Skeleton className="aspect-[2/3] w-full rounded-none" />
              <div className="flex flex-col gap-2 p-3">
                <Skeleton className="h-4 w-3/4" />
                <Skeleton className="h-3 w-1/2" />
              </div>
            </Card>
          ))}
        </div>
      ) : !history.length ? (
        <Empty>
          <EmptyHeader>
            <EmptyTitle>暂无观看历史</EmptyTitle>
            <EmptyDescription>观看记录将显示在这里</EmptyDescription>
          </EmptyHeader>
        </Empty>
      ) : (
        <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6">
          {history.map((h) => (
            <Link key={h.video_id} to={`/video/${h.video_id}`}>
              <Card className="overflow-hidden transition-all hover:ring-primary/40">
                <div
                  className="aspect-[2/3] w-full bg-cover bg-center"
                  style={h.cover ? { backgroundImage: `url(${h.cover})` } : undefined}
                >
                  {h.progress > 0 ? (
                    <div className="flex p-2">
                      <Badge variant="default" className="bg-black/65 text-white">
                        {Math.round(h.progress)}%
                      </Badge>
                    </div>
                  ) : null}
                </div>
                <div className="flex flex-col gap-1 p-3">
                  <h3 className="truncate text-sm font-medium">{h.title}</h3>
                  <p className="text-xs text-muted-foreground">
                    观看于 {new Date(h.last_played_at).toLocaleDateString()}
                  </p>
                </div>
              </Card>
            </Link>
          ))}
        </div>
      )}
    </div>
  )
}
