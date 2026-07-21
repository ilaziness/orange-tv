import { useEffect, useState } from 'react'
import type * as React from 'react'
import { z } from 'zod'
import { adminApi, errorMessage } from '@/lib/api'
import type { UserGroupItem } from '@orange-tv/shared'
import { toast } from 'sonner'

const groupSchema = z.object({
  name: z.string().min(1, '用户组名称不能为空'),
  permissions: z.string(),
  description: z.string(),
})

const emptyForm = { name: '', permissions: '', description: '' }

export type GroupForm = typeof emptyForm

export function useGroups() {
  const [list, setList] = useState<UserGroupItem[]>([])
  const [total, setTotal] = useState(0)
  const [error, setError] = useState('')
  const [showCreate, setShowCreate] = useState(false)
  const [form, setForm] = useState(emptyForm)
  const [deleteId, setDeleteId] = useState<number | null>(null)

  async function load() {
    setError('')
    try {
      const res = await adminApi.listGroups({ page_size: 100 })
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
    const result = groupSchema.safeParse(form)
    if (!result.success) {
      setError(result.error.issues[0]?.message || '表单校验失败')
      return
    }
    try {
      await adminApi.createGroup(result.data)
      toast.success('用户组已创建')
      setShowCreate(false)
      setForm(emptyForm)
      await load()
    } catch (err) {
      setError(errorMessage(err))
    }
  }

  async function confirmDelete() {
    if (deleteId === null) return
    try {
      await adminApi.deleteGroup(deleteId)
      toast.success('用户组已删除')
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
    confirmDelete,
  }
}
