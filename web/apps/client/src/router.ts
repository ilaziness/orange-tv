import { createBrowserRouter, type RouteObject } from 'react-router'
import { ClientLayout, loader as clientLayoutLoader } from '@/components/layout/ClientLayout'
import { RequireAuth } from '@/components/auth/RequireAuth'

const routes: RouteObject[] = [
  {
    path: '/',
    Component: ClientLayout,
    loader: clientLayoutLoader,
    children: [
      { index: true, lazy: () => import('@/pages/home/HomePage') },
      { path: 'videos', lazy: () => import('@/pages/videos/VideosPage') },
      { path: 'video/:id', lazy: () => import('@/pages/video/VideoDetailPage') },
      { path: 'play/:id/:sourceId/:episodeId', lazy: () => import('@/pages/video/PlayPage') },
      { path: 'live', lazy: () => import('@/pages/live/LivePage') },
      { path: 'login', lazy: () => import('@/pages/auth/LoginPage') },
      { path: 'register', lazy: () => import('@/pages/auth/RegisterPage') },
      {
        Component: RequireAuth,
        children: [
          { path: 'favorites', lazy: () => import('@/pages/user/FavoritesPage') },
          { path: 'history', lazy: () => import('@/pages/user/HistoryPage') },
          { path: 'profile', lazy: () => import('@/pages/user/ProfilePage') },
        ],
      },
    ],
  },
  {
    path: '*',
    lazy: () => import('@/pages/NotFoundPage'),
  },
]

export const router = createBrowserRouter(routes)
