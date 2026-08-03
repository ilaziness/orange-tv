import { useEffect, useState, type HTMLInputTypeAttribute } from 'react'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Field, FieldLabel } from '@/components/ui/field'
import { Spinner } from '@/components/ui/spinner'

interface PromptDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  title?: string
  description?: string
  label?: string
  placeholder?: string
  confirmText?: string
  cancelText?: string
  type?: HTMLInputTypeAttribute
  loading?: boolean
  onConfirm: (value: string) => void
}

export function PromptDialog({
  open,
  onOpenChange,
  title = '请输入',
  description,
  label,
  placeholder,
  confirmText = '确认',
  cancelText = '取消',
  type = 'text',
  loading = false,
  onConfirm,
}: PromptDialogProps) {
  const [value, setValue] = useState('')

  useEffect(() => {
    if (open) setValue('')
  }, [open])

  function handleConfirm() {
    if (loading) return
    onConfirm(value)
  }

  return (
    <Dialog
      open={open}
      onOpenChange={(v) => {
        if (!loading) onOpenChange(v)
      }}
    >
      <DialogContent className="sm:max-w-md" showCloseButton={!loading}>
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
          {description && <DialogDescription>{description}</DialogDescription>}
        </DialogHeader>
        <Field data-disabled={loading ? true : undefined}>
          {label && <FieldLabel htmlFor="prompt-input">{label}</FieldLabel>}
          <Input
            id="prompt-input"
            type={type}
            value={value}
            placeholder={placeholder}
            onChange={(e) => setValue(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter') handleConfirm()
            }}
            disabled={loading}
            autoFocus
          />
        </Field>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)} disabled={loading}>
            {cancelText}
          </Button>
          <Button onClick={handleConfirm} disabled={loading}>
            {loading && <Spinner data-icon="inline-start" />}
            {loading ? '处理中...' : confirmText}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
