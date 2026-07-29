import { useEffect, useRef } from 'react'
import Artplayer from 'artplayer'
import Hls from 'hls.js'
import artplayerPluginHlsControl from 'artplayer-plugin-hls-control'
import type { AdSettings } from '@orange-tv/shared'

// Disable the built-in right-click context menu globally
Artplayer.CONTEXTMENU = false

type Props = {
  src: string
  format?: string
  poster?: string
  autoplay?: boolean
  storageKey?: string
  adConfig?: AdSettings | null
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

export function VideoPlayer({ src, format, poster, autoplay = true, storageKey, adConfig }: Props) {
  const containerRef = useRef<HTMLDivElement | null>(null)

  // Serialize adConfig to a stable string so the effect doesn't re-run on object identity changes
  const adConfigKey = adConfig ? JSON.stringify(adConfig) : ''

  useEffect(() => {
    const el = containerRef.current
    if (!el || !src) return

    // Clear any leftover DOM from a previous instance (StrictMode remount or re-render)
    el.innerHTML = ''

    const hlsRef: { current: Hls | null } = { current: null }

    // Parse adConfig back from the serialized string
    const ad: AdSettings | null = adConfigKey ? JSON.parse(adConfigKey) : null

    const layers: NonNullable<Artplayer['option']['layers']> = []
    if (ad?.enabled && ad.url) {
      layers.push(buildAdLayer(ad) as NonNullable<Artplayer['option']['layers']>[number])
    }

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
      pip: true,
      fullscreen: true,
      fullscreenWeb: true,
      setting: false,
      hotkey: true,
      layers,
      customType: {
        m3u8: (video: HTMLVideoElement, url: string) => {
          if (Hls.isSupported()) {
            const hls = new Hls()
            hls.loadSource(url)
            hls.attachMedia(video)
            hlsRef.current = hls
          } else if (video.canPlayType('application/vnd.apple.mpegurl')) {
            video.src = url
          }
        },
      },
      plugins: [
        artplayerPluginHlsControl({
          quality: { control: true, setting: false },
        }),
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

    // Playback progress memory
    if (storageKey) {
      const saved = Number(localStorage.getItem(storageKey) || 0)
      if (saved > 0) {
        art.on('ready', () => {
          try {
            art.currentTime = saved
          } catch {
            // ignore
          }
        })
      }
      art.on('video:timeupdate', () => {
        const t = art.currentTime
        if (typeof t === 'number' && t > 1) {
          localStorage.setItem(storageKey, String(Math.floor(t)))
        }
      })
    }

    return () => {
      if (adTimeoutId) {
        clearTimeout(adTimeoutId)
        adTimeoutId = null
      }
      if (hlsRef.current) {
        hlsRef.current.destroy()
        hlsRef.current = null
      }
      // Use the local `art` variable, not artRef.current, to ensure we destroy
      // the exact instance created in this effect run (StrictMode double-invoke safe)
      if (art) {
        try { art.pause() } catch { /* ignore */ }
        try { art.video?.pause() } catch { /* ignore */ }
        try { art.video?.removeAttribute('src') } catch { /* ignore */ }
        try { art.video?.load() } catch { /* ignore */ }
        art.destroy(true) // true = remove DOM element completely
      }
    }
  }, [src, format, poster, storageKey, autoplay, adConfigKey])

  if (!src) {
    return (
      <div className="flex items-center justify-center rounded-xl border border-dashed border-border p-8 text-muted-foreground">
        无可播放地址
      </div>
    )
  }

  return <div ref={containerRef} className="w-full" style={{ aspectRatio: '16 / 9' }} />
}
