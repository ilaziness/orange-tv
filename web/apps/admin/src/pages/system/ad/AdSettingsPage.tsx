import { useEffect, useState } from 'react'
import type * as React from 'react'
import { z } from 'zod'
import { adminApi, errorMessage } from '@/lib/api'
import { PageContainer } from '@/components/shared'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Field, FieldError, FieldGroup, FieldLabel } from '@/components/ui/field'
import { Checkbox } from '@/components/ui/checkbox'
import { Skeleton } from '@/components/ui/skeleton'
import { Spinner } from '@/components/ui/spinner'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Save } from 'lucide-react'
import { toast } from 'sonner'

const adSettingsSchema = z
  .object({
    enabled: z.boolean(),
    type: z.enum(['image', 'video', 'html']),
    url: z.string(),
    link: z.string(),
    duration: z.number().min(1).max(300),
    skipable: z.boolean(),
  })
  .refine((data) => !data.enabled || data.url.trim() !== '', {
    message: '启用广告时必须填写广告素材 URL',
    path: ['url'],
  })

export default function AdSettingsPage() {
  const [form, setForm] = useState({
    enabled: false,
    type: 'image',
    url: '',
    link: '',
    duration: 5,
    skipable: true,
  })
  const [loading, setLoading] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [fieldError, setFieldError] = useState('')

  async function load(opts?: { silent?: boolean }) {
    if (!opts?.silent) setLoading(true)
    try {
      const res = await adminApi.getAdSettings()
      const ad = res.data
      setForm({
        enabled: !!ad.enabled,
        type: ad.type || 'image',
        url: ad.url || '',
        link: ad.link || '',
        duration: ad.duration || 5,
        skipable: !!ad.skipable,
      })
    } catch (err) {
      toast.error(errorMessage(err))
    } finally {
      if (!opts?.silent) setLoading(false)
    }
  }

  useEffect(() => {
    void load()
  }, [])

  async function save(e: React.SyntheticEvent<HTMLFormElement>) {
    e.preventDefault()
    if (submitting) return
    setFieldError('')
    const result = adSettingsSchema.safeParse(form)
    if (!result.success) {
      const firstError = result.error.issues[0]
      const msg = firstError?.message || '配置校验失败'
      toast.error(msg)
      if (firstError?.path[0] === 'url') setFieldError(msg)
      return
    }
    setSubmitting(true)
    try {
      await adminApi.updateSettings({
        group: 'ad',
        data: {
          enabled: result.data.enabled,
          type: result.data.type,
          url: result.data.url,
          link: result.data.link,
          duration: result.data.duration,
          skipable: result.data.skipable,
        },
      })
      toast.success('广告配置已保存')
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
          <CardTitle>视频广告配置</CardTitle>
          <CardDescription>
            配置视频播放前的 Loading 广告，支持图片、视频和 HTML
            三种类型。广告在视频加载阶段展示，播放就绪后自动移除。
          </CardDescription>
        </CardHeader>
        <CardContent>
          {loading ? (
            <div className="flex flex-col gap-4">
              {Array.from({ length: 6 }).map((_, i) => (
                <Skeleton key={i} className="h-10 w-full" />
              ))}
            </div>
          ) : (
            <form onSubmit={save} className="flex flex-col gap-4">
              <FieldGroup>
                <Field data-disabled={submitting ? true : undefined}>
                  <FieldLabel>启用广告</FieldLabel>
                  <label className="flex items-center gap-2 text-sm">
                    <Checkbox
                      checked={form.enabled}
                      onCheckedChange={(checked) =>
                        setForm((prev) => ({ ...prev, enabled: checked === true }))
                      }
                      disabled={submitting}
                    />
                    启用视频 Loading 广告
                  </label>
                </Field>
                <Field data-disabled={submitting ? true : undefined}>
                  <FieldLabel>广告类型</FieldLabel>
                  <Select
                    items={[
                      { value: 'image', label: '图片广告' },
                      { value: 'video', label: '视频广告' },
                      { value: 'html', label: 'HTML 广告' },
                    ]}
                    value={form.type}
                    onValueChange={(v) => setForm((prev) => ({ ...prev, type: v ?? 'image' }))}
                  >
                    <SelectTrigger className="w-48" disabled={submitting}>
                      <SelectValue placeholder="请选择广告类型" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="image">图片广告</SelectItem>
                      <SelectItem value="video">视频广告</SelectItem>
                      <SelectItem value="html">HTML 广告</SelectItem>
                    </SelectContent>
                  </Select>
                </Field>
                <Field
                  data-disabled={submitting ? true : undefined}
                  data-invalid={fieldError ? true : undefined}
                >
                  <FieldLabel htmlFor="ad_url">广告素材 URL</FieldLabel>
                  <Input
                    id="ad_url"
                    placeholder="请输入广告素材地址（图片/视频/HTML 页面 URL）"
                    value={form.url}
                    onChange={(e) => {
                      setForm((prev) => ({ ...prev, url: e.target.value }))
                      setFieldError('')
                    }}
                    aria-invalid={!!fieldError}
                    disabled={submitting}
                  />
                  {fieldError && <FieldError>{fieldError}</FieldError>}
                </Field>
                <Field data-disabled={submitting ? true : undefined}>
                  <FieldLabel htmlFor="ad_link">点击跳转链接</FieldLabel>
                  <Input
                    id="ad_link"
                    placeholder="点击广告跳转的链接（可选）"
                    value={form.link}
                    onChange={(e) => setForm((prev) => ({ ...prev, link: e.target.value }))}
                    disabled={submitting}
                  />
                </Field>
                <Field data-disabled={submitting ? true : undefined}>
                  <FieldLabel htmlFor="ad_duration">展示时长（秒）</FieldLabel>
                  <Input
                    id="ad_duration"
                    type="number"
                    min={1}
                    max={300}
                    placeholder="广告展示时长，1-300 秒"
                    value={form.duration}
                    onChange={(e) =>
                      setForm((prev) => ({
                        ...prev,
                        duration: e.target.value === '' ? 0 : Number(e.target.value),
                      }))
                    }
                    disabled={submitting}
                  />
                </Field>
                <Field data-disabled={submitting ? true : undefined}>
                  <FieldLabel>允许跳过</FieldLabel>
                  <label className="flex items-center gap-2 text-sm">
                    <Checkbox
                      checked={form.skipable}
                      onCheckedChange={(checked) =>
                        setForm((prev) => ({ ...prev, skipable: checked === true }))
                      }
                      disabled={submitting}
                    />
                    显示「跳过广告」按钮
                  </label>
                </Field>
              </FieldGroup>
              <div className="flex justify-end">
                <Button type="submit" disabled={submitting}>
                  {submitting ? (
                    <Spinner data-icon="inline-start" />
                  ) : (
                    <Save data-icon="inline-start" />
                  )}
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
