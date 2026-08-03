import type * as React from 'react'
import type { PlaySourceFieldErrors, PlaySourceForm } from './usePlaySources'
import { statusOptions } from './usePlaySources'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Field, FieldError, FieldGroup, FieldLabel } from '@/components/ui/field'
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

function RequiredMark() {
  return <span className="ml-0.5 text-destructive">*</span>
}

interface PlaySourceDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  title: string
  form: PlaySourceForm
  updateForm: (patch: Partial<PlaySourceForm>) => void
  submitting: boolean
  error: string
  fieldErrors: PlaySourceFieldErrors
  onSubmit: (e: React.SyntheticEvent<HTMLFormElement>) => void
}

export function PlaySourceDialog({
  open,
  onOpenChange,
  title,
  form,
  updateForm,
  submitting,
  error,
  fieldErrors,
  onSubmit,
}: PlaySourceDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md" showCloseButton={!submitting}>
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
          <DialogDescription className="sr-only">
            填写播放源名称、排序与状态后保存
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={onSubmit} className="flex flex-col gap-5">
          {error && (
            <Alert variant="destructive">
              <AlertTitle>出错了</AlertTitle>
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}
          <FieldGroup>
            <Field
              data-invalid={fieldErrors.name ? true : undefined}
              data-disabled={submitting ? true : undefined}
            >
              <FieldLabel htmlFor="play-source-name">
                名称
                <RequiredMark />
              </FieldLabel>
              <Input
                id="play-source-name"
                placeholder="请输入播放源名称"
                value={form.name}
                onChange={(e) => updateForm({ name: e.target.value })}
                maxLength={100}
                required
                disabled={submitting}
                aria-invalid={fieldErrors.name ? true : undefined}
              />
              {fieldErrors.name && <FieldError>{fieldErrors.name}</FieldError>}
            </Field>
            <Field
              data-invalid={fieldErrors.sort_order ? true : undefined}
              data-disabled={submitting ? true : undefined}
            >
              <FieldLabel htmlFor="play-source-sort-order">排序</FieldLabel>
              <Input
                id="play-source-sort-order"
                type="number"
                min="0"
                max="4294967295"
                step="1"
                placeholder="请输入排序值"
                value={form.sort_order}
                onChange={(e) => updateForm({ sort_order: e.target.value })}
                disabled={submitting}
                aria-invalid={fieldErrors.sort_order ? true : undefined}
              />
              {fieldErrors.sort_order && <FieldError>{fieldErrors.sort_order}</FieldError>}
            </Field>
            <Field
              data-invalid={fieldErrors.status ? true : undefined}
              data-disabled={submitting ? true : undefined}
            >
              <FieldLabel htmlFor="play-source-status">
                状态
                <RequiredMark />
              </FieldLabel>
              <Select
                items={statusOptions}
                value={form.status || undefined}
                onValueChange={(v) => updateForm({ status: v ?? '' })}
                disabled={submitting}
              >
                <SelectTrigger
                  id="play-source-status"
                  className="w-full"
                  aria-invalid={fieldErrors.status ? true : undefined}
                >
                  <SelectValue placeholder="请选择状态" />
                </SelectTrigger>
                <SelectContent>
                  {statusOptions.map((option) => (
                    <SelectItem key={option.value} value={option.value}>
                      {option.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              {fieldErrors.status && <FieldError>{fieldErrors.status}</FieldError>}
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
