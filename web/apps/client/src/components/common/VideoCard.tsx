import { Link } from 'react-router'
import type { VideoListItem } from '@orange-tv/shared'
import { Card } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'

export function VideoCard({ item }: { item: VideoListItem }) {
  return (
    <Link to={`/video/${item.id}`}>
      <Card className="overflow-hidden transition-all hover:ring-primary/40 hover:transition-all">
        <div
          className="aspect-[2/3] w-full bg-cover bg-center"
          style={item.cover ? { backgroundImage: `url(${item.cover})` } : undefined}
        >
          {item.rating ? (
            <div className="flex p-2">
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
