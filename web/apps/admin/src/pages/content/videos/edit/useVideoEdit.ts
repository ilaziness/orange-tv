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
  const [sources, setSources] = useState<PlaySource[]>([])
  const [selectedDirectors, setSelectedDirectors] = useState<NamedItem[]>([])
  const [selectedActors, setSelectedActors] = useState<NamedItem[]>([])
  const [selectedTags, setSelectedTags] = useState<NamedItem[]>([])
  const [episodes, setEpisodes] = useState<EpisodeDraft[]>([])
  const [form, setForm] = useState(emptyForm)
  const [submitting, setSubmitting] = useState(false)
  const [initLoading, setInitLoading] = useState(true)

  useEffect(() => {
    void (async () => {
      setInitLoading(true)
      try {
        const [cats, srcs] = await Promise.all([
          adminApi.listCategories(),
          adminApi.listPlaySources(),
        ])
        setCategories(flattenCategories(cats.data || []))
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
            release_date: d.release_date || '',
          })
          setSelectedDirectors(d.directors || [])
          setSelectedActors(d.actors || [])
          setSelectedTags(d.tags || [])
        }
      } catch (err) {
        setError(errorMessage(err))
      } finally {
        setInitLoading(false)
      }
    })()
  }, [id, isNew])

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
      director_ids: selectedDirectors.map((x) => x.id),
      actors: selectedActors.map((x) => ({ actor_id: x.id })),
      tag_ids: selectedTags.map((x) => x.id),
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
    sources,
    selectedDirectors,
    selectedActors,
    selectedTags,
    episodes,
    form,
    setForm,
    setSelectedDirectors,
    setSelectedActors,
    setSelectedTags,
    addEpisode,
    updateEpisode,
    removeEpisode,
    submit,
  }
}
