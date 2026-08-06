import { useEffect, useState } from 'react'
import { isRouteErrorResponse, useNavigate, useRouteError } from 'react-router'
import { AlertTriangleIcon, HomeIcon, RefreshCwIcon } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  Empty,
  EmptyContent,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from '@/components/ui/empty'

// Matches the errors browsers/Vite throw when a dynamically imported route
// chunk fails to load (e.g. after a new deployment invalidates old chunk
// hashes, or a transient network blip during dev HMR).
const CHUNK_LOAD_ERROR_PATTERN =
  /error loading dynamically imported module|failed to fetch dynamically imported module|importing a module script failed/i

const RELOAD_FLAG_KEY = 'orange-tv:route-chunk-reload'

function isChunkLoadError(error: unknown): boolean {
  const message = error instanceof Error ? error.message : String(error)
  return CHUNK_LOAD_ERROR_PATTERN.test(message)
}

/**
 * Route-level error boundary used as the router's `errorElement`.
 *
 * For chunk-load failures (stale/missing lazy-loaded module, common after a
 * new deploy or a flaky dev-server HMR update) it automatically reloads the
 * page once, since a hard reload re-fetches the current module graph and
 * resolves the issue. For any other error it shows a friendly fallback UI
 * instead of the default react-router crash screen.
 */
export function RouteErrorBoundary() {
  const error = useRouteError()
  const navigate = useNavigate()
  const chunkLoadError = isChunkLoadError(error)
  const [reloading, setReloading] = useState(false)

  useEffect(() => {
    if (!chunkLoadError) return
    // Avoid an infinite reload loop if the chunk genuinely can't be loaded.
    if (sessionStorage.getItem(RELOAD_FLAG_KEY)) return
    sessionStorage.setItem(RELOAD_FLAG_KEY, '1')
    setReloading(true)
    window.location.reload()
  }, [chunkLoadError])

  // Clear the guard once a route renders successfully again.
  useEffect(() => {
    return () => sessionStorage.removeItem(RELOAD_FLAG_KEY)
  }, [])

  const title = isRouteErrorResponse(error) ? `${error.status} ${error.statusText}` : '页面出错了'
  const description = reloading
    ? '检测到资源加载失败，正在自动刷新页面...'
    : chunkLoadError
      ? '资源加载失败，请刷新页面重试'
      : error instanceof Error
        ? error.message
        : '发生了未知错误，请稍后重试'

  return (
    <Empty className="min-h-[60vh]">
      <EmptyHeader>
        <EmptyMedia variant="icon">
          <AlertTriangleIcon />
        </EmptyMedia>
        <EmptyTitle>{title}</EmptyTitle>
        <EmptyDescription>{description}</EmptyDescription>
      </EmptyHeader>
      {!reloading ? (
        <EmptyContent>
          <div className="flex gap-3">
            <Button variant="outline" onClick={() => window.location.reload()}>
              <RefreshCwIcon data-icon="inline-start" />
              刷新页面
            </Button>
            <Button onClick={() => navigate('/')}>
              <HomeIcon data-icon="inline-start" />
              返回首页
            </Button>
          </div>
        </EmptyContent>
      ) : null}
    </Empty>
  )
}
