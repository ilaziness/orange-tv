import { useCallback, useEffect, useState } from 'react'
import type * as React from 'react'
import { z } from 'zod'
import { adminApi, errorMessage } from '@/lib/api'
import { flattenCategories } from '@/lib/categories'
import type {
  Category,
  CollectCategoryMap,
  CollectLog,
  CollectSource,
  PlaySource,
  RemoteCategory,
} from '@orange-tv/shared'
import { toast } from 'sonner'
import { DEFAULT_PAGE_SIZE } from '@/lib/constants'

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

  if (hour === '*') {
    return `每小时 ${m}分`
  }

  const stepMatch = hour.match(/^\*\/(\d+)$/)
  if (stepMatch) {
    return `每${stepMatch[1]}小时 ${m}分`
  }

  const hours = hour
    .split(',')
    .map((h) => h.trim())
    .filter(Boolean)
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
  play_source_id: z
    .union([z.string(), z.number()])
    .transform((v) => Number(v))
    .refine((v) => v > 0, '请选择播放源'),
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
  data_range: 'today',
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

export type CategoryBindingItem = {
  external_category_id: number
  category_id: number
}

export function useCollect() {
  const [sources, setSources] = useState<CollectSource[]>([])
  const [sourcesTotal, setSourcesTotal] = useState(0)
  const [sourcesPage, setSourcesPage] = useState(1)
  const [sourcesPageSize] = useState(DEFAULT_PAGE_SIZE)
  const [playSources, setPlaySources] = useState<PlaySource[]>([])
  const [categories, setCategories] = useState<Category[]>([])
  const [logs, setLogs] = useState<CollectLog[]>([])
  const [logsTotal, setLogsTotal] = useState(0)
  const [logsPage, setLogsPage] = useState(1)
  const [logsPageSize] = useState(DEFAULT_PAGE_SIZE)
  const [maps, setMaps] = useState<CollectCategoryMap[]>([])
  const [remoteCategories, setRemoteCategories] = useState<RemoteCategory[]>([])
  const [formError, setFormError] = useState('')
  const [loading, setLoading] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [categoryLoading, setCategoryLoading] = useState(false)
  const [savingCategories, setSavingCategories] = useState(false)
  const [collecting, setCollecting] = useState(false)
  const [deleting, setDeleting] = useState(false)
  const [logsLoading, setLogsLoading] = useState(false)
  const [schedulingId, setSchedulingId] = useState<number | null>(null)

  // dialog states
  const [formOpen, setFormOpen] = useState(false)
  const [editId, setEditId] = useState(0)
  const [form, setForm] = useState(emptyForm)
  const [categoryOpen, setCategoryOpen] = useState(false)
  const [categorySourceId, setCategorySourceId] = useState(0)
  const [collectOpen, setCollectOpen] = useState(false)
  const [collectSourceId, setCollectSourceId] = useState(0)
  const [collectDataRange, setCollectDataRange] = useState('today')
  const [deleteId, setDeleteId] = useState<number | null>(null)

  const fetchSources = useCallback(
    async (page: number) => {
      const s = await adminApi.listCollectSources({ page, page_size: sourcesPageSize })
      setSources(s.data.list || [])
      setSourcesTotal(s.data.total || 0)
      setSourcesPage(page)
    },
    [sourcesPageSize],
  )

  const fetchLogs = useCallback(
    async (page: number) => {
      const l = await adminApi.listCollectLogs({ page, page_size: logsPageSize })
      setLogs(l.data.list || [])
      setLogsTotal(l.data.total || 0)
      setLogsPage(page)
    },
    [logsPageSize],
  )

  const loadSources = useCallback(
    async (page: number) => {
      setLoading(true)
      try {
        await fetchSources(page)
      } catch (err) {
        toast.error(errorMessage(err))
      } finally {
        setLoading(false)
      }
    },
    [fetchSources],
  )

  const loadLogs = useCallback(
    async (page: number) => {
      setLogsLoading(true)
      try {
        await fetchLogs(page)
      } catch (err) {
        toast.error(errorMessage(err))
      } finally {
        setLogsLoading(false)
      }
    },
    [fetchLogs],
  )

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const [p, c] = await Promise.all([adminApi.listPlaySources(), adminApi.listCategories()])
      setPlaySources(p.data.list || [])
      setCategories(c.data || [])
      await Promise.all([fetchSources(1), fetchLogs(1)])
    } catch (err) {
      toast.error(errorMessage(err))
    } finally {
      setLoading(false)
    }
  }, [fetchSources, fetchLogs])

  useEffect(() => {
    void load()
  }, [load])

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
      data_range: item.data_range || 'today',
    })
    setFormOpen(true)
  }

  async function onSubmit(e: React.SyntheticEvent<HTMLFormElement>) {
    e.preventDefault()
    if (submitting) return
    setFormError('')
    const result = collectSchema.safeParse(form)
    if (!result.success) {
      setFormError(result.error.issues[0]?.message || '表单校验失败')
      return
    }
    const cronExpr = formatCronExpr(result.data.cron_minute, result.data.cron_hour)
    const payload = {
      name: result.data.name,
      type: result.data.type,
      collect_url: result.data.collect_url,
      api_key: result.data.api_key,
      play_source_id: result.data.play_source_id,
      data_range: result.data.data_range,
      cron_expr: cronExpr,
    }
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
    setMaps([])
    setRemoteCategories([])
    setCategoryOpen(true)
    setCategoryLoading(true)
    try {
      const [mapRes, remoteRes] = await Promise.all([
        adminApi.getCollectCategories(sourceId),
        adminApi.fetchRemoteCategories(sourceId),
      ])
      setMaps(mapRes.data || [])
      setRemoteCategories(remoteRes.data.list || [])
    } catch (err) {
      toast.error(errorMessage(err))
    } finally {
      setCategoryLoading(false)
    }
  }

  async function saveCategoryBinding(items: CategoryBindingItem[]) {
    if (!categorySourceId) return
    setSavingCategories(true)
    try {
      const res = await adminApi.setCollectCategories(categorySourceId, { items })
      setMaps(res.data || [])
      toast.success('绑定分类已保存')
      setCategoryOpen(false)
    } catch (err) {
      toast.error(errorMessage(err))
    } finally {
      setSavingCategories(false)
    }
  }

  async function enableSchedule(id: number) {
    setSchedulingId(id)
    try {
      await adminApi.enableCollectSchedule(id)
      toast.success('定时采集已启用')
      await loadSources(sourcesPage)
    } catch (err) {
      toast.error(errorMessage(err))
    } finally {
      setSchedulingId(null)
    }
  }

  async function disableSchedule(id: number) {
    setSchedulingId(id)
    try {
      await adminApi.disableCollectSchedule(id)
      toast.success('定时采集已禁用')
      await loadSources(sourcesPage)
    } catch (err) {
      toast.error(errorMessage(err))
    } finally {
      setSchedulingId(null)
    }
  }

  async function confirmDelete() {
    if (deleteId === null) return
    setDeleting(true)
    try {
      await adminApi.deleteCollectSource(deleteId)
      toast.success('采集源已删除')
      await loadSources(sourcesPage)
    } catch (err) {
      toast.error(errorMessage(err))
    } finally {
      setDeleting(false)
      setDeleteId(null)
    }
  }

  function openCollectNow(sourceId: number) {
    const source = sources.find((s) => s.id === sourceId)
    setCollectSourceId(sourceId)
    setCollectDataRange(source?.data_range || 'today')
    setCollectOpen(true)
  }

  async function submitCollectNow() {
    if (!collectSourceId || collecting) return
    setCollecting(true)
    try {
      await adminApi.collectNow(collectSourceId, { data_range: collectDataRange })
      toast.success('采集已启动')
      setCollectOpen(false)
      await Promise.all([loadSources(sourcesPage), loadLogs(1)])
    } catch (err) {
      toast.error(errorMessage(err))
    } finally {
      setCollecting(false)
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
    formError,
    loading,
    submitting,
    categoryLoading,
    savingCategories,
    collecting,
    deleting,
    logsLoading,
    schedulingId,
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
