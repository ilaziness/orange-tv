import { useEffect, useRef } from 'react'
import { useLocation } from 'react-router'
import { injectAdScripts } from '@/lib/adCode'

interface AnalyticsCodeProps {
  code: string
}

function trackPageView(path: string, href: string, title: string) {
  const w = window as unknown as Record<string, unknown>

  if (typeof w.gtag === 'function') {
    ;(w.gtag as (...args: unknown[]) => void)('event', 'page_view', {
      page_location: href,
      page_path: path,
      page_title: title,
    })
  }

  const hmt = w._hmt
  if (hmt && typeof (hmt as { push?: unknown }).push === 'function') {
    ;((hmt as { push: (...args: unknown[]) => void }).push as (...args: unknown[]) => void)([
      '_trackPageview',
      path,
    ])
  }
}

/**
 * AnalyticsCode injects third-party web analytics snippets (Baidu, Google
 * Analytics, etc.) into a hidden container and reports virtual page views on
 * SPA route changes.
 *
 * The initial page view is intentionally skipped because the analytics snippet
 * itself sends one when it loads.
 */
export function AnalyticsCode({ code }: AnalyticsCodeProps) {
  const containerRef = useRef<HTMLDivElement>(null)
  const scriptsRef = useRef<HTMLScriptElement[]>([])
  const location = useLocation()
  const initialPageViewRef = useRef(true)

  useEffect(() => {
    const container = containerRef.current
    if (!container) return

    scriptsRef.current.forEach((s) => s.remove())
    scriptsRef.current = []
    container.innerHTML = ''

    if (!code) return

    scriptsRef.current = injectAdScripts(container, code)

    return () => {
      scriptsRef.current.forEach((s) => s.remove())
      scriptsRef.current = []
      container.innerHTML = ''
    }
  }, [code])

  useEffect(() => {
    if (initialPageViewRef.current) {
      initialPageViewRef.current = false
      return
    }

    const timer = setTimeout(() => {
      trackPageView(location.pathname + location.search, window.location.href, document.title)
    }, 0)

    return () => clearTimeout(timer)
  }, [location.pathname, location.search])

  return <div ref={containerRef} className="hidden" aria-hidden="true" />
}
