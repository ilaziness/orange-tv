import type { PlaySource } from '@orange-tv/shared'
import type { EpisodeDraft } from './useVideoEdit'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Plus, Trash2 } from 'lucide-react'

interface EpisodeManagerProps {
  episodes: EpisodeDraft[]
  sources: PlaySource[]
  onAdd: () => void
  onUpdate: (index: number, patch: Partial<EpisodeDraft>) => void
  onRemove: (index: number) => void
}

export function EpisodeManager({ episodes, sources, onAdd, onUpdate, onRemove }: EpisodeManagerProps) {
  return (
    <div className="rounded-lg border p-4">
      <h3 className="mb-3 font-medium">新增剧集（保存时一并创建）</h3>
      <div className="flex flex-col gap-2">
        {episodes.map((ep, idx) => (
          <div key={idx} className="flex flex-wrap items-center gap-2">
            <Select
              items={sources.map((source) => ({ value: String(source.id), label: source.name }))}
              value={ep.source_id ? String(ep.source_id) : ''}
              onValueChange={(v) => onUpdate(idx, { source_id: Number(v ?? '0') })}
            >
              <SelectTrigger className="w-32">
                <SelectValue placeholder="选择播放源" />
              </SelectTrigger>
              <SelectContent>
                {sources.map((s) => (
                  <SelectItem key={s.id} value={String(s.id)}>{s.name}</SelectItem>
                ))}
              </SelectContent>
            </Select>
            <Input
              type="number"
              placeholder="集数，如 1"
              value={ep.episode_number}
              onChange={(e) => onUpdate(idx, { episode_number: e.target.value })}
              className="w-32"
            />
            <Input
              placeholder="剧集标题，如 第1集"
              value={ep.title}
              onChange={(e) => onUpdate(idx, { title: e.target.value })}
              className="w-44"
            />
            <Input
              placeholder="播放地址，如 https://..."
              value={ep.play_url}
              onChange={(e) => onUpdate(idx, { play_url: e.target.value })}
              className="min-w-[220px] flex-1"
            />
            <Select
              items={[{ value: 'hls', label: 'hls' }, { value: 'mp4', label: 'mp4' }, { value: 'dash', label: 'dash' }, { value: 'flv', label: 'flv' }]}
              value={ep.format}
              onValueChange={(v) => onUpdate(idx, { format: v ?? 'hls' })}
            >
              <SelectTrigger className="w-24">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="hls">hls</SelectItem>
                <SelectItem value="mp4">mp4</SelectItem>
                <SelectItem value="dash">dash</SelectItem>
                <SelectItem value="flv">flv</SelectItem>
              </SelectContent>
            </Select>
            <Button
              type="button"
              size="icon-sm"
              variant="ghost"
              onClick={() => onRemove(idx)}
              title="删除该行"
            >
              <Trash2 />
              <span className="sr-only">删除该行</span>
            </Button>
          </div>
        ))}
        <Button type="button" variant="outline" size="sm" onClick={onAdd}>
          <Plus data-icon="inline-start" />
          添加剧集行
        </Button>
      </div>
    </div>
  )
}
