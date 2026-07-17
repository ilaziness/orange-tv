import { useEffect, useState } from 'react'
import type { FormEvent } from 'react'
import { adminApi, errorMessage } from '../../lib/api'
import { ErrorAlert, PageCard, PageHeader } from '../../components/ui'
import type { UserGroupItem } from '@orange-tv/shared'

export default function UserGroupsPage() {
  const [list, setList] = useState<UserGroupItem[]>([])
  const [total, setTotal] = useState(0)
  const [error, setError] = useState('')
  const [showCreate, setShowCreate] = useState(false)
  const [form, setForm] = useState({ name: '', permissions: '', description: '' })

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

  async function onCreate(e: FormEvent) {
    e.preventDefault()
    setError('')
    try {
      await adminApi.createGroup(form)
      setShowCreate(false)
      setForm({ name: '', permissions: '', description: '' })
      await load()
    } catch (err) {
      setError(errorMessage(err))
    }
  }

  return (
    <PageCard className="stack">
      <PageHeader title="用户组管理"><button className="primary" onClick={() => setShowCreate(!showCreate)}>新增用户组</button></PageHeader>
      <ErrorAlert>{error}</ErrorAlert>
      {showCreate ? (
        <form className="inline-form" onSubmit={onCreate}>
          <input placeholder="名称" value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} required />
          <input placeholder="权限（JSON）" value={form.permissions} onChange={(e) => setForm({ ...form, permissions: e.target.value })} />
          <input placeholder="描述" value={form.description} onChange={(e) => setForm({ ...form, description: e.target.value })} />
          <button type="submit" className="primary">保存</button>
        </form>
      ) : null}
      <p className="muted">共 {total} 条</p>
      <table className="data-table">
        <thead><tr><th>ID</th><th>名称</th><th>描述</th><th>操作</th></tr></thead>
        <tbody>
          {list.map((g) => (
            <tr key={g.id}>
              <td>{g.id}</td>
              <td>{g.name}</td>
              <td>{g.description || '-'}</td>
              <td>
                {g.name !== 'super_admin' ? (
                  <button onClick={async () => { if (confirm('删除该用户组？')) { try { await adminApi.deleteGroup(g.id); await load() } catch (err) { alert(errorMessage(err)) } } }}>删除</button>
                ) : <span className="muted">系统组</span>}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </PageCard>
  )
}
