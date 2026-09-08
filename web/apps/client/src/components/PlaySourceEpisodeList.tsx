import { useEffect, useState } from 'react'
import type { VideoDetailSourceGroup } from '@orange-tv/shared'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from '@/components/ui/collapsible'
import { cn } from '@/lib/utils'
import { ChevronDownIcon } from 'lucide-react'

type PlaySourceEpisodeListProps = {
  sources: VideoDetailSourceGroup[]
  currentSourceId: number
  currentEpisodeId: number
  onSelectEpisode: (sourceId: number, episodeId: number) => void
}

export function PlaySourceEpisodeList({
  sources,
  currentSourceId,
  currentEpisodeId,
  onSelectEpisode,
}: PlaySourceEpisodeListProps) {
  const [expandedIds, setExpandedIds] = useState<Set<number>>(() => new Set([currentSourceId]))

  useEffect(() => {
    setExpandedIds((prev) => {
      if (prev.has(currentSourceId)) return prev
      const next = new Set(prev)
      next.add(currentSourceId)
      return next
    })
  }, [currentSourceId])

  const toggleSource = (sourceId: number, open: boolean) => {
    setExpandedIds((prev) => {
      const next = new Set(prev)
      if (open) {
        next.add(sourceId)
      } else {
        next.delete(sourceId)
      }
      return next
    })
  }

  return (
    <div className="flex flex-col gap-4">
      {sources.map((source) => {
        const isExpanded = expandedIds.has(source.id)
        return (
          <Collapsible
            key={source.id}
            open={isExpanded}
            onOpenChange={(open) => toggleSource(source.id, open)}
          >
            <Card>
              <CardHeader>
                <CardTitle className="w-full">
                  <CollapsibleTrigger
                    render={
                      <Button variant="ghost" className="h-auto w-full justify-between gap-2 px-0">
                        {source.name}
                        <ChevronDownIcon
                          data-icon="inline-end"
                          className={cn('transition-transform', isExpanded && 'rotate-180')}
                        />
                      </Button>
                    }
                  />
                </CardTitle>
              </CardHeader>
              <CollapsibleContent>
                <CardContent>
                  <div className="flex flex-wrap gap-2">
                    {source.episodes.map((ep) => (
                      <Button
                        key={ep.id}
                        variant={
                          source.id === currentSourceId && ep.id === currentEpisodeId
                            ? 'default'
                            : 'outline'
                        }
                        size="sm"
                        onClick={() => onSelectEpisode(source.id, ep.id)}
                      >
                        {ep.title || `第${ep.episode}集`}
                      </Button>
                    ))}
                  </div>
                </CardContent>
              </CollapsibleContent>
            </Card>
          </Collapsible>
        )
      })}
    </div>
  )
}
