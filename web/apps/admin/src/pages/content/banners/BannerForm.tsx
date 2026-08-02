import type * as React from 'react'
import type { BannerFormType } from './useBanners'
import { VideoPickerDialog } from './VideoPickerDialog'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Field, FieldDescription, FieldGroup, FieldLabel } from '@/components/ui/field'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Spinner } from '@/components/ui/spinner'
import { Film, X } from 'lucide-react'

interface BannerFormDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  editingId: number
  form: BannerFormType
  setForm: React.Dispatch<React.SetStateAction<BannerFormType>>
  submitting: boolean
  onSubmit: (e: React.SyntheticEvent<HTMLFormElement>) => void
  selectedVideo: { id: number; title: string } | null
  onPickVideo: (video: { id: number; title: string }) => void
  videoPickerOpen: boolean
  setVideoPickerOpen: (open: boolean) => void
}

export function BannerFormDialog({
  open,
  onOpenChange,
  editingId,
  form,
  setForm,
  submitting,
  onSubmit,
  selectedVideo,
  onPickVideo,
  videoPickerOpen,
  setVideoPickerOpen,
}: BannerFormDialogProps) {
  return (
    <>
      <Dialog open={open} onOpenChange={(v) => { if (!submitting) onOpenChange(v) }}>
        <DialogContent className="sm:max-w-md" showCloseButton={!submitting}>
          <DialogHeader>
            <DialogTitle>{editingId ? '编辑 Banner' : '新增 Banner'}</DialogTitle>
            <DialogDescription className="sr-only">
              填写 Banner 信息后保存
            </DialogDescription>
          </DialogHeader>
          <form onSubmit={onSubmit} className="flex flex-col gap-5">
            <FieldGroup>
              <Field data-disabled={submitting ? true : undefined}>
                <FieldLabel htmlFor="banner-title">
                  标题<span className="ml-0.5 text-destructive">*</span>
                </FieldLabel>
                <Input
                  id="banner-title"
                  placeholder="请输入标题"
                  value={form.title}
                  onChange={(e) => setForm((prev) => ({ ...prev, title: e.target.value }))}
                  maxLength={100}
                  required
                  disabled={submitting}
                  autoFocus
                />
              </Field>

              <Field data-disabled={submitting ? true : undefined}>
                <FieldLabel htmlFor="banner-cover">
                  封面URL<span className="ml-0.5 text-destructive">*</span>
                </FieldLabel>
                <Input
                  id="banner-cover"
                  placeholder="请输入封面URL"
                  value={form.cover}
                  onChange={(e) => setForm((prev) => ({ ...prev, cover: e.target.value }))}
                  disabled={submitting}
                />
                <FieldDescription>推荐21:9比例，最大尺寸1536 × 658</FieldDescription>
              </Field>

              <Field data-disabled={submitting ? true : undefined}>
                <FieldLabel htmlFor="banner-link">链接</FieldLabel>
                <Input
                  id="banner-link"
                  placeholder="请输入链接"
                  value={form.link}
                  onChange={(e) => setForm((prev) => ({ ...prev, link: e.target.value }))}
                  disabled={submitting}
                />
              </Field>

              <Field data-disabled={submitting ? true : undefined}>
                <FieldLabel>关联影视</FieldLabel>
                {selectedVideo ? (
                  <div className="flex items-center gap-2">
                    <span className="flex flex-1 items-center gap-1.5 rounded-lg border px-3 py-2 text-sm">
                      <Film className="size-4 text-muted-foreground" />
                      {selectedVideo.title || `影视 #${selectedVideo.id}`}
                    </span>
                    <Button
                      type="button"
                      size="sm"
                      variant="outline"
                      onClick={() => setVideoPickerOpen(true)}
                      disabled={submitting}
                    >
                      重新选择
                    </Button>
                    <Button
                      type="button"
                      size="sm"
                      variant="ghost"
                      onClick={() => onPickVideo({ id: 0, title: '' })}
                      disabled={submitting}
                    >
                      <X data-icon="inline-start" />
                      清除
                    </Button>
                  </div>
                ) : (
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    onClick={() => setVideoPickerOpen(true)}
                    disabled={submitting}
                    className="w-full justify-start"
                  >
                    <Film data-icon="inline-start" />
                    请选择关联影视
                  </Button>
                )}
              </Field>

              <Field data-disabled={submitting ? true : undefined}>
                <FieldLabel htmlFor="banner-sort">排序</FieldLabel>
                <Input
                  id="banner-sort"
                  type="number"
                  placeholder="请输入排序（默认0）"
                  value={form.sort}
                  onChange={(e) => setForm((prev) => ({ ...prev, sort: e.target.value }))}
                  disabled={submitting}
                />
              </Field>

              <Field data-disabled={submitting ? true : undefined}>
                <FieldLabel>状态</FieldLabel>
                <Select
                  items={[{ value: '1', label: '启用' }, { value: '0', label: '禁用' }]}
                  value={form.status}
                  onValueChange={(v) => setForm((prev) => ({ ...prev, status: v ?? '' }))}
                >
                  <SelectTrigger className="w-full">
                    <SelectValue placeholder="请选择状态" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="1">启用</SelectItem>
                    <SelectItem value="0">禁用</SelectItem>
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
                {submitting ? '保存中...' : '保存'}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      <VideoPickerDialog
        open={videoPickerOpen}
        onOpenChange={setVideoPickerOpen}
        onSelect={onPickVideo}
      />
    </>
  )
}
