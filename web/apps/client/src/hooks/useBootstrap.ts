import { useEffect, useState } from 'react'
import { useSettingsStore } from '@/store/settings'
import { useAdsStore } from '@/store/ads'

const BOOTSTRAP_TIMEOUT_MS = 30_000

/**
 * 启动加载 hook：并行预加载全局配置（settings）和广告（ads general scene），
 * 加载完成或 30s 超时后返回 ready=true，调用方据此决定是否渲染主应用。
 *
 * 失败兜底已在 store 内 catch，不会 reject；超时后用已加载部分 + 默认值兜底。
 */
export function useBootstrap() {
  const [ready, setReady] = useState(false)
  const loadSettings = useSettingsStore((s) => s.loadSettings)
  const loadAds = useAdsStore((s) => s.loadAds)

  useEffect(() => {
    let cancelled = false
    let timer: ReturnType<typeof setTimeout> | undefined

    const tasks = Promise.all([loadSettings(), loadAds()])
    const timeout = new Promise<'timeout'>((resolve) => {
      timer = setTimeout(() => resolve('timeout'), BOOTSTRAP_TIMEOUT_MS)
    })

    Promise.race([tasks, timeout]).finally(() => {
      if (cancelled) return
      if (timer) clearTimeout(timer)
      setReady(true)
    })

    return () => {
      cancelled = true
      if (timer) clearTimeout(timer)
    }
  }, [loadSettings, loadAds])

  return ready
}
