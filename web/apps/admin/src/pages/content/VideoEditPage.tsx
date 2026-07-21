import { Link, useNavigate, useParams } from 'react-router'
import { useEffect, useState } from 'react'
import type * as React from 'react'
import { adminApi, errorMessage } from '@/lib/api'
import { flattenCategories } from '@/lib/categories'
import { PageContainer } from '@/components/shared'
import type { Category, NamedItem, PlaySource, VideoDetail } from '@orange-tv/shared'
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Field, FieldGroup, FieldLabel } from '@/components/ui/field'
import { Textarea } from '@/components/ui/textarea'
import { Checkbox } from '@/components/ui/checkbox'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { ArrowLeft, Plus, Save } from 'lucide-react'
import { toast } from 'sonner'

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
    category_id: '0',
    publish_status: '0',
    serial_status: '1',
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
            category_id: String(d.category_id),
            publish_status: String(d.publish_status ?? 0),
            serial_status: String(d.serial_status),
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

  async function onSubmit(e: React.SyntheticEvent<HTMLFormElement>) {
    e.preventDefault()
    setError('')
    if (form.category_id === '0') {
      setError('请选择分类')
      return
    }
    const body = {
      title: form.title,
      subtitle: form.subtitle,
      description: form.description,
      category_id: Number(form.category_id),
      publish_status: Number(form.publish_status),
      serial_status: Number(form.serial_status),
      cover_image: form.cover_image,
      poster_image: form.poster_image,
      year: Number(form.year) || 0,
      region: form.region,
      language: form.language,
      duration: Number(form.duration) || 0,
      rating: Number(form.rating) || 0,
      release_date: form.release_date,
      director_ids: selectedDirectors,
      actors: selectedActors,
      tag_ids: selectedTags,
    }
    try {
      let videoId = Number(id)
      if (isNew) {
        const res = await adminApi.createVideo(body)
        videoId = (res.data as VideoDetail).id
        toast.success('影视已创建')
      } else {
        if (!videoId) {
          setError('无效的影视 ID')
          return
        }
        await adminApi.updateVideo(videoId, body)
        toast.success('影视已更新')
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
    <PageContainer>
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle>{isNew ? '新增影视' : `编辑影视 #${id}`}</CardTitle>
            <Button variant="outline" size="sm" render={<Link to="/content/videos" />}>
              <ArrowLeft data-icon="inline-start" />
              返回列表
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          {error && (
            <Alert variant="destructive" className="mb-4">
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}
          <form onSubmit={onSubmit} className="flex flex-col gap-6">
            <FieldGroup className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
              <Field>
                <FieldLabel htmlFor="title">标题</FieldLabel>
                <Input id="title" value={form.title} onChange={(e) => setForm({ ...form, title: e.target.value })} required />
              </Field>
              <Field>
                <FieldLabel htmlFor="subtitle">副标题</FieldLabel>
                <Input id="subtitle" value={form.subtitle} onChange={(e) => setForm({ ...form, subtitle: e.target.value })} />
              </Field>
              <Field>
                <FieldLabel>分类</FieldLabel>
                <Select items={categories.map((category) => ({ value: String(category.id), label: `${'—'.repeat(category.depth)} ${category.name}` }))} value={form.category_id} onValueChange={(v) => setForm({ ...form, category_id: v ?? '0' })}>
                  <SelectTrigger>
                    <SelectValue placeholder="请选择" />
                  </SelectTrigger>
                  <SelectContent>
                    {categories.map((c) => (
                      <SelectItem key={c.id} value={String(c.id)}>
                        {'—'.repeat(c.depth)} {c.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </Field>
              <Field>
                <FieldLabel>上下架</FieldLabel>
                <Select items={[{ value: '0', label: '下架' }, { value: '1', label: '上架' }]} value={form.publish_status} onValueChange={(v) => setForm({ ...form, publish_status: v ?? '0' })}>
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="0">下架</SelectItem>
                    <SelectItem value="1">上架</SelectItem>
                  </SelectContent>
                </Select>
              </Field>
              <Field>
                <FieldLabel>连载状态</FieldLabel>
                <Select items={[{ value: '1', label: '连载中' }, { value: '2', label: '已完结' }, { value: '3', label: '即将上线' }]} value={form.serial_status} onValueChange={(v) => setForm({ ...form, serial_status: v ?? '1' })}>
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="1">连载中</SelectItem>
                    <SelectItem value="2">已完结</SelectItem>
                    <SelectItem value="3">即将上线</SelectItem>
                  </SelectContent>
                </Select>
              </Field>
              <Field>
                <FieldLabel htmlFor="year">年份</FieldLabel>
                <Input id="year" type="number" value={form.year} onChange={(e) => setForm({ ...form, year: Number(e.target.value) })} />
              </Field>
              <Field>
                <FieldLabel htmlFor="region">地区</FieldLabel>
                <Input id="region" value={form.region} onChange={(e) => setForm({ ...form, region: e.target.value })} />
              </Field>
              <Field>
                <FieldLabel htmlFor="language">语言</FieldLabel>
                <Input id="language" value={form.language} onChange={(e) => setForm({ ...form, language: e.target.value })} />
              </Field>
              <Field>
                <FieldLabel htmlFor="duration">时长(分钟)</FieldLabel>
                <Input id="duration" type="number" value={form.duration} onChange={(e) => setForm({ ...form, duration: Number(e.target.value) })} />
              </Field>
              <Field>
                <FieldLabel htmlFor="rating">评分</FieldLabel>
                <Input id="rating" type="number" step="0.1" value={form.rating} onChange={(e) => setForm({ ...form, rating: Number(e.target.value) })} />
              </Field>
              <Field>
                <FieldLabel htmlFor="release_date">上映日期</FieldLabel>
                <Input id="release_date" type="date" value={form.release_date} onChange={(e) => setForm({ ...form, release_date: e.target.value })} />
              </Field>
              <Field className="sm:col-span-2 lg:col-span-3">
                <FieldLabel htmlFor="cover_image">封面 URL</FieldLabel>
                <Input id="cover_image" value={form.cover_image} onChange={(e) => setForm({ ...form, cover_image: e.target.value })} />
              </Field>
              <Field className="sm:col-span-2 lg:col-span-3">
                <FieldLabel htmlFor="poster_image">海报 URL</FieldLabel>
                <Input id="poster_image" value={form.poster_image} onChange={(e) => setForm({ ...form, poster_image: e.target.value })} />
              </Field>
              <Field className="sm:col-span-2 lg:col-span-3">
                <FieldLabel htmlFor="description">简介</FieldLabel>
                <Textarea id="description" rows={4} value={form.description} onChange={(e) => setForm({ ...form, description: e.target.value })} />
              </Field>
            </FieldGroup>

            <div className="rounded-lg border p-4">
              <h3 className="mb-3 font-medium">导演</h3>
              <div className="flex flex-wrap gap-4">
                {directors.map((d) => (
                  <label key={d.id} className="flex items-center gap-2 text-sm">
                    <Checkbox
                      checked={selectedDirectors.includes(d.id)}
                      onCheckedChange={() => setSelectedDirectors(toggleId(selectedDirectors, d.id))}
                    />
                    {d.name}
                  </label>
                ))}
              </div>
            </div>

            <div className="rounded-lg border p-4">
              <h3 className="mb-3 font-medium">演员</h3>
              <div className="flex flex-col gap-2">
                {actors.map((a) => {
                  const selected = selectedActors.find((x) => x.actor_id === a.id)
                  return (
                    <div key={a.id} className="flex items-center gap-2">
                      <label className="flex items-center gap-2 text-sm">
                        <Checkbox
                          checked={!!selected}
                          onCheckedChange={() => {
                            if (selected) setSelectedActors(selectedActors.filter((x) => x.actor_id !== a.id))
                            else setSelectedActors([...selectedActors, { actor_id: a.id, role: '' }])
                          }}
                        />
                        {a.name}
                      </label>
                      {selected && (
                        <Input
                          placeholder="角色名"
                          value={selected.role}
                          onChange={(e) => setSelectedActors(selectedActors.map((x) => x.actor_id === a.id ? { ...x, role: e.target.value } : x))}
                          className="max-w-[200px]"
                        />
                      )}
                    </div>
                  )
                })}
              </div>
            </div>

            <div className="rounded-lg border p-4">
              <h3 className="mb-3 font-medium">标签</h3>
              <div className="flex flex-wrap gap-4">
                {tags.map((t) => (
                  <label key={t.id} className="flex items-center gap-2 text-sm">
                    <Checkbox
                      checked={selectedTags.includes(t.id)}
                      onCheckedChange={() => setSelectedTags(toggleId(selectedTags, t.id))}
                    />
                    {t.name}
                  </label>
                ))}
              </div>
            </div>

            <div className="rounded-lg border p-4">
              <h3 className="mb-3 font-medium">新增剧集（保存时一并创建）</h3>
              <div className="flex flex-col gap-2">
                {episodes.map((ep, idx) => (
                  <div key={idx} className="flex flex-wrap items-center gap-2">
                    <Select items={sources.map((source) => ({ value: String(source.id), label: source.name }))} value={String(ep.source_id)} onValueChange={(v) => {
                      const next = [...episodes]
                      next[idx] = { ...ep, source_id: Number(v ?? '0') }
                      setEpisodes(next)
                    }}>
                      <SelectTrigger className="w-32">
                        <SelectValue placeholder="播放源" />
                      </SelectTrigger>
                      <SelectContent>
                        {sources.map((s) => (
                          <SelectItem key={s.id} value={String(s.id)}>{s.name}</SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                    <Input
                      type="number"
                      placeholder="集数"
                      value={ep.episode_number}
                      onChange={(e) => {
                        const next = [...episodes]
                        next[idx] = { ...ep, episode_number: Number(e.target.value) }
                        setEpisodes(next)
                      }}
                      className="w-20"
                    />
                    <Input
                      placeholder="标题"
                      value={ep.title}
                      onChange={(e) => {
                        const next = [...episodes]
                        next[idx] = { ...ep, title: e.target.value }
                        setEpisodes(next)
                      }}
                      className="w-32"
                    />
                    <Input
                      placeholder="播放地址"
                      value={ep.play_url}
                      onChange={(e) => {
                        const next = [...episodes]
                        next[idx] = { ...ep, play_url: e.target.value }
                        setEpisodes(next)
                      }}
                      className="min-w-[200px] flex-1"
                    />
                    <Select items={[{ value: 'hls', label: 'hls' }, { value: 'mp4', label: 'mp4' }, { value: 'dash', label: 'dash' }, { value: 'flv', label: 'flv' }]} value={ep.format} onValueChange={(v) => {
                      const next = [...episodes]
                      next[idx] = { ...ep, format: v ?? 'hls' }
                      setEpisodes(next)
                    }}>
                      <SelectTrigger className="w-24">
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="hls">hls</SelectItem>
                        <SelectItem value="mp4">mp4</SelectItem>
                        <SelectItem value="dash">dash</SelectItem>
                        <SelectItem value="flv">flv</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                ))}
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={() => setEpisodes([...episodes, { source_id: 0, episode_number: 1, title: '', play_url: '', format: 'hls' }])}
                >
                  <Plus data-icon="inline-start" />
                  添加剧集行
                </Button>
              </div>
            </div>

            <div className="flex justify-end">
              <Button type="submit">
                <Save data-icon="inline-start" />
                保存
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>
    </PageContainer>
  )
}

