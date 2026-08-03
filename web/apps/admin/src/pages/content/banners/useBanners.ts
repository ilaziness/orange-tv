import { useCallback, useEffect, useRef, useState } from 'react'
import type * as React from 'react'
import { z } from 'zod'
import { adminApi, errorMessage } from '@/lib/api'
import type { BannerItem } from '@orange-tv/shared'
import { toast } from 'sonner'
import { DEFAULT_PAGE_SIZE } from '@/lib/constants'

const bannerSchema = z.object({
  title: z.string().min(1, '标题不能为空'),
  cover: z.string(),
  link: z.string(),
  video_id: z.union([z.string(), z.number()]).transform((v) => Number(v) || 0),
  sort: z.union([z.string(), z.number()]).transform((v) => Number(v) || 0),
  status: z.union([z.string(), z.number()]).transform((v) => Number(v)),
})

const emptyForm = { title: '', cover: '', link: '', video_id: '', sort: '', status: '' }

export type BannerFormType = typeof emptyForm

export function useBanners() {
  const [list, setList] = useState<BannerItem[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [loading, setLoading] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [deleting, setDeleting] = useState(false)
  const [dialogOpen, setDialogOpen] = useState(false)
  const [editingId, setEditingId] = useState(0)
  const [form, setForm] = useState<BannerFormType>(emptyForm)
  const [selectedVideo, setSelectedVideo] = useState<{ id: number; title: string } | null>(null)
  const [videoPickerOpen, setVideoPickerOpen] = useState(false)
  const [deleteId, setDeleteId] = useState<number | null>(null)
  const pageRef = useRef(page)

  useEffect(() => {
    pageRef.current = page
  }, [page])

  const load = useCallback(async (p = pageRef.current) => {
    setLoading(true)
    try {
      const res = await adminApi.listBanners({ page: p, page_size: DEFAULT_PAGE_SIZE })
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
    setSelectedVideo(null)
    setDialogOpen(true)
  }

  function openEdit(banner: BannerItem) {
    setEditingId(banner.id)
    setForm({
      title: banner.title,
      cover: banner.cover,
      link: banner.link,
      video_id: String(banner.video_id || ''),
      sort: String(banner.sort || ''),
      status: String(banner.status),
    })
    setSelectedVideo(banner.video_id ? { id: banner.video_id, title: '' } : null)
    setDialogOpen(true)
  }

  async function onSubmit(e: React.SyntheticEvent<HTMLFormElement>) {
    e.preventDefault()
    if (submitting) return

    const formToValidate = {
      ...form,
      video_id: form.video_id || (selectedVideo?.id ?? 0),
    }
    const result = bannerSchema.safeParse(formToValidate)
    if (!result.success) {
      toast.error(result.error.issues[0]?.message || '表单校验失败')
      return
    }

    const submitData = {
      ...result.data,
      status: result.data.status || 1,
    }

    setSubmitting(true)
    try {
      if (editingId) {
        await adminApi.updateBanner(editingId, submitData)
        toast.success('更新成功')
      } else {
        await adminApi.createBanner(submitData)
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

  async function onToggle(b: BannerItem) {
    try {
      await adminApi.updateBanner(b.id, { status: b.status === 1 ? 0 : 1 })
      toast.success(b.status === 1 ? '已禁用' : '已启用')
      await load(page)
    } catch (err) {
      toast.error(errorMessage(err))
    }
  }

  async function confirmDelete() {
    if (deleteId === null) return
    setDeleting(true)
    try {
      await adminApi.deleteBanner(deleteId)
      toast.success('Banner 已删除')
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
    selectedVideo,
    setSelectedVideo,
    videoPickerOpen,
    setVideoPickerOpen,
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
