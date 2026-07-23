import { useCallback, useEffect, useState } from 'react'
import type * as React from 'react'
import { z } from 'zod'
import { adminApi, errorMessage } from '@/lib/api'
import { flattenCategories } from '@/lib/categories'
import type { Category, CollectCategoryMap, CollectLog, CollectSource, PlaySource, RemoteCategory } from '@orange-tv/shared'
import { toast } from 'sonner'

/**
 * parseCronExpr parses a cron expression like "0 8 * * *" or "30 6,18 * * *"
 * into minute and hour fields. Returns empty strings if unparseable.
 */
function parseCronExpr(expr: string): { minute: string; hour: string } {
  const parts = expr.trim().split(/\s+/)
  if (parts.length < 2) return { minute: '0', hour: '' }
  return { minute: parts[0], hour: parts[1] }
}

/**
 * formatCronExpr combines minute and hour into a cron expression.
 * Empty hour means no schedule (wildcard). Returns empty string if both empty.
 */
function formatCronExpr(minute: string, hour: string): string {
  const m = minute.trim() || '0'
  const h = hour.trim()
  if (!h) return ''
  return `${m} ${h} * * *`
}

/**
 * formatCronFriendly converts a cron expression to a human-readable string.
 * e.g. "0 8 * * *" -> "每天 08:00", "30 6,18 * * *" -> "每天 06:30, 18:30"
 */
export function formatCronFriendly(expr: string): string {
  if (!expr) return '未设置'
  const { minute, hour } = parseCronExpr(expr)
  if (!hour) return '未设置'
  const m = minute.padStart(2, '0')
  const hours = hour.split(',').map((h) => h.trim()).filter(Boolean)
  if (hours.length === 0) return '未设置'
  const times = hours.map((h) => `${h.padStart(2, '0')}:${m}`)
  return `每天 ${times.join(', ')}`
}

const collectSchema = z.object({
  name: z.string().min(1, '源名称不能为空'),
  type: z.union([z.string(), z.number()]).transform((v) => Number(v)),
  collect_url: z.string().min(1, '采集地址不能为空'),
  api_key: z.string(),
  cron_minute: z.string(),
  cron_hour: z.string(),
  play_source_id: z.union([z.string(), z.number()]).transform((v) => Number(v)).refine((v) => v > 0, '请选择播放源'),
  data_range: z.string(),
})

const emptyForm = {
  name: '',
  type: '2',
  collect_url: '',
  api_key: '',
  cron_minute: '0',
  cron_hour: '',
  play_source_id: '0',
  data_range: 'all',
}

export type CollectForm = typeof emptyForm

const DATA_RANGE_OPTIONS = [
  { value: 'today', label: '今天' },
  { value: 'last1d', label: '近1天' },
  { value: 'last3d', label: '近3天' },
  { value: 'last1w', label: '近1周' },
  { value: 'last1m', label: '近1月' },
  { value: 'all', label: '全部' },
]

export function useCollect() {
  const [sources, setSources] = useState<CollectSource[]>([])
  const [sourcesTotal, setSourcesTotal] = useState(0)
  const [sourcesPage, setSourcesPage] = useState(1)
  const [sourcesPageSize] = useState(20)
  const [playSources, setPlaySources] = useState<PlaySource[]>([])
  const [categories, setCategories] = useState<Category[]>([])
  const [logs, setLogs] = useState<CollectLog[]>([])
  const [logsTotal, setLogsTotal] = useState(0)
  const [logsPage, setLogsPage] = useState(1)
  const [logsPageSize] = useState(20)
  const [maps, setMaps] = useState<CollectCategoryMap[]>([])
  const [remoteCategories, setRemoteCategories] = useState<RemoteCategory[]>([])
  const [error, setError] = useState('')
  const [formError, setFormError] = useState('')
  const [loading, setLoading] = useState(false)
  const [submitting, setSubmitting] = useState(false)

  // dialog states
  const [formOpen, setFormOpen] = useState(false)
  const [editId, setEditId] = useState(0)
  const [form, setForm] = useState(emptyForm)
  const [categoryOpen, setCategoryOpen] = useState(false)
  const [categorySourceId, setCategorySourceId] = useState(0)
  const [collectOpen, setCollectOpen] = useState(false)
  const [collectSourceId, setCollectSourceId] = useState(0)
  const [collectDataRange, setCollectDataRange] = useState('all')
  const [deleteId, setDeleteId] = useState<number | null>(null)

  const loadSources = useCallback(async (page: number) => {
    setLoading(true)
    setError('')
    try {
      const s = await adminApi.listCollectSources({ page, page_size: 20 })
      setSources(s.data.list || [])
      setSourcesTotal(s.data.total || 0)
      setSourcesPage(page)
    } catch (err) {
      setError(errorMessage(err))
    } finally {
      setLoading(false)
    }
  }, [])

  const loadLogs = useCallback(async (page: number) => {
    try {
      const l = await adminApi.listCollectLogs({ page, page_size: 20 })
      setLogs(l.data.list || [])
      setLogsTotal(l.data.total || 0)
      setLogsPage(page)
    } catch (err) {
      setError(errorMessage(err))
    }
  }, [])

  const load = useCallback(async () => {
    setError('')
    try {
      const [p, c] = await Promise.all([
        adminApi.listPlaySources(),
        adminApi.listCategories(),
      ])
      setPlaySources(p.data.list || [])
      setCategories(c.data || [])
      await Promise.all([loadSources(1), loadLogs(1)])
    } catch (err) {
      setError(errorMessage(err))
    }
  }, [loadSources, loadLogs])

  useEffect(() => { void load() }, [load])

  const flatCats = flattenCategories(categories)

  function openCreate() {
    setEditId(0)
    setForm(emptyForm)
    setFormError('')
    setFormOpen(true)
  }

  function openEdit(item: CollectSource) {
    setEditId(item.id)
    setFormError('')
    const { minute, hour } = parseCronExpr(item.cron_expr || '')
    setForm({
      name: item.name,
      type: String(item.type),
      collect_url: item.collect_url,
      api_key: '',
      cron_minute: minute,
      cron_hour: hour,
      play_source_id: String(item.play_source_id),
      data_range: item.data_range || 'all',
    })
    setFormOpen(true)
  }

  async function onSubmit(e: React.SyntheticEvent<HTMLFormElement>) {
    e.preventDefault()
    setFormError('')
    const result = collectSchema.safeParse(form)
    if (!result.success) {
      setFormError(result.error.issues[0]?.message || '表单校验失败')
      return
    }
    const cronExpr = formatCronExpr(result.data.cron_minute, result.data.cron_hour)
    const payload = { ...result.data, cron_expr: cronExpr }
    delete (payload as Record<string, unknown>).cron_minute
    delete (payload as Record<string, unknown>).cron_hour
    setSubmitting(true)
    try {
      if (editId) {
        await adminApi.updateCollectSource(editId, payload)
        toast.success('采集源已更新')
      } else {
        await adminApi.createCollectSource(payload)
        toast.success('采集源已创建')
      }
      setFormOpen(false)
      await loadSources(sourcesPage)
    } catch (err) {
      setFormError(errorMessage(err))
    } finally {
      setSubmitting(false)
    }
  }

  async function openCategoryBinding(sourceId: number) {
    setCategorySourceId(sourceId)
    setCategoryOpen(true)
    try {
      const [mapRes, remoteRes] = await Promise.all([
        adminApi.getCollectCategories(sourceId),
        adminApi.fetchRemoteCategories(sourceId),
      ])
      setMaps(mapRes.data || [])
      setRemoteCategories(remoteRes.data.list || [])
    } catch (err) {
      setError(errorMessage(err))
    }
  }

  async function saveCategoryBinding(items: { external_category: string; category_id: number }[]) {
    if (!categorySourceId) return
    try {
      const res = await adminApi.setCollectCategories(categorySourceId, { items })
      setMaps(res.data || [])
      toast.success('分类映射已保存')
    } catch (err) {
      toast.error(errorMessage(err))
    }
  }

  async function enableSchedule(id: number) {
    try {
      await adminApi.enableCollectSchedule(id)
      toast.success('定时采集已启用')
      await loadSources(sourcesPage)
    } catch (err) {
      toast.error(errorMessage(err))
    }
  }

  async function disableSchedule(id: number) {
    try {
      await adminApi.disableCollectSchedule(id)
      toast.success('定时采集已禁用')
      await loadSources(sourcesPage)
    } catch (err) {
      toast.error(errorMessage(err))
    }
  }

  async function confirmDelete() {
    if (deleteId === null) return
    try {
      await adminApi.deleteCollectSource(deleteId)
      toast.success('采集源已删除')
      await loadSources(sourcesPage)
    } catch (err) {
      toast.error(errorMessage(err))
    } finally {
      setDeleteId(null)
    }
  }

  function openCollectNow(sourceId: number) {
    setCollectSourceId(sourceId)
    setCollectDataRange('all')
    setCollectOpen(true)
  }

  async function submitCollectNow() {
    if (!collectSourceId) return
    try {
      await adminApi.collectNow(collectSourceId, { data_range: collectDataRange })
      toast.success('采集已启动')
      setCollectOpen(false)
      await loadSources(sourcesPage)
    } catch (err) {
      toast.error(errorMessage(err))
    }
  }

  return {
    sources,
    sourcesTotal,
    sourcesPage,
    sourcesPageSize,
    playSources,
    flatCats,
    logs,
    logsTotal,
    logsPage,
    logsPageSize,
    maps,
    remoteCategories,
    error,
    formError,
    loading,
    submitting,
    formOpen,
    setFormOpen,
    editId,
    form,
    setForm,
    categoryOpen,
    setCategoryOpen,
    categorySourceId,
    collectOpen,
    setCollectOpen,
    collectSourceId,
    collectDataRange,
    setCollectDataRange,
    deleteId,
    setDeleteId,
    DATA_RANGE_OPTIONS,
    loadSources,
    loadLogs,
    openCreate,
    openEdit,
    onSubmit,
    openCategoryBinding,
    saveCategoryBinding,
    enableSchedule,
    disableSchedule,
    openCollectNow,
    submitCollectNow,
    confirmDelete,
    load,
  }
}
