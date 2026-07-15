import { useEffect, useMemo, useState } from 'react'
import { BrowserRouter, Link, Route, Routes, useNavigate, useParams, useSearchParams } from 'react-router-dom'
import type { Category, LiveChannel, VideoDetail, VideoListItem } from '@orange-tv/shared'
import { clientApi, errorMessage } from './lib/api'
import { VideoPlayer } from './components/Player'
import './App.css'

function applyThemeVars(config: Record<string, unknown>, customCss?: string) {
  const root = document.documentElement
  const map: Record<string, string> = {
    primary_color: '--theme-primary',
    secondary_color: '--theme-secondary',
    background_color: '--theme-bg',
    text_color: '--theme-text',
    header_height: '--theme-header-height',
  }
  Object.entries(map).forEach(([k, cssVar]) => {
    const v = config[k]
    if (typeof v === 'string' && v) root.style.setProperty(cssVar, v)
  })
  let styleEl = document.getElementById('orange-theme-custom') as HTMLStyleElement | null
  if (!styleEl) {
    styleEl = document.createElement('style')
    styleEl.id = 'orange-theme-custom'
    document.head.appendChild(styleEl)
  }
  styleEl.textContent = customCss || ''
}

function Layout({ children }: { children: React.ReactNode }) {
  const [keyword, setKeyword] = useState('')
  const navigate = useNavigate()
  useEffect(() => {
    void clientApi.themeCurrent().then((res) => {
      applyThemeVars(res.data?.config || {}, res.data?.custom_css)
    }).catch(() => undefined)
  }, [])
  return (
    <div className="shell">
      <header className="nav">
        <Link to="/" className="brand">ORANGE TV</Link>
        <div className="nav-links">
          <Link to="/">首页</Link>
          <Link to="/category">分类</Link>
          <Link to="/live">直播</Link>
          <form className="search-box" onSubmit={(e) => {
            e.preventDefault()
            if (keyword.trim()) navigate(`/category?keyword=${encodeURIComponent(keyword.trim())}`)
          }}>
            <input placeholder="搜索影视" value={keyword} onChange={(e) => setKeyword(e.target.value)} />
            <button className="primary" type="submit">搜索</button>
          </form>
        </div>
      </header>
      <div className="container">{children}</div>
    </div>
  )
}

function VideoCard({ item }: { item: VideoListItem }) {
  return (
    <Link className="card" to={`/video/${item.id}`}>
      <div className="cover" style={{ backgroundImage: item.cover ? `url(${item.cover})` : undefined }}>
        {item.rating ? <span>{item.rating.toFixed(1)}</span> : null}
      </div>
      <div className="card-body">
        <h3 title={item.title}>{item.title}</h3>
        <div className="meta">{item.year || '未知年份'} · {item.region || '未知地区'}</div>
      </div>
    </Link>
  )
}

function HomePage() {
  const [categories, setCategories] = useState<Category[]>([])
  const [videos, setVideos] = useState<VideoListItem[]>([])
  const [hot, setHot] = useState<VideoListItem[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    void (async () => {
      setLoading(true)
      try {
        const [cats, list, hotList] = await Promise.all([
          clientApi.categories(),
          clientApi.videos({ page: 1, page_size: 24, sort: 'created_at_desc' }),
          clientApi.videos({ page: 1, page_size: 6, sort: 'rating_desc' }),
        ])
        setCategories(cats.data || [])
        setVideos(list.data.list || [])
        setHot(hotList.data.list || [])
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
      {error ? <p className="error">{error}</p> : null}
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
      {!loading && !videos.length ? <div className="empty">暂无上架影视</div> : null}
      <div className="grid">
        {videos.map((item) => <VideoCard key={item.id} item={item} />)}
      </div>
    </>
  )
}

function CategoryPage() {
  const [params, setParams] = useSearchParams()
  const [categories, setCategories] = useState<Category[]>([])
  const [videos, setVideos] = useState<VideoListItem[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  const categoryId = Number(params.get('category_id') || 0)
  const year = Number(params.get('year') || 0)
  const region = params.get('region') || ''
  const language = params.get('language') || ''
  const sort = params.get('sort') || 'created_at_desc'
  const keyword = params.get('keyword') || ''
  const page = Number(params.get('page') || 1)

  useEffect(() => {
    void clientApi.categories().then((res) => setCategories(res.data || [])).catch(() => undefined)
  }, [])

  useEffect(() => {
    void (async () => {
      setLoading(true)
      setError('')
      try {
        const res = keyword
          ? await clientApi.search(keyword, page, {
              category_id: categoryId || undefined,
              year: year || undefined,
              region: region || undefined,
              language: language || undefined,
              sort,
            })
          : await clientApi.videos({
              category_id: categoryId || undefined,
              year: year || undefined,
              region: region || undefined,
              language: language || undefined,
              sort,
              page,
              page_size: 20,
            })
        setVideos(res.data.list || [])
        setTotal(res.data.total || 0)
      } catch (err) {
        setError(errorMessage(err))
      } finally {
        setLoading(false)
      }
    })()
  }, [categoryId, year, region, language, sort, keyword, page])

  const roots = useMemo(() => categories.filter((c) => !c.parent_id || c.parent_id === 0), [categories])

  function update(key: string, value: string) {
    const next = new URLSearchParams(params)
    if (!value) next.delete(key)
    else next.set(key, value)
    if (key !== 'page') next.delete('page')
    setParams(next)
  }

  return (
    <>
      <div className="section-title">
        <h1>{keyword ? `搜索：${keyword}` : '分类浏览'}</h1>
        <span className="muted">共 {total} 部</span>
      </div>
      {error ? <p className="error">{error}</p> : null}
      <div className="chips">
        <button className={`chip ${!categoryId ? 'active' : ''}`} onClick={() => update('category_id', '')}>全部</button>
        {roots.map((c) => (
          <button key={c.id} className={`chip ${categoryId === c.id ? 'active' : ''}`} onClick={() => update('category_id', String(c.id))}>
            {c.name}
          </button>
        ))}
      </div>
      <div className="filters">
        <input placeholder="地区" value={region} onChange={(e) => update('region', e.target.value)} />
        <input placeholder="语言" value={language} onChange={(e) => update('language', e.target.value)} />
        <input placeholder="年份" type="number" value={year || ''} onChange={(e) => update('year', e.target.value)} />
        <select value={sort} onChange={(e) => update('sort', e.target.value)}>
          <option value="created_at_desc">最新</option>
          <option value="rating_desc">评分最高</option>
          <option value="year_desc">年份最新</option>
          <option value="view_count_desc">播放最多</option>
        </select>
      </div>
      {loading ? <div className="skeleton" /> : null}
      {!loading && !videos.length ? <div className="empty">没有符合条件的影视</div> : null}
      <div className="grid">
        {videos.map((item) => <VideoCard key={item.id} item={item} />)}
      </div>
      <div className="filters" style={{ marginTop: '1rem' }}>
        <button disabled={page <= 1} onClick={() => update('page', String(page - 1))}>上一页</button>
        <span className="muted">第 {page} 页</span>
        <button disabled={videos.length < 20} onClick={() => update('page', String(page + 1))}>下一页</button>
      </div>
    </>
  )
}

function DetailPage() {
  const { id } = useParams()
  const [detail, setDetail] = useState<VideoDetail | null>(null)
  const [related, setRelated] = useState<VideoListItem[]>([])
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    void (async () => {
      setLoading(true)
      try {
        const vid = Number(id)
        const [res, rel] = await Promise.all([
          clientApi.video(vid),
          clientApi.related(vid).catch(() => ({ data: [] as VideoListItem[] })),
        ])
        setDetail(res.data)
        setRelated(rel.data || [])
      } catch (err) {
        setError(errorMessage(err))
      } finally {
        setLoading(false)
      }
    })()
  }, [id])

  if (loading) return <div className="skeleton" />
  if (error) return <p className="error">{error}</p>
  if (!detail) return <div className="empty">内容不存在</div>

  const first = detail.sources?.[0]?.episodes?.[0]

  return (
    <div className="detail">
      <div className="poster" style={{ backgroundImage: (detail.poster || detail.cover) ? `url(${detail.poster || detail.cover})` : undefined }} />
      <div>
        <h1>{detail.title}</h1>
        <p className="muted">{detail.subtitle}</p>
        <div className="meta">
          {[detail.year, detail.region, detail.language, detail.duration ? `${detail.duration} 分钟` : '', detail.rating ? `评分 ${detail.rating}` : '']
            .filter(Boolean)
            .join(' · ')}
        </div>
        <div className="tags">
          {detail.directors.map((d) => <span className="tag" key={`d-${d.id}`}>导演 {d.name}</span>)}
          {detail.actors.map((a) => <span className="tag" key={`a-${a.id}`}>{a.name}{a.role ? `/${a.role}` : ''}</span>)}
          {detail.tags.map((t) => <span className="tag" key={`t-${t.id}`}>{t.name}</span>)}
        </div>
        <p>{detail.description || '暂无简介'}</p>
        {first ? (
          <div className="toolbar" style={{ marginBottom: '1rem' }}>
            <Link to={`/play/${detail.id}?source=${detail.sources[0].id}&ep=${first.episode}`}>
              <button className="primary">立即播放</button>
            </Link>
          </div>
        ) : null}
        <div className="sources">
          <h3>播放源 / 剧集</h3>
          {!detail.sources?.length ? <div className="empty">暂无播放源</div> : null}
          {detail.sources?.map((src) => (
            <div className="source-block" key={src.id}>
              <strong>{src.name}</strong>
              <div className="episodes">
                {src.episodes.map((ep) => (
                  <Link
                    className="episode"
                    key={`${src.id}-${ep.episode}`}
                    to={`/play/${detail.id}?source=${src.id}&ep=${ep.episode}`}
                  >
                    {ep.title || `第${ep.episode}集`} · {ep.format}
                  </Link>
                ))}
              </div>
            </div>
          ))}
        </div>
        {related.length ? (
          <>
            <div className="section-title" style={{ marginTop: '1.5rem' }}><h2>相关推荐</h2></div>
            <div className="grid">
              {related.map((item) => <VideoCard key={`r-${item.id}`} item={item} />)}
            </div>
          </>
        ) : null}
      </div>
    </div>
  )
}

function PlayPage() {
  const { id } = useParams()
  const [params, setParams] = useSearchParams()
  const [detail, setDetail] = useState<VideoDetail | null>(null)
  const [error, setError] = useState('')
  const sourceId = Number(params.get('source') || 0)
  const epNo = Number(params.get('ep') || 1)

  useEffect(() => {
    setDetail(null)
    setError('')
    void clientApi.video(Number(id)).then((res) => {
      setDetail(res.data)
      if (!params.get('source') && res.data.sources?.[0]) {
        const src = res.data.sources[0]
        const ep = src.episodes[0]
        setParams({ source: String(src.id), ep: String(ep?.episode || 1) }, { replace: true })
      }
    }).catch((err) => setError(errorMessage(err)))
  }, [id])

  if (error) return <p className="error">{error}</p>
  if (!detail) return <div className="skeleton" />

  const source = detail.sources?.find((s) => s.id === sourceId) || detail.sources?.[0]
  const episode = source?.episodes.find((e) => e.episode === epNo) || source?.episodes[0]
  if (!source || !episode) return <div className="empty">暂无可播放内容</div>

  return (
    <div className="stack player-page">
      <div className="section-title">
        <h1>{detail.title} · {episode.title || `第${episode.episode}集`}</h1>
        <Link to={`/video/${detail.id}`} className="muted">返回详情</Link>
      </div>
      <VideoPlayer
        key={`${source.id}-${episode.episode}-${episode.url}`}
        src={episode.url}
        format={episode.format}
        poster={detail.poster || detail.cover}
        storageKey={`play:${detail.id}:${source.id}:${episode.episode}`}
      />
      <div className="sources">
        <h3>播放源</h3>
        <div className="chips">
          {detail.sources.map((s) => (
            <button
              key={s.id}
              className={`chip ${s.id === source.id ? 'active' : ''}`}
              onClick={() => setParams({ source: String(s.id), ep: String(s.episodes[0]?.episode || 1) })}
            >
              {s.name}
            </button>
          ))}
        </div>
        <h3>选集</h3>
        <div className="episodes">
          {source.episodes.map((ep) => (
            <button
              key={ep.episode}
              className={`episode ${ep.episode === episode.episode ? 'active' : ''}`}
              onClick={() => setParams({ source: String(source.id), ep: String(ep.episode) })}
            >
              {ep.title || `第${ep.episode}集`}
            </button>
          ))}
        </div>
      </div>
    </div>
  )
}

function LivePage() {
  const [list, setList] = useState<LiveChannel[]>([])
  const [category, setCategory] = useState('')
  const [current, setCurrent] = useState<LiveChannel | null>(null)
  const [error, setError] = useState('')

  useEffect(() => {
    void clientApi.live({ page: 1, page_size: 100, category: category || undefined })
      .then((res) => {
        const items = res.data.list || []
        setList(items)
        setCurrent((prev) => {
          if (prev && items.some((it) => it.id === prev.id)) return prev
          return items[0] || null
        })
      })
      .catch((err) => setError(errorMessage(err)))
  }, [category])

  return (
    <div className="stack">
      <div className="section-title"><h1>直播频道</h1></div>
      {error ? <p className="error">{error}</p> : null}
      <div className="filters">
        <input placeholder="分类筛选" value={category} onChange={(e) => setCategory(e.target.value)} />
      </div>
      {current ? (
        <VideoPlayer
          key={current.id}
          src={current.stream_url}
          format="hls"
          poster={current.logo}
          autoplay
        />
      ) : null}
      <div className="grid">
        {list.map((ch) => (
          <button key={ch.id} className={`card live-card ${current?.id === ch.id ? 'active' : ''}`} onClick={() => setCurrent(ch)}>
            <div className="cover" style={{ backgroundImage: ch.logo ? `url(${ch.logo})` : undefined }} />
            <div className="card-body">
              <h3>{ch.name}</h3>
              <div className="meta">{ch.category || '未分类'}</div>
            </div>
          </button>
        ))}
      </div>
      {!list.length ? <div className="empty">暂无直播频道</div> : null}
    </div>
  )
}

export default function App() {
  return (
    <BrowserRouter>
      <Layout>
        <Routes>
          <Route path="/" element={<HomePage />} />
          <Route path="/category" element={<CategoryPage />} />
          <Route path="/video/:id" element={<DetailPage />} />
          <Route path="/play/:id" element={<PlayPage />} />
          <Route path="/live" element={<LivePage />} />
        </Routes>
      </Layout>
    </BrowserRouter>
  )
}
