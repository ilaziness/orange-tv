import { lazy, Suspense } from 'react'
import type { ReactNode } from 'react'
import { Navigate, Route, Routes } from 'react-router'
import { ClientLayout } from '@/components/layout/ClientLayout'
import { RequireAuth } from '@/components/auth/RequireAuth'
import { Spinner } from '@/components/ui/spinner'

const HomePage = lazy(() => import('@/pages/home/HomePage'))
const CategoryPage = lazy(() => import('@/pages/category/CategoryPage'))
const VideoDetailPage = lazy(() => import('@/pages/video/VideoDetailPage'))
const PlayPage = lazy(() => import('@/pages/video/PlayPage'))
const LivePage = lazy(() => import('@/pages/live/LivePage'))
const LivePlayPage = lazy(() => import('@/pages/live/LivePlayPage'))
const LoginPage = lazy(() => import('@/pages/auth/LoginPage'))
const RegisterPage = lazy(() => import('@/pages/auth/RegisterPage'))
const FavoritesPage = lazy(() => import('@/pages/user/FavoritesPage'))
const HistoryPage = lazy(() => import('@/pages/user/HistoryPage'))

function Lazy({ children }: { children: ReactNode }) {
  return (
    <Suspense
      fallback={
        <div className="flex items-center justify-center py-20">
          <Spinner className="size-8" />
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
      <Route element={<ClientLayout />}>
        <Route path="/" element={<Lazy><HomePage /></Lazy>} />
        <Route path="/category" element={<Lazy><CategoryPage /></Lazy>} />
        <Route path="/video/:id" element={<Lazy><VideoDetailPage /></Lazy>} />
        <Route path="/play/:id" element={<Lazy><PlayPage /></Lazy>} />
        <Route path="/play/:id/:sourceIdx" element={<Lazy><PlayPage /></Lazy>} />
        <Route path="/live" element={<Lazy><LivePage /></Lazy>} />
        <Route path="/play/live/:id" element={<Lazy><LivePlayPage /></Lazy>} />
        <Route path="/login" element={<Lazy><LoginPage /></Lazy>} />
        <Route path="/register" element={<Lazy><RegisterPage /></Lazy>} />
        <Route element={<RequireAuth />}>
          <Route path="/favorites" element={<Lazy><FavoritesPage /></Lazy>} />
          <Route path="/history" element={<Lazy><HistoryPage /></Lazy>} />
        </Route>
      </Route>
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}
