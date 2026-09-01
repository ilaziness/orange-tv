import { useCallback, useEffect, useRef, useState } from 'react'
import type * as React from 'react'
import { z } from 'zod'
import { adminApi, errorMessage } from '@/lib/api'
import type { LiveTVChannel } from '@orange-tv/shared'
import { toast } from 'sonner'
import { DEFAULT_PAGE_SIZE } from '@/lib/constants'

const liveTVSchema = z.object({
  name: z.string().min(1, '频道名称不能为空'),
  category: z.string(),
  stream_url: z.string().min(1, '直播流地址不能为空'),
  logo: z.string(),
  description: z.string(),
  sort_order: z.union([z.string(), z.number()]).transform((v) => Number(v) || 0),
  status: z.union([z.string(), z.number()]).transform((v) => Number(v)),
})

export const emptyForm = {
  name: '',
  category: '',
  stream_url: '',
  logo: '',
  description: '',
  sort_order: '',
  status: '1',
}

export type LiveTVFormState = typeof emptyForm

export function useLiveTV() {
  const [list, setList] = useState<LiveTVChannel[]>([])
  const [error, setError] = useState('')
  const [dialogError, setDialogError] = useState('')
  const [loading, setLoading] = useState(false)
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)
  const pageRef = useRef(page)

  useEffect(() => {
    pageRef.current = page
  }, [page])
  const [form, setForm] = useState(emptyForm)
  const [editId, setEditId] = useState(0)
  const [deleteId, setDeleteId] = useState<number | null>(null)
  const [dialogOpen, setDialogOpen] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [deleting, setDeleting] = useState(false)
  const [syncing, setSyncing] = useState(false)
  const [syncDialogOpen, setSyncDialogOpen] = useState(false)
  const [syncUrl, setSyncUrl] = useState('')

  const load = useCallback(async (p = pageRef.current) => {
    setLoading(true)
    setError('')
    try {
      const res = await adminApi.listLiveTV({ page: p, page_size: DEFAULT_PAGE_SIZE })
      setList(res.data.list || [])
      setTotal(res.data.total || 0)
      setPage(res.data.page || p)
    } catch (err) {
      setError(errorMessage(err))
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void load(1)
  }, [load])

  async function onSubmit(e: React.SyntheticEvent<HTMLFormElement>) {
    e.preventDefault()
    setDialogError('')
    const result = liveTVSchema.safeParse(form)
    if (!result.success) {
      setDialogError(result.error.issues[0]?.message || '表单校验失败')
      return
    }
    setSubmitting(true)
    try {
      if (editId) {
        await adminApi.updateLiveTV(editId, result.data)
        toast.success('直播频道已更新')
      } else {
        await adminApi.createLiveTV(result.data)
        toast.success('直播频道已创建')
      }
      setForm(emptyForm)
      setEditId(0)
      setDialogOpen(false)
      await load(page)
    } catch (err) {
      setDialogError(errorMessage(err))
    } finally {
      setSubmitting(false)
    }
  }

  async function confirmDelete() {
    if (deleteId === null) return
    setDeleting(true)
    try {
      await adminApi.deleteLiveTV(deleteId)
      toast.success('直播频道已删除')
      await load(page)
    } catch (err) {
      toast.error(errorMessage(err))
    } finally {
      setDeleting(false)
      setDeleteId(null)
    }
  }

  async function confirmSync() {
    const url = syncUrl.trim()
    if (!url) {
      toast.error('请输入直播源地址')
      return
    }
    setSyncDialogOpen(false)
    setSyncing(true)
    try {
      const res = await adminApi.syncLiveTVSource(url)
      toast.success(
        `同步完成：共 ${res.data.total} 条，新增 ${res.data.created}，更新 ${res.data.updated}，删除 ${res.data.deleted}`,
      )
      setSyncUrl('')
      await load(1)
    } catch (err) {
      toast.error(errorMessage(err))
    } finally {
      setSyncing(false)
    }
  }

  async function openSyncDialog() {
    setSyncDialogOpen(true)
    try {
      const res = await adminApi.getLiveTVSyncSource()
      setSyncUrl(res.data.source_url || '')
    } catch {
      setSyncUrl('')
    }
  }

  function closeSyncDialog(open: boolean) {
    setSyncDialogOpen(open)
    if (!open) {
      setSyncUrl('')
    }
  }

  function openCreate() {
    setForm(emptyForm)
    setEditId(0)
    setDialogError('')
    setDialogOpen(true)
  }

  function openEdit(item: LiveTVChannel) {
    setEditId(item.id)
    setForm({
      name: item.name,
      category: item.category || '',
      stream_url: item.stream_url,
      logo: item.logo || '',
      description: item.description || '',
      sort_order: item.sort_order ? String(item.sort_order) : '',
      status: String(item.status ?? 1),
    })
    setDialogError('')
    setDialogOpen(true)
  }

  function closeDialog(open: boolean) {
    setDialogOpen(open)
    if (!open) {
      setForm(emptyForm)
      setEditId(0)
      setDialogError('')
    }
  }

  return {
    list,
    error,
    loading,
    page,
    total,
    form,
    setForm,
    editId,
    setEditId,
    deleteId,
    setDeleteId,
    dialogOpen,
    closeDialog,
    dialogError,
    submitting,
    deleting,
    syncing,
    syncDialogOpen,
    syncUrl,
    setSyncUrl,
    closeSyncDialog,
    confirmSync,
    openSyncDialog,
    onSubmit,
    confirmDelete,
    openCreate,
    openEdit,
    load,
  }
}
