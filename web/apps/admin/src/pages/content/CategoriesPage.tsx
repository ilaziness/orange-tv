import { useEffect, useState } from 'react'
import type { FormEvent } from 'react'
import { adminApi, errorMessage } from '../../lib/api'
import { flattenCategories } from '../../utils/categories'
import { ErrorAlert, PageCard, PageHeader, StatusBadge } from '../../components/ui'
import type { Category } from '@orange-tv/shared'

export default function CategoriesPage() {
  const [tree, setTree] = useState<Category[]>([])
  const [name, setName] = useState('')
  const [parentId, setParentId] = useState(0)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

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

  async function onCreate(e: FormEvent) {
    e.preventDefault()
    setError('')
    try {
      await adminApi.createCategory({ name, parent_id: parentId, status: 1 })
      setName('')
      await load()
    } catch (err) {
      setError(errorMessage(err))
    }
  }

  async function onDelete(id: number) {
    if (!confirm('确认删除该分类？')) return
    try {
      await adminApi.deleteCategory(id)
      await load()
    } catch (err) {
      setError(errorMessage(err))
    }
  }

  async function toggleStatus(item: Category) {
    try {
      await adminApi.updateCategory(item.id, { status: item.status === 1 ? 0 : 1 })
      await load()
    } catch (err) {
      setError(errorMessage(err))
    }
  }

  return (
    <PageCard className="stack">
      <PageHeader title="分类管理" />
      <ErrorAlert>{error}</ErrorAlert>
      <form className="toolbar" onSubmit={onCreate}>
        <input placeholder="分类名称" value={name} onChange={(e) => setName(e.target.value)} required />
        <select value={parentId} onChange={(e) => setParentId(Number(e.target.value))}>
          <option value={0}>无父级</option>
          {flat.map((c) => (
            <option key={c.id} value={c.id}>{'—'.repeat(c.depth)} {c.name}</option>
          ))}
        </select>
        <button className="primary" type="submit">新增分类</button>
        <button type="button" onClick={() => void load()} disabled={loading}>刷新</button>
      </form>
      <div className="tree">
        {flat.map((item) => (
          <div key={item.id} className="tree-item" style={{ marginLeft: item.depth * 16 }}>
            <div>
              <strong>{item.name}</strong>
              <div className="muted">ID {item.id} · 排序 {item.sort_order}</div>
            </div>
            <div className="actions">
              <StatusBadge status={item.status} />
              <button onClick={() => void toggleStatus(item)}>{item.status === 1 ? '禁用' : '启用'}</button>
              <button className="danger" onClick={() => void onDelete(item.id)}>删除</button>
            </div>
          </div>
        ))}
        {!flat.length ? <p className="muted">暂无分类</p> : null}
      </div>
    </PageCard>
  )
}
