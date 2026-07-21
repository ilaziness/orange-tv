import { useEffect, useMemo, useState } from 'react'
import type * as React from 'react'
import { z } from 'zod'
import { adminApi, errorMessage } from '@/lib/api'
import { flattenCategories } from '@/lib/categories'
import type { Category } from '@orange-tv/shared'
import { toast } from 'sonner'

export type CategoryForm = {
  name: string
  parentId: string
  sortOrder: string
  status: string
}

const categorySchema = z.object({
  name: z.string().trim().min(1, '分类名称不能为空').max(100, '分类名称不能超过 100 个字符'),
  parentId: z.string().transform((v) => Number(v)).refine((v) => Number.isSafeInteger(v) && v >= 0, '请选择有效的父级分类'),
  sortOrder: z.string().transform((v) => Number(v)).refine((v) => Number.isSafeInteger(v) && v >= 0 && v <= 4294967295, '排序必须为 0 至 4294967295 的整数'),
  status: z.string().transform((v) => Number(v)).refine((v) => v === 0 || v === 1, '请选择有效状态'),
})

const emptyForm: CategoryForm = { name: '', parentId: '0', sortOrder: '0', status: '1' }

export const statusOptions = [
  { value: '1', label: '启用' },
  { value: '0', label: '禁用' },
]

function categoryAndDescendantIDs(categories: Category[], id: number): Set<number> {
  const excluded = new Set([id])

  function collect(items: Category[]) {
    for (const item of items) {
      excluded.add(item.id)
      if (item.children) collect(item.children)
    }
  }

  function find(items: Category[]): boolean {
    for (const item of items) {
      if (item.id === id) {
        if (item.children) collect(item.children)
        return true
      }
      if (item.children && find(item.children)) return true
    }
    return false
  }

  find(categories)
  return excluded
}

export function useCategories() {
  const [tree, setTree] = useState<Category[]>([])
  const [form, setForm] = useState<CategoryForm>(emptyForm)
  const [editing, setEditing] = useState<Category | null>(null)
  const [dialogOpen, setDialogOpen] = useState(false)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [updatingId, setUpdatingId] = useState<number | null>(null)
  const [deleting, setDeleting] = useState(false)
  const [deleteId, setDeleteId] = useState<number | null>(null)

  async function load() {
    setLoading(true)
    setError('')
    try {
      const res = await adminApi.listCategories()
      setTree(res.data || [])
    } catch (err) {
      setError(errorMessage(err))
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { void load() }, [])

  const flat = flattenCategories(tree)
  const excludedParentIDs = useMemo(
    () => (editing ? categoryAndDescendantIDs(tree, editing.id) : new Set<number>()),
    [editing, tree],
  )
  const parentOptions = useMemo(() => [
    { value: '0', label: '无父级' },
    ...flat
      .filter((item) => !excludedParentIDs.has(item.id))
      .map((item) => ({ value: String(item.id), label: `${'—'.repeat(item.depth)} ${item.name}` })),
  ], [excludedParentIDs, flat])

  function openCreate() {
    setError('')
    setEditing(null)
    setForm(emptyForm)
    setDialogOpen(true)
  }

  function openEdit(item: Category) {
    setError('')
    setEditing(item)
    setForm({
      name: item.name,
      parentId: String(item.parent_id),
      sortOrder: String(item.sort_order),
      status: String(item.status),
    })
    setDialogOpen(true)
  }

  function closeDialog(open: boolean) {
    if (submitting) return
    setDialogOpen(open)
    if (!open) setEditing(null)
  }

  async function onSubmit(e: React.SyntheticEvent<HTMLFormElement>) {
    e.preventDefault()
    setSubmitting(true)
    setError('')
    const result = categorySchema.safeParse(form)
    if (!result.success) {
      toast.error(result.error.issues[0]?.message || '表单校验失败')
      setSubmitting(false)
      return
    }
    try {
      const payload = {
        name: result.data.name,
        parent_id: result.data.parentId,
        sort_order: result.data.sortOrder,
        status: result.data.status,
      }
      if (editing) {
        await adminApi.updateCategory(editing.id, payload)
        toast.success('分类已更新')
      } else {
        await adminApi.createCategory(payload)
        toast.success('分类已创建')
      }
      setDialogOpen(false)
      setEditing(null)
      setForm(emptyForm)
      await load()
    } catch (err) {
      setError(errorMessage(err))
    } finally {
      setSubmitting(false)
    }
  }

  async function confirmDelete() {
    if (deleteId === null) return
    setDeleting(true)
    try {
      await adminApi.deleteCategory(deleteId)
      toast.success('分类已删除')
      await load()
    } catch (err) {
      toast.error(errorMessage(err))
    } finally {
      setDeleting(false)
      setDeleteId(null)
    }
  }

  async function toggleStatus(item: Category) {
    setUpdatingId(item.id)
    try {
      await adminApi.updateCategory(item.id, { status: item.status === 1 ? 0 : 1 })
      toast.success(item.status === 1 ? '已禁用' : '已启用')
      await load()
    } catch (err) {
      toast.error(errorMessage(err))
    } finally {
      setUpdatingId(null)
    }
  }

  return {
    tree,
    flat,
    form,
    setForm,
    editing,
    dialogOpen,
    error,
    loading,
    submitting,
    updatingId,
    deleting,
    deleteId,
    setDeleteId,
    parentOptions,
    load,
    openCreate,
    openEdit,
    closeDialog,
    onSubmit,
    confirmDelete,
    toggleStatus,
  }
}
