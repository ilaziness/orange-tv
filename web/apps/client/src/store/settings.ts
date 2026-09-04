import { create } from 'zustand'
import type { FeatureSettings, PublicSEOSettings, SiteSettings } from '@orange-tv/shared'
import { clientApi } from '@/lib/api'

const DEFAULT_FEATURE: FeatureSettings = {
  livetv_enabled: false,
  comment_enabled: true,
  comment_review: true,
  rating_enabled: true,
}

const DEFAULT_SEO: PublicSEOSettings = {
  public_base_url: '',
  default_og_image: '',
  google_site_verification: '',
  baidu_site_verification: '',
  bing_site_verification: '',
}

interface SettingsState {
  name: string
  logo: string
  copyright: string
  icp: string
  seo_keywords: string
  description: string
  analytics_code: string
  feature: FeatureSettings
  seo: PublicSEOSettings
  loaded: boolean
  loadSettings: () => Promise<void>
}

const DEFAULT_SETTINGS: Omit<SettingsState, 'loadSettings'> = {
  name: '小橘TV',
  logo: '/logo.svg',
  copyright: '',
  icp: '',
  seo_keywords: '',
  description: '',
  analytics_code: '',
  feature: DEFAULT_FEATURE,
  seo: DEFAULT_SEO,
  loaded: false,
}

export const useSettingsStore = create<SettingsState>((set, get) => ({
  ...DEFAULT_SETTINGS,
  loadSettings: async () => {
    if (get().loaded) return
    try {
      const res = await clientApi.systemSettings(['site', 'feature', 'seo'])
      const data = res.data as Record<string, SiteSettings | FeatureSettings | PublicSEOSettings>
      const site = data.site as SiteSettings | undefined
      const feature = data.feature as FeatureSettings | undefined
      const seo = data.seo as PublicSEOSettings | undefined
      set({
        name: site?.name || DEFAULT_SETTINGS.name,
        logo: site?.logo || DEFAULT_SETTINGS.logo,
        copyright: site?.copyright || '',
        icp: site?.icp || '',
        seo_keywords: site?.seo_keywords || '',
        description: site?.description || '',
        analytics_code: site?.analytics_code || '',
        feature: feature || DEFAULT_FEATURE,
        seo: {
          public_base_url: seo?.public_base_url || '',
          default_og_image: seo?.default_og_image || '',
          google_site_verification: seo?.google_site_verification || '',
          baidu_site_verification: seo?.baidu_site_verification || '',
          bing_site_verification: seo?.bing_site_verification || '',
        },
        loaded: true,
      })
    } catch {
      set({ ...DEFAULT_SETTINGS, loaded: true })
    }
  },
}))
