import { useEffect, useState } from 'react'
import type * as React from 'react'
import { z } from 'zod'
import { adminApi, errorMessage } from '@/lib/api'
import type { BannerItem } from '@orange-tv/shared'
import { toast } from 'sonner'

const bannerSchema = z.object({
  title: z.string().min(1, '标题不能为空'),
  cover: z.string(),
  link: z.string(),
  video_id: z.union([z.string(), z.number()]).transform((v) => Number(v) || 0),
  sort: z.union([z.string(), z.number()]).transform((v) => Number(v) || 0),
  status: z.union([z.string(), z.number()]).transform((v) => Number(v)),
})

const emptyForm = { title: '', cover: '', link: '', video_id: '0', sort: '0', status: '1' }

export type BannerForm = typeof emptyForm

export function useBanners() {
  const [list, setList] = useState<BannerItem[]>([])
  const [total, setTotal] = useState(0)
  const [error, setError] = useState('')
  const [showCreate, setShowCreate] = useState(false)
  const [form, setForm] = useState(emptyForm)
  const [deleteId, setDeleteId] = useState<number | null>(null)

  async function load() {
    setError('')
    try {
      const res = await adminApi.listBanners({ page_size: 100 })
      setList(res.data.list || [])
      setTotal(res.data.total)
    } catch (err) {
      setError(errorMessage(err))
    }
  }

  useEffect(() => { void load() }, [])

  async function onCreate(e: React.SyntheticEvent<HTMLFormElement>) {
    e.preventDefault()
    setError('')
    const result = bannerSchema.safeParse(form)
    if (!result.success) {
      setError(result.error.issues[0]?.message || '表单校验失败')
      return
    }
    try {
      await adminApi.createBanner(result.data)
      toast.success('Banner 已创建')
      setShowCreate(false)
      setForm(emptyForm)
      await load()
    } catch (err) {
      setError(errorMessage(err))
    }
  }

  async function onToggle(b: BannerItem) {
    try {
      await adminApi.updateBanner(b.id, { status: b.status === 1 ? 0 : 1 })
      toast.success(b.status === 1 ? '已禁用' : '已启用')
      await load()
    } catch (err) {
      toast.error(errorMessage(err))
    }
  }

  async function confirmDelete() {
    if (deleteId === null) return
    try {
      await adminApi.deleteBanner(deleteId)
      toast.success('Banner 已删除')
      await load()
    } catch (err) {
      toast.error(errorMessage(err))
    } finally {
      setDeleteId(null)
    }
  }

  return {
    list,
    total,
    error,
    showCreate,
    setShowCreate,
    form,
    setForm,
    deleteId,
    setDeleteId,
    onCreate,
    onToggle,
    confirmDelete,
  }
}
