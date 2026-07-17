import { lazy, Suspense } from 'react'
import type { ReactNode } from 'react'
import { Navigate, Route, Routes } from 'react-router'
import { RequireAuth } from '@/components/auth/RequireAuth'
import { AdminLayout } from '@/components/layout/AdminLayout'
import { Skeleton } from '@/components/ui/skeleton'

const LoginPage = lazy(() => import('@/pages/login/LoginPage'))
const DashboardPage = lazy(() => import('@/pages/dashboard/DashboardPage'))
const CategoriesPage = lazy(() => import('@/pages/content/CategoriesPage'))
const DirectorsPage = lazy(() => import('@/pages/content/DirectorsPage'))
const ActorsPage = lazy(() => import('@/pages/content/ActorsPage'))
const TagsPage = lazy(() => import('@/pages/content/TagsPage'))
const PlaySourcesPage = lazy(() => import('@/pages/content/PlaySourcesPage'))
const VideosPage = lazy(() => import('@/pages/content/VideosPage'))
const VideoEditPage = lazy(() => import('@/pages/content/VideoEditPage'))
const LivePage = lazy(() => import('@/pages/content/LivePage'))
const CollectPage = lazy(() => import('@/pages/content/CollectPage'))
const ThemesPage = lazy(() => import('@/pages/system/ThemesPage'))
const SiteSettingsPage = lazy(() => import('@/pages/system/SiteSettingsPage'))
const APISettingsPage = lazy(() => import('@/pages/system/APISettingsPage'))
const SystemLogPage = lazy(() => import('@/pages/system/SystemLogPage'))
const AdminsPage = lazy(() => import('@/pages/system/AdminsPage'))
const UserGroupsPage = lazy(() => import('@/pages/system/UserGroupsPage'))
const UsersPage = lazy(() => import('@/pages/system/UsersPage'))
const BannersPage = lazy(() => import('@/pages/system/BannersPage'))

function Lazy({ children }: { children: ReactNode }) {
  return (
    <Suspense fallback={
      <div className="flex h-full items-center justify-center p-8">
        <Skeleton className="h-8 w-48" />
      </div>
    }>
      {children}
    </Suspense>
  )
}

export function AppRoutes() {
  return (
    <Routes>
      <Route path="/login" element={<Lazy><LoginPage /></Lazy>} />
      <Route element={<RequireAuth />}>
        <Route element={<AdminLayout />}>
          <Route path="/" element={<Lazy><DashboardPage /></Lazy>} />
          <Route path="/content/categories" element={<Lazy><CategoriesPage /></Lazy>} />
          <Route path="/content/videos" element={<Lazy><VideosPage /></Lazy>} />
          <Route path="/content/videos/new" element={<Lazy><VideoEditPage /></Lazy>} />
          <Route path="/content/videos/:id" element={<Lazy><VideoEditPage /></Lazy>} />
          <Route path="/content/directors" element={<Lazy><DirectorsPage /></Lazy>} />
          <Route path="/content/actors" element={<Lazy><ActorsPage /></Lazy>} />
          <Route path="/content/tags" element={<Lazy><TagsPage /></Lazy>} />
          <Route path="/content/play-sources" element={<Lazy><PlaySourcesPage /></Lazy>} />
          <Route path="/content/live" element={<Lazy><LivePage /></Lazy>} />
          <Route path="/content/collect" element={<Lazy><CollectPage /></Lazy>} />
          <Route path="/system/site" element={<Lazy><SiteSettingsPage /></Lazy>} />
          <Route path="/system/api" element={<Lazy><APISettingsPage /></Lazy>} />
          <Route path="/system/theme" element={<Lazy><ThemesPage /></Lazy>} />
          <Route path="/system/log" element={<Lazy><SystemLogPage /></Lazy>} />
          <Route path="/system/admins" element={<Lazy><AdminsPage /></Lazy>} />
          <Route path="/system/groups" element={<Lazy><UserGroupsPage /></Lazy>} />
          <Route path="/system/users" element={<Lazy><UsersPage /></Lazy>} />
          <Route path="/system/banners" element={<Lazy><BannersPage /></Lazy>} />
        </Route>
      </Route>
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}
