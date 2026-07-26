import { Link } from 'react-router'
import type { VideoListItem } from '@orange-tv/shared'
import { Card } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { FilmIcon } from 'lucide-react'

export function VideoCard({ item }: { item: VideoListItem }) {
  return (
    <Link to={`/video/${item.id}`} className="cursor-pointer">
      <Card className="overflow-hidden pt-0 transition-all hover:ring-primary/40 hover:transition-all">
        <div
          className="relative flex aspect-[2/3] w-full items-center justify-center bg-muted bg-cover bg-center"
          style={item.cover ? { backgroundImage: `url(${item.cover})` } : undefined}
        >
          {!item.cover ? (
            <FilmIcon className="size-10 text-muted-foreground/40" />
          ) : null}
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
          <p className="text-xs text-muted-foreground">
            {item.year || '未知年份'} · {item.region || '未知地区'}
          </p>
        </div>
      </Card>
    </Link>
  )
}
