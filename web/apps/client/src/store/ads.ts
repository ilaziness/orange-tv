import { create } from 'zustand'
import type { ClientAdItem } from '@orange-tv/shared'
import { clientApi } from '@/lib/api'

interface AdsState {
  ads: ClientAdItem[]
  loaded: boolean
  loadAds: () => Promise<void>
}

export const useAdsStore = create<AdsState>((set, get) => ({
  ads: [],
  loaded: false,
  loadAds: async () => {
    if (get().loaded) return
    try {
      const res = await clientApi.ads('general')
      set({ ads: res.data || [], loaded: true })
    } catch {
      set({ ads: [], loaded: true })
    }
  },
}))
