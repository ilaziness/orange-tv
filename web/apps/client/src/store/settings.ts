import { create } from 'zustand'
import type { AdSettings, FeatureSettings, SiteSettings } from '@orange-tv/shared'
import { clientApi } from '@/lib/api'

const DEFAULT_AD: AdSettings = {
  enabled: false,
  type: '',
  url: '',
  link: '',
  duration: 0,
  skipable: false,
}

const DEFAULT_FEATURE: FeatureSettings = {
  live_enabled: false,
  comment_enabled: true,
  comment_review: true,
  rating_enabled: true,
}

interface SettingsState {
  name: string
  logo: string
  copyright: string
  icp: string
  seo_keywords: string
  description: string
  ad: AdSettings
  feature: FeatureSettings
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
  ad: DEFAULT_AD,
  feature: DEFAULT_FEATURE,
  loaded: false,
}

export const useSettingsStore = create<SettingsState>((set, get) => ({
  ...DEFAULT_SETTINGS,
  loadSettings: async () => {
    if (get().loaded) return
    try {
      const res = await clientApi.systemSettings(['site', 'ad', 'feature'])
      const data = res.data as Record<string, SiteSettings | AdSettings | FeatureSettings>
      const site = data.site as SiteSettings | undefined
      const ad = data.ad as AdSettings | undefined
      const feature = data.feature as FeatureSettings | undefined
      set({
        name: site?.name || DEFAULT_SETTINGS.name,
        logo: site?.logo || DEFAULT_SETTINGS.logo,
        copyright: site?.copyright || '',
        icp: site?.icp || '',
        seo_keywords: site?.seo_keywords || '',
        description: site?.description || '',
        ad: ad || DEFAULT_AD,
        feature: feature || DEFAULT_FEATURE,
        loaded: true,
      })
    } catch {
      set({ ...DEFAULT_SETTINGS, loaded: true })
    }
  },
}))
