import { useEffect } from 'react'
import { useSettingsStore } from '@/store/settings'

export function usePageTitle(title: string) {
  const siteName = useSettingsStore((s) => s.name)
  useEffect(() => {
    document.title = `${title} | ${siteName}`
  }, [title, siteName])
}
