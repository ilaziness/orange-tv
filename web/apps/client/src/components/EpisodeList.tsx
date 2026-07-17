import { Button } from '@/components/ui/button'
import type { VideoSourceGroup } from '@orange-tv/shared'

type EpisodeListProps = {
  source: VideoSourceGroup
  currentEpisode?: number
  onSelect: (episode: number) => void
}

export function EpisodeList({ source, currentEpisode, onSelect }: EpisodeListProps) {
  if (!source.episodes?.length) return null

  return (
    <div className="flex flex-wrap gap-2">
      {source.episodes.map((ep) => (
        <Button
          key={ep.episode}
          variant={currentEpisode === ep.episode ? 'default' : 'outline'}
          size="sm"
          onClick={() => onSelect(ep.episode)}
        >
          {ep.title || `第${ep.episode}集`}
        </Button>
      ))}
    </div>
  )
}
