import { useEffect, useMemo, useState } from 'react'
import { BrowserRouter, Link, Route, Routes, useNavigate, useParams, useSearchParams } from 'react-router-dom'
import type { Category, ClientBanner, CommentItem, FavoriteItem, HistoryItem, LiveChannel, UserProfile, VideoDetail, VideoListItem } from '@orange-tv/shared'
import { clientApi, errorMessage, getToken, setToken } from './lib/api'
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
  const [user, setUser] = useState<UserProfile | null>(null)
  useEffect(() => {
    void clientApi.themeCurrent().then((res) => {
      applyThemeVars(res.data?.config || {}, res.data?.custom_css)
    }).catch(() => undefined)
    if (getToken()) {
      void clientApi.profile().then((res) => setUser(res.data || null)).catch(() => { setToken(null); setUser(null) })
    }
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
          {user ? (
            <>
              <Link to="/favorites">收藏</Link>
              <Link to="/history">历史</Link>
              <span className="nav-user">{user.username}</span>
              <button onClick={() => { setToken(null); setUser(null); navigate('/') }}>退出</button>
            </>
          ) : (
            <>
              <Link to="/login">登录</Link>
              <Link to="/register">注册</Link>
            </>
          )}
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
  const [banners, setBanners] = useState<ClientBanner[]>([])
  const [bannerIdx, setBannerIdx] = useState(0)
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

  // Banner auto-rotate
  useEffect(() => {
    if (banners.length <= 1) return
    const t = setInterval(() => setBannerIdx((i) => (i + 1) % banners.length), 5000)
    return () => clearInterval(t)
  }, [banners.length])

  const roots = categories.filter((c) => !c.parent_id || c.parent_id === 0)
  const banner = hot[0]
  const activeBanner = banners[bannerIdx]

  return (
    <>
      {activeBanner ? (
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
  const [favLoading, setFavLoading] = useState(false)
  const [comments, setComments] = useState<CommentItem[]>([])
  const [commentText, setCommentText] = useState('')
  const [commentError, setCommentError] = useState('')

  async function loadComments(vid: number) {
    try {
      const res = await clientApi.listComments(vid)
      setComments(res.data.list || [])
    } catch { /* ignore */ }
  }

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
        void loadComments(vid)
      } catch (err) {
        setError(errorMessage(err))
      } finally {
        setLoading(false)
      }
    })()
  }, [id])

  async function toggleFav() {
    if (!detail) return
    setFavLoading(true)
    try {
      await clientApi.addFavorite(detail.id)
      alert('已收藏')
    } catch (err) {
      // If already favorited, try to remove
      try {
        await clientApi.removeFavorite(detail.id)
        alert('已取消收藏')
      } catch (err2) {
        alert(errorMessage(err2))
      }
    } finally {
      setFavLoading(false)
    }
  }

  async function postComment(e: React.FormEvent) {
    e.preventDefault()
    if (!detail || !commentText.trim()) return
    setCommentError('')
    try {
      const res = await clientApi.createComment(detail.id, commentText.trim())
      setComments((prev) => [res.data!, ...prev])
      setCommentText('')
    } catch (err) {
      setCommentError(errorMessage(err))
    }
  }

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
            <button disabled={favLoading} onClick={toggleFav}>{favLoading ? '处理中...' : '收藏'}</button>
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

        {/* Comments section (C6) */}
        <div className="comments-section" style={{ marginTop: '1.5rem' }}>
          <h3>评论</h3>
          <form className="comment-form" onSubmit={postComment}>
            <textarea placeholder="写下你的评论..." value={commentText} onChange={(e) => setCommentText(e.target.value)} maxLength={500} />
            <button type="submit" className="primary" disabled={!commentText.trim()}>发表</button>
          </form>
          {commentError ? <p className="error">{commentError}</p> : null}
          {!comments.length ? <p className="muted">暂无评论，快来抢沙发吧</p> : null}
          <div className="comment-list">
            {comments.map((c) => (
              <div key={c.id} className="comment-item">
                <div className="comment-avatar">{c.avatar ? <img src={c.avatar} alt="" /> : <span>{c.username[0]}</span>}</div>
                <div className="comment-body">
                  <strong>{c.username}</strong>
                  <span className="muted"> · {new Date(c.created_at).toLocaleString()}</span>
                  <p>{c.content}</p>
                </div>
              </div>
            ))}
          </div>
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

  // Report play history when source/episode changes
  useEffect(() => {
    if (!detail || !sourceId || !epNo) return
    void clientApi.upsertHistory({
      video_id: Number(id),
      play_source_id: sourceId,
      episode_id: epNo,
      progress: 0,
      duration: 0,
    }).catch(() => undefined)
  }, [id, sourceId, epNo, detail])

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

// ===== Phase 5: User auth / favorites / history / comments =====

function LoginPage() {
  const navigate = useNavigate()
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      const res = await clientApi.login(username, password)
      setToken(res.data?.access_token || null)
      navigate('/', { replace: true })
    } catch (err) {
      setError(errorMessage(err))
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="auth-page">
      <form className="auth-card" onSubmit={onSubmit}>
        <h2>用户登录</h2>
        {error ? <p className="error">{error}</p> : null}
        <label>用户名<input value={username} onChange={(e) => setUsername(e.target.value)} required minLength={3} /></label>
        <label>密码<input type="password" value={password} onChange={(e) => setPassword(e.target.value)} required minLength={6} /></label>
        <button className="primary" disabled={loading}>{loading ? '登录中...' : '登录'}</button>
        <p className="muted"><Link to="/register">没有账号？去注册</Link></p>
      </form>
    </div>
  )
}

function RegisterPage() {
  const navigate = useNavigate()
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [email, setEmail] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      await clientApi.register(username, password, email || undefined)
      // Auto login after register
      const res = await clientApi.login(username, password)
      setToken(res.data?.access_token || null)
      navigate('/', { replace: true })
    } catch (err) {
      setError(errorMessage(err))
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="auth-page">
      <form className="auth-card" onSubmit={onSubmit}>
        <h2>用户注册</h2>
        {error ? <p className="error">{error}</p> : null}
        <label>用户名<input value={username} onChange={(e) => setUsername(e.target.value)} required minLength={3} /></label>
        <label>密码<input type="password" value={password} onChange={(e) => setPassword(e.target.value)} required minLength={6} /></label>
        <label>邮箱（选填）<input value={email} onChange={(e) => setEmail(e.target.value)} /></label>
        <button className="primary" disabled={loading}>{loading ? '注册中...' : '注册'}</button>
        <p className="muted"><Link to="/login">已有账号？去登录</Link></p>
      </form>
    </div>
  )
}

function FavoritesPage() {
  const [list, setList] = useState<FavoriteItem[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(true)

  async function load(p = page) {
    setLoading(true)
    setError('')
    try {
      const res = await clientApi.listFavorites(p)
      setList(res.data.list || [])
      setTotal(res.data.total || 0)
      setPage(res.data.page || p)
    } catch (err) {
      setError(errorMessage(err))
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { void load(1) }, [])

  return (
    <>
      <div className="section-title"><h2>我的收藏</h2></div>
      {error ? <p className="error">{error}</p> : null}
      {loading ? <div className="skeleton" /> : null}
      {!loading && !list.length ? <div className="empty">暂无收藏</div> : null}
      <div className="grid">
        {list.map((f) => (
          <Link key={f.video_id} className="card" to={`/video/${f.video_id}`}>
            <div className="cover" style={{ backgroundImage: f.cover ? `url(${f.cover})` : undefined }}>
              {f.rating ? <span>{f.rating.toFixed(1)}</span> : null}
            </div>
            <div className="card-body">
              <h3 title={f.title}>{f.title}</h3>
              <p className="muted">{f.year || ''}</p>
            </div>
          </Link>
        ))}
      </div>
      {total > 20 ? (
        <div className="pager">
          <button disabled={page <= 1} onClick={() => void load(page - 1)}>上一页</button>
          <span className="muted">第 {page} 页</span>
          <button disabled={list.length < 20} onClick={() => void load(page + 1)}>下一页</button>
        </div>
      ) : null}
    </>
  )
}

function HistoryPage() {
  const [list, setList] = useState<HistoryItem[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(true)

  async function load(p = page) {
    setLoading(true)
    setError('')
    try {
      const res = await clientApi.listHistory(p)
      setList(res.data.list || [])
      setTotal(res.data.total || 0)
      setPage(res.data.page || p)
    } catch (err) {
      setError(errorMessage(err))
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { void load(1) }, [])

  return (
    <>
      <div className="section-title">
        <h2>播放历史</h2>
        {list.length ? <button onClick={async () => { if (confirm('清空全部历史？')) { await clientApi.clearHistory(); void load(1) } }}>清空</button> : null}
      </div>
      {error ? <p className="error">{error}</p> : null}
      {loading ? <div className="skeleton" /> : null}
      {!loading && !list.length ? <div className="empty">暂无播放历史</div> : null}
      <div className="grid">
        {list.map((h) => (
          <Link key={h.video_id} className="card" to={`/play/${h.video_id}`}>
            <div className="cover" style={{ backgroundImage: h.cover ? `url(${h.cover})` : undefined }}>
              {h.progress > 0 && h.duration > 0 ? (
                <span className="progress-bar"><span style={{ width: `${(h.progress / h.duration) * 100}%` }} /></span>
              ) : null}
            </div>
            <div className="card-body">
              <h3 title={h.title}>{h.title}</h3>
              <p className="muted">{new Date(h.last_played_at).toLocaleDateString()}</p>
            </div>
          </Link>
        ))}
      </div>
      {total > 20 ? (
        <div className="pager">
          <button disabled={page <= 1} onClick={() => void load(page - 1)}>上一页</button>
          <span className="muted">第 {page} 页</span>
          <button disabled={list.length < 20} onClick={() => void load(page + 1)}>下一页</button>
        </div>
      ) : null}
    </>
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
          <Route path="/login" element={<LoginPage />} />
          <Route path="/register" element={<RegisterPage />} />
          <Route path="/favorites" element={<FavoritesPage />} />
          <Route path="/history" element={<HistoryPage />} />
        </Routes>
      </Layout>
    </BrowserRouter>
  )
}
