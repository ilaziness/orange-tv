import { useCallback, useEffect, useState } from 'react'
import type * as React from 'react'
import { z } from 'zod'
import { adminApi, errorMessage } from '@/lib/api'
import type { PlaySource } from '@orange-tv/shared'
import { toast } from 'sonner'

export type PlaySourceForm = {
  name: string
  sort_order: string
  status: string
}

export type PlaySourceFieldErrors = Partial<Record<keyof PlaySourceForm, string>>

const playSourceSchema = z.object({
  name: z.string().trim().min(1, '播放源名称不能为空').max(100, '播放源名称不能超过 100 个字符'),
  sort_order: z
    .string()
    .transform((v) => (v.trim() === '' ? 0 : Number(v)))
    .refine(
      (v) => Number.isSafeInteger(v) && v >= 0 && v <= 4294967295,
      '排序必须为 0 至 4294967295 的整数',
    ),
  status: z
    .string()
    .min(1, '请选择状态')
    .transform((v) => Number(v))
    .refine((v) => v === 0 || v === 1, '请选择有效状态'),
})

export const emptyForm: PlaySourceForm = { name: '', sort_order: '0', status: '1' }

export const statusOptions = [
  { value: '1', label: '启用' },
  { value: '0', label: '禁用' },
]

function fieldErrorsFromZod(error: z.ZodError): PlaySourceFieldErrors {
  const out: PlaySourceFieldErrors = {}
  for (const issue of error.issues) {
    const key = issue.path[0]
    if (key === 'name' || key === 'sort_order' || key === 'status') {
      if (!out[key]) out[key] = issue.message
    }
  }
  return out
}

export function usePlaySources() {
  const [items, setItems] = useState<PlaySource[]>([])
  const [error, setError] = useState('')
  const [dialogError, setDialogError] = useState('')
  const [fieldErrors, setFieldErrors] = useState<PlaySourceFieldErrors>({})
  const [loading, setLoading] = useState(false)
  const [form, setForm] = useState<PlaySourceForm>(emptyForm)
  const [editId, setEditId] = useState(0)
  const [deleteId, setDeleteId] = useState<number | null>(null)
  const [dialogOpen, setDialogOpen] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [deleting, setDeleting] = useState(false)
  const [togglingId, setTogglingId] = useState<number | null>(null)

  const load = useCallback(async () => {
    setLoading(true)
    setError('')
    try {
      const res = await adminApi.listPlaySources()
      setItems(res.data.list || [])
    } catch (err) {
      setError(errorMessage(err))
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void load()
  }, [load])

  function openCreate() {
    setForm(emptyForm)
    setEditId(0)
    setDialogError('')
    setFieldErrors({})
    setDialogOpen(true)
  }

  function openEdit(item: PlaySource) {
    setEditId(item.id)
    setForm({
      name: item.name,
      sort_order: String(item.sort_order ?? 0),
      status: String(item.status ?? 1),
    })
    setDialogError('')
    setFieldErrors({})
    setDialogOpen(true)
  }

  function closeDialog(open: boolean) {
    if (submitting) return
    setDialogOpen(open)
    if (!open) {
      setForm(emptyForm)
      setEditId(0)
      setDialogError('')
      setFieldErrors({})
    }
  }

  function updateForm(patch: Partial<PlaySourceForm>) {
    setForm((prev) => ({ ...prev, ...patch }))
    setFieldErrors((prev) => {
      const next = { ...prev }
      for (const key of Object.keys(patch) as (keyof PlaySourceForm)[]) {
        delete next[key]
      }
      return next
    })
  }

  async function onSubmit(e: React.SyntheticEvent<HTMLFormElement>) {
    e.preventDefault()
    setDialogError('')
    setFieldErrors({})
    const result = playSourceSchema.safeParse(form)
    if (!result.success) {
      setFieldErrors(fieldErrorsFromZod(result.error))
      return
    }
    setSubmitting(true)
    try {
      const payload = {
        name: result.data.name,
        sort_order: result.data.sort_order,
        status: result.data.status,
      }
      if (editId) {
        await adminApi.updatePlaySource(editId, payload)
        toast.success('播放源已更新')
      } else {
        await adminApi.createPlaySource(payload)
        toast.success('播放源已创建')
      }
      setForm(emptyForm)
      setEditId(0)
      setDialogOpen(false)
      await load()
    } catch (err) {
      setDialogError(errorMessage(err))
    } finally {
      setSubmitting(false)
    }
  }

  async function toggleStatus(item: PlaySource) {
    setTogglingId(item.id)
    try {
      await adminApi.updatePlaySource(item.id, { status: item.status === 1 ? 0 : 1 })
      toast.success(item.status === 1 ? '已禁用' : '已启用')
      await load()
    } catch (err) {
      toast.error(errorMessage(err))
    } finally {
      setTogglingId(null)
    }
  }

  async function confirmDelete() {
    if (deleteId === null) return
    setDeleting(true)
    try {
      await adminApi.deletePlaySource(deleteId)
      toast.success('播放源已删除')
      await load()
    } catch (err) {
      toast.error(errorMessage(err))
    } finally {
      setDeleting(false)
      setDeleteId(null)
    }
  }

  return {
    items,
    error,
    loading,
    form,
    updateForm,
    editId,
    deleteId,
    setDeleteId,
    dialogOpen,
    closeDialog,
    dialogError,
    fieldErrors,
    submitting,
    deleting,
    togglingId,
    onSubmit,
    confirmDelete,
    toggleStatus,
    openCreate,
    openEdit,
    load,
  }
}
