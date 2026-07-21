import { useEffect, useState } from 'react'
import { z } from 'zod'
import { adminApi, errorMessage } from '@/lib/api'
import type { PlaySource } from '@orange-tv/shared'
import { toast } from 'sonner'

const playSourceNameSchema = z.string().trim().min(1, '播放源名称不能为空')

export function usePlaySources() {
  const [items, setItems] = useState<PlaySource[]>([])
  const [name, setName] = useState('')
  const [error, setError] = useState('')
  const [deleteId, setDeleteId] = useState<number | null>(null)

  async function load() {
    try {
      const res = await adminApi.listPlaySources()
      setItems(res.data.list || [])
    } catch (err) {
      setError(errorMessage(err))
    }
  }

  useEffect(() => { void load() }, [])

  async function create() {
    const result = playSourceNameSchema.safeParse(name)
    if (!result.success) {
      toast.error(result.error.issues[0]?.message || '播放源名称不能为空')
      return
    }
    try {
      await adminApi.createPlaySource({ name: result.data, status: 1 })
      setName('')
      toast.success('播放源已创建')
      await load()
    } catch (err) {
      toast.error(errorMessage(err))
    }
  }

  async function toggleStatus(item: PlaySource) {
    try {
      await adminApi.updatePlaySource(item.id, { status: item.status === 1 ? 0 : 1 })
      toast.success(item.status === 1 ? '已禁用' : '已启用')
      await load()
    } catch (err) {
      toast.error(errorMessage(err))
    }
  }

  async function confirmDelete() {
    if (deleteId === null) return
    try {
      await adminApi.deletePlaySource(deleteId)
      toast.success('播放源已删除')
      await load()
    } catch (err) {
      toast.error(errorMessage(err))
    } finally {
      setDeleteId(null)
    }
  }

  return {
    items,
    name,
    setName,
    error,
    deleteId,
    setDeleteId,
    create,
    toggleStatus,
    confirmDelete,
  }
}
