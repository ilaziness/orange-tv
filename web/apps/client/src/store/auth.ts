import { create } from 'zustand'
import type { UserProfile } from '@orange-tv/shared'
import { clientApi, getToken, setToken as setApiToken } from '@/lib/api'

interface AuthState {
  token: string | null
  profile: UserProfile | null
  setToken: (token: string | null) => void
  loadProfile: () => Promise<void>
  logout: () => void
}

export const useAuthStore = create<AuthState>((set) => ({
  token: getToken(),
  profile: null,
  setToken: (token) => {
    setApiToken(token)
    set({ token, profile: null })
  },
  loadProfile: async () => {
    const token = getToken()
    if (!token) {
      set({ token: null, profile: null })
      return
    }
    try {
      const res = await clientApi.profile()
      set({ profile: res.data || null })
    } catch {
      setApiToken(null)
      set({ token: null, profile: null })
    }
  },
  logout: () => {
    setApiToken(null)
    set({ token: null, profile: null })
  },
}))
