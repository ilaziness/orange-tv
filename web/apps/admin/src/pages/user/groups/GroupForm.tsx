import type * as React from 'react'
import type { GroupForm } from './useGroups'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'

interface GroupFormProps {
  form: GroupForm
  setForm: React.Dispatch<React.SetStateAction<GroupForm>>
  onSubmit: (e: React.SyntheticEvent<HTMLFormElement>) => void
}

export function GroupForm({ form, setForm, onSubmit }: GroupFormProps) {
  return (
    <form onSubmit={onSubmit} className="mb-4 flex flex-wrap items-end gap-2 rounded-lg border p-4">
      <Input
        placeholder="名称"
        value={form.name}
        onChange={(e) => setForm((prev) => ({ ...prev, name: e.target.value }))}
        required
        className="max-w-xs"
      />
      <Input
        placeholder="权限（JSON）"
        value={form.permissions}
        onChange={(e) => setForm((prev) => ({ ...prev, permissions: e.target.value }))}
        className="max-w-xs"
      />
      <Input
        placeholder="描述"
        value={form.description}
        onChange={(e) => setForm((prev) => ({ ...prev, description: e.target.value }))}
        className="max-w-xs"
      />
      <Button type="submit" size="sm">保存</Button>
    </form>
  )
}
