import { create } from 'zustand'
import type { PublicSiteInfo } from '@orange-tv/shared'
import { clientApi } from '@/lib/api'

const DEFAULT_SITE: PublicSiteInfo = {
  name: 'Orange TV',
  logo: '/logo.svg',
  copyright: '',
  icp: '',
  seo_keywords: '',
  description: '',
}

interface SiteState {
  site: PublicSiteInfo
  loaded: boolean
  loadSite: () => Promise<void>
}

export const useSiteStore = create<SiteState>((set, get) => ({
  site: DEFAULT_SITE,
  loaded: false,
  loadSite: async () => {
    if (get().loaded) return
    try {
      const res = await clientApi.site()
      const data = res.data
      set({
        site: {
          name: data?.name || DEFAULT_SITE.name,
          logo: data?.logo || DEFAULT_SITE.logo,
          copyright: data?.copyright || '',
          icp: data?.icp || '',
          seo_keywords: data?.seo_keywords || '',
          description: data?.description || '',
        },
        loaded: true,
      })
    } catch {
      set({ site: DEFAULT_SITE, loaded: true })
    }
  },
}))
