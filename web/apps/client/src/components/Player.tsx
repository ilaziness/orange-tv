import { useEffect, useRef, useState } from 'react'
import Artplayer from 'artplayer'
import Hls from 'hls.js'
import artplayerPluginHlsControl from 'artplayer-plugin-hls-control'
import type { AdSettings } from '@orange-tv/shared'
import { cn } from '@/lib/utils'

// Disable the built-in right-click context menu globally
Artplayer.CONTEXTMENU = false

type PlaylistItem = { episodeId: number; title: string }

type Props = {
  src: string
  format?: string
  poster?: string
  autoplay?: boolean
  videoId?: number
  sourceId?: number
  episodeId?: number
  resumeAt?: number
  adConfig?: AdSettings | null
  playlist?: PlaylistItem[]
  currentEpisodeId?: number
  onEpisodeChange?: (episodeId: number) => void
  onProgress?: (currentTime: number, duration: number) => void
}

function isHlsSrc(format?: string, src?: string): boolean {
  const f = (format || '').toLowerCase()
  const u = (src || '').toLowerCase()
  return f === 'hls' || u.includes('.m3u8')
}

function escapeAttr(str: string): string {
  return str.replace(/&/g, '&amp;').replace(/"/g, '&quot;').replace(/</g, '&lt;').replace(/>/g, '&gt;')
}

function buildAdLayer(adConfig: AdSettings) {
  const url = escapeAttr(adConfig.url)
  const link = escapeAttr(adConfig.link)
  const adHTML = adConfig.type === 'image'
    ? `<img src="${url}" style="width:100%;height:100%;object-fit:contain" />`
    : adConfig.type === 'video'
      ? `<video src="${url}" autoplay muted playsinline style="width:100%;height:100%;object-fit:contain" />`
      : `<iframe src="${url}" style="width:100%;height:100%;border:0" allowfullscreen></iframe>`

  const skipBtn = adConfig.skipable
    ? `<div class="ad-skip-btn" style="position:absolute;bottom:12px;right:12px;padding:6px 16px;background:rgba(0,0,0,0.7);color:#fff;border-radius:4px;cursor:pointer;font-size:14px">跳过广告</div>`
    : ''

  const linkHTML = link
    ? `<a href="${link}" target="_blank" rel="noopener noreferrer" style="position:absolute;inset:0;display:block;z-index:1"></a>`
    : ''

  return {
    name: 'loadingAd',
    html: `<div style="position:absolute;inset:0;background:#000;display:flex;align-items:center;justify-content:center">${linkHTML}${adHTML}${skipBtn}</div>`,
    style: {
      position: 'absolute' as const,
      inset: '0',
      zIndex: '10',
    },
  }
}

const SVG_LIST = '<svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="8" y1="6" x2="21" y2="6"/><line x1="8" y1="12" x2="21" y2="12"/><line x1="8" y1="18" x2="21" y2="18"/><line x1="3" y1="6" x2="3.01" y2="6"/><line x1="3" y1="12" x2="3.01" y2="12"/><line x1="3" y1="18" x2="3.01" y2="18"/></svg>'

function playlistPlugin(option: {
  playlistRef: React.MutableRefObject<PlaylistItem[] | undefined>
  currentEpRef: React.MutableRefObject<number | undefined>
  onChangeRef: React.MutableRefObject<((episodeId: number) => void) | undefined>
  onToggleRef: React.MutableRefObject<() => void>
}) {
  return (art: Artplayer) => {
    function getItems(): PlaylistItem[] {
      return option.playlistRef.current || []
    }

    function getCurrent(): number {
      return option.currentEpRef.current || 0
    }

    function findIndex(): number {
      const items = getItems()
      const cur = getCurrent()
      return items.findIndex((it) => it.episodeId === cur)
    }

    function changeTo(episodeId: number) {
      const fn = option.onChangeRef.current
      if (fn) fn(episodeId)
    }

    function next() {
      const items = getItems()
      const idx = findIndex()
      if (idx >= 0 && idx < items.length - 1) {
        changeTo(items[idx + 1].episodeId)
      }
    }

    function prev() {
      const items = getItems()
      const idx = findIndex()
      if (idx > 0) {
        changeTo(items[idx - 1].episodeId)
      }
    }

    // Playlist toggle button — delegates to React state
    art.controls.add({
      name: 'playlistToggle',
      position: 'right',
      index: 100,
      html: SVG_LIST,
      tooltip: '播放列表',
      style: { marginRight: '6px', opacity: '0.85', cursor: 'pointer' },
      click: () => option.onToggleRef.current(),
    })

    // Auto play next episode on video end
    art.on('video:ended', () => {
      next()
    })

    return {
      name: 'playlistPlugin',
      next,
      prev,
    }
  }
}

export function VideoPlayer({ src, format, poster, autoplay = true, videoId, sourceId, episodeId, resumeAt, adConfig, playlist, currentEpisodeId, onEpisodeChange, onProgress }: Props) {
  const containerRef = useRef<HTMLDivElement | null>(null)
  const [playlistVisible, setPlaylistVisible] = useState(false)

  // Desktop shows playlist by default; mobile hides it to avoid narrowing the video
  useEffect(() => {
    const mql = window.matchMedia('(min-width: 1024px)')
    setPlaylistVisible(mql.matches)
    const handler = (e: MediaQueryListEvent) => setPlaylistVisible(e.matches)
    mql.addEventListener('change', handler)
    return () => mql.removeEventListener('change', handler)
  }, [])

  // Refs for playlist data — updated every render, read inside plugin closures
  const playlistRef = useRef<PlaylistItem[] | undefined>(playlist)
  const currentEpRef = useRef<number | undefined>(currentEpisodeId)
  const onChangeRef = useRef<((episodeId: number) => void) | undefined>(onEpisodeChange)
  const onToggleRef = useRef<() => void>(() => {})
  const onProgressRef = useRef<((currentTime: number, duration: number) => void) | undefined>(onProgress)
  playlistRef.current = playlist
  currentEpRef.current = currentEpisodeId
  onChangeRef.current = onEpisodeChange
  onToggleRef.current = () => setPlaylistVisible(v => !v)
  onProgressRef.current = onProgress

  // Serialize adConfig to a stable string so the effect doesn't re-run on object identity changes
  const adConfigKey = adConfig ? JSON.stringify(adConfig) : ''

  useEffect(() => {
    const el = containerRef.current
    if (!el || !src) return

    // Guard to prevent the customType callback from creating an HLS instance
    // after this effect has been cleaned up (race condition with async loading)
    let destroyed = false

    // Destroy any orphaned Artplayer instances that share this container
    // (e.g. from a previous cleanup that threw before completing)
    for (const inst of Artplayer.instances) {
      if (inst.template?.$container === el) {
        try { inst.muted = true } catch { /* ignore */ }
        try { if (inst.video) inst.video.muted = true } catch { /* ignore */ }
        try { inst.pause() } catch { /* ignore */ }
        try { inst.destroy(true) } catch { /* ignore */ }
      }
    }

    // Clear any leftover DOM from a previous instance (StrictMode remount or re-render)
    el.innerHTML = ''

    // Parse adConfig back from the serialized string
    const ad: AdSettings | null = adConfigKey ? JSON.parse(adConfigKey) : null

    const layers: NonNullable<Artplayer['option']['layers']> = []
    if (ad?.enabled && ad.url) {
      layers.push(buildAdLayer(ad) as NonNullable<Artplayer['option']['layers']>[number])
    }

    const playbackId = videoId != null && sourceId != null && episodeId != null
      ? `video_${videoId}_source_${sourceId}_ep_${episodeId}`
      : undefined

    const art = new Artplayer({
      container: el,
      url: src,
      type: isHlsSrc(format, src) ? 'm3u8' : '',
      poster: poster || '',
      autoplay,
      playsInline: true,
      autoSize: false,
      autoMini: false,
      loop: false,
      muted: false,
      mutex: true,
      pip: false,
      fullscreen: true,
      fullscreenWeb: true,
      setting: false,
      hotkey: true,
      ...(playbackId
        ? { autoPlayback: !resumeAt, id: playbackId }
        : { autoPlayback: false }),
      layers,
      customType: {
        m3u8: (video: HTMLVideoElement, url: string, art: Artplayer) => {
          if (destroyed) return
          if (Hls.isSupported()) {
            const prevHls = (art as unknown as { hls?: Hls }).hls
            if (prevHls) {
              prevHls.destroy()
            }
            const hls = new Hls()
            hls.loadSource(url)
            hls.attachMedia(video)
            ;(art as unknown as { hls?: Hls }).hls = hls
          } else if (video.canPlayType('application/vnd.apple.mpegurl')) {
            video.src = url
          }
        },
      },
      plugins: [
        artplayerPluginHlsControl({
          quality: { control: true, setting: false },
        }),
        ...(playlist && playlist.length > 0 && onEpisodeChange
          ? [playlistPlugin({ playlistRef, currentEpRef, onChangeRef, onToggleRef })]
          : []),
      ],
    })

    // Remove ad layer when video is ready
    let adTimeoutId: ReturnType<typeof setTimeout> | null = null
    if (ad?.enabled && ad.url) {
      const removeAd = () => {
        if (adTimeoutId) {
          clearTimeout(adTimeoutId)
          adTimeoutId = null
        }
        if (art.layers?.loadingAd) {
          art.layers.loadingAd.remove()
        }
      }
      art.once('video:canplay', removeAd)
      art.once('video:playing', removeAd)

      // Auto-remove after duration
      if (ad.duration > 0) {
        adTimeoutId = setTimeout(() => {
          adTimeoutId = null
          if (art.layers?.loadingAd) {
            removeAd()
          }
        }, ad.duration * 1000)
      }

      // Skip button click handler
      art.on('ready', () => {
        const skipBtn = el.querySelector('.ad-skip-btn')
        if (skipBtn) {
          skipBtn.addEventListener('click', (e) => {
            e.preventDefault()
            e.stopPropagation()
            removeAd()
          })
        }
      })
    }

    // Resume playback from a saved position (e.g. remote history) — once only
    if (resumeAt && resumeAt > 0) {
      const seekToResume = () => {
        try {
          if (art.duration && art.duration > resumeAt) {
            art.currentTime = resumeAt
          }
        } catch { /* ignore */ }
      }
      art.once('video:loadedmetadata', seekToResume)
      art.once('video:canplay', seekToResume)
    }

    // Hide the HLS quality selector when only one resolution is available
    art.on('ready', () => {
      const hlsInstance = (art as unknown as { hls?: Hls }).hls
      if (hlsInstance && hlsInstance.levels.length <= 1) {
        try { art.controls.remove('hls-quality') } catch { /* ignore */ }
      }
    })

    // Progress callback with 3-second throttle — only record after 10 seconds
    let lastSavedTime = 0
    art.on('video:timeupdate', () => {
      const t = art.currentTime
      if (typeof t !== 'number') return
      const fn = onProgressRef.current
      if (!fn) return
      if (t > 10 && t - lastSavedTime >= 3) {
        lastSavedTime = t
        const d = typeof art.duration === 'number' ? art.duration : 0
        fn(Math.floor(t), Math.floor(d))
      }
    })

    return () => {
      destroyed = true

      if (adTimeoutId) {
        clearTimeout(adTimeoutId)
        adTimeoutId = null
      }

      // Mute first to immediately stop audio, even if later steps fail
      try { art.muted = true } catch { /* ignore */ }
      try { if (art.video) art.video.muted = true } catch { /* ignore */ }

      const artHls = (art as unknown as { hls?: Hls }).hls
      if (artHls) {
        artHls.destroy()
        ;(art as unknown as { hls?: Hls }).hls = undefined
      }
      // Use the local `art` variable, not artRef.current, to ensure we destroy
      // the exact instance created in this effect run (StrictMode double-invoke safe)
      if (art) {
        try { art.pause() } catch { /* ignore */ }
        try { art.video?.pause() } catch { /* ignore */ }
        try { art.video?.removeAttribute('src') } catch { /* ignore */ }
        try { art.video?.load() } catch { /* ignore */ }
        try { art.destroy(true) } catch { /* ignore */ }
      }

      // Final safety net: clear any remaining DOM
      el.innerHTML = ''
    }
    // playlist/onEpisodeChange are read via refs (playlistRef/onChangeRef) updated every render,
    // so they are intentionally omitted from the dependency array.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [src, format, poster, autoplay, adConfigKey, videoId, sourceId, episodeId, resumeAt])

  // Resize ArtPlayer when sidebar toggles (container width changes)
  useEffect(() => {
    const el = containerRef.current
    if (!el) return
    // Find the ArtPlayer instance for this container
    for (const inst of Artplayer.instances) {
      if (inst.template?.$container === el) {
        try { (inst as unknown as { resize?: () => void }).resize?.() } catch { /* ignore */ }
        break
      }
    }
  }, [playlistVisible])

  const hasPlaylist = !!(playlist && playlist.length > 0 && onEpisodeChange)

  if (!src) {
    return (
      <div className="flex items-center justify-center rounded-xl border border-dashed border-border p-8 text-muted-foreground">
        无可播放地址
      </div>
    )
  }

  return (
    <div className="relative flex h-[40vh] w-full overflow-hidden sm:h-[55vh] lg:h-[65vh]">
      <div ref={containerRef} className="min-w-0 flex-1" />
      {hasPlaylist && (
        <div
          className={cn(
            'overflow-y-auto overflow-x-hidden bg-zinc-900 text-white transition-all duration-300 ease-in-out',
            'absolute inset-y-0 right-0 z-20 shadow-lg lg:static lg:h-full',
            playlistVisible ? 'w-44 opacity-100' : 'w-0 opacity-0',
          )}
        >
          <div className="w-44 p-2">
            <div className="flex flex-col gap-0.5">
              {playlist!.map((it) => (
                <div
                  key={it.episodeId}
                  className={cn(
                    'cursor-pointer truncate rounded px-3.5 py-2 text-xs transition-colors',
                    it.episodeId === currentEpisodeId
                      ? 'bg-white/15 text-white'
                      : 'text-white/70 hover:bg-white/10 hover:text-white',
                  )}
                  onClick={() => {
                    onEpisodeChange?.(it.episodeId)
                    // Auto-hide overlay playlist on mobile after switching episode
                    if (typeof window !== 'undefined' && window.innerWidth < 1024) {
                      setPlaylistVisible(false)
                    }
                  }}
                >
                  {it.title || `第${it.episodeId}集`}
                </div>
              ))}
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
