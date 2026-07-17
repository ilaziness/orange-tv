import { Link, useLocation, useNavigate } from 'react-router'
import { useAuthStore } from '../../store/auth'

export function Topbar() {
  const location = useLocation()
  const navigate = useNavigate()
  const profile = useAuthStore((s) => s.profile)
  const logout = useAuthStore((s) => s.logout)

  return (
    <header className="topbar">
      <div className="brand">Orange TV Admin</div>
      <nav className="topnav">
        <Link className={location.pathname === '/' ? 'active' : ''} to="/">首页</Link>
        <Link className={location.pathname.startsWith('/content') ? 'active' : ''} to="/content/categories">内容管理</Link>
        <Link className={location.pathname.startsWith('/system') ? 'active' : ''} to="/system/site">系统设置</Link>
      </nav>
      <div className="userbox">
        <span>{profile?.username} ({profile?.role})</span>
        <button onClick={async () => { await logout(); navigate('/login') }}>退出</button>
      </div>
    </header>
  )
}
