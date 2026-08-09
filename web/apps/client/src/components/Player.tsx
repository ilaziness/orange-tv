import { useEffect, useMemo, useRef, useState } from 'react'
import Artplayer from 'artplayer'
import Hls from 'hls.js'
import flvjs from 'flv.js'
import artplayerPluginHlsControl from 'artplayer-plugin-hls-control'
import type { ClientAdItem } from '@orange-tv/shared'
import { cn } from '@/lib/utils'
import { injectAdScripts } from '@/lib/adCode'

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
  ads?: ClientAdItem[]
  playlist?: PlaylistItem[]
  currentEpisodeId?: number
  onEpisodeChange?: (episodeId: number) => void
  onProgress?: (currentTime: number, duration: number) => void
}

/** Detect player type from a concrete Content-Type header. */
function detectPlayerTypeFromContentType(contentType?: string): '' | 'mp4' | 'm3u8' | 'flv' {
  const ct = (contentType || '').toLowerCase()
  if (ct.includes('flv') || ct.includes('x-flv')) return 'flv'
  if (ct.includes('mpegurl') || ct.includes('m3u8')) return 'm3u8'
  if (ct.includes('mp4')) return 'mp4'
  return ''
}

/** Detect the Artplayer customType key for the given format/src. */
function detectPlayerType(
  format?: string,
  src?: string,
  contentType?: string,
): '' | 'mp4' | 'm3u8' | 'flv' {
  const f = (format || '').toLowerCase()
  const u = (src || '').toLowerCase()
  if (f === 'flv' || u.includes('.flv')) return 'flv'
  if (f === 'hls' || f === 'm3u8' || u.includes('.m3u8')) return 'm3u8'
  if (f === 'mp4' || u.includes('.mp4')) return 'mp4'
  // For extensionless URLs, trust the probed Content-Type (e.g. video/x-flv).
  return detectPlayerTypeFromContentType(contentType)
}

function escapeAttr(str: string): string {
  return str
    .replace(/&/g, '&amp;')
    .replace(/"/g, '&quot;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
}

function buildAdLayerFromItem(item: ClientAdItem) {
  const url = escapeAttr(item.content_url)
  const link = escapeAttr(item.link_url)
  const adHTML =
    item.type === 'image'
      ? `<img src="${url}" style="width:100%;height:100%;object-fit:contain" />`
      : item.type === 'video'
        ? `<video src="${url}" autoplay muted playsinline style="width:100%;height:100%;object-fit:contain" />`
        : item.type === 'code'
          ? `<div class="ad-code-container" style="width:100%;height:100%;display:flex;align-items:center;justify-content:center;overflow:hidden"></div>`
          : `<iframe src="${url}" style="width:100%;height:100%;border:0" allowfullscreen sandbox="allow-scripts allow-same-origin allow-popups allow-forms"></iframe>`

  const skipBtn = `<div class="ad-skip-btn" style="position:absolute;bottom:12px;right:12px;padding:6px 16px;background:rgba(0,0,0,0.7);color:#fff;border-radius:4px;cursor:pointer;font-size:14px">跳过广告</div>`

  const countdown = `<div class="ad-countdown" style="position:absolute;top:12px;right:12px;padding:4px 12px;background:rgba(0,0,0,0.7);color:#fff;border-radius:4px;font-size:14px">广告 ${item.duration}s</div>`

  const linkHTML = link
    ? `<a href="${link}" target="_blank" rel="noopener noreferrer" style="position:absolute;inset:0;display:block;z-index:1"></a>`
    : ''

  return {
    name: 'loadingAd',
    html: `<div style="position:absolute;inset:0;background:#000;display:flex;align-items:center;justify-content:center">${countdown}${linkHTML}${adHTML}${skipBtn}</div>`,
    style: {
      position: 'absolute' as const,
      inset: '0',
      zIndex: '10',
    },
  }
}

const SVG_LIST =
  '<svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="8" y1="6" x2="21" y2="6"/><line x1="8" y1="12" x2="21" y2="12"/><line x1="8" y1="18" x2="21" y2="18"/><line x1="3" y1="6" x2="3.01" y2="6"/><line x1="3" y1="12" x2="3.01" y2="12"/><line x1="3" y1="18" x2="3.01" y2="18"/></svg>'

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

export function VideoPlayer({
  src,
  format,
  poster,
  autoplay = true,
  videoId,
  sourceId,
  episodeId,
  resumeAt,
  ads,
  playlist,
  currentEpisodeId,
  onEpisodeChange,
  onProgress,
}: Props) {
  const containerRef = useRef<HTMLDivElement | null>(null)
  const [playlistVisible, setPlaylistVisible] = useState(false)
  // Probed Content-Type for extensionless URLs (e.g. https://huyazhibo.de5.net/?id=xxx).
  const [probedContentType, setProbedContentType] = useState<string | undefined>(undefined)
  const [probing, setProbing] = useState(false)

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
  const onProgressRef = useRef<((currentTime: number, duration: number) => void) | undefined>(
    onProgress,
  )
  playlistRef.current = playlist
  currentEpRef.current = currentEpisodeId
  onChangeRef.current = onEpisodeChange
  onToggleRef.current = () => setPlaylistVisible((v) => !v)
  onProgressRef.current = onProgress

  // Determine whether we need to probe the actual Content-Type for extensionless URLs.
  // We probe regardless of the declared format, because LivePage defaults to 'hls'
  // and the backend format field is often empty for extensionless URLs.
  const needsProbe = useMemo(() => {
    const u = (src || '').toLowerCase()
    return !u.includes('.m3u8') && !u.includes('.flv') && !u.includes('.mp4')
  }, [src])

  // Probe Content-Type via HEAD so the player can pick the right customType.
  useEffect(() => {
    if (!src || !needsProbe) {
      setProbedContentType(undefined)
      return
    }
    setProbing(true)
    const controller = new AbortController()
    fetch(src, { method: 'HEAD', signal: controller.signal })
      .then(async (res) => {
        const ct = res.headers.get('content-type') || ''
        setProbedContentType(ct)
      })
      .catch(() => {
        setProbedContentType('')
      })
      .finally(() => {
        setProbing(false)
      })
    return () => controller.abort()
  }, [src, needsProbe])

  // Serialize ads to a stable string so the effect doesn't re-run on array identity changes
  const adsKey = ads ? JSON.stringify(ads) : ''

  useEffect(() => {
    const el = containerRef.current
    if (!el || !src) return
    // Wait for the HEAD probe to finish before creating the player.
    if (probing) return

    // Guard to prevent the customType callback from creating an HLS instance
    // after this effect has been cleaned up (race condition with async loading)
    let destroyed = false

    // Destroy any orphaned Artplayer instances that share this container
    // (e.g. from a previous cleanup that threw before completing)
    for (const inst of Artplayer.instances) {
      if (inst.template?.$container === el) {
        try {
          inst.muted = true
        } catch {
          /* ignore */
        }
        try {
          if (inst.video) inst.video.muted = true
        } catch {
          /* ignore */
        }
        try {
          inst.pause()
        } catch {
          /* ignore */
        }
        try {
          inst.destroy(true)
        } catch {
          /* ignore */
        }
      }
    }

    // Clear any leftover DOM from a previous instance (StrictMode remount or re-render)
    el.innerHTML = ''

    // Parse ads back from the serialized string
    const adList: ClientAdItem[] = adsKey ? JSON.parse(adsKey) : []

    const layers: NonNullable<Artplayer['option']['layers']> = []
    if (adList.length > 0) {
      const first = adList[0]
      layers.push(buildAdLayerFromItem(first) as NonNullable<Artplayer['option']['layers']>[number])
    }

    const playbackId =
      videoId != null && sourceId != null && episodeId != null
        ? `video_${videoId}_source_${sourceId}_ep_${episodeId}`
        : undefined

    const playerType = detectPlayerType(format, src, probedContentType)

    const art = new Artplayer({
      container: el,
      url: src,
      type: playerType,
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
      ...(playbackId ? { autoPlayback: !resumeAt, id: playbackId } : { autoPlayback: false }),
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

            // Error recovery: attempt to recover from fatal HLS errors with a
            // bounded retry count. After maxRetries, destroy hls and show a
            // user-visible error notice instead of looping indefinitely.
            let retryCount = 0
            const maxRetries = 3
            hls.on(Hls.Events.ERROR, (_event, data) => {
              if (!data.fatal) return
              if (retryCount >= maxRetries) {
                hls.destroy()
                ;(art as unknown as { hls?: Hls }).hls = undefined
                art.notice.show = '直播源加载失败，请稍后重试或切换频道'
                return
              }
              retryCount++
              switch (data.type) {
                case Hls.ErrorTypes.NETWORK_ERROR:
                  hls.startLoad()
                  break
                case Hls.ErrorTypes.MEDIA_ERROR:
                  hls.recoverMediaError()
                  break
                default:
                  hls.destroy()
                  ;(art as unknown as { hls?: Hls }).hls = undefined
                  art.notice.show = '直播源播放错误，请稍后重试或切换频道'
                  break
              }
            })
          } else if (video.canPlayType('application/vnd.apple.mpegurl')) {
            video.src = url
          }
        },
        flv: (video: HTMLVideoElement, url: string, art: Artplayer) => {
          if (destroyed) return
          if (flvjs.isSupported()) {
            const prevFlv = (art as unknown as { flv?: flvjs.Player }).flv
            if (prevFlv) {
              try {
                prevFlv.pause()
                prevFlv.unload()
                prevFlv.detachMediaElement()
                prevFlv.destroy()
              } catch {
                /* ignore */
              }
            }
            // isLive is determined by whether this is a live stream (no videoId)
            // or a VOD episode (has videoId). Live streams use isLive=true for
            // low-latency playback; VOD uses isLive=false for seekable playback.
            const isLive = videoId == null
            const player = flvjs.createPlayer({
              type: 'flv',
              isLive,
              url,
            })
            player.attachMediaElement(video)
            player.load()
            ;(art as unknown as { flv?: flvjs.Player }).flv = player
          } else {
            // Fallback: set src directly (won't work in most browsers without MSE)
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

    // Multi-ad rotation: play ads in sequence, remove layer when video is ready.
    // Minimum total display: 5 seconds. If video loads faster, keep showing until 5s elapsed.
    let adTimeoutId: ReturnType<typeof setTimeout> | null = null
    let adIntervalId: ReturnType<typeof setInterval> | null = null
    let videoReady = false
    const adStartTime = Date.now()
    const MIN_DISPLAY_MS = 5000
    let onLayerClick: ((e: Event) => void) | null = null

    if (adList.length > 0) {
      const removeAd = () => {
        if (adTimeoutId) {
          clearTimeout(adTimeoutId)
          adTimeoutId = null
        }
        if (adIntervalId) {
          clearInterval(adIntervalId)
          adIntervalId = null
        }
        if (art.layers?.loadingAd) {
          art.layers.loadingAd.remove()
        }
      }

      const tryRemove = () => {
        if (videoReady && Date.now() - adStartTime >= MIN_DISPLAY_MS) {
          removeAd()
        }
      }

      // Start countdown timer for the current ad
      const startCountdown = (duration: number) => {
        if (adIntervalId) {
          clearInterval(adIntervalId)
        }
        let remaining = duration
        const updateText = () => {
          const countdownEl = el.querySelector('.ad-countdown')
          if (countdownEl) {
            countdownEl.textContent = `广告 ${remaining}s`
          }
        }
        updateText()
        adIntervalId = setInterval(() => {
          remaining--
          if (remaining <= 0) {
            if (adIntervalId) {
              clearInterval(adIntervalId)
              adIntervalId = null
            }
            return
          }
          updateText()
        }, 1000)
      }

      const showAd = (index: number) => {
        const item = adList[index]
        if (!item) {
          // All ads played, stay on last until video ready
          return
        }
        const layer = art.layers?.loadingAd as HTMLElement | undefined
        if (layer) {
          const newLayer = buildAdLayerFromItem(item)
          layer.innerHTML = newLayer.html
          // Inject code for type=code ads
          if (item.type === 'code' && item.content_code) {
            const codeContainer = layer.querySelector('.ad-code-container') as HTMLElement | null
            if (codeContainer) {
              injectAdScripts(codeContainer, item.content_code)
            }
          }
        }
        // Start countdown for this ad
        startCountdown(item.duration)
        // Schedule next ad or removal
        adTimeoutId = setTimeout(() => {
          if (index < adList.length - 1) {
            showAd(index + 1)
          } else {
            // Last ad played, wait for video ready + min display
            tryRemove()
          }
        }, item.duration * 1000)
      }

      // Mark video ready and try to remove ad
      const onVideoReady = () => {
        videoReady = true
        tryRemove()
      }
      art.once('video:canplay', onVideoReady)
      art.once('video:playing', onVideoReady)

      // Start countdown for first ad (already in layers)
      startCountdown(adList[0].duration)
      const firstAdCode = adList[0].content_code
      if (adList[0].type === 'code' && firstAdCode) {
        art.on('ready', () => {
          const codeContainer = el.querySelector('.ad-code-container') as HTMLElement | null
          if (codeContainer) {
            injectAdScripts(codeContainer, firstAdCode)
          }
        })
      }

      // If only one ad, schedule its removal
      if (adList.length === 1) {
        adTimeoutId = setTimeout(() => {
          tryRemove()
        }, adList[0].duration * 1000)
      } else {
        // Multiple ads: schedule rotation from second ad
        adTimeoutId = setTimeout(() => {
          showAd(1)
        }, adList[0].duration * 1000)
      }

      // Skip button: use event delegation on the player container so it
      // survives innerHTML replacements during ad rotation.
      onLayerClick = (e: Event) => {
        const target = e.target as HTMLElement
        if (target.classList.contains('ad-skip-btn')) {
          e.preventDefault()
          e.stopPropagation()
          removeAd()
        }
      }
      el.addEventListener('click', onLayerClick)
    }

    // Resume playback from a saved position (e.g. remote history) — once only
    if (resumeAt && resumeAt > 0) {
      const seekToResume = () => {
        try {
          if (art.duration && art.duration > resumeAt) {
            art.currentTime = resumeAt
          }
        } catch {
          /* ignore */
        }
      }
      art.once('video:loadedmetadata', seekToResume)
      art.once('video:canplay', seekToResume)
    }

    // Hide the HLS quality selector when only one resolution is available
    art.on('ready', () => {
      const hlsInstance = (art as unknown as { hls?: Hls }).hls
      if (hlsInstance && hlsInstance.levels.length <= 1) {
        try {
          art.controls.remove('hls-quality')
        } catch {
          /* ignore */
        }
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
      if (adIntervalId) {
        clearInterval(adIntervalId)
        adIntervalId = null
      }
      if (onLayerClick) {
        el.removeEventListener('click', onLayerClick)
        onLayerClick = null
      }

      // Mute first to immediately stop audio, even if later steps fail
      try {
        art.muted = true
      } catch {
        /* ignore */
      }
      try {
        if (art.video) art.video.muted = true
      } catch {
        /* ignore */
      }

      const artHls = (art as unknown as { hls?: Hls }).hls
      if (artHls) {
        artHls.destroy()
        ;(art as unknown as { hls?: Hls }).hls = undefined
      }
      const artFlv = (art as unknown as { flv?: flvjs.Player }).flv
      if (artFlv) {
        try {
          artFlv.pause()
          artFlv.unload()
          artFlv.detachMediaElement()
          artFlv.destroy()
        } catch {
          /* ignore */
        }
        ;(art as unknown as { flv?: flvjs.Player }).flv = undefined
      }
      // Use the local `art` variable, not artRef.current, to ensure we destroy
      // the exact instance created in this effect run (StrictMode double-invoke safe)
      if (art) {
        try {
          art.pause()
        } catch {
          /* ignore */
        }
        try {
          art.video?.pause()
        } catch {
          /* ignore */
        }
        try {
          art.video?.removeAttribute('src')
        } catch {
          /* ignore */
        }
        try {
          art.video?.load()
        } catch {
          /* ignore */
        }
        try {
          art.destroy(true)
        } catch {
          /* ignore */
        }
      }

      // Final safety net: clear any remaining DOM
      el.innerHTML = ''
    }
    // playlist/onEpisodeChange are read via refs (playlistRef/onChangeRef) updated every render,
    // so they are intentionally omitted from the dependency array.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [
    src,
    format,
    poster,
    autoplay,
    adsKey,
    videoId,
    sourceId,
    episodeId,
    resumeAt,
    probedContentType,
    probing,
  ])

  // Resize ArtPlayer when sidebar toggles (container width changes)
  useEffect(() => {
    const el = containerRef.current
    if (!el) return
    // Find the ArtPlayer instance for this container
    for (const inst of Artplayer.instances) {
      if (inst.template?.$container === el) {
        try {
          ;(inst as unknown as { resize?: () => void }).resize?.()
        } catch {
          /* ignore */
        }
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
