import { useEffect } from 'react'
import { useSettingsStore } from '@/store/settings'

export function useSettings() {
  const name = useSettingsStore((s) => s.name)
  const logo = useSettingsStore((s) => s.logo)
  const copyright = useSettingsStore((s) => s.copyright)
  const icp = useSettingsStore((s) => s.icp)
  const seo_keywords = useSettingsStore((s) => s.seo_keywords)
  const description = useSettingsStore((s) => s.description)
  const analytics_code = useSettingsStore((s) => s.analytics_code)
  const loaded = useSettingsStore((s) => s.loaded)
  const feature = useSettingsStore((s) => s.feature)
  const seo = useSettingsStore((s) => s.seo)
  const loadSettings = useSettingsStore((s) => s.loadSettings)

  useEffect(() => {
    if (!loaded) {
      void loadSettings()
    }
  }, [loaded, loadSettings])

  return {
    site: { name, logo, copyright, icp, seo_keywords, description, analytics_code },
    feature,
    seo,
    loaded,
  }
}
