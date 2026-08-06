import type * as React from 'react'
import type { AdFormType } from './useAds'
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
import { Textarea } from '@/components/ui/textarea'
import { Spinner } from '@/components/ui/spinner'

interface AdFormDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  editingId: number
  form: AdFormType
  setForm: React.Dispatch<React.SetStateAction<AdFormType>>
  submitting: boolean
  onSubmit: (e: React.SyntheticEvent<HTMLFormElement>) => void
}

export function AdFormDialog({
  open,
  onOpenChange,
  editingId,
  form,
  setForm,
  submitting,
  onSubmit,
}: AdFormDialogProps) {
  const isCodeType = form.type === 'code'

  return (
    <Dialog
      open={open}
      onOpenChange={(v) => {
        if (!submitting) onOpenChange(v)
      }}
    >
      <DialogContent className="sm:max-w-md" showCloseButton={!submitting}>
        <DialogHeader>
          <DialogTitle>{editingId ? '编辑广告' : '新增广告'}</DialogTitle>
          <DialogDescription className="sr-only">填写广告信息后保存</DialogDescription>
        </DialogHeader>
        <form onSubmit={onSubmit} className="flex flex-col gap-5">
          <FieldGroup>
            <Field data-disabled={submitting ? true : undefined}>
              <FieldLabel htmlFor="ad-key">
                广告标识<span className="ml-0.5 text-destructive">*</span>
              </FieldLabel>
              <Input
                id="ad-key"
                placeholder="如 home_sidebar，前端据此区分广告位置"
                value={form.ad_key}
                onChange={(e) => setForm((prev) => ({ ...prev, ad_key: e.target.value }))}
                maxLength={50}
                required
                disabled={submitting}
                autoFocus
              />
              <FieldDescription>唯一标识，前端通过此 key 匹配广告位</FieldDescription>
            </Field>

            <Field data-disabled={submitting ? true : undefined}>
              <FieldLabel htmlFor="ad-title">
                标题<span className="ml-0.5 text-destructive">*</span>
              </FieldLabel>
              <Input
                id="ad-title"
                placeholder="请输入标题"
                value={form.title}
                onChange={(e) => setForm((prev) => ({ ...prev, title: e.target.value }))}
                maxLength={100}
                required
                disabled={submitting}
              />
            </Field>

            <Field data-disabled={submitting ? true : undefined}>
              <FieldLabel>广告场景</FieldLabel>
              <Select
                items={[
                  { value: 'video_loading', label: '播放前广告' },
                  { value: 'general', label: '一般广告' },
                ]}
                value={form.scene}
                onValueChange={(v) => setForm((prev) => ({ ...prev, scene: v ?? '' }))}
              >
                <SelectTrigger className="w-full">
                  <SelectValue placeholder="请选择广告场景" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="video_loading">播放前广告</SelectItem>
                  <SelectItem value="general">一般广告</SelectItem>
                </SelectContent>
              </Select>
            </Field>

            <Field data-disabled={submitting ? true : undefined}>
              <FieldLabel>广告类型</FieldLabel>
              <Select
                items={[
                  { value: 'image', label: '图片' },
                  { value: 'video', label: '视频' },
                  { value: 'html', label: 'HTML' },
                  { value: 'code', label: '广告代码' },
                ]}
                value={form.type}
                onValueChange={(v) => setForm((prev) => ({ ...prev, type: v ?? '' }))}
              >
                <SelectTrigger className="w-full">
                  <SelectValue placeholder="请选择广告类型" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="image">图片</SelectItem>
                  <SelectItem value="video">视频</SelectItem>
                  <SelectItem value="html">HTML</SelectItem>
                  <SelectItem value="code">广告代码</SelectItem>
                </SelectContent>
              </Select>
            </Field>

            {isCodeType ? (
              <Field data-disabled={submitting ? true : undefined}>
                <FieldLabel htmlFor="ad-code">
                  广告代码<span className="ml-0.5 text-destructive">*</span>
                </FieldLabel>
                <Textarea
                  id="ad-code"
                  placeholder="粘贴广告平台代码（如 AdSense 的 script 代码片段）"
                  value={form.content_code}
                  onChange={(e) => setForm((prev) => ({ ...prev, content_code: e.target.value }))}
                  rows={6}
                  required
                  disabled={submitting}
                  className="font-mono text-xs"
                />
                <FieldDescription>支持含 &lt;script&gt; 标签的广告平台代码</FieldDescription>
              </Field>
            ) : (
              <>
                <Field data-disabled={submitting ? true : undefined}>
                  <FieldLabel htmlFor="ad-url">
                    素材URL<span className="ml-0.5 text-destructive">*</span>
                  </FieldLabel>
                  <Input
                    id="ad-url"
                    placeholder="请输入素材URL"
                    value={form.content_url}
                    onChange={(e) => setForm((prev) => ({ ...prev, content_url: e.target.value }))}
                    required
                    disabled={submitting}
                  />
                </Field>

                <Field data-disabled={submitting ? true : undefined}>
                  <FieldLabel htmlFor="ad-link">跳转链接</FieldLabel>
                  <Input
                    id="ad-link"
                    placeholder="请输入点击跳转链接"
                    value={form.link_url}
                    onChange={(e) => setForm((prev) => ({ ...prev, link_url: e.target.value }))}
                    disabled={submitting}
                  />
                </Field>
              </>
            )}

            <Field data-disabled={submitting ? true : undefined}>
              <FieldLabel htmlFor="ad-duration">展示时长（秒）</FieldLabel>
              <Input
                id="ad-duration"
                type="number"
                placeholder="默认5秒"
                value={form.duration}
                onChange={(e) => setForm((prev) => ({ ...prev, duration: e.target.value }))}
                disabled={submitting}
              />
            </Field>

            <Field data-disabled={submitting ? true : undefined}>
              <FieldLabel htmlFor="ad-sort">排序</FieldLabel>
              <Input
                id="ad-sort"
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
                items={[
                  { value: '1', label: '启用' },
                  { value: '0', label: '禁用' },
                ]}
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
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
              disabled={submitting}
            >
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
  )
}
