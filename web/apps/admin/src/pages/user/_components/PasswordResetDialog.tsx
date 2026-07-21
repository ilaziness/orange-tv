import { PromptDialog } from '@/components/shared'
import { z } from 'zod'
import { toast } from 'sonner'

const pwdSchema = z.string().min(6, '密码至少 6 位')

interface PasswordResetDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onConfirm: (password: string) => void
}

export function PasswordResetDialog({ open, onOpenChange, onConfirm }: PasswordResetDialogProps) {
  function handleConfirm(pwd: string) {
    const result = pwdSchema.safeParse(pwd)
    if (!result.success) {
      toast.error(result.error.issues[0]?.message || '密码至少 6 位')
      return
    }
    onConfirm(result.data)
  }

  return (
    <PromptDialog
      open={open}
      onOpenChange={onOpenChange}
      title="重置密码"
      description="输入新密码（至少 6 位）"
      label="新密码"
      placeholder="请输入新密码"
      confirmText="重置"
      onConfirm={handleConfirm}
    />
  )
}
