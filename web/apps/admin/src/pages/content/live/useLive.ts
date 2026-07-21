import { useEffect, useState } from 'react'
import type * as React from 'react'
import { z } from 'zod'
import { adminApi, errorMessage } from '@/lib/api'
import type { LiveChannel } from '@orange-tv/shared'
import { toast } from 'sonner'

const liveSchema = z.object({
  name: z.string().min(1, '频道名称不能为空'),
  category: z.string(),
  stream_url: z.string().min(1, '直播流地址不能为空'),
  logo: z.string(),
  description: z.string(),
  sort_order: z.union([z.string(), z.number()]).transform((v) => Number(v) || 0),
  status: z.union([z.string(), z.number()]).transform((v) => Number(v)),
})

export const emptyForm = { name: '', category: '', stream_url: '', logo: '', description: '', sort_order: '0', status: '1' }

export type LiveFormState = typeof emptyForm

export function useLive() {
  const [list, setList] = useState<LiveChannel[]>([])
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const [form, setForm] = useState(emptyForm)
  const [editId, setEditId] = useState(0)
  const [deleteId, setDeleteId] = useState<number | null>(null)

  async function load() {
    setLoading(true)
    setError('')
    try {
      const res = await adminApi.listLive({ page: 1, page_size: 100 })
      setList(res.data.list || [])
    } catch (err) {
      setError(errorMessage(err))
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { void load() }, [])

  async function onSubmit(e: React.SyntheticEvent<HTMLFormElement>) {
    e.preventDefault()
    setError('')
    const result = liveSchema.safeParse(form)
    if (!result.success) {
      setError(result.error.issues[0]?.message || '表单校验失败')
      return
    }
    try {
      if (editId) {
        await adminApi.updateLive(editId, result.data)
        toast.success('直播频道已更新')
      } else {
        await adminApi.createLive(result.data)
        toast.success('直播频道已创建')
      }
      setForm(emptyForm)
      setEditId(0)
      await load()
    } catch (err) {
      setError(errorMessage(err))
    }
  }

  async function confirmDelete() {
    if (deleteId === null) return
    try {
      await adminApi.deleteLive(deleteId)
      toast.success('直播频道已删除')
      await load()
    } catch (err) {
      toast.error(errorMessage(err))
    } finally {
      setDeleteId(null)
    }
  }

  function startEdit(item: LiveChannel) {
    setEditId(item.id)
    setForm({
      name: item.name,
      category: item.category || '',
      stream_url: item.stream_url,
      logo: item.logo || '',
      description: item.description || '',
      sort_order: String(item.sort_order ?? 0),
      status: String(item.status ?? 1),
    })
  }

  return {
    list,
    error,
    loading,
    form,
    setForm,
    editId,
    setEditId,
    deleteId,
    setDeleteId,
    onSubmit,
    confirmDelete,
    startEdit,
    load,
  }
}
