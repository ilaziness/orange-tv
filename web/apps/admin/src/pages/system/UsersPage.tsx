import { useCallback, useEffect, useRef, useState } from 'react'
import { adminApi, errorMessage } from '../../lib/api'
import { ErrorAlert, PageCard, PageHeader, StatusBadge } from '../../components/ui'
import type { UserItem } from '@orange-tv/shared'

export default function UsersPage() {
  const [list, setList] = useState<UserItem[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [keyword, setKeyword] = useState('')
  const [error, setError] = useState('')
  const [queryKey, setQueryKey] = useState(0)
  const keywordRef = useRef(keyword)
  const pageRef = useRef(page)

  useEffect(() => { keywordRef.current = keyword }, [keyword])
  useEffect(() => { pageRef.current = page }, [page])

  const load = useCallback(async (p = pageRef.current, k = keywordRef.current) => {
    setError('')
    try {
      const res = await adminApi.listUsers({ page: p, page_size: 20, keyword: k || undefined })
      setList(res.data.list || [])
      setTotal(res.data.total)
    } catch (err) {
      setError(errorMessage(err))
    }
  }, [])

  useEffect(() => { void load() }, [page, queryKey, load])

  async function onToggleStatus(u: UserItem) {
    try {
      await adminApi.updateUser(u.id, { status: u.status === 1 ? 0 : 1 })
      await load()
    } catch (err) {
      alert(errorMessage(err))
    }
  }

  async function onResetPwd(id: number) {
    const pwd = window.prompt('输入新密码（≥6位）')
    if (!pwd || pwd.length < 6) return
    try {
      await adminApi.resetUserPassword(id, pwd)
      alert('密码已重置')
    } catch (err) {
      alert(errorMessage(err))
    }
  }

  return (
    <PageCard className="stack">
      <PageHeader title="用户管理" />
      <ErrorAlert>{error}</ErrorAlert>
      <div className="toolbar">
        <input placeholder="用户名/邮箱" value={keyword} onChange={(e) => setKeyword(e.target.value)} />
        <button onClick={() => { setPage(1); setQueryKey((q) => q + 1) }}>查询</button>
      </div>
      <p className="muted">共 {total} 条</p>
      <table className="data-table">
        <thead><tr><th>ID</th><th>用户名</th><th>邮箱</th><th>状态</th><th>最后登录</th><th>注册时间</th><th>操作</th></tr></thead>
        <tbody>
          {list.map((u) => (
            <tr key={u.id}>
              <td>{u.id}</td>
              <td>{u.username}</td>
              <td>{u.email || '-'}</td>
              <td><StatusBadge status={u.status} /></td>
              <td>{u.last_login_at || '-'}</td>
              <td>{u.created_at || '-'}</td>
              <td>
                <button onClick={() => onToggleStatus(u)}>{u.status === 1 ? '禁用' : '启用'}</button>
                <button onClick={() => onResetPwd(u.id)}>重置密码</button>
                <button onClick={async () => { if (confirm('删除该用户？')) { try { await adminApi.deleteUser(u.id); await load() } catch (err) { alert(errorMessage(err)) } } }}>删除</button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </PageCard>
  )
}
