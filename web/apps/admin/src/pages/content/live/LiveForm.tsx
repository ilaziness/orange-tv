import type * as React from 'react'
import { emptyForm, type LiveFormState } from './useLive'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { RefreshCw } from 'lucide-react'

interface LiveFormProps {
  form: LiveFormState
  setForm: React.Dispatch<React.SetStateAction<LiveFormState>>
  editId: number
  setEditId: (id: number) => void
  onSubmit: (e: React.SyntheticEvent<HTMLFormElement>) => void
  loading: boolean
  onRefresh: () => void
}

export function LiveForm({
  form,
  setForm,
  editId,
  setEditId,
  onSubmit,
  loading,
  onRefresh,
}: LiveFormProps) {
  return (
    <form onSubmit={onSubmit} className="mb-6 flex flex-col gap-4">
      <div className="flex flex-wrap gap-2">
        <Input
          placeholder="频道名称"
          value={form.name}
          onChange={(e) => setForm((prev) => ({ ...prev, name: e.target.value }))}
          required
          className="max-w-xs"
        />
        <Input
          placeholder="分类"
          value={form.category}
          onChange={(e) => setForm((prev) => ({ ...prev, category: e.target.value }))}
          className="max-w-xs"
        />
        <Input
          placeholder="直播流地址"
          value={form.stream_url}
          onChange={(e) => setForm((prev) => ({ ...prev, stream_url: e.target.value }))}
          required
          className="min-w-[280px] flex-1"
        />
      </div>
      <div className="flex flex-wrap items-center gap-2">
        <Input
          placeholder="Logo URL"
          value={form.logo}
          onChange={(e) => setForm((prev) => ({ ...prev, logo: e.target.value }))}
          className="max-w-xs"
        />
        <Input
          placeholder="简介"
          value={form.description}
          onChange={(e) => setForm((prev) => ({ ...prev, description: e.target.value }))}
          className="max-w-xs"
        />
        <Input
          type="number"
          placeholder="排序"
          value={form.sort_order}
          onChange={(e) => setForm((prev) => ({ ...prev, sort_order: e.target.value }))}
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
        <Button type="submit" size="sm">
          {editId ? '保存修改' : '新增频道'}
        </Button>
        {editId ? (
          <Button type="button" variant="outline" size="sm" onClick={() => { setEditId(0); setForm(emptyForm) }}>
            取消
          </Button>
        ) : null}
        <Button type="button" variant="outline" size="sm" onClick={onRefresh} disabled={loading}>
          <RefreshCw data-icon="inline-start" />
          刷新
        </Button>
      </div>
    </form>
  )
}
