export interface PlaybackHistoryItem {
  videoId: number
  sourceId: number
  episodeId: number
  progress: number
  title: string
  updatedAt: number
}

const STORAGE_KEY = 'orange_tv_playback_history'
const MAX_ITEMS = 10

export function getHistory(): PlaybackHistoryItem[] {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (!raw) return []
    const parsed = JSON.parse(raw) as PlaybackHistoryItem[]
    if (!Array.isArray(parsed)) return []
    return parsed
  } catch {
    return []
  }
}

export function saveHistory(item: PlaybackHistoryItem): void {
  try {
    const list = getHistory()
    const key = (it: PlaybackHistoryItem) => `${it.videoId}_${it.sourceId}_${it.episodeId}`
    const filtered = list.filter((it) => key(it) !== key(item))
    filtered.unshift(item)
    const trimmed = filtered.slice(0, MAX_ITEMS)
    localStorage.setItem(STORAGE_KEY, JSON.stringify(trimmed))
  } catch {
    // ignore
  }
}

export function formatTime(seconds: number): string {
  const s = Math.floor(seconds)
  const h = Math.floor(s / 3600)
  const m = Math.floor((s % 3600) / 60)
  const sec = s % 60
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${pad(h)}:${pad(m)}:${pad(sec)}`
}
