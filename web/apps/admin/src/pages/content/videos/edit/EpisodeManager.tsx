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
import { Plus } from 'lucide-react'

interface EpisodeManagerProps {
  episodes: EpisodeDraft[]
  sources: PlaySource[]
  onAdd: () => void
  onUpdate: (index: number, patch: Partial<EpisodeDraft>) => void
}

export function EpisodeManager({ episodes, sources, onAdd, onUpdate }: EpisodeManagerProps) {
  return (
    <div className="rounded-lg border p-4">
      <h3 className="mb-3 font-medium">新增剧集（保存时一并创建）</h3>
      <div className="flex flex-col gap-2">
        {episodes.map((ep, idx) => (
          <div key={idx} className="flex flex-wrap items-center gap-2">
            <Select
              items={sources.map((source) => ({ value: String(source.id), label: source.name }))}
              value={String(ep.source_id)}
              onValueChange={(v) => onUpdate(idx, { source_id: Number(v ?? '0') })}
            >
              <SelectTrigger className="w-32">
                <SelectValue placeholder="播放源" />
              </SelectTrigger>
              <SelectContent>
                {sources.map((s) => (
                  <SelectItem key={s.id} value={String(s.id)}>{s.name}</SelectItem>
                ))}
              </SelectContent>
            </Select>
            <Input
              type="number"
              placeholder="集数"
              value={ep.episode_number}
              onChange={(e) => onUpdate(idx, { episode_number: e.target.value })}
              className="w-20"
            />
            <Input
              placeholder="标题"
              value={ep.title}
              onChange={(e) => onUpdate(idx, { title: e.target.value })}
              className="w-32"
            />
            <Input
              placeholder="播放地址"
              value={ep.play_url}
              onChange={(e) => onUpdate(idx, { play_url: e.target.value })}
              className="min-w-[200px] flex-1"
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
