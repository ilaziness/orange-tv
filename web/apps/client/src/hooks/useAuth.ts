import { useEffect } from 'react'
import { useAuthStore } from '@/store/auth'

export function useAuth() {
  const token = useAuthStore((s) => s.token)
  const profile = useAuthStore((s) => s.profile)
  const loadProfile = useAuthStore((s) => s.loadProfile)

  useEffect(() => {
    if (token && !profile) {
      void loadProfile()
    }
  }, [token, profile, loadProfile])

  return { token, profile, logout: useAuthStore((s) => s.logout) }
}
