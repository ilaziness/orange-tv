import { useEffect, useState } from 'react'
import { useParams } from 'react-router'
import type * as React from 'react'
import { z } from 'zod'
import { adminApi, errorMessage } from '@/lib/api'
import { flattenCategories } from '@/lib/categories'
import type { Category, NamedItem, PlaySource, VideoDetail } from '@orange-tv/shared'
import { toast } from 'sonner'

const videoFormSchema = z.object({
  title: z.string().min(1, '请输入影视标题'),
  subtitle: z.string(),
  description: z.string(),
  category_id: z.string().refine((v) => v !== '' && v !== '0', '请选择分类').transform((v) => Number(v)),
  publish_status: z.union([z.string(), z.number()]).transform((v) => Number(v)),
  serial_status: z.union([z.string(), z.number()]).transform((v) => Number(v)),
  cover_image: z.string(),
  poster_image: z.string(),
  year: z.union([z.string(), z.number()]).transform((v) => Number(v) || 0),
  region: z.string(),
  language: z.string(),
  duration: z.union([z.string(), z.number()]).transform((v) => Number(v) || 0),
  rating: z.union([z.string(), z.number()]).transform((v) => Number(v) || 0),
  release_date: z.string(),
})

const emptyForm = {
  title: '',
  subtitle: '',
  description: '',
  category_id: '',
  publish_status: '0',
  serial_status: '1',
  cover_image: '',
  poster_image: '',
  year: '',
  region: '',
  language: '',
  duration: '',
  rating: '',
  release_date: '',
}

const emptyEpisode = { source_id: 0, episode_number: '', title: '', play_url: '', format: 'hls' }

export type VideoForm = typeof emptyForm
export type EpisodeDraft = typeof emptyEpisode

export function useVideoEdit() {
  const { id } = useParams()
  const isNew = !id || id === 'new'
  const [error, setError] = useState('')
  const [categories, setCategories] = useState<Array<Category & { depth: number }>>([])
  const [directors, setDirectors] = useState<NamedItem[]>([])
  const [actors, setActors] = useState<NamedItem[]>([])
  const [tags, setTags] = useState<NamedItem[]>([])
  const [sources, setSources] = useState<PlaySource[]>([])
  const [selectedDirectors, setSelectedDirectors] = useState<number[]>([])
  const [selectedActors, setSelectedActors] = useState<number[]>([])
  const [selectedTags, setSelectedTags] = useState<number[]>([])
  const [episodes, setEpisodes] = useState<EpisodeDraft[]>([])
  const [form, setForm] = useState(emptyForm)
  const [submitting, setSubmitting] = useState(false)
  const [initLoading, setInitLoading] = useState(true)

  useEffect(() => {
    void (async () => {
      setInitLoading(true)
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
            category_id: d.category_id ? String(d.category_id) : '',
            publish_status: String(d.publish_status ?? 0),
            serial_status: String(d.serial_status),
            cover_image: d.cover,
            poster_image: d.poster,
            year: d.year ? String(d.year) : '',
            region: d.region,
            language: d.language,
            duration: d.duration ? String(d.duration) : '',
            rating: d.rating ? String(d.rating) : '',
            release_date: d.release_date || '',
          })
          setSelectedDirectors(d.directors.map((x) => x.id))
          setSelectedActors(d.actors.map((x) => x.id))
          setSelectedTags(d.tags.map((x) => x.id))
        }
      } catch (err) {
        setError(errorMessage(err))
      } finally {
        setInitLoading(false)
      }
    })()
  }, [id, isNew])

  function toggleDirector(id: number) {
    setSelectedDirectors((prev) => prev.includes(id) ? prev.filter((x) => x !== id) : [...prev, id])
  }

  function toggleTag(id: number) {
    setSelectedTags((prev) => prev.includes(id) ? prev.filter((x) => x !== id) : [...prev, id])
  }

  function toggleActor(id: number) {
    setSelectedActors((prev) => prev.includes(id) ? prev.filter((x) => x !== id) : [...prev, id])
  }

  function addEpisode() {
    setEpisodes((prev) => [...prev, { ...emptyEpisode }])
  }

  function updateEpisode(index: number, patch: Partial<EpisodeDraft>) {
    setEpisodes((prev) => {
      const next = [...prev]
      next[index] = { ...next[index], ...patch }
      return next
    })
  }

  function removeEpisode(index: number) {
    setEpisodes((prev) => prev.filter((_, i) => i !== index))
  }

  async function submit(e: React.SyntheticEvent<HTMLFormElement>): Promise<number | null> {
    e.preventDefault()
    setError('')
    const result = videoFormSchema.safeParse(form)
    if (!result.success) {
      setError(result.error.issues[0]?.message || '表单校验失败')
      return null
    }
    const body = {
      ...result.data,
      director_ids: selectedDirectors,
      actors: selectedActors.map((id) => ({ actor_id: id })),
      tag_ids: selectedTags,
    }
    setSubmitting(true)
    try {
      let videoId = Number(id)
      if (isNew) {
        const res = await adminApi.createVideo(body)
        videoId = (res.data as VideoDetail).id
        toast.success('影视已创建')
      } else {
        if (!videoId) {
          setError('无效的影视 ID')
          return null
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
      return videoId
    } catch (err) {
      setError(errorMessage(err))
      return null
    } finally {
      setSubmitting(false)
    }
  }

  return {
    error,
    initLoading,
    submitting,
    categories,
    directors,
    actors,
    tags,
    sources,
    selectedDirectors,
    selectedActors,
    selectedTags,
    episodes,
    form,
    setForm,
    toggleDirector,
    toggleActor,
    toggleTag,
    addEpisode,
    updateEpisode,
    removeEpisode,
    submit,
  }
}
