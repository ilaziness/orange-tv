import { useEffect, useMemo, useState } from 'react'
import { BrowserRouter, Link, Navigate, Route, Routes, useLocation, useNavigate, useParams } from 'react-router'
import type { Category, NamedItem, PlaySource, VideoDetail, VideoListItem, DashboardData, AdminItem, UserGroupItem, UserItem, BannerItem } from '@orange-tv/shared'
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
        { to: '/content/live', label: '直播管理' },
        { to: '/content/collect', label: '数据采集' },
        { to: '/content/directors', label: '导演管理' },
        { to: '/content/actors', label: '演员管理' },
        { to: '/content/tags', label: '标签管理' },
        { to: '/content/play-sources', label: '播放源管理' },
      ]
    }
    if (location.pathname.startsWith('/system')) {
      return [
        { to: '/system/site', label: '站点设置' },
        { to: '/system/api', label: 'API配置' },
        { to: '/system/theme', label: '主题管理' },
        { to: '/system/admins', label: '管理员' },
        { to: '/system/groups', label: '用户组' },
        { to: '/system/users', label: '用户' },
        { to: '/system/banners', label: 'Banner' },
        { to: '/system/log', label: '系统日志' },
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
          <Link className={location.pathname.startsWith('/system') ? 'active' : ''} to="/system/site">系统设置</Link>
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
  const [data, setData] = useState<DashboardData | null>(null)
  const [error, setError] = useState('')

  useEffect(() => {
    void (async () => {
      try {
        const res = await adminApi.dashboard()
        setData(res.data || null)
      } catch (err) {
        setError(errorMessage(err))
      }
    })()
  }, [])

  const stats: Array<{ label: string; value: number | undefined }> = [
    { label: '影视总数', value: data?.total_videos },
    { label: '今日新增', value: data?.today_videos },
    { label: '上线中', value: data?.online_videos },
    { label: '已下线', value: data?.offline_videos },
    { label: '分类数', value: data?.total_categories },
    { label: '管理员', value: data?.total_admins },
    { label: '注册用户', value: data?.total_users },
    { label: '在线用户', value: data?.online_count },
    { label: '今日PV', value: data?.today_pv },
    { label: '今日UV', value: data?.today_uv },
  ]

  return (
    <div className="page-card">
      <div className="page-header">
        <h1>概况</h1>
      </div>
      {error ? <p className="error">{error}</p> : null}
      <div className="dashboard-grid">
        {stats.map((s) => (
          <div key={s.label} className="stat-card">
            <div className="stat-value">{s.value ?? '-'}</div>
            <div className="stat-label">{s.label}</div>
          </div>
        ))}
      </div>
      <div className="toolbar" style={{ marginTop: 16 }}>
        <Link to="/content/videos"><button className="primary">影视管理</button></Link>
        <Link to="/content/categories"><button>分类管理</button></Link>
        <Link to="/system/site"><button>站点设置</button></Link>
        <Link to="/system/log"><button>系统日志</button></Link>
        <Link to="/system/admins"><button>管理员</button></Link>
        <Link to="/system/users"><button>用户</button></Link>
        <Link to="/system/banners"><button>Banner</button></Link>
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
  const [selected, setSelected] = useState<Set<number>>(new Set())

  async function load(p = page) {
    setError('')
    try {
      const res = await adminApi.listVideos({ keyword, page: p, page_size: 20 })
      setItems(res.data.list || [])
      setTotal(res.data.total || 0)
      setPage(res.data.page || p)
      setSelected(new Set())
    } catch (err) {
      setError(errorMessage(err))
    }
  }

  useEffect(() => { void load(1) }, [])

  function toggleSelect(id: number) {
    setSelected((prev) => {
      const next = new Set(prev)
      if (next.has(id)) next.delete(id)
      else next.add(id)
      return next
    })
  }

  function toggleSelectAll() {
    setSelected((prev) => {
      if (prev.size === items.length) return new Set()
      return new Set(items.map((i) => i.id))
    })
  }

  async function batchPublish(status: number) {
    if (selected.size === 0) return
    if (!confirm(`确认批量${status === 1 ? '上架' : '下架'} ${selected.size} 条影视？`)) return
    try {
      await adminApi.batchUpdatePublishStatus(Array.from(selected), status)
      await load(page)
    } catch (err) {
      setError(errorMessage(err))
    }
  }

  async function batchDelete() {
    if (selected.size === 0) return
    if (!confirm(`确认批量删除 ${selected.size} 条影视？`)) return
    try {
      await adminApi.batchDeleteVideos(Array.from(selected))
      await load(page)
    } catch (err) {
      setError(errorMessage(err))
    }
  }

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
      {selected.size > 0 ? (
        <div className="toolbar">
          <span className="muted">已选 {selected.size} 项</span>
          <button onClick={() => batchPublish(1)}>批量上架</button>
          <button onClick={() => batchPublish(0)}>批量下架</button>
          <button className="danger" onClick={batchDelete}>批量删除</button>
        </div>
      ) : null}
      <table className="table">
        <thead>
          <tr>
            <th><input type="checkbox" checked={selected.size === items.length && items.length > 0} onChange={toggleSelectAll} /></th>
            <th>ID</th><th>标题</th><th>年份</th><th>评分</th><th>状态</th><th>操作</th>
          </tr>
        </thead>
        <tbody>
          {items.map((item) => (
            <tr key={item.id}>
              <td><input type="checkbox" checked={selected.has(item.id)} onChange={() => toggleSelect(item.id)} /></td>
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


function LivePage() {
  const [list, setList] = useState<import('@orange-tv/shared').LiveChannel[]>([])
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const [form, setForm] = useState({ name: '', category: '', stream_url: '', logo: '', description: '', sort_order: 0, status: 1 })
  const [editId, setEditId] = useState(0)

  async function load() {
    setLoading(true)
    setError('')
    try {
      const res = await adminApi.listLive({ page: 1, page_size: 100 })
      setList(res.data.list || [])
    } catch (err) {
      setError(errorMessage(err))
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { void load() }, [])

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError('')
    try {
      if (editId) await adminApi.updateLive(editId, form)
      else await adminApi.createLive(form)
      setForm({ name: '', category: '', stream_url: '', logo: '', description: '', sort_order: 0, status: 1 })
      setEditId(0)
      await load()
    } catch (err) {
      setError(errorMessage(err))
    }
  }

  async function onDelete(id: number) {
    if (!confirm('确认删除该直播频道？')) return
    try {
      await adminApi.deleteLive(id)
      await load()
    } catch (err) {
      setError(errorMessage(err))
    }
  }

  function startEdit(item: import('@orange-tv/shared').LiveChannel) {
    setEditId(item.id)
    setForm({
      name: item.name,
      category: item.category || '',
      stream_url: item.stream_url,
      logo: item.logo || '',
      description: item.description || '',
      sort_order: item.sort_order || 0,
      status: item.status ?? 1,
    })
  }

  return (
    <div className="page-card stack">
      <div className="page-header"><h1>直播管理</h1></div>
      {error ? <p className="error">{error}</p> : null}
      <form className="stack" onSubmit={onSubmit}>
        <div className="toolbar">
          <input placeholder="频道名称" value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} required />
          <input placeholder="分类" value={form.category} onChange={(e) => setForm({ ...form, category: e.target.value })} />
          <input placeholder="直播流地址" value={form.stream_url} onChange={(e) => setForm({ ...form, stream_url: e.target.value })} required style={{ minWidth: 280 }} />
        </div>
        <div className="toolbar">
          <input placeholder="Logo URL" value={form.logo} onChange={(e) => setForm({ ...form, logo: e.target.value })} />
          <input placeholder="简介" value={form.description} onChange={(e) => setForm({ ...form, description: e.target.value })} />
          <input type="number" placeholder="排序" value={form.sort_order} onChange={(e) => setForm({ ...form, sort_order: Number(e.target.value) })} />
          <select value={form.status} onChange={(e) => setForm({ ...form, status: Number(e.target.value) })}>
            <option value={1}>启用</option>
            <option value={0}>禁用</option>
          </select>
          <button className="primary" type="submit">{editId ? '保存修改' : '新增频道'}</button>
          {editId ? <button type="button" onClick={() => { setEditId(0); setForm({ name: '', category: '', stream_url: '', logo: '', description: '', sort_order: 0, status: 1 }) }}>取消</button> : null}
          <button type="button" onClick={() => void load()} disabled={loading}>刷新</button>
        </div>
      </form>
      <table className="table">
        <thead>
          <tr>
            <th>ID</th><th>名称</th><th>分类</th><th>状态</th><th>排序</th><th>操作</th>
          </tr>
        </thead>
        <tbody>
          {list.map((item) => (
            <tr key={item.id}>
              <td>{item.id}</td>
              <td>
                <strong>{item.name}</strong>
                <div className="muted" style={{ maxWidth: 360, overflow: 'hidden', textOverflow: 'ellipsis' }}>{item.stream_url}</div>
              </td>
              <td>{item.category || '-'}</td>
              <td><span className={`badge ${item.status === 1 ? 'ok' : 'off'}`}>{item.status === 1 ? '启用' : '禁用'}</span></td>
              <td>{item.sort_order}</td>
              <td className="actions">
                <button onClick={() => startEdit(item)}>编辑</button>
                <button className="danger" onClick={() => void onDelete(item.id)}>删除</button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
      {!list.length ? <p className="muted">暂无直播频道</p> : null}
    </div>
  )
}


function CollectPage() {
  const [sources, setSources] = useState<import('@orange-tv/shared').CollectSource[]>([])
  const [playSources, setPlaySources] = useState<PlaySource[]>([])
  const [categories, setCategories] = useState<Category[]>([])
  const [logs, setLogs] = useState<import('@orange-tv/shared').CollectLog[]>([])
  const [maps, setMaps] = useState<import('@orange-tv/shared').CollectCategoryMap[]>([])
  const [error, setError] = useState('')
  const [selectedId, setSelectedId] = useState(0)
  const [form, setForm] = useState({
    name: '', type: 2, collect_url: '', api_key: '', cron_expr: '', play_source_id: 0, status: 1, config: '',
  })
  const [mapText, setMapText] = useState('[]')

  async function load() {
    setError('')
    try {
      const [s, p, c, l] = await Promise.all([
        adminApi.listCollectSources(),
        adminApi.listPlaySources(),
        adminApi.listCategories(),
        adminApi.listCollectLogs({ page: 1, page_size: 30 }),
      ])
      setSources(s.data.list || [])
      setPlaySources(p.data.list || [])
      setCategories(c.data || [])
      setLogs(l.data.list || [])
    } catch (err) {
      setError(errorMessage(err))
    }
  }

  useEffect(() => { void load() }, [])

  const flatCats = flattenCategories(categories)

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault()
    try {
      const body = { ...form, play_source_id: Number(form.play_source_id), type: Number(form.type), status: Number(form.status) }
      if (selectedId) await adminApi.updateCollectSource(selectedId, body)
      else await adminApi.createCollectSource(body)
      setSelectedId(0)
      setForm({ name: '', type: 2, collect_url: '', api_key: '', cron_expr: '', play_source_id: 0, status: 1, config: '' })
      await load()
    } catch (err) {
      setError(errorMessage(err))
    }
  }

  async function editSource(item: import('@orange-tv/shared').CollectSource) {
    setSelectedId(item.id)
    setForm({
      name: item.name,
      type: item.type,
      collect_url: item.collect_url,
      api_key: '',
      cron_expr: item.cron_expr || '',
      play_source_id: item.play_source_id,
      status: item.status,
      config: item.config || '',
    })
    try {
      const res = await adminApi.getCollectCategories(item.id)
      setMaps(res.data || [])
      setMapText(JSON.stringify((res.data || []).map((m) => ({ external_category: m.external_category, category_id: m.category_id })), null, 2))
    } catch (err) {
      setError(errorMessage(err))
    }
  }

  async function saveMaps() {
    if (!selectedId) return
    try {
      const items = JSON.parse(mapText || '[]')
      const res = await adminApi.setCollectCategories(selectedId, { items })
      setMaps(res.data || [])
      setMapText(JSON.stringify((res.data || []).map((m) => ({
        external_category: m.external_category,
        category_id: m.category_id,
      })), null, 2))
      await load()
    } catch (err) {
      setError(errorMessage(err))
    }
  }

  async function start(id: number) {
    try {
      await adminApi.startCollect(id)
      await load()
    } catch (err) {
      setError(errorMessage(err))
    }
  }

  async function stop(id: number) {
    try {
      await adminApi.stopCollect(id)
      await load()
    } catch (err) {
      setError(errorMessage(err))
    }
  }

  async function remove(id: number) {
    if (!confirm('确认删除采集源？')) return
    try {
      await adminApi.deleteCollectSource(id)
      await load()
    } catch (err) {
      setError(errorMessage(err))
    }
  }

  return (
    <div className="page-card stack">
      <div className="page-header"><h1>数据采集</h1></div>
      <p className="muted">支持默认 JSON 与苹果 CMS；手动触发异步执行，可配置 cron 定时采集。请先配置分类映射并绑定播放源。</p>
      {error ? <p className="error">{error}</p> : null}
      <form className="stack" onSubmit={onSubmit}>
        <div className="toolbar">
          <input placeholder="源名称" value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} required />
          <select value={form.type} onChange={(e) => setForm({ ...form, type: Number(e.target.value) })}>
            <option value={1}>默认格式</option>
            <option value={2}>苹果CMS</option>
          </select>
          <input placeholder="采集地址" style={{ minWidth: 280 }} value={form.collect_url} onChange={(e) => setForm({ ...form, collect_url: e.target.value })} required />
        </div>
        <div className="toolbar">
          <input placeholder="API Key" value={form.api_key} onChange={(e) => setForm({ ...form, api_key: e.target.value })} />
          <input placeholder="cron 表达式(空=不定时)" value={form.cron_expr} onChange={(e) => setForm({ ...form, cron_expr: e.target.value })} />
          <select value={form.play_source_id} onChange={(e) => setForm({ ...form, play_source_id: Number(e.target.value) })} required>
            <option value={0}>绑定播放源</option>
            {playSources.map((ps) => <option key={ps.id} value={ps.id}>{ps.name}</option>)}
          </select>
          <select value={form.status} onChange={(e) => setForm({ ...form, status: Number(e.target.value) })}>
            <option value={1}>启用</option>
            <option value={0}>禁用</option>
          </select>
          <button className="primary" type="submit">{selectedId ? '保存源' : '新增源'}</button>
          <button type="button" onClick={() => void load()}>刷新</button>
        </div>
      </form>

      <table className="table">
        <thead>
          <tr><th>ID</th><th>名称</th><th>类型</th><th>状态</th><th>最后采集</th><th>操作</th></tr>
        </thead>
        <tbody>
          {sources.map((s) => (
            <tr key={s.id}>
              <td>{s.id}</td>
              <td>
                <strong>{s.name}</strong>
                <div className="muted" style={{ maxWidth: 320, overflow: 'hidden', textOverflow: 'ellipsis' }}>{s.collect_url}</div>
              </td>
              <td>{s.type === 2 ? '苹果CMS' : '默认'}</td>
              <td><span className={`badge ${s.status === 1 ? 'ok' : 'off'}`}>{s.status === 1 ? '启用' : '禁用'}</span></td>
              <td className="muted">{s.last_collect_at || '-'}</td>
              <td className="actions">
                <button onClick={() => void editSource(s)}>编辑/映射</button>
                <button className="primary" onClick={() => void start(s.id)}>开始</button>
                <button onClick={() => void stop(s.id)}>停止</button>
                <button className="danger" onClick={() => void remove(s.id)}>删除</button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>

      {selectedId ? (
        <div className="stack">
          <h3>分类映射（JSON 数组）— 源 #{selectedId}</h3>
          <p className="muted">示例：{`[{"external_category":"1","category_id":11}]`}。外部分类键对应苹果 type_id 或默认 category 字段。</p>
          <div className="muted">可用系统分类：{flatCats.map((c) => `${c.id}:${c.name}`).join('， ') || '暂无'}</div>
          <textarea rows={8} value={mapText} onChange={(e) => setMapText(e.target.value)} style={{ width: '100%' }} />
          <button className="primary" onClick={() => void saveMaps()}>保存映射</button>
          {maps.length ? <div className="muted">当前映射 {maps.length} 条</div> : null}
        </div>
      ) : null}

      <div className="stack">
        <h3>采集日志</h3>
        <table className="table">
          <thead>
            <tr><th>ID</th><th>源</th><th>状态</th><th>总数</th><th>成功</th><th>失败</th><th>耗时ms</th><th>时间</th></tr>
          </thead>
          <tbody>
            {logs.map((l) => (
              <tr key={l.id}>
                <td>{l.id}</td>
                <td>{l.source_id}</td>
                <td>{l.status}</td>
                <td>{l.total_count}</td>
                <td>{l.success_count}</td>
                <td>{l.failed_count}</td>
                <td>{l.duration_ms}</td>
                <td className="muted">{l.created_at || '-'}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}

function ThemesPage() {
  const [list, setList] = useState<import('@orange-tv/shared').ThemeItem[]>([])
  const [error, setError] = useState('')
  const [selected, setSelected] = useState<import('@orange-tv/shared').ThemeItem | null>(null)
  const [configText, setConfigText] = useState('{}')
  const [customCss, setCustomCss] = useState('')
  const [customJs, setCustomJs] = useState('')

  async function load() {
    setError('')
    try {
      const res = await adminApi.listThemes()
      setList(res.data || [])
    } catch (err) {
      setError(errorMessage(err))
    }
  }

  useEffect(() => { void load() }, [])

  function pick(item: import('@orange-tv/shared').ThemeItem) {
    setSelected(item)
    setConfigText(JSON.stringify(item.config || {}, null, 2))
    setCustomCss(item.custom_css || '')
    setCustomJs(item.custom_js || '')
  }

  async function save() {
    if (!selected) return
    try {
      const config = JSON.parse(configText || '{}')
      await adminApi.updateTheme(selected.id, { config, custom_css: customCss, custom_js: customJs })
      await load()
    } catch (err) {
      setError(errorMessage(err))
    }
  }

  async function activate(id: number) {
    try {
      await adminApi.activateTheme(id)
      await load()
    } catch (err) {
      setError(errorMessage(err))
    }
  }

  return (
    <div className="page-card stack">
      <div className="page-header"><h1>主题管理</h1></div>
      <p className="muted">最小可用：切换激活主题、覆盖 config / custom_css / custom_js。上传第三方主题包不在本阶段。</p>
      {error ? <p className="error">{error}</p> : null}
      <div className="tree">
        {list.map((item) => (
          <div key={item.id} className="tree-item">
            <div>
              <strong>{item.name}</strong>
              <div className="muted">{item.identifier} · v{item.version}</div>
            </div>
            <div className="actions">
              <span className={`badge ${item.is_active ? 'ok' : 'off'}`}>{item.is_active ? '使用中' : '未激活'}</span>
              <button onClick={() => pick(item)}>编辑</button>
              {!item.is_active ? <button className="primary" onClick={() => void activate(item.id)}>激活</button> : null}
            </div>
          </div>
        ))}
      </div>
      {selected ? (
        <div className="stack">
          <h3>编辑：{selected.name}</h3>
          <label>Config JSON<textarea rows={8} value={configText} onChange={(e) => setConfigText(e.target.value)} style={{ width: '100%' }} /></label>
          <label>Custom CSS<textarea rows={4} value={customCss} onChange={(e) => setCustomCss(e.target.value)} style={{ width: '100%' }} /></label>
          <label>Custom JS<textarea rows={3} value={customJs} onChange={(e) => setCustomJs(e.target.value)} style={{ width: '100%' }} /></label>
          <button className="primary" onClick={() => void save()}>保存配置</button>
        </div>
      ) : null}
    </div>
  )
}

function SiteSettingsPage() {
  const [form, setForm] = useState({
    name: '', logo: '', copyright: '', icp: '', seo_keywords: '', description: '',
  })
  const [error, setError] = useState('')
  const [msg, setMsg] = useState('')
  const [loading, setLoading] = useState(false)

  async function load() {
    setLoading(true)
    setError('')
    try {
      const res = await adminApi.getSettings()
      const site = res.data.site
      setForm({
        name: site.name || '',
        logo: site.logo || '',
        copyright: site.copyright || '',
        icp: site.icp || '',
        seo_keywords: site.seo_keywords || '',
        description: site.description || '',
      })
    } catch (err) {
      setError(errorMessage(err))
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { void load() }, [])

  async function save(e: React.FormEvent) {
    e.preventDefault()
    setError('')
    setMsg('')
    try {
      await adminApi.updateSettings({ site: form })
      setMsg('已保存')
      await load()
    } catch (err) {
      setError(errorMessage(err))
    }
  }

  return (
    <div className="page-card stack">
      <div className="page-header"><h1>站点设置</h1></div>
      {error ? <p className="error">{error}</p> : null}
      {msg ? <p className="muted">{msg}</p> : null}
      {loading ? <p className="muted">加载中...</p> : (
        <form className="stack" onSubmit={(e) => void save(e)}>
          <label>站点名称<input value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} /></label>
          <label>Logo URL<input value={form.logo} onChange={(e) => setForm({ ...form, logo: e.target.value })} /></label>
          <label>版权信息<input value={form.copyright} onChange={(e) => setForm({ ...form, copyright: e.target.value })} /></label>
          <label>备案号<input value={form.icp} onChange={(e) => setForm({ ...form, icp: e.target.value })} /></label>
          <label>SEO 关键词<input value={form.seo_keywords} onChange={(e) => setForm({ ...form, seo_keywords: e.target.value })} /></label>
          <label>站点描述<textarea rows={3} value={form.description} onChange={(e) => setForm({ ...form, description: e.target.value })} style={{ width: '100%' }} /></label>
          <button className="primary" type="submit">保存</button>
        </form>
      )}
    </div>
  )
}

function APISettingsPage() {
  const [form, setForm] = useState({
    site_mode: 'video_site',
    api_output_format: 'default',
    enable_third_party_collect: true,
    resource_api_key: '',
  })
  const [keySet, setKeySet] = useState(false)
  const [error, setError] = useState('')
  const [msg, setMsg] = useState('')

  async function load() {
    setError('')
    try {
      const res = await adminApi.getSettings()
      const api = res.data.api
      setForm({
        site_mode: api.site_mode || 'video_site',
        api_output_format: api.api_output_format || 'default',
        enable_third_party_collect: !!api.enable_third_party_collect,
        resource_api_key: '',
      })
      setKeySet(!!api.resource_api_key_set)
    } catch (err) {
      setError(errorMessage(err))
    }
  }

  useEffect(() => { void load() }, [])

  async function save(e: React.FormEvent) {
    e.preventDefault()
    setError('')
    setMsg('')
    try {
      await adminApi.updateSettings({
        api: {
          site_mode: form.site_mode,
          api_output_format: form.api_output_format,
          enable_third_party_collect: form.enable_third_party_collect,
          ...(form.resource_api_key.trim() ? { resource_api_key: form.resource_api_key.trim() } : {}),
        },
      })
      setMsg('已保存')
      setForm((f) => ({ ...f, resource_api_key: '' }))
      await load()
    } catch (err) {
      setError(errorMessage(err))
    }
  }

  return (
    <div className="page-card stack">
      <div className="page-header"><h1>API 配置</h1></div>
      <p className="muted">资源站开放接口：`/api/open/v1/*`。密钥通过 Header X-API-Key 或 query key 传递；密钥输入框留空表示不修改。</p>
      {error ? <p className="error">{error}</p> : null}
      {msg ? <p className="muted">{msg}</p> : null}
      <form className="stack" onSubmit={(e) => void save(e)}>
        <label>站点模式
          <select value={form.site_mode} onChange={(e) => setForm({ ...form, site_mode: e.target.value })}>
            <option value="video_site">影视站</option>
            <option value="resource_site">资源站</option>
          </select>
        </label>
        <label>API 输出格式
          <select value={form.api_output_format} onChange={(e) => setForm({ ...form, api_output_format: e.target.value })}>
            <option value="default">系统默认格式</option>
            <option value="apple_cms">苹果 CMS</option>
          </select>
        </label>
        <label className="row">
          <input type="checkbox" checked={form.enable_third_party_collect} onChange={(e) => setForm({ ...form, enable_third_party_collect: e.target.checked })} />
          允许第三方采集
        </label>
        <label>资源站 API 密钥 {keySet ? <span className="muted">（已配置）</span> : <span className="muted">（未配置）</span>}
          <input type="password" placeholder={keySet ? '****** 留空不修改' : '设置密钥'} value={form.resource_api_key} onChange={(e) => setForm({ ...form, resource_api_key: e.target.value })} />
        </label>
        <button className="primary" type="submit">保存</button>
      </form>
    </div>
  )
}

function SystemLogPage() {
  const [tab, setTab] = useState<'system' | 'login'>('system')
  const [systemLogs, setSystemLogs] = useState<import('@orange-tv/shared').SystemLogItem[]>([])
  const [loginLogs, setLoginLogs] = useState<import('@orange-tv/shared').LoginLogItem[]>([])
  const [module, setModule] = useState('')
  const [username, setUsername] = useState('')
  const [error, setError] = useState('')
  const [total, setTotal] = useState(0)

  async function load() {
    setError('')
    try {
      if (tab === 'system') {
        const res = await adminApi.listSystemLogs({ page: 1, page_size: 50, module: module || undefined })
        setSystemLogs(res.data.list || [])
        setTotal(res.data.total)
      } else {
        const res = await adminApi.listLoginLogs({ page: 1, page_size: 50, username: username || undefined })
        setLoginLogs(res.data.list || [])
        setTotal(res.data.total)
      }
    } catch (err) {
      setError(errorMessage(err))
    }
  }

  useEffect(() => { void load() }, [tab])

  return (
    <div className="page-card stack">
      <div className="page-header"><h1>系统日志</h1></div>
      <div className="toolbar">
        <button className={tab === 'system' ? 'primary' : ''} onClick={() => setTab('system')}>操作日志</button>
        <button className={tab === 'login' ? 'primary' : ''} onClick={() => setTab('login')}>登录日志</button>
      </div>
      {error ? <p className="error">{error}</p> : null}
      {tab === 'system' ? (
        <>
          <div className="toolbar">
            <input placeholder="模块筛选" value={module} onChange={(e) => setModule(e.target.value)} />
            <button onClick={() => void load()}>查询</button>
          </div>
          <p className="muted">共 {total} 条</p>
          <div className="tree">
            {systemLogs.map((l) => (
              <div key={l.id} className="tree-item">
                <div>
                  <strong>[{l.level}] {l.module}/{l.action}</strong>
                  <div className="muted">admin={l.admin_id} · {l.ip_address} · {l.created_at}</div>
                  <div>{l.content}</div>
                </div>
              </div>
            ))}
          </div>
        </>
      ) : (
        <>
          <div className="toolbar">
            <input placeholder="用户名" value={username} onChange={(e) => setUsername(e.target.value)} />
            <button onClick={() => void load()}>查询</button>
          </div>
          <p className="muted">共 {total} 条</p>
          <div className="tree">
            {loginLogs.map((l) => (
              <div key={l.id} className="tree-item">
                <div>
                  <strong>{l.username}</strong>
                  <div className="muted">{l.status === 1 ? '成功' : '失败'} · {l.ip_address} · {l.created_at}</div>
                  <div className="muted">{l.user_agent}</div>
                </div>
              </div>
            ))}
          </div>
        </>
      )}
    </div>
  )
}

// ===== Phase 5: Admins / User Groups / Users / Banners =====

function AdminsPage() {
  const [list, setList] = useState<AdminItem[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [keyword, setKeyword] = useState('')
  const [error, setError] = useState('')
  const [showCreate, setShowCreate] = useState(false)
  const [form, setForm] = useState({ username: '', password: '', email: '', group_id: 1, status: 1 })
  const [groups, setGroups] = useState<UserGroupItem[]>([])

  async function load() {
    setError('')
    try {
      const [res, gRes] = await Promise.all([
        adminApi.listAdmins({ page, page_size: 20, keyword: keyword || undefined }),
        adminApi.listGroups({ page_size: 100 }),
      ])
      setList(res.data.list || [])
      setTotal(res.data.total)
      setGroups(gRes.data.list || [])
    } catch (err) {
      setError(errorMessage(err))
    }
  }

  useEffect(() => { void load() }, [page])

  async function onCreate(e: React.FormEvent) {
    e.preventDefault()
    setError('')
    try {
      await adminApi.createAdmin(form)
      setShowCreate(false)
      setForm({ username: '', password: '', email: '', group_id: 1, status: 1 })
      await load()
    } catch (err) {
      setError(errorMessage(err))
    }
  }

  async function onResetPwd(id: number) {
    const pwd = window.prompt('输入新密码（≥6位）')
    if (!pwd || pwd.length < 6) return
    try {
      await adminApi.resetAdminPassword(id, pwd)
      alert('密码已重置')
    } catch (err) {
      alert(errorMessage(err))
    }
  }

  return (
    <div className="page-card stack">
      <div className="page-header">
        <h1>管理员管理</h1>
        <button className="primary" onClick={() => setShowCreate(!showCreate)}>新增管理员</button>
      </div>
      {error ? <p className="error">{error}</p> : null}
      {showCreate ? (
        <form className="inline-form" onSubmit={onCreate}>
          <input placeholder="用户名" value={form.username} onChange={(e) => setForm({ ...form, username: e.target.value })} required minLength={3} />
          <input type="password" placeholder="密码" value={form.password} onChange={(e) => setForm({ ...form, password: e.target.value })} required minLength={6} />
          <input placeholder="邮箱" value={form.email} onChange={(e) => setForm({ ...form, email: e.target.value })} />
          <select value={form.group_id} onChange={(e) => setForm({ ...form, group_id: Number(e.target.value) })}>
            {groups.map((g) => <option key={g.id} value={g.id}>{g.name}</option>)}
          </select>
          <select value={form.status} onChange={(e) => setForm({ ...form, status: Number(e.target.value) })}>
            <option value={1}>启用</option>
            <option value={0}>禁用</option>
          </select>
          <button type="submit" className="primary">保存</button>
        </form>
      ) : null}
      <div className="toolbar">
        <input placeholder="用户名/邮箱" value={keyword} onChange={(e) => setKeyword(e.target.value)} />
        <button onClick={() => { setPage(1); void load() }}>查询</button>
      </div>
      <p className="muted">共 {total} 条</p>
      <table className="data-table">
        <thead><tr><th>ID</th><th>用户名</th><th>邮箱</th><th>用户组</th><th>状态</th><th>最后登录</th><th>操作</th></tr></thead>
        <tbody>
          {list.map((a) => (
            <tr key={a.id}>
              <td>{a.id}</td>
              <td>{a.username}</td>
              <td>{a.email || '-'}</td>
              <td>{a.group_name}</td>
              <td>{a.status === 1 ? '启用' : '禁用'}</td>
              <td>{a.last_login_at || '-'}</td>
              <td>
                <button onClick={() => onResetPwd(a.id)}>重置密码</button>
                <button onClick={async () => { if (confirm('删除该管理员？')) { try { await adminApi.deleteAdmin(a.id); await load() } catch (err) { alert(errorMessage(err)) } } }}>删除</button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

function UserGroupsPage() {
  const [list, setList] = useState<UserGroupItem[]>([])
  const [total, setTotal] = useState(0)
  const [error, setError] = useState('')
  const [showCreate, setShowCreate] = useState(false)
  const [form, setForm] = useState({ name: '', permissions: '', description: '' })

  async function load() {
    setError('')
    try {
      const res = await adminApi.listGroups({ page_size: 100 })
      setList(res.data.list || [])
      setTotal(res.data.total)
    } catch (err) {
      setError(errorMessage(err))
    }
  }

  useEffect(() => { void load() }, [])

  async function onCreate(e: React.FormEvent) {
    e.preventDefault()
    setError('')
    try {
      await adminApi.createGroup(form)
      setShowCreate(false)
      setForm({ name: '', permissions: '', description: '' })
      await load()
    } catch (err) {
      setError(errorMessage(err))
    }
  }

  return (
    <div className="page-card stack">
      <div className="page-header">
        <h1>用户组管理</h1>
        <button className="primary" onClick={() => setShowCreate(!showCreate)}>新增用户组</button>
      </div>
      {error ? <p className="error">{error}</p> : null}
      {showCreate ? (
        <form className="inline-form" onSubmit={onCreate}>
          <input placeholder="名称" value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} required />
          <input placeholder="权限（JSON）" value={form.permissions} onChange={(e) => setForm({ ...form, permissions: e.target.value })} />
          <input placeholder="描述" value={form.description} onChange={(e) => setForm({ ...form, description: e.target.value })} />
          <button type="submit" className="primary">保存</button>
        </form>
      ) : null}
      <p className="muted">共 {total} 条</p>
      <table className="data-table">
        <thead><tr><th>ID</th><th>名称</th><th>描述</th><th>操作</th></tr></thead>
        <tbody>
          {list.map((g) => (
            <tr key={g.id}>
              <td>{g.id}</td>
              <td>{g.name}</td>
              <td>{g.description || '-'}</td>
              <td>
                {g.name !== 'super_admin' ? (
                  <button onClick={async () => { if (confirm('删除该用户组？')) { try { await adminApi.deleteGroup(g.id); await load() } catch (err) { alert(errorMessage(err)) } } }}>删除</button>
                ) : <span className="muted">系统组</span>}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

function UsersPage() {
  const [list, setList] = useState<UserItem[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [keyword, setKeyword] = useState('')
  const [error, setError] = useState('')

  async function load() {
    setError('')
    try {
      const res = await adminApi.listUsers({ page, page_size: 20, keyword: keyword || undefined })
      setList(res.data.list || [])
      setTotal(res.data.total)
    } catch (err) {
      setError(errorMessage(err))
    }
  }

  useEffect(() => { void load() }, [page])

  async function onToggleStatus(u: UserItem) {
    try {
      await adminApi.updateUser(u.id, { status: u.status === 1 ? 0 : 1 })
      await load()
    } catch (err) {
      alert(errorMessage(err))
    }
  }

  async function onResetPwd(id: number) {
    const pwd = window.prompt('输入新密码（≥6位）')
    if (!pwd || pwd.length < 6) return
    try {
      await adminApi.resetUserPassword(id, pwd)
      alert('密码已重置')
    } catch (err) {
      alert(errorMessage(err))
    }
  }

  return (
    <div className="page-card stack">
      <div className="page-header"><h1>用户管理</h1></div>
      {error ? <p className="error">{error}</p> : null}
      <div className="toolbar">
        <input placeholder="用户名/邮箱" value={keyword} onChange={(e) => setKeyword(e.target.value)} />
        <button onClick={() => { setPage(1); void load() }}>查询</button>
      </div>
      <p className="muted">共 {total} 条</p>
      <table className="data-table">
        <thead><tr><th>ID</th><th>用户名</th><th>邮箱</th><th>状态</th><th>最后登录</th><th>注册时间</th><th>操作</th></tr></thead>
        <tbody>
          {list.map((u) => (
            <tr key={u.id}>
              <td>{u.id}</td>
              <td>{u.username}</td>
              <td>{u.email || '-'}</td>
              <td>{u.status === 1 ? '启用' : '禁用'}</td>
              <td>{u.last_login_at || '-'}</td>
              <td>{u.created_at || '-'}</td>
              <td>
                <button onClick={() => onToggleStatus(u)}>{u.status === 1 ? '禁用' : '启用'}</button>
                <button onClick={() => onResetPwd(u.id)}>重置密码</button>
                <button onClick={async () => { if (confirm('删除该用户？')) { try { await adminApi.deleteUser(u.id); await load() } catch (err) { alert(errorMessage(err)) } } }}>删除</button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

function BannersPage() {
  const [list, setList] = useState<BannerItem[]>([])
  const [total, setTotal] = useState(0)
  const [error, setError] = useState('')
  const [showCreate, setShowCreate] = useState(false)
  const [form, setForm] = useState({ title: '', cover: '', link: '', video_id: 0, sort: 0, status: 1 })

  async function load() {
    setError('')
    try {
      const res = await adminApi.listBanners({ page_size: 100 })
      setList(res.data.list || [])
      setTotal(res.data.total)
    } catch (err) {
      setError(errorMessage(err))
    }
  }

  useEffect(() => { void load() }, [])

  async function onCreate(e: React.FormEvent) {
    e.preventDefault()
    setError('')
    try {
      await adminApi.createBanner(form)
      setShowCreate(false)
      setForm({ title: '', cover: '', link: '', video_id: 0, sort: 0, status: 1 })
      await load()
    } catch (err) {
      setError(errorMessage(err))
    }
  }

  async function onToggle(b: BannerItem) {
    try {
      await adminApi.updateBanner(b.id, { status: b.status === 1 ? 0 : 1 })
      await load()
    } catch (err) {
      alert(errorMessage(err))
    }
  }

  return (
    <div className="page-card stack">
      <div className="page-header">
        <h1>Banner 管理</h1>
        <button className="primary" onClick={() => setShowCreate(!showCreate)}>新增 Banner</button>
      </div>
      {error ? <p className="error">{error}</p> : null}
      {showCreate ? (
        <form className="inline-form" onSubmit={onCreate}>
          <input placeholder="标题" value={form.title} onChange={(e) => setForm({ ...form, title: e.target.value })} required />
          <input placeholder="封面URL" value={form.cover} onChange={(e) => setForm({ ...form, cover: e.target.value })} />
          <input placeholder="链接" value={form.link} onChange={(e) => setForm({ ...form, link: e.target.value })} />
          <input type="number" placeholder="影视ID" value={form.video_id || ''} onChange={(e) => setForm({ ...form, video_id: Number(e.target.value) })} />
          <input type="number" placeholder="排序" value={form.sort} onChange={(e) => setForm({ ...form, sort: Number(e.target.value) })} />
          <select value={form.status} onChange={(e) => setForm({ ...form, status: Number(e.target.value) })}>
            <option value={1}>启用</option>
            <option value={0}>禁用</option>
          </select>
          <button type="submit" className="primary">保存</button>
        </form>
      ) : null}
      <p className="muted">共 {total} 条</p>
      <table className="data-table">
        <thead><tr><th>ID</th><th>标题</th><th>封面</th><th>排序</th><th>状态</th><th>操作</th></tr></thead>
        <tbody>
          {list.map((b) => (
            <tr key={b.id}>
              <td>{b.id}</td>
              <td>{b.title}</td>
              <td>{b.cover ? <img src={b.cover} alt="" style={{ width: 60, height: 34, objectFit: 'cover' }} /> : '-'}</td>
              <td>{b.sort}</td>
              <td>{b.status === 1 ? '启用' : '禁用'}</td>
              <td>
                <button onClick={() => onToggle(b)}>{b.status === 1 ? '禁用' : '启用'}</button>
                <button onClick={async () => { if (confirm('删除该Banner？')) { try { await adminApi.deleteBanner(b.id); await load() } catch (err) { alert(errorMessage(err)) } } }}>删除</button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
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
                <Route path="/content/live" element={<LivePage />} />
                <Route path="/content/collect" element={<CollectPage />} />
                <Route path="/system/site" element={<SiteSettingsPage />} />
                <Route path="/system/api" element={<APISettingsPage />} />
                <Route path="/system/theme" element={<ThemesPage />} />
                <Route path="/system/log" element={<SystemLogPage />} />
                <Route path="/system/admins" element={<AdminsPage />} />
                <Route path="/system/groups" element={<UserGroupsPage />} />
                <Route path="/system/users" element={<UsersPage />} />
                <Route path="/system/banners" element={<BannersPage />} />
                <Route path="*" element={<Navigate to="/" replace />} />
              </Routes>
            </AdminLayout>
          </RequireAuth>
        } />
      </Routes>
    </BrowserRouter>
  )
}
