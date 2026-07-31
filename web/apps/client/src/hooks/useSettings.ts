import { useEffect } from 'react'
import { useSettingsStore } from '@/store/settings'

export function useSettings() {
  const name = useSettingsStore((s) => s.name)
  const logo = useSettingsStore((s) => s.logo)
  const copyright = useSettingsStore((s) => s.copyright)
  const icp = useSettingsStore((s) => s.icp)
  const loaded = useSettingsStore((s) => s.loaded)
  const feature = useSettingsStore((s) => s.feature)
  const loadSettings = useSettingsStore((s) => s.loadSettings)

  useEffect(() => {
    if (!loaded) {
      void loadSettings()
    }
  }, [loaded, loadSettings])

  useEffect(() => {
    if (name) {
      document.title = name
    }
  }, [name])

  return { site: { name, logo, copyright, icp }, feature, loaded }
}
