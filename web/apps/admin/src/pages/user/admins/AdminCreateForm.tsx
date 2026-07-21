import type * as React from 'react'
import type { AdminForm } from './useAdmins'
import type { UserGroupItem } from '@orange-tv/shared'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'

interface AdminCreateFormProps {
  form: AdminForm
  setForm: React.Dispatch<React.SetStateAction<AdminForm>>
  groups: UserGroupItem[]
  onSubmit: (e: React.SyntheticEvent<HTMLFormElement>) => void
}

export function AdminCreateForm({ form, setForm, groups, onSubmit }: AdminCreateFormProps) {
  return (
    <form onSubmit={onSubmit} className="mb-4 flex flex-wrap items-end gap-2 rounded-lg border p-4">
      <Input
        placeholder="用户名"
        value={form.username}
        onChange={(e) => setForm((prev) => ({ ...prev, username: e.target.value }))}
        required
        minLength={3}
        className="max-w-xs"
      />
      <Input
        type="password"
        placeholder="密码"
        value={form.password}
        onChange={(e) => setForm((prev) => ({ ...prev, password: e.target.value }))}
        required
        minLength={6}
        className="max-w-xs"
      />
      <Input
        type="email"
        placeholder="邮箱"
        value={form.email}
        onChange={(e) => setForm((prev) => ({ ...prev, email: e.target.value }))}
        className="max-w-xs"
      />
      <Select
        items={groups.map((group) => ({ value: String(group.id), label: group.name }))}
        value={form.group_id}
        onValueChange={(v) => setForm((prev) => ({ ...prev, group_id: v ?? '1' }))}
      >
        <SelectTrigger className="w-32">
          <SelectValue placeholder="用户组" />
        </SelectTrigger>
        <SelectContent>
          {groups.map((g) => (
            <SelectItem key={g.id} value={String(g.id)}>{g.name}</SelectItem>
          ))}
        </SelectContent>
      </Select>
      <Select
        items={[{ value: '1', label: '启用' }, { value: '0', label: '禁用' }]}
        value={form.status}
        onValueChange={(v) => setForm((prev) => ({ ...prev, status: v ?? '1' }))}
      >
        <SelectTrigger className="w-28">
          <SelectValue />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="1">启用</SelectItem>
          <SelectItem value="0">禁用</SelectItem>
        </SelectContent>
      </Select>
      <Button type="submit" size="sm">保存</Button>
    </form>
  )
}
