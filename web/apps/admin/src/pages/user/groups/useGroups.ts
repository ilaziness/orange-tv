import { useCallback, useEffect, useRef, useState } from 'react'
import type * as React from 'react'
import { z } from 'zod'
import { adminApi, errorMessage } from '@/lib/api'
import type { UserGroupItem } from '@orange-tv/shared'
import { toast } from 'sonner'
import { DEFAULT_PAGE_SIZE } from '@/lib/constants'

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
  const [page, setPage] = useState(1)
  const [loading, setLoading] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [deleting, setDeleting] = useState(false)
  const [dialogOpen, setDialogOpen] = useState(false)
  const [editId, setEditId] = useState(0)
  const [form, setForm] = useState(emptyForm)
  const [deleteId, setDeleteId] = useState<number | null>(null)
  const pageRef = useRef(page)

  useEffect(() => {
    pageRef.current = page
  }, [page])

  const load = useCallback(async (p = pageRef.current) => {
    setLoading(true)
    try {
      const res = await adminApi.listGroups({ page: p, page_size: DEFAULT_PAGE_SIZE })
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
  }, [load])

  function openCreate() {
    setEditId(0)
    setForm(emptyForm)
    setDialogOpen(true)
  }

  function openEdit(item: UserGroupItem) {
    setEditId(item.id)
    setForm({
      name: item.name,
      permissions: item.permissions || '',
      description: item.description || '',
    })
    setDialogOpen(true)
  }

  function closeDialog(open: boolean) {
    setDialogOpen(open)
    if (!open) {
      setEditId(0)
      setForm(emptyForm)
    }
  }

  async function onSubmit(e: React.SyntheticEvent<HTMLFormElement>) {
    e.preventDefault()
    if (submitting) return
    const result = groupSchema.safeParse(form)
    if (!result.success) {
      toast.error(result.error.issues[0]?.message || '表单校验失败')
      return
    }
    setSubmitting(true)
    try {
      const isEdit = editId > 0
      if (isEdit) {
        await adminApi.updateGroup(editId, result.data)
        toast.success('用户组已更新')
      } else {
        await adminApi.createGroup(result.data)
        toast.success('用户组已创建')
      }
      closeDialog(false)
      await load(isEdit ? page : 1)
    } catch (err) {
      toast.error(errorMessage(err))
    } finally {
      setSubmitting(false)
    }
  }

  async function confirmDelete() {
    if (deleteId === null) return
    setDeleting(true)
    try {
      await adminApi.deleteGroup(deleteId)
      toast.success('用户组已删除')
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
    closeDialog,
    editId,
    form,
    setForm,
    deleteId,
    setDeleteId,
    hasNext,
    openCreate,
    openEdit,
    onSubmit,
    confirmDelete,
    load,
  }
}
