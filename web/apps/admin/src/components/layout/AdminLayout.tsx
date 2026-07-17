import { Outlet } from 'react-router'
import { Sidebar } from './Sidebar.tsx'
import { Topbar } from './Topbar.tsx'

export function AdminLayout() {
  return (
    <div className="layout">
      <Topbar />
      <div className="body">
        <Sidebar />
        <main className="content"><Outlet /></main>
      </div>
    </div>
  )
}
