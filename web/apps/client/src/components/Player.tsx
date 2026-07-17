import { useEffect, useRef } from 'react'
import videojs from 'video.js'
import type Player from 'video.js/dist/types/player'
import 'video.js/dist/video-js.css'

type Props = {
  src: string
  format?: string
  poster?: string
  autoplay?: boolean
  storageKey?: string
}

function mimeFromFormat(format?: string, src?: string): string {
  const f = (format || '').toLowerCase()
  const u = (src || '').toLowerCase()
  if (f === 'mp4' || u.includes('.mp4')) return 'video/mp4'
  if (f === 'dash' || u.includes('.mpd')) return 'application/dash+xml'
  if (f === 'flv' || u.includes('.flv')) return 'video/x-flv'
  return 'application/x-mpegURL'
}

export function VideoPlayer({ src, format, poster, autoplay = true, storageKey }: Props) {
  const videoRef = useRef<HTMLDivElement | null>(null)
  const playerRef = useRef<Player | null>(null)

  useEffect(() => {
    const el = videoRef.current
    if (!el || !src) return

    // video.js expects a <video> element it can own; create one under the container
    const videoEl = document.createElement('video')
    videoEl.className = 'video-js vjs-big-play-centered'
    videoEl.setAttribute('playsinline', 'true')
    el.innerHTML = ''
    el.appendChild(videoEl)

    const player = videojs(videoEl, {
      controls: true,
      autoplay,
      preload: 'auto',
      fluid: true,
      poster,
      sources: [{ src, type: mimeFromFormat(format, src) }],
    })
    playerRef.current = player

    if (storageKey) {
      const saved = Number(localStorage.getItem(storageKey) || 0)
      if (saved > 0) {
        player.ready(() => {
          try {
            player.currentTime(saved)
          } catch {
            // ignore
          }
        })
      }
      player.on('timeupdate', () => {
        const t = player.currentTime()
        if (typeof t === 'number' && t > 1) {
          localStorage.setItem(storageKey, String(Math.floor(t)))
        }
      })
    }

    return () => {
      if (playerRef.current && !playerRef.current.isDisposed()) {
        playerRef.current.dispose()
      }
      playerRef.current = null
      if (el) el.innerHTML = ''
    }
  }, [src, format, poster, storageKey, autoplay])

  if (!src) {
    return (
      <div className="flex items-center justify-center rounded-xl border border-dashed border-border p-8 text-muted-foreground">
        无可播放地址
      </div>
    )
  }

  return (
    <div data-vjs-player>
      <div ref={videoRef} />
    </div>
  )
}
