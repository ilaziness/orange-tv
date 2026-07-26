import { useEffect } from 'react'
import { useSiteStore } from '@/store/site'

export function useSite() {
  const site = useSiteStore((s) => s.site)
  const loaded = useSiteStore((s) => s.loaded)
  const loadSite = useSiteStore((s) => s.loadSite)

  useEffect(() => {
    if (!loaded) {
      void loadSite()
    }
  }, [loaded, loadSite])

  useEffect(() => {
    if (site.name) {
      document.title = site.name
    }
  }, [site.name])

  return { site, loaded }
}
