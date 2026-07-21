import { useCallback, useEffect, useRef, useState } from 'react'
import { adminApi, errorMessage } from '@/lib/api'
import type { UserItem } from '@orange-tv/shared'
import { toast } from 'sonner'

export function useUsers() {
  const [list, setList] = useState<UserItem[]>([])
  const [total, setTotal] = useState(0)
  const [keyword, setKeyword] = useState('')
  const [error, setError] = useState('')
  const [queryKey, setQueryKey] = useState(0)
  const [deleteId, setDeleteId] = useState<number | null>(null)
  const [resetId, setResetId] = useState<number | null>(null)
  const keywordRef = useRef(keyword)

  useEffect(() => { keywordRef.current = keyword }, [keyword])

  const load = useCallback(async (k = keywordRef.current) => {
    setError('')
    try {
      const res = await adminApi.listUsers({ page: 1, page_size: 20, keyword: k || undefined })
      setList(res.data.list || [])
      setTotal(res.data.total)
    } catch (err) {
      setError(errorMessage(err))
    }
  }, [])

  useEffect(() => { void load() }, [queryKey, load])

  async function onToggleStatus(u: UserItem) {
    try {
      await adminApi.updateUser(u.id, { status: u.status === 1 ? 0 : 1 })
      toast.success(u.status === 1 ? '已禁用' : '已启用')
      await load()
    } catch (err) {
      toast.error(errorMessage(err))
    }
  }

  async function confirmResetPwd(pwd: string) {
    if (resetId === null) return
    try {
      await adminApi.resetUserPassword(resetId, pwd)
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
      await adminApi.deleteUser(deleteId)
      toast.success('用户已删除')
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
    queryKey,
    setQueryKey,
    deleteId,
    setDeleteId,
    resetId,
    setResetId,
    onToggleStatus,
    confirmResetPwd,
    confirmDelete,
  }
}
