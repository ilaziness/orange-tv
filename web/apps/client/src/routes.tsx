import { lazy, Suspense } from 'react'
import type { ReactNode } from 'react'
import { Navigate, Route, Routes } from 'react-router'
import { ClientLayout } from './components/layout/ClientLayout.tsx'
import { RequireAuth } from './components/auth/RequireAuth.tsx'
import { Loading } from './components/ui/Loading'

const HomePage = lazy(() => import('./pages/home/HomePage.tsx'))
const CategoryPage = lazy(() => import('./pages/category/CategoryPage.tsx'))
const VideoDetailPage = lazy(() => import('./pages/video/VideoDetailPage.tsx'))
const PlayPage = lazy(() => import('./pages/video/PlayPage.tsx'))
const LivePage = lazy(() => import('./pages/live/LivePage.tsx'))
const LivePlayPage = lazy(() => import('./pages/live/LivePlayPage.tsx'))
const LoginPage = lazy(() => import('./pages/auth/LoginPage.tsx'))
const RegisterPage = lazy(() => import('./pages/auth/RegisterPage.tsx'))
const FavoritesPage = lazy(() => import('./pages/user/FavoritesPage.tsx'))
const HistoryPage = lazy(() => import('./pages/user/HistoryPage.tsx'))

function Lazy({ children }: { children: ReactNode }) {
  return <Suspense fallback={<Loading />}>{children}</Suspense>
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
