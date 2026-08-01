import { useState } from 'react'
import { Link } from 'react-router'
import type { VideoListItem } from '@orange-tv/shared'
import { Card } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { FilmIcon } from 'lucide-react'

export function VideoCard({ item }: { item: VideoListItem }) {
  const [error, setError] = useState(false)
  const hasCover = item.cover && !error
  const metaParts = [item.year ? String(item.year) : null, item.region || null].filter(
    (v): v is string => v !== null,
  )
  const tags = (item.tags || []).slice(0, 2)
  return (
    <Link to={`/video/${item.id}`} className="cursor-pointer">
      <Card className="overflow-hidden pt-0 transition-all hover:ring-primary/40 hover:transition-all">
        <div className="relative flex aspect-[2/3] w-full items-center justify-center bg-muted">
          {hasCover ? (
            <img
              src={item.cover}
              alt={item.title}
              className="absolute inset-0 h-full w-full object-cover"
              onError={() => setError(true)}
            />
          ) : (
            <FilmIcon className="size-10 text-muted-foreground/40" />
          )}
          {item.rating ? (
            <div className="absolute top-0 left-0 flex p-2">
              <Badge variant="default" className="bg-black/65 text-white">
                {item.rating.toFixed(1)}
              </Badge>
            </div>
          ) : null}
        </div>
        <div className="flex flex-col gap-1 p-3">
          <h3 className="truncate text-sm font-medium">{item.title}</h3>
          {metaParts.length > 0 && (
            <p className="text-xs text-muted-foreground">{metaParts.join(' · ')}</p>
          )}
          {tags.length > 0 && (
            <div className="flex flex-wrap gap-1">
              {tags.map((tag) => (
                <Badge key={tag.id} variant="secondary">
                  {tag.name}
                </Badge>
              ))}
            </div>
          )}
        </div>
      </Card>
    </Link>
  )
}
