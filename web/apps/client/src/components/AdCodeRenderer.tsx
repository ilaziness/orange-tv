import { useEffect, useRef } from 'react'
import { injectAdScripts } from '@/lib/adCode'

interface AdCodeRendererProps {
  code: string
  className?: string
}

/**
 * AdCodeRenderer executes third-party ad platform code (e.g. AdSense).
 * It delegates script injection to the shared `injectAdScripts` utility
 * and removes injected scripts on cleanup.
 */
export function AdCodeRenderer({ code, className }: AdCodeRendererProps) {
  const containerRef = useRef<HTMLDivElement>(null)
  const scriptsRef = useRef<HTMLScriptElement[]>([])

  useEffect(() => {
    const container = containerRef.current
    if (!container || !code) return

    // Clean up previous scripts.
    scriptsRef.current.forEach((s) => s.remove())
    scriptsRef.current = []
    container.innerHTML = ''

    scriptsRef.current = injectAdScripts(container, code)

    return () => {
      scriptsRef.current.forEach((s) => s.remove())
      scriptsRef.current = []
    }
  }, [code])

  return <div ref={containerRef} className={className} />
}
