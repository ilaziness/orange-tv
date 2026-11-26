import { useEffect, useState } from 'react'
import { Link } from 'react-router'
import type { ClientBanner } from '@orange-tv/shared'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { ChevronLeftIcon, ChevronRightIcon, PlayIcon } from 'lucide-react'

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

  if (!banners.length) return null

  const activeBanner = banners[bannerIdx] ?? banners[0]

  const goPrev = () => setBannerIdx((i) => (i - 1 + banners.length) % banners.length)
  const goNext = () => setBannerIdx((i) => (i + 1) % banners.length)

  return (
    <section className="group relative w-full overflow-hidden rounded-2xl border border-border bg-background shadow-sm aspect-24/9">
      {banners.map((banner, i) => (
        <div
          key={banner.id}
          className={cn(
            'absolute inset-0 transition-opacity duration-700 ease-in-out',
            i === bannerIdx ? 'opacity-100' : 'opacity-0 pointer-events-none',
          )}
        >
          {banner.cover ? (
            <img src={banner.cover} alt={banner.title} className="h-full w-full object-cover" />
          ) : (
            <div className="h-full w-full bg-muted" />
          )}
        </div>
      ))}

      <div className="absolute inset-0 flex flex-col justify-end gap-6 p-6 md:p-10">
        <div
          className={cn('flex max-w-2xl flex-col gap-4', banners.length > 1 && 'max-md:pb-12')}
        >
          <h1 className="text-3xl font-bold tracking-tight text-gray-300 md:text-4xl lg:text-5xl">
            {activeBanner.title}
          </h1>
          <div className="flex flex-wrap gap-3">
            {activeBanner.video_id ? (
              <Button
                nativeButton={false}
                size="lg"
                render={<Link to={`/video/${activeBanner.video_id}`} />}
              >
                <PlayIcon data-icon="inline-start" />
                立即播放
              </Button>
            ) : null}
            {activeBanner.link ? (
              <Button
                nativeButton={false}
                variant="secondary"
                size="lg"
                render={<a href={activeBanner.link} />}
              >
                查看详情
              </Button>
            ) : null}
          </div>
        </div>

        {banners.length > 1 ? (
          <>
            <Button
              type="button"
              variant="outline"
              size="icon-lg"
              className="absolute left-4 top-1/2 z-10 -translate-y-1/2 opacity-0 shadow-sm group-hover:opacity-100"
              aria-label="上一个"
              onClick={goPrev}
            >
              <ChevronLeftIcon />
            </Button>
            <Button
              type="button"
              variant="outline"
              size="icon-lg"
              className="absolute right-4 top-1/2 z-10 -translate-y-1/2 opacity-0 shadow-sm group-hover:opacity-100"
              aria-label="下一个"
              onClick={goNext}
            >
              <ChevronRightIcon />
            </Button>
            <div className="absolute bottom-4 right-4 z-10 flex gap-2 md:bottom-6 md:right-6">
              {banners.map((banner, i) => (
                <button
                  key={banner.id}
                  type="button"
                  className={cn(
                    'h-10 w-16 overflow-hidden rounded-md border-2 transition-all duration-300 md:h-14 md:w-24',
                    i === bannerIdx
                      ? 'border-primary opacity-100 ring-1 ring-primary/40'
                      : 'border-muted-foreground/40 opacity-70 hover:border-muted-foreground/70 hover:opacity-100',
                  )}
                  onClick={() => setBannerIdx(i)}
                  aria-label={banner.title || `Banner ${i + 1}`}
                  aria-current={i === bannerIdx ? 'true' : undefined}
                >
                  {banner.cover ? (
                    <img
                      src={banner.cover}
                      alt=""
                      className="h-full w-full object-cover"
                      draggable={false}
                    />
                  ) : (
                    <div className="h-full w-full bg-muted" />
                  )}
                </button>
              ))}
            </div>
          </>
        ) : null}
      </div>
    </section>
  )
}
