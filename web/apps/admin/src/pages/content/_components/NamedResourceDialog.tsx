import type * as React from 'react'
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
import { Spinner } from '@/components/ui/spinner'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'

interface NamedResourceDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  title: string
  name: string
  onNameChange: (value: string) => void
  submitting: boolean
  error: string
  fieldError: string
  onSubmit: (e: React.SyntheticEvent<HTMLFormElement>) => void
}

export function NamedResourceDialog({
  open,
  onOpenChange,
  title,
  name,
  onNameChange,
  submitting,
  error,
  fieldError,
  onSubmit,
}: NamedResourceDialogProps) {
  return (
    <Dialog
      open={open}
      onOpenChange={(v) => {
        if (!submitting) onOpenChange(v)
      }}
    >
      <DialogContent className="sm:max-w-md" showCloseButton={!submitting}>
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
          <DialogDescription className="sr-only">输入名称后保存</DialogDescription>
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
              data-invalid={fieldError ? true : undefined}
              data-disabled={submitting ? true : undefined}
            >
              <FieldLabel htmlFor="named-resource-name">
                名称<span className="ml-0.5 text-destructive">*</span>
              </FieldLabel>
              <Input
                id="named-resource-name"
                placeholder="请输入名称"
                value={name}
                onChange={(e) => onNameChange(e.target.value)}
                maxLength={100}
                required
                disabled={submitting}
                aria-invalid={fieldError ? true : undefined}
                autoFocus
              />
              {fieldError && <FieldError>{fieldError}</FieldError>}
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
