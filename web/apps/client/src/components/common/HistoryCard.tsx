import { useState } from 'react'
import { Link } from 'react-router'
import { Card } from '@/components/ui/card'
import { FilmIcon } from 'lucide-react'
import { formatTime } from '@/lib/playbackHistory'

type Props = {
  videoId: number
  sourceId: number
  episodeId: number
  title: string
  categoryName?: string
  year?: string
  cover?: string
  progress: number
}

export function HistoryCard({ videoId, sourceId, episodeId, title, categoryName, year, cover, progress }: Props) {
  const [error, setError] = useState(false)
  const hasCover = cover && !error
  const to = `/play/${videoId}/${sourceId}/${episodeId}`
  const metaParts = [categoryName || null, year || null].filter((v): v is string => v !== null)
  return (
    <Link to={to} className="block w-28 shrink-0 cursor-pointer">
      <Card className="gap-0 overflow-hidden pb-1.5 pt-0 transition-all hover:ring-primary/40 hover:transition-all">
        <div className="relative flex aspect-[2/3] w-full items-center justify-center bg-muted">
          {hasCover ? (
            <img
              src={cover}
              alt={title}
              className="absolute inset-0 h-full w-full object-cover"
              onError={() => setError(true)}
            />
          ) : (
            <FilmIcon className="size-8 text-muted-foreground/40" />
          )}
          {progress > 0 ? (
            <div className="absolute bottom-0 left-0 right-0 flex bg-black/65 px-2 py-1">
              <span className="truncate text-xs text-white">{formatTime(progress)}</span>
            </div>
          ) : null}
        </div>
        <div className="flex flex-col gap-0.5 p-1.5">
          <h3 className="truncate text-xs font-medium">{title}</h3>
          {metaParts.length > 0 && <p className="truncate text-xs text-muted-foreground">{metaParts.join(' · ')}</p>}
        </div>
      </Card>
    </Link>
  )
}
