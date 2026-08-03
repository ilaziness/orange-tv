import type * as React from 'react'
import type { CategoryForm } from './useCategories'

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

type Option = { value: string; label: string }

interface CategoryDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  title: string
  form: CategoryForm
  setForm: React.Dispatch<React.SetStateAction<CategoryForm>>
  parentOptions: Option[]
  statusOptions: Option[]
  submitting: boolean
  error: string
  onSubmit: (e: React.SyntheticEvent<HTMLFormElement>) => void
}

export function CategoryDialog({
  open,
  onOpenChange,
  title,
  form,
  setForm,
  parentOptions,
  statusOptions,
  submitting,
  error,
  onSubmit,
}: CategoryDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md" showCloseButton={!submitting}>
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
              <FieldLabel htmlFor="category-name">分类名称</FieldLabel>
              <Input
                id="category-name"
                value={form.name}
                onChange={(e) => setForm((prev) => ({ ...prev, name: e.target.value }))}
                maxLength={100}
                required
                disabled={submitting}
              />
            </Field>
            <Field>
              <FieldLabel htmlFor="category-parent">父级分类</FieldLabel>
              <Select
                items={parentOptions}
                value={form.parentId}
                onValueChange={(value) => setForm((prev) => ({ ...prev, parentId: value ?? '0' }))}
                disabled={submitting}
              >
                <SelectTrigger id="category-parent">
                  <SelectValue placeholder="请选择父级分类" />
                </SelectTrigger>
                <SelectContent>
                  {parentOptions.map((option) => (
                    <SelectItem key={option.value} value={option.value}>
                      {option.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </Field>
            <Field>
              <FieldLabel htmlFor="category-sort-order">排序</FieldLabel>
              <Input
                id="category-sort-order"
                type="number"
                min="0"
                max="4294967295"
                step="1"
                value={form.sortOrder}
                onChange={(e) => setForm((prev) => ({ ...prev, sortOrder: e.target.value }))}
                required
                disabled={submitting}
              />
            </Field>
            <Field>
              <FieldLabel htmlFor="category-status">状态</FieldLabel>
              <Select
                items={statusOptions}
                value={form.status}
                onValueChange={(value) => setForm((prev) => ({ ...prev, status: value ?? '1' }))}
                disabled={submitting}
              >
                <SelectTrigger id="category-status">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {statusOptions.map((option) => (
                    <SelectItem key={option.value} value={option.value}>
                      {option.label}
                    </SelectItem>
                  ))}
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
