import { RouterProvider } from 'react-router'
import { router } from '@/router'
import { useBootstrap } from '@/hooks/useBootstrap'
import { AppLoading } from '@/components/AppLoading'

export default function App() {
  const ready = useBootstrap()
  if (!ready) return <AppLoading />
  return <RouterProvider router={router} />
}
