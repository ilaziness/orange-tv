import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom'
import { ADMIN_API_BASE } from '@orange-tv/shared'
import './App.css'

function LoginPage() {
  return (
    <main className="page">
      <h1>管理后台登录</h1>
      <p className="muted">Phase 1 占位页（认证在第二阶段实现）</p>
      <p className="muted">API base: {ADMIN_API_BASE}</p>
    </main>
  )
}

function DashboardPage() {
  return (
    <div className="layout">
      <aside className="sidebar">
        <div className="brand">Orange TV Admin</div>
        <nav>
          <a href="/">概况</a>
          <a href="/">内容管理</a>
          <a href="/">系统设置</a>
        </nav>
      </aside>
      <main className="page">
        <h1>概况</h1>
        <p>管理端布局骨架（Phase 1）</p>
      </main>
    </div>
  )
}

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route path="/" element={<DashboardPage />} />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </BrowserRouter>
  )
}
