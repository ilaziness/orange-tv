import { useEffect } from 'react'
import { useSiteStore } from '@/store/site'

export function useSite() {
  const name = useSiteStore((s) => s.name)
  const logo = useSiteStore((s) => s.logo)
  const copyright = useSiteStore((s) => s.copyright)
  const icp = useSiteStore((s) => s.icp)
  const loaded = useSiteStore((s) => s.loaded)
  const loadSite = useSiteStore((s) => s.loadSite)

  useEffect(() => {
    if (!loaded) {
      void loadSite()
    }
  }, [loaded, loadSite])

  useEffect(() => {
    if (name) {
      document.title = name
    }
  }, [name])

  return { site: { name, logo, copyright, icp }, loaded }
}
