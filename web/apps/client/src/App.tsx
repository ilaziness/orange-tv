import { useEffect, useMemo, useState } from 'react'
import { BrowserRouter, Link, Route, Routes, useNavigate, useParams, useSearchParams } from 'react-router-dom'
import type { Category, VideoDetail, VideoListItem } from '@orange-tv/shared'
import { clientApi, errorMessage } from './lib/api'
import './App.css'

function Layout({ children }: { children: React.ReactNode }) {
  const [keyword, setKeyword] = useState('')
  const navigate = useNavigate()
  return (
    <div className="shell">
      <header className="nav">
        <Link to="/" className="brand">ORANGE TV</Link>
        <div className="nav-links">
          <Link to="/">首页</Link>
          <Link to="/category">分类</Link>
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
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    void (async () => {
      setLoading(true)
      try {
        const [cats, list] = await Promise.all([
          clientApi.categories(),
          clientApi.videos({ page: 1, page_size: 24, sort: 'created_at_desc' }),
        ])
        setCategories(cats.data || [])
        setVideos(list.data.list || [])
      } catch (err) {
        setError(errorMessage(err))
      } finally {
        setLoading(false)
      }
    })()
  }, [])

  const roots = categories.filter((c) => !c.parent_id || c.parent_id === 0)

  return (
    <>
      <section className="hero">
        <h1>发现精彩影视内容</h1>
        <p className="muted">支持分类浏览、筛选与详情查看。播放器将在后续阶段接入。</p>
      </section>
      {error ? <p className="error">{error}</p> : null}
      <div className="section-title"><h2>分类入口</h2></div>
      <div className="chips">
        {roots.map((c) => (
          <Link key={c.id} className="chip" to={`/category?category_id=${c.id}`}>{c.name}</Link>
        ))}
        {!roots.length && !loading ? <span className="muted">暂无分类</span> : null}
      </div>
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
          ? await clientApi.search(keyword, page)
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
    // Only reset page when filters change; keep page when navigating pages.
    if (key !== 'page') {
      next.delete('page')
    }
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
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    void (async () => {
      setLoading(true)
      try {
        const res = await clientApi.video(Number(id))
        setDetail(res.data)
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
        <div className="sources">
          <h3>播放源 / 剧集</h3>
          {!detail.sources?.length ? <div className="empty">暂无播放源</div> : null}
          {detail.sources?.map((src) => (
            <div className="source-block" key={src.id}>
              <strong>{src.name}</strong>
              <div className="episodes">
                {src.episodes.map((ep) => (
                  <span className="episode" key={`${src.id}-${ep.episode}`} title={ep.url}>
                    {ep.title || `第${ep.episode}集`} · {ep.format}
                  </span>
                ))}
              </div>
            </div>
          ))}
          <p className="muted">当前阶段仅展示播放地址信息，真实播放器在第三阶段接入。</p>
        </div>
      </div>
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
        </Routes>
      </Layout>
    </BrowserRouter>
  )
}
