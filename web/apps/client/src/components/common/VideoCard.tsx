import { useState } from 'react'
import { Link } from 'react-router'
import type { ClientVideoListItem } from '@orange-tv/shared'
import { Card, CardContent, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { FilmIcon } from 'lucide-react'

export function VideoCard({ item }: { item: ClientVideoListItem }) {
  const [error, setError] = useState(false)
  const hasCover = item.cover && !error
  const metaParts = [item.year ? String(item.year) : null, item.region || null].filter(
    (v): v is string => v !== null,
  )
  const tags = (item.tags || []).slice(0, 2)
  return (
    <Link to={`/video/${item.id}`} className="block h-full cursor-pointer">
      <Card
        size="sm"
        className="h-full gap-0 overflow-hidden pb-2 pt-0 transition-all hover:ring-primary/40 hover:transition-all"
      >
        <div className="relative flex aspect-[2/3] w-full items-center justify-center bg-muted">
          {hasCover ? (
            <img
              src={item.cover}
              alt={item.title}
              className="absolute inset-0 size-full object-cover"
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
        <CardContent className="flex flex-col gap-0.5 p-2">
          <CardTitle className="truncate">{item.title}</CardTitle>
          <p className="truncate text-xs text-muted-foreground">
            {metaParts.length > 0 ? metaParts.join(' · ') : '\u00A0'}
          </p>
          <div className="flex h-5 flex-wrap gap-1 overflow-hidden">
            {tags.map((tag) => (
              <Badge key={tag.id} variant="secondary">
                {tag.name}
              </Badge>
            ))}
          </div>
        </CardContent>
      </Card>
    </Link>
  )
}
