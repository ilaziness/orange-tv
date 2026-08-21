import type * as React from 'react'
import type { UserCreateForm, UserEditForm } from './useUsers'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Field, FieldGroup, FieldLabel } from '@/components/ui/field'
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

interface UserFormDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  editId: number
  createForm: UserCreateForm
  setCreateForm: React.Dispatch<React.SetStateAction<UserCreateForm>>
  editForm: UserEditForm
  setEditForm: React.Dispatch<React.SetStateAction<UserEditForm>>
  submitting: boolean
  onSubmit: (e: React.SyntheticEvent<HTMLFormElement>) => void
}

export function UserFormDialog({
  open,
  onOpenChange,
  editId,
  createForm,
  setCreateForm,
  editForm,
  setEditForm,
  submitting,
  onSubmit,
}: UserFormDialogProps) {
  return (
    <Dialog
      open={open}
      onOpenChange={(v) => {
        if (!submitting) onOpenChange(v)
      }}
    >
      <DialogContent className="sm:max-w-md" showCloseButton={!submitting}>
        <DialogHeader>
          <DialogTitle>{editId > 0 ? '编辑用户' : '新增用户'}</DialogTitle>
          <DialogDescription className="sr-only">
            {editId > 0 ? '修改用户信息' : '创建一个新的用户账号'}
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={onSubmit} className="flex flex-col gap-5">
          <FieldGroup>
            {editId === 0 && (
              <>
                <Field data-disabled={submitting ? true : undefined}>
                  <FieldLabel htmlFor="email">邮箱</FieldLabel>
                  <Input
                    id="email"
                    type="email"
                    placeholder="请输入邮箱"
                    value={createForm.email}
                    onChange={(e) => setCreateForm((prev) => ({ ...prev, email: e.target.value }))}
                    required
                    maxLength={128}
                    disabled={submitting}
                    autoFocus
                  />
                </Field>
                <Field data-disabled={submitting ? true : undefined}>
                  <FieldLabel htmlFor="password">密码</FieldLabel>
                  <Input
                    id="password"
                    type="password"
                    placeholder="请输入密码（至少6位）"
                    value={createForm.password}
                    onChange={(e) =>
                      setCreateForm((prev) => ({ ...prev, password: e.target.value }))
                    }
                    required
                    minLength={6}
                    disabled={submitting}
                  />
                </Field>
                <Field data-disabled={submitting ? true : undefined}>
                  <FieldLabel htmlFor="nickname">昵称（可选）</FieldLabel>
                  <Input
                    id="nickname"
                    placeholder="3-15 位，不填则使用邮箱前缀"
                    value={createForm.nickname}
                    onChange={(e) =>
                      setCreateForm((prev) => ({ ...prev, nickname: e.target.value }))
                    }
                    minLength={3}
                    maxLength={15}
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
                    value={createForm.status}
                    onValueChange={(v) => setCreateForm((prev) => ({ ...prev, status: v ?? '1' }))}
                  >
                    <SelectTrigger disabled={submitting}>
                      <SelectValue placeholder="请选择状态" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="1">启用</SelectItem>
                      <SelectItem value="0">禁用</SelectItem>
                    </SelectContent>
                  </Select>
                </Field>
              </>
            )}
            {editId > 0 && (
              <>
                <Field data-disabled={submitting ? true : undefined}>
                  <FieldLabel htmlFor="edit-email">邮箱</FieldLabel>
                  <Input
                    id="edit-email"
                    type="email"
                    placeholder="请输入邮箱（可选）"
                    value={editForm.email}
                    onChange={(e) => setEditForm((prev) => ({ ...prev, email: e.target.value }))}
                    maxLength={128}
                    disabled={submitting}
                    autoFocus
                  />
                </Field>
                <Field data-disabled={submitting ? true : undefined}>
                  <FieldLabel htmlFor="edit-nickname">昵称（可选）</FieldLabel>
                  <Input
                    id="edit-nickname"
                    placeholder="3-15 位"
                    value={editForm.nickname}
                    onChange={(e) => setEditForm((prev) => ({ ...prev, nickname: e.target.value }))}
                    minLength={3}
                    maxLength={15}
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
                    value={editForm.status}
                    onValueChange={(v) => setEditForm((prev) => ({ ...prev, status: v ?? '1' }))}
                  >
                    <SelectTrigger disabled={submitting}>
                      <SelectValue placeholder="请选择状态" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="1">启用</SelectItem>
                      <SelectItem value="0">禁用</SelectItem>
                    </SelectContent>
                  </Select>
                </Field>
              </>
            )}
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
