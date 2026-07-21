import { useEffect, useState } from 'react'
import type * as React from 'react'
import { z } from 'zod'
import { adminApi, errorMessage } from '@/lib/api'
import { flattenCategories } from '@/lib/categories'
import type { Category, CollectCategoryMap, CollectLog, CollectSource, PlaySource } from '@orange-tv/shared'
import { toast } from 'sonner'

const collectMapSchema = z.array(z.object({
  external_category: z.string().min(1, '外部分类键不能为空'),
  category_id: z.union([z.string(), z.number()]).transform((v) => Number(v)),
}))

const collectSchema = z.object({
  name: z.string().min(1, '源名称不能为空'),
  type: z.union([z.string(), z.number()]).transform((v) => Number(v)),
  collect_url: z.string().min(1, '采集地址不能为空'),
  api_key: z.string(),
  cron_expr: z.string(),
  play_source_id: z.union([z.string(), z.number()]).transform((v) => Number(v)),
  status: z.union([z.string(), z.number()]).transform((v) => Number(v)),
  config: z.string(),
})

const emptyForm = { name: '', type: '2', collect_url: '', api_key: '', cron_expr: '', play_source_id: '0', status: '1', config: '' }

export type CollectForm = typeof emptyForm

export function useCollect() {
  const [sources, setSources] = useState<CollectSource[]>([])
  const [playSources, setPlaySources] = useState<PlaySource[]>([])
  const [categories, setCategories] = useState<Category[]>([])
  const [logs, setLogs] = useState<CollectLog[]>([])
  const [maps, setMaps] = useState<CollectCategoryMap[]>([])
  const [error, setError] = useState('')
  const [selectedId, setSelectedId] = useState(0)
  const [form, setForm] = useState(emptyForm)
  const [mapText, setMapText] = useState('[]')
  const [deleteId, setDeleteId] = useState<number | null>(null)

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

  async function onSubmit(e: React.SyntheticEvent<HTMLFormElement>) {
    e.preventDefault()
    setError('')
    const result = collectSchema.safeParse(form)
    if (!result.success) {
      setError(result.error.issues[0]?.message || '表单校验失败')
      return
    }
    try {
      if (selectedId) {
        await adminApi.updateCollectSource(selectedId, result.data)
        toast.success('采集源已更新')
      } else {
        await adminApi.createCollectSource(result.data)
        toast.success('采集源已创建')
      }
      setSelectedId(0)
      setForm(emptyForm)
      await load()
    } catch (err) {
      setError(errorMessage(err))
    }
  }

  async function editSource(item: CollectSource) {
    setSelectedId(item.id)
    setForm({
      name: item.name,
      type: String(item.type),
      collect_url: item.collect_url,
      api_key: '',
      cron_expr: item.cron_expr || '',
      play_source_id: String(item.play_source_id),
      status: String(item.status),
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
      const parsed = JSON.parse(mapText || '[]')
      const result = collectMapSchema.safeParse(parsed)
      if (!result.success) {
        toast.error(result.error.issues[0]?.message || '映射 JSON 格式错误')
        return
      }
      const res = await adminApi.setCollectCategories(selectedId, { items: result.data })
      setMaps(res.data || [])
      setMapText(JSON.stringify((res.data || []).map((m) => ({
        external_category: m.external_category,
        category_id: m.category_id,
      })), null, 2))
      toast.success('映射已保存')
      await load()
    } catch (err) {
      toast.error(errorMessage(err))
    }
  }

  async function start(id: number) {
    try {
      await adminApi.startCollect(id)
      toast.success('采集已启动')
      await load()
    } catch (err) {
      toast.error(errorMessage(err))
    }
  }

  async function stop(id: number) {
    try {
      await adminApi.stopCollect(id)
      toast.success('采集已停止')
      await load()
    } catch (err) {
      toast.error(errorMessage(err))
    }
  }

  async function confirmDelete() {
    if (deleteId === null) return
    try {
      await adminApi.deleteCollectSource(deleteId)
      toast.success('采集源已删除')
      await load()
    } catch (err) {
      toast.error(errorMessage(err))
    } finally {
      setDeleteId(null)
    }
  }

  return {
    sources,
    playSources,
    flatCats,
    logs,
    maps,
    error,
    selectedId,
    setSelectedId,
    form,
    setForm,
    mapText,
    setMapText,
    deleteId,
    setDeleteId,
    onSubmit,
    editSource,
    saveMaps,
    start,
    stop,
    confirmDelete,
    load,
  }
}
