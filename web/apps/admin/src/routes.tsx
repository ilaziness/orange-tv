import { lazy, Suspense } from 'react'
import type { ReactNode } from 'react'
import { Navigate, Route, Routes } from 'react-router'
import { RequireAuth } from '@/components/auth/RequireAuth'
import { AdminLayout } from '@/components/layout/AdminLayout'
import { Skeleton } from '@/components/ui/skeleton'

const LoginPage = lazy(() => import('@/pages/login/LoginPage'))
const DashboardPage = lazy(() => import('@/pages/dashboard/DashboardPage'))
const CategoriesPage = lazy(() => import('@/pages/content/categories/CategoriesPage'))
const DirectorsPage = lazy(() => import('@/pages/content/directors/DirectorsPage'))
const ActorsPage = lazy(() => import('@/pages/content/actors/ActorsPage'))
const TagsPage = lazy(() => import('@/pages/content/tags/TagsPage'))
const PlaySourcesPage = lazy(() => import('@/pages/content/playSources/PlaySourcesPage'))
const VideosPage = lazy(() => import('@/pages/content/videos/VideosPage'))
const VideoEditPage = lazy(() => import('@/pages/content/videos/edit/VideoEditPage'))
const LiveTVPage = lazy(() => import('@/pages/content/livetv/LiveTVPage'))
const CommentsPage = lazy(() => import('@/pages/content/comments/CommentsPage'))
const CollectPage = lazy(() => import('@/pages/content/collect/CollectPage'))
const BannersPage = lazy(() => import('@/pages/content/banners/BannersPage'))
const AdminsPage = lazy(() => import('@/pages/user/admins/AdminsPage'))
const UserGroupsPage = lazy(() => import('@/pages/user/groups/UserGroupsPage'))
const UsersPage = lazy(() => import('@/pages/user/users/UsersPage'))
const LoginLogsPage = lazy(() => import('@/pages/user/loginLogs/LoginLogsPage'))
const SiteSettingsPage = lazy(() => import('@/pages/system/site/SiteSettingsPage'))
const APISettingsPage = lazy(() => import('@/pages/system/api/APISettingsPage'))
const AdsPage = lazy(() => import('@/pages/system/ad/AdsPage'))
const SystemLogPage = lazy(() => import('@/pages/system/log/SystemLogPage'))
const DataManagementPage = lazy(() => import('@/pages/system/datamgmt/DataManagementPage'))
const SettingsPage = lazy(() => import('@/pages/settings/SettingsPage'))

function Lazy({ children }: { children: ReactNode }) {
  return (
    <Suspense
      fallback={
        <div className="flex h-full items-center justify-center p-8">
          <Skeleton className="h-8 w-48" />
        </div>
      }
    >
      {children}
    </Suspense>
  )
}

export function AppRoutes() {
  return (
    <Routes>
      <Route
        path="/login"
        element={
          <Lazy>
            <LoginPage />
          </Lazy>
        }
      />
      <Route element={<RequireAuth />}>
        <Route element={<AdminLayout />}>
          <Route
            path="/"
            element={
              <Lazy>
                <DashboardPage />
              </Lazy>
            }
          />
          <Route
            path="/content/categories"
            element={
              <Lazy>
                <CategoriesPage />
              </Lazy>
            }
          />
          <Route
            path="/content/videos"
            element={
              <Lazy>
                <VideosPage />
              </Lazy>
            }
          />
          <Route
            path="/content/videos/new"
            element={
              <Lazy>
                <VideoEditPage />
              </Lazy>
            }
          />
          <Route
            path="/content/videos/:id"
            element={
              <Lazy>
                <VideoEditPage />
              </Lazy>
            }
          />
          <Route
            path="/content/directors"
            element={
              <Lazy>
                <DirectorsPage />
              </Lazy>
            }
          />
          <Route
            path="/content/actors"
            element={
              <Lazy>
                <ActorsPage />
              </Lazy>
            }
          />
          <Route
            path="/content/tags"
            element={
              <Lazy>
                <TagsPage />
              </Lazy>
            }
          />
          <Route
            path="/content/play-sources"
            element={
              <Lazy>
                <PlaySourcesPage />
              </Lazy>
            }
          />
          <Route
            path="/content/livetv"
            element={
              <Lazy>
                <LiveTVPage />
              </Lazy>
            }
          />
          <Route
            path="/content/comments"
            element={
              <Lazy>
                <CommentsPage />
              </Lazy>
            }
          />
          <Route
            path="/content/collect"
            element={
              <Lazy>
                <CollectPage />
              </Lazy>
            }
          />
          <Route
            path="/content/banners"
            element={
              <Lazy>
                <BannersPage />
              </Lazy>
            }
          />
          <Route
            path="/user/admins"
            element={
              <Lazy>
                <AdminsPage />
              </Lazy>
            }
          />
          <Route
            path="/user/groups"
            element={
              <Lazy>
                <UserGroupsPage />
              </Lazy>
            }
          />
          <Route
            path="/user/users"
            element={
              <Lazy>
                <UsersPage />
              </Lazy>
            }
          />
          <Route
            path="/user/login-logs"
            element={
              <Lazy>
                <LoginLogsPage />
              </Lazy>
            }
          />
          <Route
            path="/system/site"
            element={
              <Lazy>
                <SiteSettingsPage />
              </Lazy>
            }
          />
          <Route
            path="/system/api"
            element={
              <Lazy>
                <APISettingsPage />
              </Lazy>
            }
          />
          <Route
            path="/system/ad"
            element={
              <Lazy>
                <AdsPage />
              </Lazy>
            }
          />
          <Route
            path="/system/log"
            element={
              <Lazy>
                <SystemLogPage />
              </Lazy>
            }
          />
          <Route
            path="/system/data-management"
            element={
              <Lazy>
                <DataManagementPage />
              </Lazy>
            }
          />
          <Route
            path="/settings"
            element={
              <Lazy>
                <SettingsPage />
              </Lazy>
            }
          />
        </Route>
      </Route>
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}
