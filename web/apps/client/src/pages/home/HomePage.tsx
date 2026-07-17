import { useEffect, useState } from 'react'
import { Link } from 'react-router'
import type { Category, ClientBanner, VideoListItem } from '@orange-tv/shared'
import { clientApi, errorMessage } from '@/lib/api'
import { BannerCarousel, VideoGrid } from '@/components/common'
import { Button } from '@/components/ui/button'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Empty, EmptyDescription, EmptyHeader, EmptyTitle } from '@/components/ui/empty'
import { AlertCircleIcon } from 'lucide-react'

export default function HomePage() {
  const [categories, setCategories] = useState<Category[]>([])
  const [videos, setVideos] = useState<VideoListItem[]>([])
  const [hot, setHot] = useState<VideoListItem[]>([])
  const [banners, setBanners] = useState<ClientBanner[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    void (async () => {
      setLoading(true)
      try {
        const [cats, list, hotList, bannerRes] = await Promise.all([
          clientApi.categories(),
          clientApi.videos({ page: 1, page_size: 24, sort: 'created_at_desc' }),
          clientApi.videos({ page: 1, page_size: 6, sort: 'rating_desc' }),
          clientApi.banners().catch(() => ({ data: [] as ClientBanner[] })),
        ])
        setCategories(cats.data || [])
        setVideos(list.data.list || [])
        setHot(hotList.data.list || [])
        setBanners(bannerRes.data || [])
      } catch (err) {
        setError(errorMessage(err))
      } finally {
        setLoading(false)
      }
    })()
  }, [])

  const roots = categories.filter((c) => !c.parent_id || c.parent_id === 0)

  return (
    <div className="flex flex-col gap-8">
      {banners.length ? (
        <BannerCarousel banners={banners} />
      ) : null}

      {error ? (
        <Alert variant="destructive">
          <AlertCircleIcon />
          <AlertTitle>加载失败</AlertTitle>
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      ) : null}

      <section className="flex flex-col gap-3">
        <h2 className="text-lg font-semibold">分类入口</h2>
        <div className="flex flex-wrap gap-2">
          {roots.map((c) => (
            <Button key={c.id} variant="outline" size="sm" nativeButton={false} render={<Link to={`/category?category_id=${c.id}`} />}>
              {c.name}
            </Button>
          ))}
          {!roots.length && !loading ? (
            <span className="text-sm text-muted-foreground">暂无分类</span>
          ) : null}
        </div>
      </section>

      {hot.length ? (
        <section className="flex flex-col gap-4">
          <h2 className="text-lg font-semibold">高分推荐</h2>
          <VideoGrid items={hot} />
        </section>
      ) : null}

      <section className="flex flex-col gap-4">
        <h2 className="text-lg font-semibold">最新上架</h2>
        {loading ? (
          <VideoGrid items={[]} loading />
        ) : !videos.length ? (
          <Empty>
            <EmptyHeader>
              <EmptyTitle>暂无上架影视</EmptyTitle>
              <EmptyDescription>暂无最新上架视频</EmptyDescription>
            </EmptyHeader>
          </Empty>
        ) : (
          <VideoGrid items={videos} />
        )}
      </section>
    </div>
  )
}
