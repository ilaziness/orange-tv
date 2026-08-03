import { useEffect, useState } from 'react'
import type * as React from 'react'
import { z } from 'zod'
import { adminApi, errorMessage } from '@/lib/api'
import { PageContainer } from '@/components/shared'
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import { Spinner } from '@/components/ui/spinner'
import { Field, FieldGroup, FieldLabel } from '@/components/ui/field'
import { Checkbox } from '@/components/ui/checkbox'
import { Save } from 'lucide-react'
import { toast } from 'sonner'

const apiSettingsSchema = z.object({
  enable_third_party_collect: z.boolean(),
})

export default function APISettingsPage() {
  const [form, setForm] = useState({
    enable_third_party_collect: true,
  })
  const [loading, setLoading] = useState(false)
  const [submitting, setSubmitting] = useState(false)

  async function load(opts?: { silent?: boolean }) {
    if (!opts?.silent) setLoading(true)
    try {
      const res = await adminApi.getAPISettings()
      const api = res.data
      setForm({
        enable_third_party_collect: !!api.enable_third_party_collect,
      })
    } catch (err) {
      toast.error(errorMessage(err))
    } finally {
      if (!opts?.silent) setLoading(false)
    }
  }

  useEffect(() => { void load() }, [])

  async function save(e: React.SyntheticEvent<HTMLFormElement>) {
    e.preventDefault()
    if (submitting) return
    const result = apiSettingsSchema.safeParse(form)
    if (!result.success) {
      toast.error(result.error.issues[0]?.message || '配置校验失败')
      return
    }
    setSubmitting(true)
    try {
      await adminApi.updateSettings({
        group: 'api',
        data: {
          enable_third_party_collect: result.data.enable_third_party_collect,
        },
      })
      toast.success('API 配置已保存')
      await load({ silent: true })
    } catch (err) {
      toast.error(errorMessage(err))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <PageContainer>
      <Card>
        <CardHeader>
          <CardTitle>API 配置</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="mb-4 text-sm text-muted-foreground">
            资源站开放接口：/api/open/v1/*。开启后允许第三方采集本站公开数据。
          </p>
          {loading ? (
            <div className="flex flex-col gap-4">
              <Skeleton className="h-10 w-full" />
            </div>
          ) : (
            <form onSubmit={save} className="flex flex-col gap-4">
              <FieldGroup>
                <Field data-disabled={submitting ? true : undefined}>
                  <FieldLabel htmlFor="enable_third_party_collect">第三方采集</FieldLabel>
                  <label htmlFor="enable_third_party_collect" className="flex items-center gap-2 text-sm">
                    <Checkbox
                      id="enable_third_party_collect"
                      checked={form.enable_third_party_collect}
                      onCheckedChange={(checked) =>
                        setForm((prev) => ({ ...prev, enable_third_party_collect: checked === true }))
                      }
                      disabled={submitting}
                    />
                    允许第三方采集
                  </label>
                </Field>
              </FieldGroup>
              <div className="flex justify-end">
                <Button type="submit" disabled={submitting}>
                  {submitting ? <Spinner data-icon="inline-start" /> : <Save data-icon="inline-start" />}
                  {submitting ? '保存中...' : '保存'}
                </Button>
              </div>
            </form>
          )}
        </CardContent>
      </Card>
    </PageContainer>
  )
}
