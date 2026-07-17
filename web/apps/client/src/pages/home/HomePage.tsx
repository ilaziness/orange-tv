import { useEffect, useState } from 'react'
import { Link } from 'react-router'
import type { Category, ClientBanner, VideoListItem } from '@orange-tv/shared'
import { clientApi, errorMessage } from '../../lib/api'
import { VideoCard } from '../../components/common/VideoCard'
import { BannerCarousel } from '../../components/common/BannerCarousel'
import { ErrorAlert } from '../../components/ui/ErrorAlert'
import { Empty } from '../../components/ui/Empty'

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
  const banner = hot[0]

  return (
    <>
      {banners.length ? (
        <BannerCarousel banners={banners} />
      ) : (
        <section className="hero banner">
          {banner ? (
            <>
              <div className="banner-cover" style={{ backgroundImage: (banner.poster || banner.cover) ? `url(${banner.poster || banner.cover})` : undefined }} />
              <div className="banner-body">
                <h1>{banner.title}</h1>
                <p className="muted">{banner.subtitle || '高分推荐'}</p>
                <Link to={`/play/${banner.id}`}><button className="primary">立即播放</button></Link>
              </div>
            </>
          ) : (
            <>
              <h1>发现精彩影视内容</h1>
              <p className="muted">支持分类浏览、筛选、详情播放与相关推荐。</p>
            </>
          )}
        </section>
      )}
      <ErrorAlert message={error} />
      <div className="section-title"><h2>分类入口</h2></div>
      <div className="chips">
        {roots.map((c) => (
          <Link key={c.id} className="chip" to={`/category?category_id=${c.id}`}>{c.name}</Link>
        ))}
        {!roots.length && !loading ? <span className="muted">暂无分类</span> : null}
      </div>
      {hot.length ? (
        <>
          <div className="section-title"><h2>高分推荐</h2></div>
          <div className="grid">
            {hot.map((item) => <VideoCard key={`hot-${item.id}`} item={item} />)}
          </div>
        </>
      ) : null}
      <div className="section-title"><h2>最新上架</h2></div>
      {loading ? <div className="skeleton" /> : null}
      {!loading && !videos.length ? <Empty message="暂无上架影视" /> : null}
      <div className="grid">
        {videos.map((item) => <VideoCard key={item.id} item={item} />)}
      </div>
    </>
  )
}
