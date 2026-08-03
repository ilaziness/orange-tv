import { useCallback, useEffect, useRef, useState } from 'react'
import type * as React from 'react'
import { z } from 'zod'
import { adminApi, errorMessage } from '@/lib/api'
import type { AdminItem, UserGroupItem } from '@orange-tv/shared'
import { toast } from 'sonner'
import { DEFAULT_PAGE_SIZE } from '@/lib/constants'

const optionalEmail = z
  .string()
  .refine((v) => !v || z.email().safeParse(v).success, '邮箱格式不正确')

const createSchema = z.object({
  username: z.string().min(3, '用户名至少 3 个字符'),
  password: z.string().min(6, '密码至少 6 位'),
  email: optionalEmail,
  group_id: z
    .union([z.string(), z.number()])
    .refine((v) => Number(v) > 0, '请选择用户组')
    .transform((v) => Number(v)),
  status: z.union([z.string(), z.number()]).transform((v) => Number(v)),
})

const editSchema = z.object({
  username: z.string().min(3, '用户名至少 3 个字符'),
  email: optionalEmail,
  group_id: z
    .union([z.string(), z.number()])
    .refine((v) => Number(v) > 0, '请选择用户组')
    .transform((v) => Number(v)),
  status: z.union([z.string(), z.number()]).transform((v) => Number(v)),
})

const emptyCreateForm = { username: '', password: '', email: '', group_id: '', status: '1' }
const emptyEditForm = { username: '', email: '', group_id: '', status: '1' }

export type AdminCreateForm = typeof emptyCreateForm
export type AdminEditForm = typeof emptyEditForm

export function useAdmins() {
  const [list, setList] = useState<AdminItem[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [keyword, setKeyword] = useState('')
  const [loading, setLoading] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [deleting, setDeleting] = useState(false)
  const [resetting, setResetting] = useState(false)
  const [groups, setGroups] = useState<UserGroupItem[]>([])
  const [dialogOpen, setDialogOpen] = useState(false)
  const [editId, setEditId] = useState(0)
  const [createForm, setCreateForm] = useState(emptyCreateForm)
  const [editForm, setEditForm] = useState(emptyEditForm)
  const [deleteId, setDeleteId] = useState<number | null>(null)
  const [resetId, setResetId] = useState<number | null>(null)
  const keywordRef = useRef(keyword)
  const pageRef = useRef(page)

  useEffect(() => {
    keywordRef.current = keyword
  }, [keyword])
  useEffect(() => {
    pageRef.current = page
  }, [page])

  const loadGroups = useCallback(async () => {
    try {
      const gRes = await adminApi.listGroups({ page_size: 100 })
      setGroups(gRes.data.list || [])
    } catch {
      // groups are best-effort
    }
  }, [])

  const load = useCallback(async (p = pageRef.current, k = keywordRef.current) => {
    setLoading(true)
    try {
      const res = await adminApi.listAdmins({
        page: p,
        page_size: DEFAULT_PAGE_SIZE,
        keyword: k || undefined,
      })
      setList(res.data.list || [])
      setTotal(res.data.total || 0)
      setPage(res.data.page || p)
    } catch (err) {
      toast.error(errorMessage(err))
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void load(1)
    void loadGroups()
  }, [load, loadGroups])

  function openCreate() {
    setEditId(0)
    setCreateForm(emptyCreateForm)
    setDialogOpen(true)
  }

  function openEdit(item: AdminItem) {
    setEditId(item.id)
    setEditForm({
      username: item.username || '',
      email: item.email || '',
      group_id: String(item.group_id || ''),
      status: String(item.status),
    })
    setDialogOpen(true)
  }

  function closeDialog(open: boolean) {
    setDialogOpen(open)
    if (!open) {
      setEditId(0)
      setCreateForm(emptyCreateForm)
      setEditForm(emptyEditForm)
    }
  }

  async function onSubmit(e: React.SyntheticEvent<HTMLFormElement>) {
    e.preventDefault()
    if (submitting) return
    if (editId > 0) {
      const result = editSchema.safeParse(editForm)
      if (!result.success) {
        toast.error(result.error.issues[0]?.message || '表单校验失败')
        return
      }
      setSubmitting(true)
      try {
        await adminApi.updateAdmin(editId, result.data)
        toast.success('管理员已更新')
        closeDialog(false)
        await load(page)
      } catch (err) {
        toast.error(errorMessage(err))
      } finally {
        setSubmitting(false)
      }
    } else {
      const result = createSchema.safeParse(createForm)
      if (!result.success) {
        toast.error(result.error.issues[0]?.message || '表单校验失败')
        return
      }
      setSubmitting(true)
      try {
        await adminApi.createAdmin(result.data)
        toast.success('管理员已创建')
        closeDialog(false)
        await load(1)
      } catch (err) {
        toast.error(errorMessage(err))
      } finally {
        setSubmitting(false)
      }
    }
  }

  async function confirmResetPwd(pwd: string) {
    if (resetId === null || resetting) return
    setResetting(true)
    try {
      await adminApi.resetAdminPassword(resetId, pwd)
      toast.success('密码已重置')
      setResetId(null)
    } catch (err) {
      toast.error(errorMessage(err))
    } finally {
      setResetting(false)
    }
  }

  async function confirmDelete() {
    if (deleteId === null) return
    setDeleting(true)
    try {
      await adminApi.deleteAdmin(deleteId)
      toast.success('管理员已删除')
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
    keyword,
    setKeyword,
    resetting,
    loading,
    submitting,
    deleting,
    groups,
    dialogOpen,
    closeDialog,
    editId,
    createForm,
    setCreateForm,
    editForm,
    setEditForm,
    deleteId,
    setDeleteId,
    resetId,
    setResetId,
    hasNext,
    openCreate,
    openEdit,
    onSubmit,
    confirmResetPwd,
    confirmDelete,
    load,
  }
}
