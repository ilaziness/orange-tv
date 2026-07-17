import { useEffect, useState } from 'react'
import type { FormEvent } from 'react'
import { adminApi, errorMessage } from '../../lib/api'
import { ErrorAlert, PageCard, PageHeader, StatusBadge } from '../../components/ui'
import type { LiveChannel } from '@orange-tv/shared'

export default function LivePage() {
  const [list, setList] = useState<LiveChannel[]>([])
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const [form, setForm] = useState({ name: '', category: '', stream_url: '', logo: '', description: '', sort_order: 0, status: 1 })
  const [editId, setEditId] = useState(0)

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

  async function onSubmit(e: FormEvent) {
    e.preventDefault()
    setError('')
    try {
      if (editId) await adminApi.updateLive(editId, form)
      else await adminApi.createLive(form)
      setForm({ name: '', category: '', stream_url: '', logo: '', description: '', sort_order: 0, status: 1 })
      setEditId(0)
      await load()
    } catch (err) {
      setError(errorMessage(err))
    }
  }

  async function onDelete(id: number) {
    if (!confirm('确认删除该直播频道？')) return
    try {
      await adminApi.deleteLive(id)
      await load()
    } catch (err) {
      setError(errorMessage(err))
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
      sort_order: item.sort_order || 0,
      status: item.status ?? 1,
    })
  }

  return (
    <PageCard className="stack">
      <PageHeader title="直播管理" />
      <ErrorAlert>{error}</ErrorAlert>
      <form className="stack" onSubmit={onSubmit}>
        <div className="toolbar">
          <input placeholder="频道名称" value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} required />
          <input placeholder="分类" value={form.category} onChange={(e) => setForm({ ...form, category: e.target.value })} />
          <input placeholder="直播流地址" value={form.stream_url} onChange={(e) => setForm({ ...form, stream_url: e.target.value })} required style={{ minWidth: 280 }} />
        </div>
        <div className="toolbar">
          <input placeholder="Logo URL" value={form.logo} onChange={(e) => setForm({ ...form, logo: e.target.value })} />
          <input placeholder="简介" value={form.description} onChange={(e) => setForm({ ...form, description: e.target.value })} />
          <input type="number" placeholder="排序" value={form.sort_order} onChange={(e) => setForm({ ...form, sort_order: Number(e.target.value) })} />
          <select value={form.status} onChange={(e) => setForm({ ...form, status: Number(e.target.value) })}>
            <option value={1}>启用</option>
            <option value={0}>禁用</option>
          </select>
          <button className="primary" type="submit">{editId ? '保存修改' : '新增频道'}</button>
          {editId ? <button type="button" onClick={() => { setEditId(0); setForm({ name: '', category: '', stream_url: '', logo: '', description: '', sort_order: 0, status: 1 }) }}>取消</button> : null}
          <button type="button" onClick={() => void load()} disabled={loading}>刷新</button>
        </div>
      </form>
      <table className="table">
        <thead>
          <tr>
            <th>ID</th><th>名称</th><th>分类</th><th>状态</th><th>排序</th><th>操作</th>
          </tr>
        </thead>
        <tbody>
          {list.map((item) => (
            <tr key={item.id}>
              <td>{item.id}</td>
              <td>
                <strong>{item.name}</strong>
                <div className="muted" style={{ maxWidth: 360, overflow: 'hidden', textOverflow: 'ellipsis' }}>{item.stream_url}</div>
              </td>
              <td>{item.category || '-'}</td>
              <td><StatusBadge status={item.status} /></td>
              <td>{item.sort_order}</td>
              <td className="actions">
                <button onClick={() => startEdit(item)}>编辑</button>
                <button className="danger" onClick={() => void onDelete(item.id)}>删除</button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
      {!list.length ? <p className="muted">暂无直播频道</p> : null}
    </PageCard>
  )
}

