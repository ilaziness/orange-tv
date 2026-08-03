import type * as React from 'react'
import type { AdminCreateForm, AdminEditForm } from './useAdmins'
import type { UserGroupItem } from '@orange-tv/shared'
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

interface AdminFormDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  editId: number
  groups: UserGroupItem[]
  createForm: AdminCreateForm
  setCreateForm: React.Dispatch<React.SetStateAction<AdminCreateForm>>
  editForm: AdminEditForm
  setEditForm: React.Dispatch<React.SetStateAction<AdminEditForm>>
  submitting: boolean
  onSubmit: (e: React.SyntheticEvent<HTMLFormElement>) => void
}

export function AdminFormDialog({
  open,
  onOpenChange,
  editId,
  groups,
  createForm,
  setCreateForm,
  editForm,
  setEditForm,
  submitting,
  onSubmit,
}: AdminFormDialogProps) {
  return (
    <Dialog
      open={open}
      onOpenChange={(v) => {
        if (!submitting) onOpenChange(v)
      }}
    >
      <DialogContent className="sm:max-w-md" showCloseButton={!submitting}>
        <DialogHeader>
          <DialogTitle>{editId > 0 ? '编辑管理员' : '新增管理员'}</DialogTitle>
          <DialogDescription className="sr-only">
            {editId > 0 ? '修改管理员信息' : '创建一个新的管理员账号'}
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={onSubmit} className="flex flex-col gap-5">
          <FieldGroup>
            {editId === 0 && (
              <>
                <Field data-disabled={submitting ? true : undefined}>
                  <FieldLabel htmlFor="username">用户名</FieldLabel>
                  <Input
                    id="username"
                    placeholder="请输入用户名（3-50个字符）"
                    value={createForm.username}
                    onChange={(e) =>
                      setCreateForm((prev) => ({ ...prev, username: e.target.value }))
                    }
                    required
                    minLength={3}
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
                  <FieldLabel htmlFor="email">邮箱</FieldLabel>
                  <Input
                    id="email"
                    type="email"
                    placeholder="请输入邮箱（可选）"
                    value={createForm.email}
                    onChange={(e) => setCreateForm((prev) => ({ ...prev, email: e.target.value }))}
                    disabled={submitting}
                  />
                </Field>
                <Field data-disabled={submitting ? true : undefined}>
                  <FieldLabel>用户组</FieldLabel>
                  <Select
                    items={groups.map((g) => ({ value: String(g.id), label: g.name }))}
                    value={createForm.group_id}
                    onValueChange={(v) => setCreateForm((prev) => ({ ...prev, group_id: v ?? '' }))}
                  >
                    <SelectTrigger disabled={submitting}>
                      <SelectValue placeholder="请选择用户组" />
                    </SelectTrigger>
                    <SelectContent>
                      {groups.map((g) => (
                        <SelectItem key={g.id} value={String(g.id)}>
                          {g.name}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
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
                  <FieldLabel htmlFor="edit-username">用户名</FieldLabel>
                  <Input
                    id="edit-username"
                    placeholder="请输入用户名（3-50个字符）"
                    value={editForm.username}
                    onChange={(e) => setEditForm((prev) => ({ ...prev, username: e.target.value }))}
                    required
                    minLength={3}
                    disabled={submitting}
                    autoFocus
                  />
                </Field>
                <Field data-disabled={submitting ? true : undefined}>
                  <FieldLabel htmlFor="edit-email">邮箱</FieldLabel>
                  <Input
                    id="edit-email"
                    type="email"
                    placeholder="请输入邮箱（可选）"
                    value={editForm.email}
                    onChange={(e) => setEditForm((prev) => ({ ...prev, email: e.target.value }))}
                    disabled={submitting}
                  />
                </Field>
                <Field data-disabled={submitting ? true : undefined}>
                  <FieldLabel>用户组</FieldLabel>
                  <Select
                    items={groups.map((g) => ({ value: String(g.id), label: g.name }))}
                    value={editForm.group_id}
                    onValueChange={(v) => setEditForm((prev) => ({ ...prev, group_id: v ?? '' }))}
                  >
                    <SelectTrigger disabled={submitting}>
                      <SelectValue placeholder="请选择用户组" />
                    </SelectTrigger>
                    <SelectContent>
                      {groups.map((g) => (
                        <SelectItem key={g.id} value={String(g.id)}>
                          {g.name}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
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
