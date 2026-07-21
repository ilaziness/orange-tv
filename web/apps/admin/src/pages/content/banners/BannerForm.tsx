import type * as React from 'react'
import type { BannerForm } from './useBanners'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'

interface BannerFormProps {
  form: BannerForm
  setForm: React.Dispatch<React.SetStateAction<BannerForm>>
  onSubmit: (e: React.SyntheticEvent<HTMLFormElement>) => void
}

export function BannerForm({ form, setForm, onSubmit }: BannerFormProps) {
  return (
    <form onSubmit={onSubmit} className="mb-4 flex flex-wrap items-end gap-2 rounded-lg border p-4">
      <Input
        placeholder="标题"
        value={form.title}
        onChange={(e) => setForm((prev) => ({ ...prev, title: e.target.value }))}
        required
        className="max-w-xs"
      />
      <Input
        placeholder="封面URL"
        value={form.cover}
        onChange={(e) => setForm((prev) => ({ ...prev, cover: e.target.value }))}
        className="max-w-xs"
      />
      <Input
        placeholder="链接"
        value={form.link}
        onChange={(e) => setForm((prev) => ({ ...prev, link: e.target.value }))}
        className="max-w-xs"
      />
      <Input
        type="number"
        placeholder="影视ID"
        value={form.video_id || ''}
        onChange={(e) => setForm((prev) => ({ ...prev, video_id: e.target.value }))}
        className="w-28"
      />
      <Input
        type="number"
        placeholder="排序"
        value={form.sort}
        onChange={(e) => setForm((prev) => ({ ...prev, sort: e.target.value }))}
        className="w-24"
      />
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
      <Button type="submit" size="sm">保存</Button>
    </form>
  )
}
