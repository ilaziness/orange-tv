import { create } from 'zustand'
import type { UserProfile } from '@orange-tv/shared'
import {
  clientApi,
  clearTokens,
  getToken,
  isAuthSessionExpired,
  registerTokenListener,
  setToken as persistAccessToken,
  setRefreshToken as persistRefreshToken,
} from '@/lib/api'

interface AuthState {
  token: string | null
  profile: UserProfile | null
  setToken: (token: string | null) => void
  setTokens: (accessToken: string, refreshToken: string) => void
  completeLogin: (accessToken: string, refreshToken: string, profile: UserProfile) => void
  loadProfile: () => Promise<void>
  logout: () => void
}

export const useAuthStore = create<AuthState>((set) => ({
  token: getToken(),
  profile: null,
  setToken: (token) => {
    if (token) {
      persistAccessToken(token)
    } else {
      clearTokens()
    }
    set({ token, profile: null })
  },
  setTokens: (accessToken, refreshToken) => {
    persistAccessToken(accessToken)
    persistRefreshToken(refreshToken)
    set({ token: accessToken, profile: null })
  },
  completeLogin: (accessToken, refreshToken, profile) => {
    persistAccessToken(accessToken)
    persistRefreshToken(refreshToken)
    set({ token: accessToken, profile })
  },
  loadProfile: async () => {
    const token = getToken()
    if (!token) {
      clearTokens()
      set({ token: null, profile: null })
      return
    }
    try {
      const res = await clientApi.profile()
      set({ profile: res.data || null })
    } catch (err) {
      if (isAuthSessionExpired(err)) {
        clearTokens()
        set({ token: null, profile: null })
      }
    }
  },
  logout: () => {
    clearTokens()
    set({ token: null, profile: null })
  },
}))

registerTokenListener((tokens) => {
  if (tokens) {
    useAuthStore.setState({ token: tokens.accessToken, profile: null })
  } else {
    useAuthStore.setState({ token: null, profile: null })
  }
})
