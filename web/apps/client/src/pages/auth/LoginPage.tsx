import { useState } from 'react'
import { useNavigate, Link } from 'react-router'
import { clientApi, errorMessage, setToken } from '../../lib/api'
import { ErrorAlert } from '../../components/ui/ErrorAlert'

export default function LoginPage() {
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const navigate = useNavigate()

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    try {
      const res = await clientApi.login(username, password)
      setToken(res.data.access_token)
      navigate('/')
    } catch (err) {
      setError(errorMessage(err))
    }
  }

  return (
    <div className="auth-page">
      <div className="auth-card">
        <h1>登录</h1>
        <ErrorAlert message={error} />
        <form onSubmit={handleSubmit}>
          <input placeholder="用户名" value={username} onChange={(e) => setUsername(e.target.value)} />
          <input type="password" placeholder="密码" value={password} onChange={(e) => setPassword(e.target.value)} />
          <button className="primary" type="submit">登录</button>
        </form>
        <p className="muted">还没有账号？ <Link to="/register">注册</Link></p>
      </div>
    </div>
  )
}
