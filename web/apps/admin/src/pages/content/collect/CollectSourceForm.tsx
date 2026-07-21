import type * as React from 'react'
import type { CollectForm } from './useCollect'
import type { PlaySource } from '@orange-tv/shared'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { RefreshCw } from 'lucide-react'

interface CollectSourceFormProps {
  form: CollectForm
  setForm: React.Dispatch<React.SetStateAction<CollectForm>>
  playSources: PlaySource[]
  selectedId: number
  onSubmit: (e: React.SyntheticEvent<HTMLFormElement>) => void
  onRefresh: () => void
}

export function CollectSourceForm({
  form,
  setForm,
  playSources,
  selectedId,
  onSubmit,
  onRefresh,
}: CollectSourceFormProps) {
  return (
    <form onSubmit={onSubmit} className="mb-6 flex flex-col gap-4">
      <div className="flex flex-wrap gap-2">
        <Input
          placeholder="源名称"
          value={form.name}
          onChange={(e) => setForm((prev) => ({ ...prev, name: e.target.value }))}
          required
          className="max-w-xs"
        />
        <Select
          items={[{ value: '1', label: '默认格式' }, { value: '2', label: '苹果CMS' }]}
          value={form.type}
          onValueChange={(v) => setForm((prev) => ({ ...prev, type: v ?? '2' }))}
        >
          <SelectTrigger className="w-32">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="1">默认格式</SelectItem>
            <SelectItem value="2">苹果CMS</SelectItem>
          </SelectContent>
        </Select>
        <Input
          placeholder="采集地址"
          value={form.collect_url}
          onChange={(e) => setForm((prev) => ({ ...prev, collect_url: e.target.value }))}
          required
          className="min-w-[280px] flex-1"
        />
      </div>
      <div className="flex flex-wrap items-center gap-2">
        <Input
          placeholder="API Key"
          value={form.api_key}
          onChange={(e) => setForm((prev) => ({ ...prev, api_key: e.target.value }))}
          className="max-w-xs"
        />
        <Input
          placeholder="cron 表达式(空=不定时)"
          value={form.cron_expr}
          onChange={(e) => setForm((prev) => ({ ...prev, cron_expr: e.target.value }))}
          className="max-w-xs"
        />
        <Select
          items={playSources.map((source) => ({ value: String(source.id), label: source.name }))}
          value={form.play_source_id}
          onValueChange={(v) => setForm((prev) => ({ ...prev, play_source_id: v ?? '0' }))}
        >
          <SelectTrigger className="w-40">
            <SelectValue placeholder="绑定播放源" />
          </SelectTrigger>
          <SelectContent>
            {playSources.map((ps) => (
              <SelectItem key={ps.id} value={String(ps.id)}>{ps.name}</SelectItem>
            ))}
          </SelectContent>
        </Select>
        <Select
          items={[{ value: '1', label: '启用' }, { value: '0', label: '禁用' }]}
          value={form.status}
          onValueChange={(v) => setForm((prev) => ({ ...prev, status: v ?? '1' }))}
        >
          <SelectTrigger className="w-28">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="1">启用</SelectItem>
            <SelectItem value="0">禁用</SelectItem>
          </SelectContent>
        </Select>
        <Textarea
          placeholder="额外配置(JSON)"
          value={form.config}
          onChange={(e) => setForm((prev) => ({ ...prev, config: e.target.value }))}
          className="max-w-xs"
          rows={1}
        />
        <Button type="submit" size="sm">{selectedId ? '保存源' : '新增源'}</Button>
        <Button type="button" variant="outline" size="sm" onClick={() => void onRefresh()}>
          <RefreshCw data-icon="inline-start" />
          刷新
        </Button>
      </div>
    </form>
  )
}
