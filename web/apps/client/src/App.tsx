import { useEffect } from 'react'
import { BrowserRouter } from 'react-router'
import { AppRoutes } from '@/routes'
import { useSettingsStore } from '@/store/settings'

export default function App() {
  const loadSettings = useSettingsStore((s) => s.loadSettings)

  useEffect(() => {
    void loadSettings()
  }, [loadSettings])

  return (
    <BrowserRouter>
      <AppRoutes />
    </BrowserRouter>
  )
}
