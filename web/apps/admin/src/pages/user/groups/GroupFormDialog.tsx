import type * as React from 'react'
import type { GroupForm } from './useGroups'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Field, FieldGroup, FieldLabel } from '@/components/ui/field'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Spinner } from '@/components/ui/spinner'

interface GroupFormDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  editId: number
  form: GroupForm
  setForm: React.Dispatch<React.SetStateAction<GroupForm>>
  submitting: boolean
  onSubmit: (e: React.SyntheticEvent<HTMLFormElement>) => void
}

export function GroupFormDialog({
  open,
  onOpenChange,
  editId,
  form,
  setForm,
  submitting,
  onSubmit,
}: GroupFormDialogProps) {
  return (
    <Dialog open={open} onOpenChange={(v) => { if (!submitting) onOpenChange(v) }}>
      <DialogContent className="sm:max-w-md" showCloseButton={!submitting}>
        <DialogHeader>
          <DialogTitle>{editId > 0 ? '编辑用户组' : '新增用户组'}</DialogTitle>
          <DialogDescription className="sr-only">
            {editId > 0 ? '修改用户组信息' : '创建一个新的用户组'}
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={onSubmit} className="flex flex-col gap-5">
          <FieldGroup>
            <Field data-disabled={submitting ? true : undefined}>
              <FieldLabel htmlFor="name">名称</FieldLabel>
              <Input
                id="name"
                placeholder="请输入用户组名称"
                value={form.name}
                onChange={(e) => setForm((prev) => ({ ...prev, name: e.target.value }))}
                required
                disabled={submitting}
                autoFocus
              />
            </Field>
            <Field data-disabled={submitting ? true : undefined}>
              <FieldLabel htmlFor="permissions">权限（JSON）</FieldLabel>
              <Input
                id="permissions"
                placeholder='请输入权限 JSON，如 ["*"] 或留空'
                value={form.permissions}
                onChange={(e) => setForm((prev) => ({ ...prev, permissions: e.target.value }))}
                disabled={submitting}
              />
            </Field>
            <Field data-disabled={submitting ? true : undefined}>
              <FieldLabel htmlFor="description">描述</FieldLabel>
              <Textarea
                id="description"
                rows={3}
                placeholder="请输入用户组描述（可选）"
                value={form.description}
                onChange={(e) => setForm((prev) => ({ ...prev, description: e.target.value }))}
                disabled={submitting}
              />
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
  )
}
