import { create } from 'zustand'
import type { AdminProfile } from '@orange-tv/shared'
import { adminApi, getToken, setToken } from '../lib/api'

type AuthState = {
  token: string | null
  profile: AdminProfile | null
  loading: boolean
  login: (username: string, password: string) => Promise<void>
  logout: () => Promise<void>
  loadProfile: () => Promise<void>
}

export const useAuthStore = create<AuthState>((set, get) => ({
  token: getToken(),
  profile: null,
  loading: false,
  async login(username, password) {
    set({ loading: true })
    try {
      const res = await adminApi.login(username, password)
      setToken(res.data.access_token)
      set({ token: res.data.access_token, profile: res.data.admin, loading: false })
    } catch (err) {
      set({ loading: false })
      throw err
    }
  },
  async logout() {
    try {
      if (get().token) await adminApi.logout()
    } catch {
      // ignore
    }
    setToken(null)
    set({ token: null, profile: null })
  },
  async loadProfile() {
    if (!get().token) return
    set({ loading: true })
    try {
      const res = await adminApi.profile()
      set({ profile: res.data, loading: false })
    } catch (err) {
      setToken(null)
      set({ token: null, profile: null, loading: false })
      throw err
    }
  },
}))
