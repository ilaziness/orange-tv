import { useEffect, useState } from 'react'
import type * as React from 'react'
import { adminApi, errorMessage } from '@/lib/api'
import { PageContainer } from '@/components/shared'
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Field, FieldGroup, FieldLabel } from '@/components/ui/field'
import { Checkbox } from '@/components/ui/checkbox'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Save } from 'lucide-react'
import { toast } from 'sonner'

export default function APISettingsPage() {
  const [form, setForm] = useState({
    site_mode: 'video_site',
    api_output_format: 'default',
    enable_third_party_collect: true,
    resource_api_key: '',
  })
  const [keySet, setKeySet] = useState(false)
  const [error, setError] = useState('')

  async function load() {
    setError('')
    try {
      const res = await adminApi.getSettings()
      const api = res.data.api
      setForm({
        site_mode: api.site_mode || 'video_site',
        api_output_format: api.api_output_format || 'default',
        enable_third_party_collect: !!api.enable_third_party_collect,
        resource_api_key: '',
      })
      setKeySet(!!api.resource_api_key_set)
    } catch (err) {
      setError(errorMessage(err))
    }
  }

  useEffect(() => { void load() }, [])

  async function save(e: React.SyntheticEvent<HTMLFormElement>) {
    e.preventDefault()
    setError('')
    try {
      await adminApi.updateSettings({
        api: {
          site_mode: form.site_mode,
          api_output_format: form.api_output_format,
          enable_third_party_collect: form.enable_third_party_collect,
          ...(form.resource_api_key.trim() ? { resource_api_key: form.resource_api_key.trim() } : {}),
        },
      })
      toast.success('API 配置已保存')
      setForm((f) => ({ ...f, resource_api_key: '' }))
      await load()
    } catch (err) {
      setError(errorMessage(err))
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
          {error && (
            <Alert variant="destructive" className="mb-4">
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}
          <form onSubmit={save} className="flex flex-col gap-4">
            <FieldGroup>
              <Field>
                <FieldLabel>站点模式</FieldLabel>
                <Select value={form.site_mode} onValueChange={(v) => setForm({ ...form, site_mode: v ?? 'video_site' })}>
                  <SelectTrigger className="w-48">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="video_site">影视站</SelectItem>
                    <SelectItem value="resource_site">资源站</SelectItem>
                  </SelectContent>
                </Select>
              </Field>
              <Field>
                <FieldLabel>API 输出格式</FieldLabel>
                <Select value={form.api_output_format} onValueChange={(v) => setForm({ ...form, api_output_format: v ?? 'default' })}>
                  <SelectTrigger className="w-48">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="default">系统默认格式</SelectItem>
                    <SelectItem value="apple_cms">苹果 CMS</SelectItem>
                  </SelectContent>
                </Select>
              </Field>
            </FieldGroup>
            <label className="flex items-center gap-2 text-sm">
              <Checkbox
                checked={form.enable_third_party_collect}
                onCheckedChange={(checked) => setForm({ ...form, enable_third_party_collect: checked === true })}
              />
              允许第三方采集
            </label>
            <FieldGroup>
              <Field>
                <FieldLabel htmlFor="resource_api_key">
                  资源站 API 密钥{' '}
                  <span className="text-xs text-muted-foreground">
                    （{keySet ? '已配置' : '未配置'}）
                  </span>
                </FieldLabel>
                <Input
                  id="resource_api_key"
                  type="password"
                  placeholder={keySet ? '****** 留空不修改' : '设置密钥'}
                  value={form.resource_api_key}
                  onChange={(e) => setForm({ ...form, resource_api_key: e.target.value })}
                />
              </Field>
            </FieldGroup>
            <div className="flex justify-end">
              <Button type="submit">
                <Save data-icon="inline-start" />
                保存
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>
    </PageContainer>
  )
}
