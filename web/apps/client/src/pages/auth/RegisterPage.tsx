import { useState } from 'react'
import { useNavigate, Link } from 'react-router'
import { clientApi, errorMessage } from '../../lib/api'
import { ErrorAlert } from '../../components/ui/ErrorAlert'

export default function RegisterPage() {
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [error, setError] = useState('')
  const navigate = useNavigate()

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    if (password !== confirmPassword) {
      setError('两次密码不一致')
      return
    }
    try {
      await clientApi.register(username, password)
      navigate('/login')
    } catch (err) {
      setError(errorMessage(err))
    }
  }

  return (
    <div className="auth-page">
      <div className="auth-card">
        <h1>注册</h1>
        <ErrorAlert message={error} />
        <form onSubmit={handleSubmit}>
          <input placeholder="用户名" value={username} onChange={(e) => setUsername(e.target.value)} />
          <input type="password" placeholder="密码" value={password} onChange={(e) => setPassword(e.target.value)} />
          <input type="password" placeholder="确认密码" value={confirmPassword} onChange={(e) => setConfirmPassword(e.target.value)} />
          <button className="primary" type="submit">注册</button>
        </form>
        <p className="muted">已有账号？ <Link to="/login">登录</Link></p>
      </div>
    </div>
  )
}
