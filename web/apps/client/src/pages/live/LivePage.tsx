import { useEffect, useMemo, useState } from 'react'
import type { LiveChannel } from '@orange-tv/shared'
import { clientApi, errorMessage } from '@/lib/api'
import { VideoPlayer } from '@/components/Player'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Empty, EmptyDescription, EmptyHeader, EmptyTitle } from '@/components/ui/empty'
import { Skeleton } from '@/components/ui/skeleton'
import { Button } from '@/components/ui/button'
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from '@/components/ui/collapsible'
import { cn } from '@/lib/utils'
import { AlertCircleIcon, ChevronRightIcon, TvIcon } from 'lucide-react'

type ChannelGroup = {
  category: string
  channels: LiveChannel[]
}

export default function LivePage() {
  const [channels, setChannels] = useState<LiveChannel[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [expanded, setExpanded] = useState<Set<string>>(new Set())
  const [selectedId, setSelectedId] = useState<number | null>(null)

  // 一次性加载所有在线频道
  useEffect(() => {
    void (async () => {
      setLoading(true)
      try {
        const res = await clientApi.live()
        const list = res.data.list || []
        setChannels(list)
        if (list.length > 0) {
          setSelectedId(list[0].id)
          // 默认展开第一个频道所属分类
          setExpanded(new Set([list[0].category]))
        }
      } catch (err) {
        setError(errorMessage(err))
      } finally {
        setLoading(false)
      }
    })()
  }, [])

  // 前端按 category 分组
  const groups = useMemo<ChannelGroup[]>(() => {
    const map = new Map<string, LiveChannel[]>()
    for (const ch of channels) {
      if (!map.has(ch.category)) {
        map.set(ch.category, [])
      }
      map.get(ch.category)!.push(ch)
    }
    for (const arr of map.values()) {
      arr.sort((a, b) => a.sort_order - b.sort_order)
    }
    return Array.from(map.keys()).map((key) => ({ category: key, channels: map.get(key)! }))
  }, [channels])

  const selectedChannel = useMemo<LiveChannel | null>(() => {
    return channels.find((ch) => ch.id === selectedId) || null
  }, [channels, selectedId])

  const toggleCategory = (category: string) => {
    setExpanded((prev) => {
      const next = new Set(prev)
      if (next.has(category)) {
        next.delete(category)
      } else {
        next.add(category)
      }
      return next
    })
  }

  const handleChannelClick = (ch: LiveChannel) => {
    setSelectedId(ch.id)
    // 播放某频道时确保其分类展开
    setExpanded((prev) => {
      if (prev.has(ch.category)) return prev
      const next = new Set(prev)
      next.add(ch.category)
      return next
    })
  }

  if (loading) {
    return (
      <div className="flex h-[calc(100vh-8rem)] flex-col gap-4 lg:flex-row">
        <aside className="flex h-auto max-h-48 w-full shrink-0 flex-col gap-2 lg:h-full lg:max-h-none lg:w-44">
          <Skeleton className="h-6 w-32" />
          {Array.from({ length: 8 }).map((_, i) => (
            <Skeleton key={i} className="h-9 w-full" />
          ))}
        </aside>
        <section className="flex min-h-0 flex-1 flex-col gap-4 lg:h-full">
          <Skeleton className="aspect-video w-full rounded-xl" />
          <Skeleton className="h-7 w-1/3" />
        </section>
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

  if (!channels.length) {
    return (
      <Empty>
        <EmptyHeader>
          <EmptyTitle>暂无电视频道</EmptyTitle>
          <EmptyDescription>暂无可用的直播频道</EmptyDescription>
        </EmptyHeader>
      </Empty>
    )
  }

  return (
    <div className="flex h-[calc(100vh-8rem)] flex-col gap-4 lg:flex-row">
      <aside className="flex h-auto max-h-48 w-full shrink-0 flex-col gap-2 lg:h-full lg:max-h-none lg:w-44">
        <h2 className="text-base font-semibold">电视频道</h2>
        <div className="flex flex-1 flex-col gap-1 overflow-y-auto pr-1">
          {groups.map((group) => {
            const isExpanded = expanded.has(group.category)
            return (
              <Collapsible
                key={group.category}
                open={isExpanded}
                onOpenChange={(open) => {
                  if (open) toggleCategory(group.category)
                  else toggleCategory(group.category)
                }}
              >
                <CollapsibleTrigger
                  render={
                    <Button
                      variant="ghost"
                      size="sm"
                      className={cn('w-full justify-start gap-2', isExpanded && 'bg-muted')}
                    >
                      <ChevronRightIcon
                        className={cn(
                          'size-4 shrink-0 transition-transform',
                          isExpanded && 'rotate-90',
                        )}
                      />
                      <span className="truncate">{group.category}</span>
                      <span className="ml-auto text-xs text-muted-foreground">
                        {group.channels.length}
                      </span>
                    </Button>
                  }
                />
                <CollapsibleContent>
                  <div className="flex flex-col pl-6">
                    {group.channels.map((ch) => (
                      <Button
                        key={ch.id}
                        variant="ghost"
                        size="sm"
                        className={cn(
                          'justify-start gap-2',
                          ch.id === selectedId && 'bg-primary/10 text-primary',
                        )}
                        onClick={() => handleChannelClick(ch)}
                      >
                        {ch.logo ? (
                          <img
                            src={ch.logo}
                            alt={ch.name}
                            className="size-6 rounded object-cover"
                            onError={(e) => {
                              ;(e.currentTarget as HTMLImageElement).style.display = 'none'
                            }}
                          />
                        ) : (
                          <TvIcon className="size-4 shrink-0 text-muted-foreground" />
                        )}
                        <span className="truncate">{ch.name}</span>
                      </Button>
                    ))}
                  </div>
                </CollapsibleContent>
              </Collapsible>
            )
          })}
        </div>
      </aside>
      <section className="flex min-h-0 flex-1 flex-col gap-4 overflow-y-auto lg:h-full">
        {selectedChannel ? (
          <>
            <div className="overflow-hidden rounded-xl border border-border">
              <VideoPlayer src={clientApi.liveStreamUrl(selectedChannel.id)} format="hls" />
            </div>
            <div className="flex flex-col gap-2">
              <h1 className="text-xl font-bold">{selectedChannel.name}</h1>
              {selectedChannel.description ? (
                <p className="text-sm text-muted-foreground">{selectedChannel.description}</p>
              ) : null}
            </div>
          </>
        ) : (
          <div className="flex flex-1 items-center justify-center rounded-xl border border-dashed border-border text-muted-foreground">
            请选择电视频道
          </div>
        )}
      </section>
    </div>
  )
}
