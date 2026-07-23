import type * as React from 'react'
import type { CollectForm } from './useCollect'
import type { PlaySource } from '@orange-tv/shared'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/dialog'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Field, FieldGroup, FieldLabel, FieldSeparator } from '@/components/ui/field'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Spinner } from '@/components/ui/spinner'

interface CollectSourceFormProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  form: CollectForm
  setForm: React.Dispatch<React.SetStateAction<CollectForm>>
  playSources: PlaySource[]
  editId: number
  onSubmit: (e: React.SyntheticEvent<HTMLFormElement>) => void
  submitting: boolean
  dataRangeOptions: { value: string; label: string }[]
  formError?: string
}

function RequiredMark() {
  return <span className="ml-0.5 text-destructive">*</span>
}

export function CollectSourceForm({
  open,
  onOpenChange,
  form,
  setForm,
  playSources,
  editId,
  onSubmit,
  submitting,
  dataRangeOptions,
  formError,
}: CollectSourceFormProps) {
  const typeOptions = [
    { value: '1', label: '默认格式' },
    { value: '2', label: '苹果CMS' },
  ]
  const playSourceOptions = playSources.map((source) => ({
    value: String(source.id),
    label: source.name,
  }))
  const rangeOptions = dataRangeOptions.map((o) => ({ value: o.value, label: o.label }))

  return (
    <Dialog open={open} onOpenChange={(v) => { if (!submitting) onOpenChange(v) }}>
      <DialogContent className="sm:max-w-2xl" showCloseButton={!submitting}>
        <DialogHeader>
          <DialogTitle>{editId ? '编辑采集源' : '新增采集源'}</DialogTitle>
        </DialogHeader>
        <form onSubmit={onSubmit} className="flex flex-col gap-4">
          {formError && (
            <Alert variant="destructive">
              <AlertDescription>{formError}</AlertDescription>
            </Alert>
          )}
          <FieldGroup className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <Field>
              <FieldLabel htmlFor="name">源名称<RequiredMark /></FieldLabel>
              <Input
                id="name"
                placeholder="请输入源名称"
                value={form.name}
                onChange={(e) => setForm((prev) => ({ ...prev, name: e.target.value }))}
                required
                disabled={submitting}
              />
            </Field>
            <Field>
              <FieldLabel htmlFor="type">类型<RequiredMark /></FieldLabel>
              <Select
                items={typeOptions}
                value={form.type}
                onValueChange={(v) => setForm((prev) => ({ ...prev, type: v ?? '2' }))}
                disabled={submitting}
              >
                <SelectTrigger id="type" disabled={submitting}>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {typeOptions.map((o) => (
                    <SelectItem key={o.value} value={o.value}>{o.label}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </Field>
            <Field className="sm:col-span-2">
              <FieldLabel htmlFor="collect_url">采集地址<RequiredMark /></FieldLabel>
              <Input
                id="collect_url"
                placeholder="如 https://example.com/api.php/provide/vod/from/m3u8/at/json/"
                value={form.collect_url}
                onChange={(e) => setForm((prev) => ({ ...prev, collect_url: e.target.value }))}
                required
                disabled={submitting}
              />
            </Field>
            <Field>
              <FieldLabel htmlFor="api_key">API Key</FieldLabel>
              <Input
                id="api_key"
                placeholder="可选，API密钥"
                value={form.api_key}
                onChange={(e) => setForm((prev) => ({ ...prev, api_key: e.target.value }))}
                disabled={submitting}
              />
            </Field>
            <Field>
              <FieldLabel htmlFor="play_source_id">绑定播放源<RequiredMark /></FieldLabel>
              <Select
                items={playSourceOptions}
                value={form.play_source_id === '0' ? '' : form.play_source_id}
                onValueChange={(v) => setForm((prev) => ({ ...prev, play_source_id: v ?? '0' }))}
                disabled={submitting}
              >
                <SelectTrigger id="play_source_id" disabled={submitting}>
                  <SelectValue placeholder="选择播放源" />
                </SelectTrigger>
                <SelectContent>
                  {playSources.map((ps) => (
                    <SelectItem key={ps.id} value={String(ps.id)}>{ps.name}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </Field>
          </FieldGroup>

          <FieldSeparator>定时配置</FieldSeparator>

          <FieldGroup className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <Field>
              <FieldLabel htmlFor="cron_minute">定时-分钟</FieldLabel>
              <Input
                id="cron_minute"
                type="number"
                min={0}
                max={59}
                placeholder="0-59，如 0"
                value={form.cron_minute}
                onChange={(e) => setForm((prev) => ({ ...prev, cron_minute: e.target.value }))}
                disabled={submitting}
              />
            </Field>
            <Field>
              <FieldLabel htmlFor="cron_hour">定时-小时</FieldLabel>
              <Input
                id="cron_hour"
                placeholder="如 8 或 6,18（多个用英文逗号分隔）"
                value={form.cron_hour}
                onChange={(e) => setForm((prev) => ({ ...prev, cron_hour: e.target.value }))}
                disabled={submitting}
              />
            </Field>
            <Field className="sm:col-span-2">
              <FieldLabel htmlFor="data_range">数据范围</FieldLabel>
              <Select
                items={rangeOptions}
                value={form.data_range}
                onValueChange={(v) => setForm((prev) => ({ ...prev, data_range: v ?? 'all' }))}
                disabled={submitting}
              >
                <SelectTrigger id="data_range" disabled={submitting}>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {dataRangeOptions.map((o) => (
                    <SelectItem key={o.value} value={o.value}>{o.label}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </Field>
          </FieldGroup>

          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)} disabled={submitting}>
              取消
            </Button>
            <Button type="submit" disabled={submitting}>
              {submitting && <Spinner data-icon="inline-start" />}
              {submitting ? '保存中...' : editId ? '保存' : '新增'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
