import { useCallback, useEffect, useRef, useState } from 'react'
import type * as React from 'react'
import { z } from 'zod'
import { adminApi, errorMessage } from '@/lib/api'
import type { AdItem } from '@orange-tv/shared'
import { toast } from 'sonner'
import { DEFAULT_PAGE_SIZE } from '@/lib/constants'

const adSchema = z.object({
  ad_key: z.string().min(1, '广告标识不能为空'),
  title: z.string().min(1, '标题不能为空'),
  scene: z.string().min(1, '请选择广告场景'),
  type: z.string().min(1, '请选择广告类型'),
  content_url: z.string(),
  content_code: z.string(),
  link_url: z.string(),
  duration: z.union([z.string(), z.number()]).transform((v) => Number(v) || 5),
  sort: z.union([z.string(), z.number()]).transform((v) => Number(v) || 0),
  status: z.union([z.string(), z.number()]).transform((v) => Number(v)),
})

const emptyForm = {
  ad_key: '',
  title: '',
  scene: '',
  type: '',
  content_url: '',
  content_code: '',
  link_url: '',
  duration: '5',
  sort: '0',
  status: '1',
}

export type AdFormType = typeof emptyForm

export function useAds() {
  const [list, setList] = useState<AdItem[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [loading, setLoading] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [deleting, setDeleting] = useState(false)
  const [dialogOpen, setDialogOpen] = useState(false)
  const [editingId, setEditingId] = useState(0)
  const [form, setForm] = useState<AdFormType>(emptyForm)
  const [deleteId, setDeleteId] = useState<number | null>(null)
  const pageRef = useRef(page)

  useEffect(() => {
    pageRef.current = page
  }, [page])

  const load = useCallback(async (p = pageRef.current) => {
    setLoading(true)
    try {
      const res = await adminApi.listAds({ page: p, page_size: DEFAULT_PAGE_SIZE })
      setList(res.data.list || [])
      setTotal(res.data.total || 0)
      setPage(p)
    } catch (err) {
      toast.error(errorMessage(err))
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void load(1)
  }, [load])

  function openCreate() {
    setEditingId(0)
    setForm(emptyForm)
    setDialogOpen(true)
  }

  function openEdit(ad: AdItem) {
    setEditingId(ad.id)
    setForm({
      ad_key: ad.ad_key,
      title: ad.title,
      scene: ad.scene,
      type: ad.type,
      content_url: ad.content_url,
      content_code: ad.content_code || '',
      link_url: ad.link_url,
      duration: String(ad.duration || 5),
      sort: String(ad.sort || 0),
      status: String(ad.status),
    })
    setDialogOpen(true)
  }

  async function onSubmit(e: React.SyntheticEvent<HTMLFormElement>) {
    e.preventDefault()
    if (submitting) return

    const result = adSchema.safeParse(form)
    if (!result.success) {
      toast.error(result.error.issues[0]?.message || '表单校验失败')
      return
    }

    const data = result.data
    // Business validation: type=code requires content_code; otherwise requires content_url.
    if (data.type === 'code') {
      if (!data.content_code.trim()) {
        toast.error('code类型广告必须提供广告代码')
        return
      }
    } else {
      if (!data.content_url.trim()) {
        toast.error('非code类型广告必须提供素材URL')
        return
      }
    }

    const submitData: Record<string, unknown> = {
      ad_key: data.ad_key,
      title: data.title,
      scene: data.scene,
      type: data.type,
      duration: data.duration,
      sort: data.sort,
      status: data.status,
    }
    if (data.type === 'code') {
      submitData.content_code = data.content_code
    } else {
      submitData.content_url = data.content_url
      submitData.link_url = data.link_url
    }

    setSubmitting(true)
    try {
      if (editingId) {
        await adminApi.updateAd(editingId, submitData)
        toast.success('更新成功')
      } else {
        await adminApi.createAd(submitData)
        toast.success('创建成功')
      }
      setDialogOpen(false)
      await load(editingId ? page : 1)
    } catch (err) {
      toast.error(errorMessage(err))
    } finally {
      setSubmitting(false)
    }
  }

  async function onToggle(ad: AdItem) {
    try {
      await adminApi.updateAd(ad.id, { status: ad.status === 1 ? 0 : 1 })
      toast.success(ad.status === 1 ? '已禁用' : '已启用')
      await load(page)
    } catch (err) {
      toast.error(errorMessage(err))
    }
  }

  async function confirmDelete() {
    if (deleteId === null) return
    setDeleting(true)
    try {
      await adminApi.deleteAd(deleteId)
      toast.success('广告已删除')
      await load(page)
    } catch (err) {
      toast.error(errorMessage(err))
    } finally {
      setDeleting(false)
      setDeleteId(null)
    }
  }

  const hasNext = page * DEFAULT_PAGE_SIZE < total

  return {
    list,
    total,
    page,
    loading,
    submitting,
    deleting,
    dialogOpen,
    setDialogOpen,
    editingId,
    form,
    setForm,
    deleteId,
    setDeleteId,
    hasNext,
    openCreate,
    openEdit,
    onSubmit,
    onToggle,
    confirmDelete,
    load,
    pageSize: DEFAULT_PAGE_SIZE,
  }
}
