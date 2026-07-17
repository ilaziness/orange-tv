import { useCallback, useEffect, useRef, useState } from 'react'
import { adminApi, errorMessage } from '../../lib/api'
import { ErrorAlert, PageCard, PageHeader } from '../../components/ui'
import type { LoginLogItem, SystemLogItem } from '@orange-tv/shared'

export default function SystemLogPage() {
  const [tab, setTab] = useState<'system' | 'login'>('system')
  const [systemLogs, setSystemLogs] = useState<SystemLogItem[]>([])
  const [loginLogs, setLoginLogs] = useState<LoginLogItem[]>([])
  const [module, setModule] = useState('')
  const [username, setUsername] = useState('')
  const [error, setError] = useState('')
  const [total, setTotal] = useState(0)
  const tabRef = useRef(tab)
  const moduleRef = useRef(module)
  const usernameRef = useRef(username)

  useEffect(() => { tabRef.current = tab }, [tab])
  useEffect(() => { moduleRef.current = module }, [module])
  useEffect(() => { usernameRef.current = username }, [username])

  const load = useCallback(async () => {
    setError('')
    try {
      if (tabRef.current === 'system') {
        const res = await adminApi.listSystemLogs({ page: 1, page_size: 50, module: moduleRef.current || undefined })
        setSystemLogs(res.data.list || [])
        setTotal(res.data.total)
      } else {
        const res = await adminApi.listLoginLogs({ page: 1, page_size: 50, username: usernameRef.current || undefined })
        setLoginLogs(res.data.list || [])
        setTotal(res.data.total)
      }
    } catch (err) {
      setError(errorMessage(err))
    }
  }, [])

  useEffect(() => { void load() }, [tab, load])

  return (
    <PageCard className="stack">
      <PageHeader title="系统日志" />
      <div className="toolbar">
        <button className={tab === 'system' ? 'primary' : ''} onClick={() => setTab('system')}>操作日志</button>
        <button className={tab === 'login' ? 'primary' : ''} onClick={() => setTab('login')}>登录日志</button>
      </div>
      <ErrorAlert>{error}</ErrorAlert>
      {tab === 'system' ? (
        <>
          <div className="toolbar">
            <input placeholder="模块筛选" value={module} onChange={(e) => setModule(e.target.value)} />
            <button onClick={() => void load()}>查询</button>
          </div>
          <p className="muted">共 {total} 条</p>
          <div className="tree">
            {systemLogs.map((l) => (
              <div key={l.id} className="tree-item">
                <div>
                  <strong>[{l.level}] {l.module}/{l.action}</strong>
                  <div className="muted">admin={l.admin_id} · {l.ip_address} · {l.created_at}</div>
                  <div>{l.content}</div>
                </div>
              </div>
            ))}
          </div>
        </>
      ) : (
        <>
          <div className="toolbar">
            <input placeholder="用户名" value={username} onChange={(e) => setUsername(e.target.value)} />
            <button onClick={() => void load()}>查询</button>
          </div>
          <p className="muted">共 {total} 条</p>
          <div className="tree">
            {loginLogs.map((l) => (
              <div key={l.id} className="tree-item">
                <div>
                  <strong>{l.username}</strong>
                  <div className="muted">{l.status === 1 ? '成功' : '失败'} · {l.ip_address} · {l.created_at}</div>
                  <div className="muted">{l.user_agent}</div>
                </div>
              </div>
            ))}
          </div>
        </>
      )}
    </PageCard>
  )
}
