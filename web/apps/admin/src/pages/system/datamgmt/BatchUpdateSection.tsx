import { useState } from 'react'
import { z } from 'zod'
import { adminApi, errorMessage } from '@/lib/api'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Field, FieldError, FieldGroup, FieldLabel } from '@/components/ui/field'
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import { Spinner } from '@/components/ui/spinner'
import { toast } from 'sonner'
import { Search, Play } from 'lucide-react'

const targetOptions = [
  { value: 'video_cover', label: '影视封面（videos.cover_image）' },
  { value: 'episode_url', label: '播放链接（play_episodes.play_url）' },
]

const previewSchema = z.object({
  target: z.string().min(1, '请选择目标字段'),
  oldValue: z.string().min(1, '查找字符串不能为空').max(2000, '查找字符串过长'),
  newValue: z.string().max(2000, '替换字符串过长').optional(),
})

const executeSchema = z.object({
  target: z.string().min(1, '请选择目标字段'),
  oldValue: z.string().min(1, '查找字符串不能为空').max(2000, '查找字符串过长'),
  newValue: z.string().min(1, '替换字符串不能为空').max(2000, '替换字符串过长'),
})

type FormState = {
  target: string
  oldValue: string
  newValue: string
}

type FormErrors = {
  target?: string
  oldValue?: string
  newValue?: string
}

export default function BatchUpdateSection() {
  const [form, setForm] = useState<FormState>({
    target: '',
    oldValue: '',
    newValue: '',
  })
  const [errors, setErrors] = useState<FormErrors>({})
  const [previewRows, setPreviewRows] = useState<number | null>(null)
  const [previewing, setPreviewing] = useState(false)
  const [executing, setExecuting] = useState(false)
  const [openConfirm, setOpenConfirm] = useState(false)

  function validate(schema: typeof previewSchema | typeof executeSchema): boolean {
    const result = schema.safeParse(form)
    if (result.success) {
      setErrors({})
      return true
    }
    const next: FormErrors = {}
    for (const issue of result.error.issues) {
      const path = issue.path[0] as keyof FormErrors
      if (!next[path]) {
        next[path] = issue.message
      }
    }
    setErrors(next)
    return false
  }

  function handlePreview() {
    if (!validate(previewSchema)) {
      return
    }
    setPreviewing(true)
    setPreviewRows(null)
    adminApi
      .batchUpdatePreview({
        target: form.target,
        old_value: form.oldValue,
      })
      .then((res) => {
        setPreviewRows(res.data.matched_rows)
      })
      .catch((err) => toast.error(errorMessage(err)))
      .finally(() => setPreviewing(false))
  }

  function handleExecute() {
    if (!validate(executeSchema)) {
      return
    }
    setExecuting(true)
    adminApi
      .batchUpdateExecute({
        target: form.target,
        old_value: form.oldValue,
        new_value: form.newValue,
      })
      .then((res) => {
        toast.success(`已成功更新 ${res.data.updated_rows} 条记录`)
        setPreviewRows(null)
        setForm({ target: '', oldValue: '', newValue: '' })
      })
      .catch((err) => toast.error(errorMessage(err)))
      .finally(() => {
        setExecuting(false)
        setOpenConfirm(false)
      })
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>批量更新</CardTitle>
        <CardDescription>
          对指定字段进行字符串替换式批量更新，目前仅开放影视封面与剧集播放链接。
        </CardDescription>
      </CardHeader>
      <CardContent>
        <FieldGroup>
          <Field data-invalid={!!errors.target}>
            <FieldLabel htmlFor="target">目标字段</FieldLabel>
            <Select
              value={form.target}
              onValueChange={(value) => {
                setForm((prev) => ({ ...prev, target: value ?? '' }))
                setErrors((prev) => ({ ...prev, target: undefined }))
              }}
            >
              <SelectTrigger id="target" className="w-full" aria-invalid={!!errors.target}>
                <SelectValue>
                  {form.target
                    ? (targetOptions.find((o) => o.value === form.target)?.label ??
                      '请选择目标字段')
                    : '请选择目标字段'}
                </SelectValue>
              </SelectTrigger>
              <SelectContent>
                <SelectGroup>
                  {targetOptions.map((opt) => (
                    <SelectItem key={opt.value} value={opt.value}>
                      {opt.label}
                    </SelectItem>
                  ))}
                </SelectGroup>
              </SelectContent>
            </Select>
            {errors.target && <FieldError>{errors.target}</FieldError>}
          </Field>
          <Field data-invalid={!!errors.oldValue}>
            <FieldLabel htmlFor="oldValue">查找字符串</FieldLabel>
            <Input
              id="oldValue"
              placeholder="例如：https://www.example.com"
              value={form.oldValue}
              onChange={(e) => {
                setForm((prev) => ({ ...prev, oldValue: e.target.value }))
                setErrors((prev) => ({ ...prev, oldValue: undefined }))
              }}
              aria-invalid={!!errors.oldValue}
            />
            {errors.oldValue && <FieldError>{errors.oldValue}</FieldError>}
          </Field>
          <Field data-invalid={!!errors.newValue}>
            <FieldLabel htmlFor="newValue">替换字符串</FieldLabel>
            <Input
              id="newValue"
              placeholder="例如：https://www.new-example.com（留空会替换为空字符串）"
              value={form.newValue}
              onChange={(e) => {
                setForm((prev) => ({ ...prev, newValue: e.target.value }))
                setErrors((prev) => ({ ...prev, newValue: undefined }))
              }}
              aria-invalid={!!errors.newValue}
            />
            {errors.newValue && <FieldError>{errors.newValue}</FieldError>}
          </Field>
        </FieldGroup>

        <div className="mt-4 flex flex-col gap-4">
          <div className="flex items-center gap-2">
            <Button onClick={handlePreview} disabled={previewing}>
              {previewing ? (
                <Spinner data-icon="inline-start" />
              ) : (
                <Search data-icon="inline-start" />
              )}
              {previewing ? '正在计算...' : '预览影响行数'}
            </Button>
            {previewRows !== null && (
              <span className="text-sm text-muted-foreground">
                预计影响 <strong>{previewRows}</strong> 条记录
              </span>
            )}
          </div>

          {previewRows !== null && previewRows > 0 && (
            <AlertDialog open={openConfirm} onOpenChange={setOpenConfirm}>
              <Button
                variant="default"
                onClick={() => {
                  if (validate(executeSchema)) {
                    setOpenConfirm(true)
                  }
                }}
              >
                <Play data-icon="inline-start" />
                确认执行更新
              </Button>
              <AlertDialogContent>
                <AlertDialogHeader>
                  <AlertDialogTitle>确认执行批量更新？</AlertDialogTitle>
                  <AlertDialogDescription>
                    该操作将更新 <strong>{previewRows}</strong> 条记录，把包含 「{form.oldValue}
                    」的字段值替换为「{form.newValue}」。此操作不可撤销，请确认。
                  </AlertDialogDescription>
                </AlertDialogHeader>
                <AlertDialogFooter>
                  <AlertDialogCancel>取消</AlertDialogCancel>
                  <AlertDialogAction onClick={handleExecute} disabled={executing}>
                    {executing && <Spinner data-icon="inline-start" />}
                    {executing ? '执行中...' : '确认执行'}
                  </AlertDialogAction>
                </AlertDialogFooter>
              </AlertDialogContent>
            </AlertDialog>
          )}
        </div>
      </CardContent>
    </Card>
  )
}
