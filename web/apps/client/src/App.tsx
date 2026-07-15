import { BrowserRouter, Route, Routes } from 'react-router-dom'
import { CLIENT_API_BASE } from '@orange-tv/shared'
import './App.css'

function HomePage() {
  return (
    <main className="page">
      <h1>Orange TV</h1>
      <p>用户端骨架（Phase 1）</p>
      <p className="muted">API base: {CLIENT_API_BASE}</p>
      <ul>
        <li>首页 / 分类 / 详情 / 播放页将在后续阶段实现</li>
        <li>后端接口前缀: /api/client/v1</li>
      </ul>
    </main>
  )
}

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<HomePage />} />
      </Routes>
    </BrowserRouter>
  )
}
