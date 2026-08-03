import { useAuthStore } from '@/store/auth'
import { useEffect, useState } from 'react'
import { Navigate, Outlet } from 'react-router'
import { Spinner } from '@/components/ui/spinner'

export function RequireAuth() {
  const token = useAuthStore((s) => s.token)
  const profile = useAuthStore((s) => s.profile)
  const [ready, setReady] = useState(false)

  useEffect(() => {
    if (!token) {
      setReady(true)
      return
    }
    const loadProfile = useAuthStore.getState().loadProfile
    loadProfile().finally(() => setReady(true))
  }, [token])

  if (!ready)
    return (
      <div className="flex items-center justify-center py-20">
        <Spinner className="size-6" />
      </div>
    )
  if (!token) return <Navigate to="/login" replace />
  if (!profile) return <Navigate to="/login" replace />
  return <Outlet />
}
