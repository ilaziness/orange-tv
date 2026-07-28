import type { VideoListItem } from '@orange-tv/shared'
import { VideoCard } from '@/components/common/VideoCard'
import { Skeleton } from '@/components/ui/skeleton'
import { Card } from '@/components/ui/card'

export function VideoGrid({ items, loading }: { items: VideoListItem[]; loading?: boolean }) {
  if (loading) {
    return (
      <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 2xl:grid-cols-8">
        {Array.from({ length: 12 }).map((_, i) => (
          <Card key={i} className="overflow-hidden">
            <Skeleton className="aspect-[2/3] w-full rounded-none" />
            <div className="flex flex-col gap-2 p-3">
              <Skeleton className="h-4 w-3/4" />
              <Skeleton className="h-3 w-1/2" />
            </div>
          </Card>
        ))}
      </div>
    )
  }

  return (
    <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 2xl:grid-cols-8">
      {items.map((item) => (
        <VideoCard key={item.id} item={item} />
      ))}
    </div>
  )
}
