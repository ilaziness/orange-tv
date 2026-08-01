import { useEffect, useState, useMemo } from 'react'
import type { HistoryItem } from '@orange-tv/shared'
import { clientApi, errorMessage } from '@/lib/api'
import { HistoryCard } from '@/components/common/HistoryCard'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Empty, EmptyDescription, EmptyHeader, EmptyTitle } from '@/components/ui/empty'
import { Skeleton } from '@/components/ui/skeleton'
import { Card } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { AlertCircleIcon } from 'lucide-react'

function dayKey(iso: string): string {
  // YYYY-MM-DD (local)
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return '未知日期'
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${y}-${m}-${day}`
}

function dayLabel(key: string): string {
  if (key === '未知日期') return key
  const today = new Date()
  const todayKey = dayKey(today.toISOString())
  const yesterday = new Date(today)
  yesterday.setDate(today.getDate() - 1)
  const yesterdayKey = dayKey(yesterday.toISOString())
  if (key === todayKey) return '今天'
  if (key === yesterdayKey) return '昨天'
  return key
}

export default function HistoryPage() {
  const [history, setHistory] = useState<HistoryItem[]>([])
  const [loading, setLoading] = useState(true)
  const [loadingMore, setLoadingMore] = useState(false)
  const [error, setError] = useState('')
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)

  useEffect(() => {
    void (async () => {
      setLoading(true)
      try {
        const res = await clientApi.listHistory(1)
        setHistory(res.data.list || [])
        setTotal(res.data.total || 0)
      } catch (err) {
        setError(errorMessage(err))
      } finally {
        setLoading(false)
      }
    })()
  }, [])

  const loadMore = async () => {
    const next = page + 1
    setLoadingMore(true)
    try {
      const res = await clientApi.listHistory(next)
      setHistory((prev) => [...prev, ...(res.data.list || [])])
      setTotal(res.data.total || total)
      setPage(next)
    } catch (err) {
      setError(errorMessage(err))
    } finally {
      setLoadingMore(false)
    }
  }

  // 按天分组，最近的天在最上；同一天内按 last_played_at 倒序（越近越左）
  const groups = useMemo(() => {
    const map = new Map<string, HistoryItem[]>()
    for (const h of history) {
      const k = dayKey(h.last_played_at)
      const arr = map.get(k) || []
      arr.push(h)
      map.set(k, arr)
    }
    // 每组内按时间倒序
    for (const arr of map.values()) {
      arr.sort((a, b) => new Date(b.last_played_at).getTime() - new Date(a.last_played_at).getTime())
    }
    // 组按天倒序（最近在上）
    return Array.from(map.entries()).sort((a, b) => (a[0] < b[0] ? 1 : -1))
  }, [history])

  const hasMore = history.length < total

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
        <div className="flex flex-col gap-6">
          {Array.from({ length: 2 }).map((_, gi) => (
            <div key={gi} className="flex flex-col gap-3">
              <Skeleton className="h-5 w-24" />
              <div className="flex flex-wrap gap-3">
                {Array.from({ length: 6 }).map((_, i) => (
                  <div key={i} className="w-28 shrink-0">
                    <Card className="gap-0 overflow-hidden pb-1.5 pt-0">
                      <Skeleton className="aspect-[2/3] w-full rounded-t-xl" />
                      <div className="flex flex-col gap-0.5 p-1.5">
                        <Skeleton className="h-3 w-full" />
                        <Skeleton className="h-2.5 w-2/3" />
                      </div>
                    </Card>
                  </div>
                ))}
              </div>
            </div>
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
        <div className="flex flex-col gap-6">
          {groups.map(([key, items]) => (
            <div key={key} className="flex gap-4">
              {/* 左侧时间轴 */}
              <div className="flex w-20 shrink-0 flex-col items-end">
                <span className="text-sm font-medium text-foreground">{dayLabel(key)}</span>
                <div className="mt-2 w-px flex-1 bg-border" />
              </div>
              {/* 右侧卡片横向排列 */}
              <div className="flex flex-1 flex-wrap gap-3 pb-2">
                {items.map((h) => (
                  <HistoryCard
                    key={`${h.video_id}_${h.episode_id}`}
                    videoId={h.video_id}
                    sourceId={h.play_source_id}
                    episodeId={h.episode_id}
                    title={h.title}
                    categoryName={h.category_name}
                    year={h.year}
                    cover={h.cover}
                    progress={h.progress}
                  />
                ))}
              </div>
            </div>
          ))}

          {hasMore ? (
            <div className="flex justify-center">
              <Button variant="outline" size="sm" onClick={() => void loadMore()} disabled={loadingMore}>
                {loadingMore ? '加载中…' : '加载更多'}
              </Button>
            </div>
          ) : null}
        </div>
      )}
    </div>
  )
}
