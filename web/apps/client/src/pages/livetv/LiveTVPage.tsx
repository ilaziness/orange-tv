import { useMemo, useState } from 'react'
import { useLoaderData } from 'react-router'
import type { ClientLiveTVChannel } from '@orange-tv/shared'
import { clientApi, errorMessage } from '@/lib/api'
import { VideoPlayer } from '@/components/Player'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Empty, EmptyDescription, EmptyHeader, EmptyTitle } from '@/components/ui/empty'
import { Button } from '@/components/ui/button'
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from '@/components/ui/collapsible'
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip'
import { cn } from '@/lib/utils'
import { AlertCircleIcon, ChevronRightIcon, TvIcon } from 'lucide-react'
import { usePageTitle } from '@/hooks/usePageTitle'
import { usePageSeo } from '@/hooks/usePageSeo'

type LiveLoaderData = {
  channels: ClientLiveTVChannel[]
  error: string
}

type ChannelGroup = {
  category: string
  channels: ClientLiveTVChannel[]
}

function ChannelLogo({ name, logo }: { name: string; logo?: string }) {
  const [error, setError] = useState(false)
  if (!logo || error) {
    return (
      <span className="flex size-6 items-center justify-center rounded-md bg-muted">
        <TvIcon className="text-muted-foreground" />
      </span>
    )
  }
  return (
    <span className="flex size-6 items-center justify-center rounded-md bg-muted p-1">
      <img
        src={logo}
        alt={name}
        className="max-h-full max-w-full object-contain"
        onError={() => setError(true)}
      />
    </span>
  )
}

export async function loader(): Promise<LiveLoaderData> {
  try {
    const res = await clientApi.livetv()
    const list = res.data.list || []
    return { channels: list, error: '' }
  } catch (err) {
    return { channels: [], error: errorMessage(err) }
  }
}

export function Component() {
  const data = useLoaderData<LiveLoaderData>()
  const { channels, error } = data
  const [expanded, setExpanded] = useState<Set<string>>(() => {
    const first = channels[0]
    return new Set(first ? [first.category] : [])
  })
  const [selectedId, setSelectedId] = useState<number | null>(() => channels[0]?.id ?? null)

  usePageTitle('电视直播')
  usePageSeo({ title: '电视直播', path: '/livetv' })

  const groups = useMemo<ChannelGroup[]>(() => {
    const map = new Map<string, ClientLiveTVChannel[]>()
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

  const selectedChannel = useMemo<ClientLiveTVChannel | null>(() => {
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

  const handleChannelClick = (ch: ClientLiveTVChannel) => {
    setSelectedId(ch.id)
    setExpanded((prev) => {
      if (prev.has(ch.category)) return prev
      const next = new Set(prev)
      next.add(ch.category)
      return next
    })
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
      <aside className="flex h-auto max-h-48 w-full shrink-0 flex-col gap-2 lg:h-full lg:max-h-none lg:w-56">
        <h2 className="text-base font-semibold">电视频道</h2>
        <div className="flex flex-1 flex-col gap-1 overflow-y-auto pr-1">
          {groups.map((group) => {
            const isExpanded = expanded.has(group.category)
            return (
              <Collapsible
                key={group.category}
                open={isExpanded}
                onOpenChange={() => toggleCategory(group.category)}
              >
                <CollapsibleTrigger
                  render={
                    <Button
                      variant="ghost"
                      className={cn('w-full justify-start gap-2', isExpanded && 'bg-muted')}
                    >
                      <ChevronRightIcon
                        data-icon="inline-start"
                        className={cn('transition-transform', isExpanded && 'rotate-90')}
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
                        className={cn(
                          'w-full justify-start gap-2.5 px-2',
                          ch.id === selectedId &&
                            'bg-primary/10 text-primary hover:bg-primary/10 hover:text-primary',
                        )}
                        onClick={() => handleChannelClick(ch)}
                      >
                        <ChannelLogo name={ch.name} logo={ch.logo} />
                        <Tooltip>
                          <TooltipTrigger render={<span className="truncate">{ch.name}</span>} />
                          <TooltipContent>{ch.name}</TooltipContent>
                        </Tooltip>
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
              <VideoPlayer
                src={clientApi.livetvStreamUrl(selectedChannel.id)}
                format={selectedChannel.format || 'hls'}
              />
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
