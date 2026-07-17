import { useEffect, useState } from 'react'
import { Link } from 'react-router'
import type { ClientBanner } from '@orange-tv/shared'

interface BannerCarouselProps {
  banners: ClientBanner[]
}

export function BannerCarousel({ banners }: BannerCarouselProps) {
  const [bannerIdx, setBannerIdx] = useState(0)

  // Banner auto-rotate
  useEffect(() => {
    if (banners.length <= 1) return
    const t = setInterval(() => setBannerIdx((i) => (i + 1) % banners.length), 5000)
    return () => clearInterval(t)
  }, [banners.length])

  const activeBanner = banners[bannerIdx]

  if (!activeBanner) return null

  return (
    <section className="hero banner carousel">
      <div className="banner-cover" style={{ backgroundImage: activeBanner.cover ? `url(${activeBanner.cover})` : undefined }} />
      <div className="banner-body">
        <h1>{activeBanner.title}</h1>
        {activeBanner.link ? <a href={activeBanner.link}><button className="primary">查看详情</button></a> : null}
        {activeBanner.video_id ? <Link to={`/video/${activeBanner.video_id}`}><button className="primary">立即播放</button></Link> : null}
      </div>
      {banners.length > 1 ? (
        <div className="carousel-dots">
          {banners.map((_, i) => (
            <span key={i} className={i === bannerIdx ? 'dot active' : 'dot'} onClick={() => setBannerIdx(i)} />
          ))}
        </div>
      ) : null}
    </section>
  )
}
