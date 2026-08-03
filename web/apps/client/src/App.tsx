import { useEffect } from 'react'
import { RouterProvider } from 'react-router'
import { router } from '@/router'
import { useSettingsStore } from '@/store/settings'

export default function App() {
  const loadSettings = useSettingsStore((s) => s.loadSettings)

  useEffect(() => {
    void loadSettings()
  }, [loadSettings])

  return <RouterProvider router={router} />
}
