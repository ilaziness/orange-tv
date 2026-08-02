import { useMemo, useState } from 'react'
import { useLoaderData } from 'react-router'
import type { HistoryItem } from '@orange-tv/shared'
import { clientApi, errorMessage } from '@/lib/api'
import { HistoryCard } from '@/components/common/HistoryCard'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Empty, EmptyDescription, EmptyHeader, EmptyTitle } from '@/components/ui/empty'
import { Button } from '@/components/ui/button'
import { AlertCircleIcon } from 'lucide-react'
import { usePageTitle } from '@/hooks/usePageTitle'

function dayKey(iso: string): string {
  const d = new Date(iso.replace(' ', 'T'))
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

type HistoryLoaderData = {
  history: HistoryItem[]
  total: number
  error: string
}

export async function loader(): Promise<HistoryLoaderData> {
  try {
    const res = await clientApi.listHistory(1)
    return {
      history: res.data.list || [],
      total: res.data.total || 0,
      error: '',
    }
  } catch (err) {
    return { history: [], total: 0, error: errorMessage(err) }
  }
}

export function Component() {
  const data = useLoaderData<HistoryLoaderData>()
  const { total: initialTotal, error } = data
  const [history, setHistory] = useState<HistoryItem[]>(data.history)
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(initialTotal)
  const [loadingMore, setLoadingMore] = useState(false)

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

  const [displayError, setError] = useState(error)

  usePageTitle('观看历史')

  const groups = useMemo(() => {
    const map = new Map<string, HistoryItem[]>()
    for (const h of history) {
      const k = dayKey(h.last_played_at)
      const arr = map.get(k) || []
      arr.push(h)
      map.set(k, arr)
    }
    for (const arr of map.values()) {
      arr.sort((a, b) => new Date(b.last_played_at.replace(' ', 'T')).getTime() - new Date(a.last_played_at.replace(' ', 'T')).getTime())
    }
    return Array.from(map.entries()).sort((a, b) => (a[0] < b[0] ? 1 : -1))
  }, [history])

  const hasMore = history.length < total

  if (displayError) {
    return (
      <Alert variant="destructive">
        <AlertCircleIcon />
        <AlertTitle>加载失败</AlertTitle>
        <AlertDescription>{displayError}</AlertDescription>
      </Alert>
    )
  }

  if (!history.length) {
    return (
      <Empty>
        <EmptyHeader>
          <EmptyTitle>暂无观看历史</EmptyTitle>
          <EmptyDescription>观看记录将显示在这里</EmptyDescription>
        </EmptyHeader>
      </Empty>
    )
  }

  return (
    <div className="flex flex-col gap-6">
      <h2 className="text-lg font-semibold">观看历史</h2>

      <div className="flex flex-col gap-6">
        {groups.map(([key, items]) => (
          <div key={key} className="flex gap-4">
            <div className="flex w-20 shrink-0 flex-col items-end">
              <span className="text-sm font-medium text-foreground">{dayLabel(key)}</span>
              <div className="mt-2 w-px flex-1 bg-border" />
            </div>
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
    </div>
  )
}
