import { Link, useNavigate, useParams } from 'react-router'
import { useEffect, useState } from 'react'
import type { FormEvent } from 'react'
import { adminApi, errorMessage } from '../../lib/api'
import { flattenCategories } from '../../utils/categories'
import { ErrorAlert, PageCard, PageHeader } from '../../components/ui'
import type { Category, NamedItem, PlaySource, VideoDetail } from '@orange-tv/shared'

export default function VideoEditPage() {
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

  async function onSubmit(e: FormEvent) {
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
    <PageCard className="stack">
      <PageHeader title={isNew ? '新增影视' : `编辑影视 #${id}`}><Link to="/content/videos"><button>返回列表</button></Link></PageHeader>
      <ErrorAlert>{error}</ErrorAlert>
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
    </PageCard>
  )
}

