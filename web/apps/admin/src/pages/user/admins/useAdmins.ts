import { useCallback, useEffect, useRef, useState } from 'react'
import type * as React from 'react'
import { z } from 'zod'
import { adminApi, errorMessage } from '@/lib/api'
import type { AdminItem, UserGroupItem } from '@orange-tv/shared'
import { toast } from 'sonner'

const adminSchema = z.object({
  username: z.string().min(1, '用户名不能为空'),
  password: z.string().min(6, '密码至少 6 位'),
  email: z.string(),
  group_id: z.union([z.string(), z.number()]).transform((v) => Number(v)),
  status: z.union([z.string(), z.number()]).transform((v) => Number(v)),
})

const emptyForm = { username: '', password: '', email: '', group_id: '1', status: '1' }

export type AdminForm = typeof emptyForm

export function useAdmins() {
  const [list, setList] = useState<AdminItem[]>([])
  const [total, setTotal] = useState(0)
  const [keyword, setKeyword] = useState('')
  const [error, setError] = useState('')
  const [showCreate, setShowCreate] = useState(false)
  const [form, setForm] = useState(emptyForm)
  const [groups, setGroups] = useState<UserGroupItem[]>([])
  const [queryKey, setQueryKey] = useState(0)
  const [deleteId, setDeleteId] = useState<number | null>(null)
  const [resetId, setResetId] = useState<number | null>(null)
  const keywordRef = useRef(keyword)

  useEffect(() => { keywordRef.current = keyword }, [keyword])

  const load = useCallback(async (k = keywordRef.current) => {
    setError('')
    try {
      const [res, gRes] = await Promise.all([
        adminApi.listAdmins({ page: 1, page_size: 20, keyword: k || undefined }),
        adminApi.listGroups({ page_size: 100 }),
      ])
      setList(res.data.list || [])
      setTotal(res.data.total)
      setGroups(gRes.data.list || [])
    } catch (err) {
      setError(errorMessage(err))
    }
  }, [])

  useEffect(() => { void load() }, [queryKey, load])

  async function onCreate(e: React.SyntheticEvent<HTMLFormElement>) {
    e.preventDefault()
    setError('')
    const result = adminSchema.safeParse(form)
    if (!result.success) {
      setError(result.error.issues[0]?.message || '表单校验失败')
      return
    }
    try {
      await adminApi.createAdmin(result.data)
      toast.success('管理员已创建')
      setShowCreate(false)
      setForm(emptyForm)
      await load()
    } catch (err) {
      setError(errorMessage(err))
    }
  }

  async function confirmResetPwd(pwd: string) {
    if (resetId === null) return
    try {
      await adminApi.resetAdminPassword(resetId, pwd)
      toast.success('密码已重置')
    } catch (err) {
      toast.error(errorMessage(err))
    } finally {
      setResetId(null)
    }
  }

  async function confirmDelete() {
    if (deleteId === null) return
    try {
      await adminApi.deleteAdmin(deleteId)
      toast.success('管理员已删除')
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
    keyword,
    setKeyword,
    error,
    showCreate,
    setShowCreate,
    form,
    setForm,
    groups,
    setQueryKey,
    deleteId,
    setDeleteId,
    resetId,
    setResetId,
    onCreate,
    confirmResetPwd,
    confirmDelete,
  }
}
