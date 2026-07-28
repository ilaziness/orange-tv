import { create } from 'zustand'
import type { AdSettings } from '@orange-tv/shared'
import { clientApi } from '@/lib/api'

const DEFAULT_AD: AdSettings = {
  enabled: false,
  type: '',
  url: '',
  link: '',
  duration: 0,
  skipable: false,
}

interface SiteState {
  name: string
  logo: string
  copyright: string
  icp: string
  seo_keywords: string
  description: string
  ad: AdSettings
  loaded: boolean
  loadSite: () => Promise<void>
}

const DEFAULT_SITE: Omit<SiteState, 'loadSite'> = {
  name: 'Orange TV',
  logo: '/logo.svg',
  copyright: '',
  icp: '',
  seo_keywords: '',
  description: '',
  ad: DEFAULT_AD,
  loaded: false,
}

export const useSiteStore = create<SiteState>((set, get) => ({
  ...DEFAULT_SITE,
  loadSite: async () => {
    if (get().loaded) return
    try {
      const [siteRes, adRes] = await Promise.all([
        clientApi.siteSettings(),
        clientApi.adSettings(),
      ])
      const site = siteRes.data
      const ad = adRes.data
      set({
        name: site?.name || DEFAULT_SITE.name,
        logo: site?.logo || DEFAULT_SITE.logo,
        copyright: site?.copyright || '',
        icp: site?.icp || '',
        seo_keywords: site?.seo_keywords || '',
        description: site?.description || '',
        ad: ad || DEFAULT_AD,
        loaded: true,
      })
    } catch {
      set({ ...DEFAULT_SITE, loaded: true })
    }
  },
}))
