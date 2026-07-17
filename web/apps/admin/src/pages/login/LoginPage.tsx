import { useNavigate } from 'react-router'
import { useAuthStore } from '../../store/auth'
import { useEffect, useState } from 'react'
import type { FormEvent } from 'react'
import { errorMessage } from '../../lib/api'
import { ErrorAlert } from '../../components/ui'

export default function LoginPage() {
  const navigate = useNavigate()
  const login = useAuthStore((s) => s.login)
  const token = useAuthStore((s) => s.token)
  const [username, setUsername] = useState('admin')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    if (token) navigate('/', { replace: true })
  }, [token, navigate])

  async function onSubmit(e: FormEvent) {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      await login(username, password)
      navigate('/', { replace: true })
    } catch (err) {
      setError(errorMessage(err))
    } finally {
      setLoading(false)
    }
  }

  return (
    <main className="login-page">
      <div className="login-card">
        <h1>Orange TV 管理后台</h1>
        <p className="muted">使用 super_admin 账号登录</p>
        <form onSubmit={onSubmit}>
          <label>
            用户名
            <input value={username} onChange={(e) => setUsername(e.target.value)} required minLength={3} />
          </label>
          <label>
            密码
            <input type="password" value={password} onChange={(e) => setPassword(e.target.value)} required minLength={6} />
          </label>
          <ErrorAlert>{error}</ErrorAlert>
          <button className="primary" disabled={loading}>{loading ? '登录中...' : '登录'}</button>
        </form>
      </div>
    </main>
  )
}
