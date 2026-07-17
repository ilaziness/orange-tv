import { useAuthStore } from '@/store/auth'
import { useEffect, useState } from 'react'
import { Navigate, Outlet } from 'react-router'
import { Skeleton } from '@/components/ui/skeleton'

export function RequireAuth() {
  const token = useAuthStore((s) => s.token)
  const loadProfile = useAuthStore((s) => s.loadProfile)
  const profile = useAuthStore((s) => s.profile)
  const [ready, setReady] = useState(false)

  useEffect(() => {
    if (!token) {
      setReady(true)
      return
    }
    loadProfile().finally(() => setReady(true))
  }, [token, loadProfile])

  if (!ready) {
    return (
      <div className="flex h-screen w-full items-center justify-center gap-4">
        <Skeleton className="size-12 rounded-full" />
        <div className="flex flex-col gap-2">
          <Skeleton className="h-4 w-32" />
          <Skeleton className="h-3 w-24" />
        </div>
      </div>
    )
  }
  if (!token) return <Navigate to="/login" replace />
  if (!profile) return <Navigate to="/login" replace />
  return <Outlet />
}
