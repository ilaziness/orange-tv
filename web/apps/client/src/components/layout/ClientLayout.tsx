import { useState } from 'react'
import { Link, Outlet, useNavigate } from 'react-router'
import { useAuth } from '../../hooks/useAuth'
import { useTheme } from '../../hooks/useTheme'

export function ClientLayout() {
  const [keyword, setKeyword] = useState('')
  const navigate = useNavigate()
  const { profile, logout } = useAuth()
  useTheme()

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault()
    if (keyword.trim()) navigate(`/category?keyword=${encodeURIComponent(keyword.trim())}`)
  }

  return (
    <div className="shell">
      <header className="nav">
        <Link to="/" className="brand">ORANGE TV</Link>
        <div className="nav-links">
          <Link to="/">首页</Link>
          <Link to="/category">分类</Link>
          <Link to="/live">直播</Link>
          <form className="search-box" onSubmit={handleSearch}>
            <input placeholder="搜索影视" value={keyword} onChange={(e) => setKeyword(e.target.value)} />
            <button className="primary" type="submit">搜索</button>
          </form>
          {profile ? (
            <>
              <Link to="/favorites">收藏</Link>
              <Link to="/history">历史</Link>
              <span className="nav-user">{profile.username || profile.email}</span>
              <button onClick={() => { logout(); navigate('/') }}>退出</button>
            </>
          ) : (
            <>
              <Link to="/login">登录</Link>
              <Link to="/register">注册</Link>
            </>
          )}
        </div>
      </header>
      <div className="container"><Outlet /></div>
    </div>
  )
}
