import type * as React from 'react'
import type { LiveFormState } from './useLive'

import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Field, FieldGroup, FieldLabel } from '@/components/ui/field'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Spinner } from '@/components/ui/spinner'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'

interface LiveDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  title: string
  form: LiveFormState
  setForm: React.Dispatch<React.SetStateAction<LiveFormState>>
  submitting: boolean
  error: string
  onSubmit: (e: React.SyntheticEvent<HTMLFormElement>) => void
}

export function LiveDialog({
  open,
  onOpenChange,
  title,
  form,
  setForm,
  submitting,
  error,
  onSubmit,
}: LiveDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-lg" showCloseButton={!submitting}>
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
        </DialogHeader>
        <form onSubmit={onSubmit} className="flex flex-col gap-5">
          {error && (
            <Alert variant="destructive">
              <AlertTitle>出错了</AlertTitle>
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}
          <FieldGroup>
            <Field>
              <FieldLabel htmlFor="live-name">频道名称</FieldLabel>
              <Input
                id="live-name"
                placeholder="请输入频道名称"
                value={form.name}
                onChange={(e) => setForm((prev) => ({ ...prev, name: e.target.value }))}
                maxLength={100}
                required
                disabled={submitting}
              />
            </Field>
            <Field>
              <FieldLabel htmlFor="live-category">分类</FieldLabel>
              <Input
                id="live-category"
                placeholder="请输入分类"
                value={form.category}
                onChange={(e) => setForm((prev) => ({ ...prev, category: e.target.value }))}
                maxLength={50}
                disabled={submitting}
              />
            </Field>
            <Field>
              <FieldLabel htmlFor="live-stream-url">直播流地址</FieldLabel>
              <Input
                id="live-stream-url"
                placeholder="请输入直播流地址"
                value={form.stream_url}
                onChange={(e) => setForm((prev) => ({ ...prev, stream_url: e.target.value }))}
                maxLength={1000}
                required
                disabled={submitting}
              />
            </Field>
            <Field>
              <FieldLabel htmlFor="live-logo">Logo URL</FieldLabel>
              <Input
                id="live-logo"
                placeholder="请输入 Logo URL"
                value={form.logo}
                onChange={(e) => setForm((prev) => ({ ...prev, logo: e.target.value }))}
                maxLength={500}
                disabled={submitting}
              />
            </Field>
            <Field>
              <FieldLabel htmlFor="live-description">简介</FieldLabel>
              <Input
                id="live-description"
                placeholder="请输入简介"
                value={form.description}
                onChange={(e) => setForm((prev) => ({ ...prev, description: e.target.value }))}
                maxLength={2000}
                disabled={submitting}
              />
            </Field>
            <Field>
              <FieldLabel htmlFor="live-sort-order">排序</FieldLabel>
              <Input
                id="live-sort-order"
                type="number"
                min="0"
                max="4294967295"
                step="1"
                placeholder="排序"
                value={form.sort_order}
                onChange={(e) => setForm((prev) => ({ ...prev, sort_order: e.target.value }))}
                disabled={submitting}
              />
            </Field>
            <Field>
              <FieldLabel htmlFor="live-status">状态</FieldLabel>
              <Select
                items={[
                  { value: '1', label: '启用' },
                  { value: '0', label: '禁用' },
                ]}
                value={form.status}
                onValueChange={(v) => setForm((prev) => ({ ...prev, status: v ?? '1' }))}
                disabled={submitting}
              >
                <SelectTrigger id="live-status">
                  <SelectValue />
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
