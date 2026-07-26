import { useEffect } from 'react'
import { BrowserRouter } from 'react-router'
import { AppRoutes } from '@/routes'
import { useSiteStore } from '@/store/site'

export default function App() {
  const loadSite = useSiteStore((s) => s.loadSite)

  useEffect(() => {
    void loadSite()
  }, [loadSite])

  return (
    <BrowserRouter>
      <AppRoutes />
    </BrowserRouter>
  )
}
