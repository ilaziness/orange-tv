import { useEffect, useMemo, useState } from 'react'
import { BrowserRouter, Link, Navigate, Route, Routes, useLocation, useNavigate, useParams } from 'react-router-dom'
import type { Category, NamedItem, PlaySource, VideoDetail, VideoListItem } from '@orange-tv/shared'
import { adminApi, errorMessage } from './lib/api'
import { useAuthStore } from './store/auth'
import './App.css'

function LoginPage() {
  const navigate = useNavigate()
  const login = useAuthStore((s) => s.login)
  const token = useAuthStore((s) => s.token)
  const [username, setUsername] = useState('admin')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    if (token) navigate('/', { replace: true })
  }, [token, navigate])

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      await login(username, password)
      navigate('/', { replace: true })
    } catch (err) {
      setError(errorMessage(err))
    } finally {
      setLoading(false)
    }
  }

  return (
    <main className="login-page">
      <div className="login-card">
        <h1>Orange TV 管理后台</h1>
        <p className="muted">使用 super_admin 账号登录</p>
        <form onSubmit={onSubmit}>
          <label>
            用户名
            <input value={username} onChange={(e) => setUsername(e.target.value)} required minLength={3} />
          </label>
          <label>
            密码
            <input type="password" value={password} onChange={(e) => setPassword(e.target.value)} required minLength={6} />
          </label>
          {error ? <p className="error">{error}</p> : null}
          <button className="primary" disabled={loading}>{loading ? '登录中...' : '登录'}</button>
        </form>
      </div>
    </main>
  )
}

function RequireAuth({ children }: { children: React.ReactNode }) {
  const token = useAuthStore((s) => s.token)
  const loadProfile = useAuthStore((s) => s.loadProfile)
  const profile = useAuthStore((s) => s.profile)
  const [ready, setReady] = useState(false)

  useEffect(() => {
    if (!token) {
      setReady(true)
      return
    }
    loadProfile().finally(() => setReady(true))
  }, [token, loadProfile])

  if (!ready) return <main className="login-page">加载中...</main>
  if (!token) return <Navigate to="/login" replace />
  if (!profile) return <Navigate to="/login" replace />
  return <>{children}</>
}

function AdminLayout({ children }: { children: React.ReactNode }) {
  const location = useLocation()
  const profile = useAuthStore((s) => s.profile)
  const logout = useAuthStore((s) => s.logout)
  const navigate = useNavigate()
  const side = useMemo(() => {
    if (location.pathname.startsWith('/content')) {
      return [
        { to: '/content/categories', label: '分类管理' },
        { to: '/content/videos', label: '影视管理' },
        { to: '/content/directors', label: '导演管理' },
        { to: '/content/actors', label: '演员管理' },
        { to: '/content/tags', label: '标签管理' },
        { to: '/content/play-sources', label: '播放源管理' },
      ]
    }
    return []
  }, [location.pathname])

  return (
    <div className="layout">
      <header className="topbar">
        <div className="brand">Orange TV Admin</div>
        <nav className="topnav">
          <Link className={location.pathname === '/' ? 'active' : ''} to="/">首页</Link>
          <Link className={location.pathname.startsWith('/content') ? 'active' : ''} to="/content/categories">内容管理</Link>
        </nav>
        <div className="userbox">
          <span>{profile?.username} ({profile?.role})</span>
          <button onClick={async () => { await logout(); navigate('/login') }}>退出</button>
        </div>
      </header>
      <div className="body">
        {side.length ? (
          <aside className="sidebar">
            <nav>
              {side.map((item) => (
                <Link key={item.to} className={location.pathname.startsWith(item.to) ? 'active' : ''} to={item.to}>
                  {item.label}
                </Link>
              ))}
            </nav>
          </aside>
        ) : null}
        <main className="content">{children}</main>
      </div>
    </div>
  )
}

function DashboardPage() {
  return (
    <div className="page-card">
      <div className="page-header">
        <h1>概况</h1>
      </div>
      <p className="muted">第二阶段已接入认证、分类与影视内容管理。请从顶部「内容管理」进入业务页面。</p>
      <div className="toolbar">
        <Link to="/content/videos"><button className="primary">影视管理</button></Link>
        <Link to="/content/categories"><button>分类管理</button></Link>
      </div>
    </div>
  )
}

function flattenCategories(tree: Category[], depth = 0): Array<Category & { depth: number }> {
  const out: Array<Category & { depth: number }> = []
  for (const item of tree) {
    out.push({ ...item, depth })
    if (item.children?.length) out.push(...flattenCategories(item.children, depth + 1))
  }
  return out
}

function CategoriesPage() {
  const [tree, setTree] = useState<Category[]>([])
  const [name, setName] = useState('')
  const [parentId, setParentId] = useState(0)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  async function load() {
    setLoading(true)
    setError('')
    try {
      const res = await adminApi.listCategories()
      setTree(res.data || [])
    } catch (err) {
      setError(errorMessage(err))
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { void load() }, [])

  const flat = flattenCategories(tree)

  async function onCreate(e: React.FormEvent) {
    e.preventDefault()
    setError('')
    try {
      await adminApi.createCategory({ name, parent_id: parentId, status: 1 })
      setName('')
      await load()
    } catch (err) {
      setError(errorMessage(err))
    }
  }

  async function onDelete(id: number) {
    if (!confirm('确认删除该分类？')) return
    try {
      await adminApi.deleteCategory(id)
      await load()
    } catch (err) {
      setError(errorMessage(err))
    }
  }

  async function toggleStatus(item: Category) {
    try {
      await adminApi.updateCategory(item.id, { status: item.status === 1 ? 0 : 1 })
      await load()
    } catch (err) {
      setError(errorMessage(err))
    }
  }

  return (
    <div className="page-card stack">
      <div className="page-header"><h1>分类管理</h1></div>
      {error ? <p className="error">{error}</p> : null}
      <form className="toolbar" onSubmit={onCreate}>
        <input placeholder="分类名称" value={name} onChange={(e) => setName(e.target.value)} required />
        <select value={parentId} onChange={(e) => setParentId(Number(e.target.value))}>
          <option value={0}>无父级</option>
          {flat.map((c) => (
            <option key={c.id} value={c.id}>{'—'.repeat(c.depth)} {c.name}</option>
          ))}
        </select>
        <button className="primary" type="submit">新增分类</button>
        <button type="button" onClick={() => void load()} disabled={loading}>刷新</button>
      </form>
      <div className="tree">
        {flat.map((item) => (
          <div key={item.id} className="tree-item" style={{ marginLeft: item.depth * 16 }}>
            <div>
              <strong>{item.name}</strong>
              <div className="muted">ID {item.id} · 排序 {item.sort_order}</div>
            </div>
            <div className="actions">
              <span className={`badge ${item.status === 1 ? 'ok' : 'off'}`}>{item.status === 1 ? '启用' : '禁用'}</span>
              <button onClick={() => void toggleStatus(item)}>{item.status === 1 ? '禁用' : '启用'}</button>
              <button className="danger" onClick={() => void onDelete(item.id)}>删除</button>
            </div>
          </div>
        ))}
        {!flat.length ? <p className="muted">暂无分类</p> : null}
      </div>
    </div>
  )
}

function NamedResourcePage({
  title,
  list,
  create,
  remove,
}: {
  title: string
  list: (keyword?: string) => Promise<{ data: { list: NamedItem[] } }>
  create: (name: string) => Promise<unknown>
  remove: (id: number) => Promise<unknown>
}) {
  const [items, setItems] = useState<NamedItem[]>([])
  const [keyword, setKeyword] = useState('')
  const [name, setName] = useState('')
  const [error, setError] = useState('')

  async function load(k = keyword) {
    setError('')
    try {
      const res = await list(k)
      setItems(res.data.list || [])
    } catch (err) {
      setError(errorMessage(err))
    }
  }

  useEffect(() => { void load('') }, [])

  return (
    <div className="page-card stack">
      <div className="page-header"><h1>{title}</h1></div>
      {error ? <p className="error">{error}</p> : null}
      <div className="toolbar">
        <input placeholder="搜索" value={keyword} onChange={(e) => setKeyword(e.target.value)} />
        <button onClick={() => void load()}>查询</button>
        <input placeholder="新名称" value={name} onChange={(e) => setName(e.target.value)} />
        <button className="primary" onClick={async () => {
          try {
            await create(name)
            setName('')
            await load()
          } catch (err) {
            setError(errorMessage(err))
          }
        }}>新增</button>
      </div>
      <table className="table">
        <thead><tr><th>ID</th><th>名称</th><th>操作</th></tr></thead>
        <tbody>
          {items.map((item) => (
            <tr key={item.id}>
              <td>{item.id}</td>
              <td>{item.name}</td>
              <td>
                <button className="danger" onClick={async () => {
                  if (!confirm('确认删除？')) return
                  try {
                    await remove(item.id)
                    await load()
                  } catch (err) {
                    setError(errorMessage(err))
                  }
                }}>删除</button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

function PlaySourcesPage() {
  const [items, setItems] = useState<PlaySource[]>([])
  const [name, setName] = useState('')
  const [error, setError] = useState('')

  async function load() {
    try {
      const res = await adminApi.listPlaySources()
      setItems(res.data.list || [])
    } catch (err) {
      setError(errorMessage(err))
    }
  }
  useEffect(() => { void load() }, [])

  return (
    <div className="page-card stack">
      <div className="page-header"><h1>播放源管理</h1></div>
      {error ? <p className="error">{error}</p> : null}
      <div className="toolbar">
        <input placeholder="播放源名称" value={name} onChange={(e) => setName(e.target.value)} />
        <button className="primary" onClick={async () => {
          try {
            await adminApi.createPlaySource({ name, status: 1 })
            setName('')
            await load()
          } catch (err) {
            setError(errorMessage(err))
          }
        }}>新增</button>
      </div>
      <table className="table">
        <thead><tr><th>ID</th><th>名称</th><th>状态</th><th>操作</th></tr></thead>
        <tbody>
          {items.map((item) => (
            <tr key={item.id}>
              <td>{item.id}</td>
              <td>{item.name}</td>
              <td><span className={`badge ${item.status === 1 ? 'ok' : 'off'}`}>{item.status === 1 ? '启用' : '禁用'}</span></td>
              <td className="actions">
                <button onClick={async () => {
                  try {
                    await adminApi.updatePlaySource(item.id, { status: item.status === 1 ? 0 : 1 })
                    await load()
                  } catch (err) {
                    setError(errorMessage(err))
                  }
                }}>{item.status === 1 ? '禁用' : '启用'}</button>
                <button className="danger" onClick={async () => {
                  if (!confirm('确认删除播放源？')) return
                  try {
                    await adminApi.deletePlaySource(item.id)
                    await load()
                  } catch (err) {
                    setError(errorMessage(err))
                  }
                }}>删除</button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

function VideosPage() {
  const [items, setItems] = useState<VideoListItem[]>([])
  const [keyword, setKeyword] = useState('')
  const [error, setError] = useState('')
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)

  async function load(p = page) {
    setError('')
    try {
      const res = await adminApi.listVideos({ keyword, page: p, page_size: 20 })
      setItems(res.data.list || [])
      setTotal(res.data.total || 0)
      setPage(res.data.page || p)
    } catch (err) {
      setError(errorMessage(err))
    }
  }

  useEffect(() => { void load(1) }, [])

  return (
    <div className="page-card stack">
      <div className="page-header">
        <h1>影视管理</h1>
        <Link to="/content/videos/new"><button className="primary">新增影视</button></Link>
      </div>
      {error ? <p className="error">{error}</p> : null}
      <div className="toolbar">
        <input placeholder="关键词" value={keyword} onChange={(e) => setKeyword(e.target.value)} />
        <button onClick={() => void load(1)}>搜索</button>
      </div>
      <table className="table">
        <thead>
          <tr>
            <th>ID</th><th>标题</th><th>年份</th><th>评分</th><th>状态</th><th>操作</th>
          </tr>
        </thead>
        <tbody>
          {items.map((item) => (
            <tr key={item.id}>
              <td>{item.id}</td>
              <td>{item.title}</td>
              <td>{item.year || '-'}</td>
              <td>{item.rating}</td>
              <td><span className={`badge ${item.publish_status === 1 ? 'ok' : 'off'}`}>{item.publish_status === 1 ? '上架' : '下架'}</span></td>
              <td className="actions">
                <Link to={`/content/videos/${item.id}`}><button>编辑</button></Link>
                <button className="danger" onClick={async () => {
                  if (!confirm('确认删除该影视？')) return
                  try {
                    await adminApi.deleteVideo(item.id)
                    await load(page)
                  } catch (err) {
                    setError(errorMessage(err))
                  }
                }}>删除</button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
      <div className="toolbar">
        <span className="muted">共 {total} 条 · 第 {page} 页</span>
        <button disabled={page <= 1} onClick={() => void load(page - 1)}>上一页</button>
        <button disabled={items.length < 20} onClick={() => void load(page + 1)}>下一页</button>
      </div>
    </div>
  )
}

function VideoEditPage() {
  const { id } = useParams()
  const isNew = !id || id === 'new'
  const navigate = useNavigate()
  const [error, setError] = useState('')
  const [categories, setCategories] = useState<Array<Category & { depth: number }>>([])
  const [directors, setDirectors] = useState<NamedItem[]>([])
  const [actors, setActors] = useState<NamedItem[]>([])
  const [tags, setTags] = useState<NamedItem[]>([])
  const [sources, setSources] = useState<PlaySource[]>([])
  const [selectedDirectors, setSelectedDirectors] = useState<number[]>([])
  const [selectedActors, setSelectedActors] = useState<Array<{ actor_id: number; role: string }>>([])
  const [selectedTags, setSelectedTags] = useState<number[]>([])
  const [episodes, setEpisodes] = useState<Array<{ source_id: number; episode_number: number; title: string; play_url: string; format: string }>>([])
  const [form, setForm] = useState({
    title: '',
    subtitle: '',
    description: '',
    category_id: 0,
    publish_status: 0,
    serial_status: 1,
    cover_image: '',
    poster_image: '',
    year: 0,
    region: '',
    language: '',
    duration: 0,
    rating: 0,
    release_date: '',
  })

  useEffect(() => {
    void (async () => {
      try {
        const [cats, dirs, acts, tgs, srcs] = await Promise.all([
          adminApi.listCategories(),
          adminApi.listDirectors(),
          adminApi.listActors(),
          adminApi.listTags(),
          adminApi.listPlaySources(),
        ])
        setCategories(flattenCategories(cats.data || []))
        setDirectors(dirs.data.list || [])
        setActors(acts.data.list || [])
        setTags(tgs.data.list || [])
        setSources(srcs.data.list || [])
        if (!isNew && id) {
          const detail = await adminApi.getVideo(Number(id))
          const d = detail.data
          setForm({
            title: d.title,
            subtitle: d.subtitle,
            description: d.description,
            category_id: d.category_id,
            publish_status: d.publish_status ?? 0,
            serial_status: d.serial_status,
            cover_image: d.cover,
            poster_image: d.poster,
            year: d.year,
            region: d.region,
            language: d.language,
            duration: d.duration,
            rating: d.rating,
            release_date: d.release_date || '',
          })
          setSelectedDirectors(d.directors.map((x) => x.id))
          setSelectedActors(d.actors.map((x) => ({ actor_id: x.id, role: x.role })))
          setSelectedTags(d.tags.map((x) => x.id))
        }
      } catch (err) {
        setError(errorMessage(err))
      }
    })()
  }, [id, isNew])

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError('')
    if (!form.category_id) {
      setError('请选择分类')
      return
    }
    const body = {
      ...form,
      category_id: Number(form.category_id),
      director_ids: selectedDirectors,
      actors: selectedActors,
      tag_ids: selectedTags,
      year: Number(form.year) || 0,
      duration: Number(form.duration) || 0,
      rating: Number(form.rating) || 0,
      publish_status: Number(form.publish_status),
      serial_status: Number(form.serial_status),
    }
    try {
      let videoId = Number(id)
      if (isNew) {
        const res = await adminApi.createVideo(body)
        videoId = (res.data as VideoDetail).id
      } else {
        if (!videoId) {
          setError('无效的影视 ID')
          return
        }
        await adminApi.updateVideo(videoId, body)
      }
      for (const ep of episodes) {
        if (!ep.play_url || !ep.source_id) continue
        await adminApi.createEpisode({
          video_id: videoId,
          source_id: Number(ep.source_id),
          episode_number: Number(ep.episode_number) || 1,
          title: ep.title,
          play_url: ep.play_url,
          format: ep.format || 'hls',
          status: 1,
        })
      }
      navigate('/content/videos')
    } catch (err) {
      setError(errorMessage(err))
    }
  }

  function toggleId(list: number[], id: number): number[] {
    return list.includes(id) ? list.filter((x) => x !== id) : [...list, id]
  }

  return (
    <div className="page-card stack">
      <div className="page-header">
        <h1>{isNew ? '新增影视' : `编辑影视 #${id}`}</h1>
        <Link to="/content/videos"><button>返回列表</button></Link>
      </div>
      {error ? <p className="error">{error}</p> : null}
      <form className="stack" onSubmit={onSubmit}>
        <div className="form-grid">
          <label>标题<input value={form.title} onChange={(e) => setForm({ ...form, title: e.target.value })} required /></label>
          <label>副标题<input value={form.subtitle} onChange={(e) => setForm({ ...form, subtitle: e.target.value })} /></label>
          <label>分类
            <select value={form.category_id} onChange={(e) => setForm({ ...form, category_id: Number(e.target.value) })} required>
              <option value={0}>请选择</option>
              {categories.map((c) => <option key={c.id} value={c.id}>{'—'.repeat(c.depth)} {c.name}</option>)}
            </select>
          </label>
          <label>上下架
            <select value={form.publish_status} onChange={(e) => setForm({ ...form, publish_status: Number(e.target.value) })}>
              <option value={0}>下架</option>
              <option value={1}>上架</option>
            </select>
          </label>
          <label>连载状态
            <select value={form.serial_status} onChange={(e) => setForm({ ...form, serial_status: Number(e.target.value) })}>
              <option value={1}>连载中</option>
              <option value={2}>已完结</option>
              <option value={3}>即将上线</option>
            </select>
          </label>
          <label>年份<input type="number" value={form.year} onChange={(e) => setForm({ ...form, year: Number(e.target.value) })} /></label>
          <label>地区<input value={form.region} onChange={(e) => setForm({ ...form, region: e.target.value })} /></label>
          <label>语言<input value={form.language} onChange={(e) => setForm({ ...form, language: e.target.value })} /></label>
          <label>时长(分钟)<input type="number" value={form.duration} onChange={(e) => setForm({ ...form, duration: Number(e.target.value) })} /></label>
          <label>评分<input type="number" step="0.1" value={form.rating} onChange={(e) => setForm({ ...form, rating: Number(e.target.value) })} /></label>
          <label>上映日期<input type="date" value={form.release_date} onChange={(e) => setForm({ ...form, release_date: e.target.value })} /></label>
          <label className="full">封面<input value={form.cover_image} onChange={(e) => setForm({ ...form, cover_image: e.target.value })} /></label>
          <label className="full">海报<input value={form.poster_image} onChange={(e) => setForm({ ...form, poster_image: e.target.value })} /></label>
          <label className="full">简介<textarea rows={4} value={form.description} onChange={(e) => setForm({ ...form, description: e.target.value })} /></label>
        </div>

        <div className="panel">
          <h3>导演</h3>
          <div className="toolbar">
            {directors.map((d) => (
              <label key={d.id}>
                <input type="checkbox" checked={selectedDirectors.includes(d.id)} onChange={() => setSelectedDirectors(toggleId(selectedDirectors, d.id))} /> {d.name}
              </label>
            ))}
          </div>
        </div>

        <div className="panel">
          <h3>演员</h3>
          <div className="stack">
            {actors.map((a) => {
              const selected = selectedActors.find((x) => x.actor_id === a.id)
              return (
                <div key={a.id} className="toolbar">
                  <label>
                    <input
                      type="checkbox"
                      checked={!!selected}
                      onChange={() => {
                        if (selected) setSelectedActors(selectedActors.filter((x) => x.actor_id !== a.id))
                        else setSelectedActors([...selectedActors, { actor_id: a.id, role: '' }])
                      }}
                    /> {a.name}
                  </label>
                  {selected ? (
                    <input
                      placeholder="角色名"
                      value={selected.role}
                      onChange={(e) => setSelectedActors(selectedActors.map((x) => x.actor_id === a.id ? { ...x, role: e.target.value } : x))}
                    />
                  ) : null}
                </div>
              )
            })}
          </div>
        </div>

        <div className="panel">
          <h3>标签</h3>
          <div className="toolbar">
            {tags.map((t) => (
              <label key={t.id}>
                <input type="checkbox" checked={selectedTags.includes(t.id)} onChange={() => setSelectedTags(toggleId(selectedTags, t.id))} /> {t.name}
              </label>
            ))}
          </div>
        </div>

        <div className="panel">
          <h3>新增剧集（保存时一并创建）</h3>
          <div className="stack">
            {episodes.map((ep, idx) => (
              <div className="toolbar" key={idx}>
                <select value={ep.source_id} onChange={(e) => {
                  const next = [...episodes]
                  next[idx] = { ...ep, source_id: Number(e.target.value) }
                  setEpisodes(next)
                }}>
                  <option value={0}>播放源</option>
                  {sources.map((s) => <option key={s.id} value={s.id}>{s.name}</option>)}
                </select>
                <input type="number" placeholder="集数" value={ep.episode_number} onChange={(e) => {
                  const next = [...episodes]
                  next[idx] = { ...ep, episode_number: Number(e.target.value) }
                  setEpisodes(next)
                }} />
                <input placeholder="标题" value={ep.title} onChange={(e) => {
                  const next = [...episodes]
                  next[idx] = { ...ep, title: e.target.value }
                  setEpisodes(next)
                }} />
                <input placeholder="播放地址" value={ep.play_url} onChange={(e) => {
                  const next = [...episodes]
                  next[idx] = { ...ep, play_url: e.target.value }
                  setEpisodes(next)
                }} />
                <select value={ep.format} onChange={(e) => {
                  const next = [...episodes]
                  next[idx] = { ...ep, format: e.target.value }
                  setEpisodes(next)
                }}>
                  <option value="hls">hls</option>
                  <option value="mp4">mp4</option>
                  <option value="dash">dash</option>
                  <option value="flv">flv</option>
                </select>
              </div>
            ))}
            <button type="button" onClick={() => setEpisodes([...episodes, { source_id: 0, episode_number: 1, title: '', play_url: '', format: 'hls' }])}>
              添加剧集行
            </button>
          </div>
        </div>

        <div className="toolbar">
          <button className="primary" type="submit">保存</button>
        </div>
      </form>
    </div>
  )
}

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route path="/*" element={
          <RequireAuth>
            <AdminLayout>
              <Routes>
                <Route path="/" element={<DashboardPage />} />
                <Route path="/content/categories" element={<CategoriesPage />} />
                <Route path="/content/videos" element={<VideosPage />} />
                <Route path="/content/videos/new" element={<VideoEditPage />} />
                <Route path="/content/videos/:id" element={<VideoEditPage />} />
                <Route path="/content/directors" element={
                  <NamedResourcePage title="导演管理" list={adminApi.listDirectors} create={adminApi.createDirector} remove={adminApi.deleteDirector} />
                } />
                <Route path="/content/actors" element={
                  <NamedResourcePage title="演员管理" list={adminApi.listActors} create={adminApi.createActor} remove={adminApi.deleteActor} />
                } />
                <Route path="/content/tags" element={
                  <NamedResourcePage title="标签管理" list={adminApi.listTags} create={adminApi.createTag} remove={adminApi.deleteTag} />
                } />
                <Route path="/content/play-sources" element={<PlaySourcesPage />} />
                <Route path="*" element={<Navigate to="/" replace />} />
              </Routes>
            </AdminLayout>
          </RequireAuth>
        } />
      </Routes>
    </BrowserRouter>
  )
}
