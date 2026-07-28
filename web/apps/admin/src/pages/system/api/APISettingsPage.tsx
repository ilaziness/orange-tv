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
import { Input } from '@/components/ui/input'
import { Field, FieldGroup, FieldLabel } from '@/components/ui/field'
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

const apiSettingsSchema = z.object({
  site_mode: z.enum(['video_site', 'resource_site']),
  api_output_format: z.enum(['default', 'apple_cms']),
  enable_third_party_collect: z.boolean(),
  resource_api_key: z.string(),
})

export default function APISettingsPage() {
  const [form, setForm] = useState({
    site_mode: 'video_site',
    api_output_format: 'default',
    enable_third_party_collect: true,
    resource_api_key: '',
  })
  const [keySet, setKeySet] = useState(false)
  const [loading, setLoading] = useState(false)
  const [submitting, setSubmitting] = useState(false)

  async function load(opts?: { silent?: boolean }) {
    if (!opts?.silent) setLoading(true)
    try {
      const res = await adminApi.getAPISettings()
      const api = res.data
      setForm({
        site_mode: api.site_mode || 'video_site',
        api_output_format: api.api_output_format || 'default',
        enable_third_party_collect: !!api.enable_third_party_collect,
        resource_api_key: '',
      })
      setKeySet(!!api.resource_api_key_set)
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
          site_mode: result.data.site_mode,
          api_output_format: result.data.api_output_format,
          enable_third_party_collect: result.data.enable_third_party_collect,
          ...(result.data.resource_api_key.trim() ? { resource_api_key: result.data.resource_api_key.trim() } : {}),
        },
      })
      toast.success('API 配置已保存')
      setForm((f) => ({ ...f, resource_api_key: '' }))
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
            资源站开放接口：{`/api/open/v1/*`}。密钥通过 Header X-API-Key 或 query key 传递；密钥输入框留空表示不修改。
          </p>
          {loading ? (
            <div className="flex flex-col gap-4">
              {Array.from({ length: 5 }).map((_, i) => (
                <Skeleton key={i} className="h-10 w-full" />
              ))}
            </div>
          ) : (
            <form onSubmit={save} className="flex flex-col gap-4">
              <FieldGroup>
                <Field data-disabled={submitting ? true : undefined}>
                  <FieldLabel>站点模式</FieldLabel>
                  <Select items={[{ value: 'video_site', label: '影视站' }, { value: 'resource_site', label: '资源站' }]} value={form.site_mode} onValueChange={(v) => setForm((prev) => ({ ...prev, site_mode: v ?? 'video_site' }))}>
                    <SelectTrigger className="w-48" disabled={submitting}>
                      <SelectValue placeholder="请选择站点模式" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="video_site">影视站</SelectItem>
                      <SelectItem value="resource_site">资源站</SelectItem>
                    </SelectContent>
                  </Select>
                </Field>
                <Field data-disabled={submitting ? true : undefined}>
                  <FieldLabel>API 输出格式</FieldLabel>
                  <Select items={[{ value: 'default', label: '系统默认格式' }, { value: 'apple_cms', label: '苹果 CMS' }]} value={form.api_output_format} onValueChange={(v) => setForm((prev) => ({ ...prev, api_output_format: v ?? 'default' }))}>
                    <SelectTrigger className="w-48" disabled={submitting}>
                      <SelectValue placeholder="请选择输出格式" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="default">系统默认格式</SelectItem>
                      <SelectItem value="apple_cms">苹果 CMS</SelectItem>
                    </SelectContent>
                  </Select>
                </Field>
                <Field data-disabled={submitting ? true : undefined}>
                  <FieldLabel>第三方采集</FieldLabel>
                  <label className="flex items-center gap-2 text-sm">
                    <Checkbox
                      checked={form.enable_third_party_collect}
                      onCheckedChange={(checked) => setForm((prev) => ({ ...prev, enable_third_party_collect: checked === true }))}
                      disabled={submitting}
                    />
                    允许第三方采集
                  </label>
                </Field>
              </FieldGroup>
              <FieldGroup>
                <Field data-disabled={submitting ? true : undefined}>
                  <FieldLabel htmlFor="resource_api_key">
                    资源站 API 密钥{' '}
                    <span className="text-xs text-muted-foreground">
                      （{keySet ? '已配置' : '未配置'}）
                    </span>
                  </FieldLabel>
                  <Input
                    id="resource_api_key"
                    type="password"
                    placeholder={keySet ? '****** 留空不修改' : '请设置资源站 API 密钥'}
                    value={form.resource_api_key}
                    onChange={(e) => setForm((prev) => ({ ...prev, resource_api_key: e.target.value }))}
                    disabled={submitting}
                  />
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
