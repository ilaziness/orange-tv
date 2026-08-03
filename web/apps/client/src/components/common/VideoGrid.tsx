import type { ClientVideoListItem } from '@orange-tv/shared'
import { VideoCard } from '@/components/common/VideoCard'
import { Skeleton } from '@/components/ui/skeleton'
import { Card } from '@/components/ui/card'

export function VideoGrid({ items, loading }: { items: ClientVideoListItem[]; loading?: boolean }) {
  if (loading) {
    return (
      <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 2xl:grid-cols-8">
        {Array.from({ length: 12 }).map((_, i) => (
          <Card key={i} className="gap-0 overflow-hidden pb-2 pt-0">
            <Skeleton className="aspect-[2/3] w-full rounded-t-xl" />
            <div className="flex flex-col gap-0.5 p-2">
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
