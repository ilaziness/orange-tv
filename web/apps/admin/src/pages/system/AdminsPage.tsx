import { useCallback, useEffect, useRef, useState } from 'react'
import type { FormEvent } from 'react'
import { adminApi, errorMessage } from '../../lib/api'
import { ErrorAlert, PageCard, PageHeader, StatusBadge } from '../../components/ui'
import type { AdminItem, UserGroupItem } from '@orange-tv/shared'

export default function AdminsPage() {
  const [list, setList] = useState<AdminItem[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [keyword, setKeyword] = useState('')
  const [error, setError] = useState('')
  const [showCreate, setShowCreate] = useState(false)
  const [form, setForm] = useState({ username: '', password: '', email: '', group_id: 1, status: 1 })
  const [groups, setGroups] = useState<UserGroupItem[]>([])
  const [queryKey, setQueryKey] = useState(0)
  const keywordRef = useRef(keyword)
  const pageRef = useRef(page)

  useEffect(() => { keywordRef.current = keyword }, [keyword])
  useEffect(() => { pageRef.current = page }, [page])

  const load = useCallback(async (p = pageRef.current, k = keywordRef.current) => {
    setError('')
    try {
      const [res, gRes] = await Promise.all([
        adminApi.listAdmins({ page: p, page_size: 20, keyword: k || undefined }),
        adminApi.listGroups({ page_size: 100 }),
      ])
      setList(res.data.list || [])
      setTotal(res.data.total)
      setGroups(gRes.data.list || [])
    } catch (err) {
      setError(errorMessage(err))
    }
  }, [])

  useEffect(() => { void load() }, [page, queryKey, load])

  async function onCreate(e: FormEvent) {
    e.preventDefault()
    setError('')
    try {
      await adminApi.createAdmin(form)
      setShowCreate(false)
      setForm({ username: '', password: '', email: '', group_id: 1, status: 1 })
      await load()
    } catch (err) {
      setError(errorMessage(err))
    }
  }

  async function onResetPwd(id: number) {
    const pwd = window.prompt('输入新密码（≥6位）')
    if (!pwd || pwd.length < 6) return
    try {
      await adminApi.resetAdminPassword(id, pwd)
      alert('密码已重置')
    } catch (err) {
      alert(errorMessage(err))
    }
  }

  return (
    <PageCard className="stack">
      <PageHeader title="管理员管理"><button className="primary" onClick={() => setShowCreate(!showCreate)}>新增管理员</button></PageHeader>
      <ErrorAlert>{error}</ErrorAlert>
      {showCreate ? (
        <form className="inline-form" onSubmit={onCreate}>
          <input placeholder="用户名" value={form.username} onChange={(e) => setForm({ ...form, username: e.target.value })} required minLength={3} />
          <input type="password" placeholder="密码" value={form.password} onChange={(e) => setForm({ ...form, password: e.target.value })} required minLength={6} />
          <input placeholder="邮箱" value={form.email} onChange={(e) => setForm({ ...form, email: e.target.value })} />
          <select value={form.group_id} onChange={(e) => setForm({ ...form, group_id: Number(e.target.value) })}>
            {groups.map((g) => <option key={g.id} value={g.id}>{g.name}</option>)}
          </select>
          <select value={form.status} onChange={(e) => setForm({ ...form, status: Number(e.target.value) })}>
            <option value={1}>启用</option>
            <option value={0}>禁用</option>
          </select>
          <button type="submit" className="primary">保存</button>
        </form>
      ) : null}
      <div className="toolbar">
        <input placeholder="用户名/邮箱" value={keyword} onChange={(e) => setKeyword(e.target.value)} />
        <button onClick={() => { setPage(1); setQueryKey((q) => q + 1) }}>查询</button>
      </div>
      <p className="muted">共 {total} 条</p>
      <table className="data-table">
        <thead><tr><th>ID</th><th>用户名</th><th>邮箱</th><th>用户组</th><th>状态</th><th>最后登录</th><th>操作</th></tr></thead>
        <tbody>
          {list.map((a) => (
            <tr key={a.id}>
              <td>{a.id}</td>
              <td>{a.username}</td>
              <td>{a.email || '-'}</td>
              <td>{a.group_name}</td>
              <td><StatusBadge status={a.status} /></td>
              <td>{a.last_login_at || '-'}</td>
              <td>
                <button onClick={() => onResetPwd(a.id)}>重置密码</button>
                <button onClick={async () => { if (confirm('删除该管理员？')) { try { await adminApi.deleteAdmin(a.id); await load() } catch (err) { alert(errorMessage(err)) } } }}>删除</button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </PageCard>
  )
}
