import { useEffect, useState } from 'react'
import { Link } from 'react-router'
import type { ClientBanner } from '@orange-tv/shared'
import { Button } from '@/components/ui/button'
import { PlayIcon } from 'lucide-react'

interface BannerCarouselProps {
  banners: ClientBanner[]
}

export function BannerCarousel({ banners }: BannerCarouselProps) {
  const [bannerIdx, setBannerIdx] = useState(0)

  useEffect(() => {
    if (banners.length <= 1) return
    const t = setInterval(() => setBannerIdx((i) => (i + 1) % banners.length), 5000)
    return () => clearInterval(t)
  }, [banners.length])

  const activeBanner = banners[bannerIdx]

  if (!activeBanner) return null

  return (
    <section className="relative flex min-h-64 items-center overflow-hidden rounded-xl border border-border bg-gradient-to-br from-primary/20 via-card to-card p-6">
      <div
        className="absolute inset-0 bg-cover bg-center opacity-30"
        style={activeBanner.cover ? { backgroundImage: `url(${activeBanner.cover})` } : undefined}
      />
      <div className="relative flex flex-col gap-4">
        <h1 className="text-2xl font-bold tracking-tight">{activeBanner.title}</h1>
        <div className="flex gap-2">
          {activeBanner.link ? (
            <Button nativeButton={false} render={<a href={activeBanner.link} />}>
              查看详情
            </Button>
          ) : null}
          {activeBanner.video_id ? (
            <Button nativeButton={false} render={<Link to={`/video/${activeBanner.video_id}`} />}>
              <PlayIcon data-icon="inline-start" />
              立即播放
            </Button>
          ) : null}
        </div>
      </div>
      {banners.length > 1 ? (
        <div className="absolute bottom-4 left-1/2 flex -translate-x-1/2 gap-2">
          {banners.map((_, i) => (
            <button
              key={i}
              className={`size-2.5 rounded-full transition-colors ${i === bannerIdx ? 'bg-primary' : 'bg-muted-foreground/40'}`}
              onClick={() => setBannerIdx(i)}
            >
              <span className="sr-only">Banner {i + 1}</span>
            </button>
          ))}
        </div>
      ) : null}
    </section>
  )
}
