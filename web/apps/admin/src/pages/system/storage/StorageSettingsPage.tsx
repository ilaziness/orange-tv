import { useEffect, useState } from 'react'
import type * as React from 'react'
import type { StorageProvider, StorageProviderConfig, StorageSettings } from '@orange-tv/shared'
import { z } from 'zod'
import { adminApi, errorMessage } from '@/lib/api'
import { PageContainer } from '@/components/shared'
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Field, FieldDescription, FieldGroup, FieldLabel } from '@/components/ui/field'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Skeleton } from '@/components/ui/skeleton'
import { Spinner } from '@/components/ui/spinner'
import { PlugZap, Save } from 'lucide-react'
import { toast } from 'sonner'

const emptyProvider = (): StorageProviderConfig => ({
  bucket: '',
  region: '',
  endpoint: '',
  access_key: '',
  secret_key: '',
  access_configured: false,
  secret_configured: false,
  cdn_domain: '',
})

const providerSchema = z.object({
  bucket: z.string(),
  region: z.string(),
  endpoint: z.string(),
  access_key: z.string(),
  secret_key: z.string(),
  access_configured: z.boolean(),
  secret_configured: z.boolean(),
  cdn_domain: z.string().max(256, '加速域名过长'),
})

const storageSchema = z.object({
  provider: z.enum(['none', 'aliyun', 'tencent', 'qiniu']),
  aliyun: providerSchema,
  tencent: providerSchema,
  qiniu: providerSchema,
})

type FormState = z.infer<typeof storageSchema>
type VendorTab = 'aliyun' | 'tencent' | 'qiniu'

const VENDOR_TABS: {
  key: VendorTab
  label: string
  hint: string
}[] = [
  { key: 'aliyun', label: '阿里云 OSS', hint: 'Region 示例：oss-cn-hangzhou（或填写 Endpoint）' },
  {
    key: 'tencent',
    label: '腾讯云 COS',
    hint: 'Region 示例：ap-guangzhou；Bucket 通常含 APPID 后缀',
  },
  { key: 'qiniu', label: '七牛云 Kodo', hint: 'Region 示例：z0 / z1 / z2（可选，可自动识别）' },
]

function hasProviderInput(p: StorageProviderConfig): boolean {
  return !!(
    p.bucket.trim() ||
    p.region.trim() ||
    p.endpoint.trim() ||
    p.access_key.trim() ||
    p.secret_key.trim() ||
    p.cdn_domain.trim()
  )
}

function isValidCDNDomain(raw: string): boolean {
  const v = raw.trim()
  if (!v || v.length > 256) return false
  try {
    const u = new URL(v)
    if (u.protocol !== 'https:') return false
    if (!u.hostname) return false
    if (u.username || u.password) return false
    const path = u.pathname.replace(/\/+$/, '')
    if (path) return false
    if (u.search || u.hash) return false
    return true
  } catch {
    return false
  }
}

function validateProvider(
  label: string,
  vendor: VendorTab,
  p: StorageProviderConfig,
  requireComplete: boolean,
): string | null {
  if (!requireComplete && !hasProviderInput(p) && !p.access_configured && !p.secret_configured) {
    return null
  }
  const needsFields = requireComplete || hasProviderInput(p) || p.access_configured || p.secret_configured
  if (!p.cdn_domain.trim() && needsFields) {
    return `${label}：加速域名不能为空`
  }
  if (p.cdn_domain.trim() && !isValidCDNDomain(p.cdn_domain)) {
    return `${label}：加速域名须为 https 主机名（无路径/查询参数）`
  }
  if (needsFields) {
    if (!p.bucket.trim()) return `${label}：Bucket 不能为空`
    if (!p.access_key.trim() && !p.access_configured) return `${label}：AccessKey 不能为空`
    if (!p.secret_key.trim() && !p.secret_configured) return `${label}：SecretKey 不能为空`
    if (vendor === 'aliyun' && !p.region.trim() && !p.endpoint.trim()) {
      return `${label}：Region 或 Endpoint 至少填写一项`
    }
    if (vendor === 'tencent' && !p.region.trim()) {
      return `${label}：Region 不能为空`
    }
  }
  return null
}

function ProviderFields({
  idPrefix,
  hint,
  value,
  onChange,
  disabled,
}: {
  idPrefix: string
  hint: string
  value: StorageProviderConfig
  onChange: (next: StorageProviderConfig) => void
  disabled: boolean
}) {
  return (
    <FieldGroup>
      <FieldDescription>{hint}</FieldDescription>
      <Field data-disabled={disabled ? true : undefined}>
        <FieldLabel htmlFor={`${idPrefix}-bucket`}>Bucket / 空间名</FieldLabel>
        <Input
          id={`${idPrefix}-bucket`}
          value={value.bucket}
          onChange={(e) => onChange({ ...value, bucket: e.target.value })}
          disabled={disabled}
          placeholder="bucket-name"
        />
      </Field>
      <Field data-disabled={disabled ? true : undefined}>
        <FieldLabel htmlFor={`${idPrefix}-region`}>
          区域
          {idPrefix === 'tencent' ? <span className="ml-0.5 text-destructive">*</span> : null}
          {idPrefix === 'aliyun' ? (
            <span className="ml-0.5 text-muted-foreground">（或填 Endpoint）</span>
          ) : null}
        </FieldLabel>
        <Input
          id={`${idPrefix}-region`}
          value={value.region}
          onChange={(e) => onChange({ ...value, region: e.target.value })}
          disabled={disabled}
          placeholder="如 oss-cn-hangzhou / ap-guangzhou / z0"
        />
      </Field>
      <Field data-disabled={disabled ? true : undefined}>
        <FieldLabel htmlFor={`${idPrefix}-endpoint`}>Endpoint（可选）</FieldLabel>
        <Input
          id={`${idPrefix}-endpoint`}
          value={value.endpoint}
          onChange={(e) => onChange({ ...value, endpoint: e.target.value })}
          disabled={disabled}
          placeholder="自定义 API 地址，一般可留空"
        />
      </Field>
      <Field data-disabled={disabled ? true : undefined}>
        <FieldLabel htmlFor={`${idPrefix}-access-key`}>AccessKey / SecretId</FieldLabel>
        <Input
          id={`${idPrefix}-access-key`}
          value={value.access_key}
          onChange={(e) => onChange({ ...value, access_key: e.target.value })}
          disabled={disabled}
          placeholder={
            value.access_configured || value.access_key
              ? '已配置（留空表示不修改）'
              : '请输入 AccessKey'
          }
          autoComplete="off"
        />
      </Field>
      <Field data-disabled={disabled ? true : undefined}>
        <FieldLabel htmlFor={`${idPrefix}-secret-key`}>SecretKey</FieldLabel>
        <Input
          id={`${idPrefix}-secret-key`}
          type="password"
          value={value.secret_key}
          onChange={(e) => onChange({ ...value, secret_key: e.target.value })}
          disabled={disabled}
          placeholder={value.secret_configured ? '已配置（留空表示不修改）' : '请输入 SecretKey'}
          autoComplete="new-password"
        />
        <FieldDescription>
          {value.secret_configured ? '密钥已保存，留空则保持不变' : '密钥仅保存在服务端'}
        </FieldDescription>
      </Field>
      <Field data-disabled={disabled ? true : undefined}>
        <FieldLabel htmlFor={`${idPrefix}-cdn-domain`}>
          加速域名<span className="ml-0.5 text-destructive">*</span>
        </FieldLabel>
        <Input
          id={`${idPrefix}-cdn-domain`}
          value={value.cdn_domain}
          onChange={(e) => onChange({ ...value, cdn_domain: e.target.value })}
          disabled={disabled}
          placeholder="https://cdn.example.com"
        />
        <FieldDescription>必须为 https，无尾斜杠，最长 256；上传返回地址使用此域名</FieldDescription>
      </Field>
    </FieldGroup>
  )
}

function normalizeProvider(raw?: StorageProviderConfig): StorageProviderConfig {
  return {
    ...emptyProvider(),
    ...raw,
    access_key: '',
    secret_key: '',
    access_configured: !!raw?.access_configured,
    secret_configured: !!raw?.secret_configured,
  }
}

function snapshotForm(state: FormState): string {
  return JSON.stringify(state)
}

export default function StorageSettingsPage() {
  const [form, setForm] = useState<FormState>({
    provider: 'none',
    aliyun: emptyProvider(),
    tencent: emptyProvider(),
    qiniu: emptyProvider(),
  })
  const [savedSnapshot, setSavedSnapshot] = useState('')
  const [activeTab, setActiveTab] = useState<VendorTab>('aliyun')
  const [loading, setLoading] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [pinging, setPinging] = useState(false)

  async function load(opts?: { silent?: boolean }) {
    if (!opts?.silent) setLoading(true)
    try {
      const res = await adminApi.getStorageSettings()
      const data = res.data as StorageSettings
      const provider = (data.provider || 'none') as StorageProvider
      const next: FormState = {
        provider,
        aliyun: normalizeProvider(data.aliyun),
        tencent: normalizeProvider(data.tencent),
        qiniu: normalizeProvider(data.qiniu),
      }
      setForm(next)
      setSavedSnapshot(snapshotForm(next))
      if (provider === 'aliyun' || provider === 'tencent' || provider === 'qiniu') {
        setActiveTab(provider)
      }
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
    const parsed = storageSchema.safeParse(form)
    if (!parsed.success) {
      toast.error(parsed.error.issues[0]?.message || '表单校验失败')
      return
    }
    const data = parsed.data
    for (const [label, vendor, cfg, enabled] of [
      ['阿里云', 'aliyun', data.aliyun, data.provider === 'aliyun'],
      ['腾讯云', 'tencent', data.tencent, data.provider === 'tencent'],
      ['七牛云', 'qiniu', data.qiniu, data.provider === 'qiniu'],
    ] as const) {
      const msg = validateProvider(label, vendor, cfg, enabled)
      if (msg) {
        setActiveTab(vendor)
        toast.error(msg)
        return
      }
    }
    setSubmitting(true)
    try {
      const strip = (p: StorageProviderConfig) => ({
        bucket: p.bucket,
        region: p.region,
        endpoint: p.endpoint,
        access_key: p.access_key,
        secret_key: p.secret_key,
        cdn_domain: p.cdn_domain,
      })
      await adminApi.updateSettings({
        group: 'storage',
        data: {
          provider: data.provider,
          aliyun: strip(data.aliyun),
          tencent: strip(data.tencent),
          qiniu: strip(data.qiniu),
        },
      })
      toast.success('云存储设置已保存')
      await load({ silent: true })
    } catch (err) {
      toast.error(errorMessage(err))
    } finally {
      setSubmitting(false)
    }
  }

  async function ping() {
    if (pinging || form.provider === 'none') return
    if (savedSnapshot && snapshotForm(form) !== savedSnapshot) {
      toast.error('请先保存后再测试连通性')
      return
    }
    setPinging(true)
    try {
      const res = await adminApi.pingStorage()
      toast.success(`连通性正常（${res.data.provider}）`)
    } catch (err) {
      toast.error(errorMessage(err))
    } finally {
      setPinging(false)
    }
  }

  function updateVendor(key: VendorTab, next: StorageProviderConfig) {
    setForm((prev) => ({ ...prev, [key]: next }))
  }

  return (
    <PageContainer>
      <form onSubmit={save} className="flex flex-col gap-4">
        <Card>
          <CardHeader>
            <CardTitle>云存储设置</CardTitle>
            <CardDescription>
              同一时间只能启用一家云存储。下方按厂商切换填写密钥与加速域名；上传返回地址一律使用加速域名。
            </CardDescription>
          </CardHeader>
          <CardContent className="flex flex-col gap-6">
            {loading ? (
              <div className="flex flex-col gap-4">
                <Skeleton className="h-10 w-full max-w-sm" />
                <Skeleton className="h-8 w-72" />
                <Skeleton className="h-48 w-full" />
              </div>
            ) : (
              <>
                <FieldGroup>
                  <Field data-disabled={submitting ? true : undefined}>
                    <FieldLabel htmlFor="storage-provider">启用厂商</FieldLabel>
                    <Select
                      items={[
                        { value: 'none', label: '不启用' },
                        { value: 'aliyun', label: '阿里云 OSS' },
                        { value: 'tencent', label: '腾讯云 COS' },
                        { value: 'qiniu', label: '七牛云 Kodo' },
                      ]}
                      value={form.provider}
                      onValueChange={(v) => {
                        const provider = ((v as StorageProvider) || 'none') as StorageProvider
                        setForm((prev) => ({ ...prev, provider }))
                        if (provider === 'aliyun' || provider === 'tencent' || provider === 'qiniu') {
                          setActiveTab(provider)
                        }
                      }}
                      disabled={submitting}
                    >
                      <SelectTrigger id="storage-provider" className="w-full max-w-sm cursor-pointer">
                        <SelectValue placeholder="选择厂商" />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="none">不启用</SelectItem>
                        <SelectItem value="aliyun">阿里云 OSS</SelectItem>
                        <SelectItem value="tencent">腾讯云 COS</SelectItem>
                        <SelectItem value="qiniu">七牛云 Kodo</SelectItem>
                      </SelectContent>
                    </Select>
                    <FieldDescription>
                      {form.provider === 'none'
                        ? '当前未启用云存储，媒体上传不可用'
                        : `当前启用：${VENDOR_TABS.find((t) => t.key === form.provider)?.label ?? form.provider}`}
                    </FieldDescription>
                  </Field>
                </FieldGroup>

                <Tabs
                  value={activeTab}
                  onValueChange={(v) => {
                    if (v === 'aliyun' || v === 'tencent' || v === 'qiniu') {
                      setActiveTab(v)
                    }
                  }}
                >
                  <TabsList>
                    {VENDOR_TABS.map((tab) => (
                      <TabsTrigger key={tab.key} value={tab.key} className="cursor-pointer">
                        {tab.label}
                        {form.provider === tab.key ? (
                          <Badge variant="secondary">启用</Badge>
                        ) : null}
                      </TabsTrigger>
                    ))}
                  </TabsList>
                  {VENDOR_TABS.map((tab) => (
                    <TabsContent key={tab.key} value={tab.key} className="mt-4">
                      <ProviderFields
                        idPrefix={tab.key}
                        hint={tab.hint}
                        value={form[tab.key]}
                        onChange={(next) => updateVendor(tab.key, next)}
                        disabled={submitting}
                      />
                    </TabsContent>
                  ))}
                </Tabs>
              </>
            )}
          </CardContent>
          <CardFooter className="justify-end gap-2">
            <Button
              type="button"
              variant="outline"
              disabled={loading || submitting || pinging || form.provider === 'none'}
              onClick={() => void ping()}
              className="cursor-pointer"
            >
              {pinging ? <Spinner data-icon="inline-start" /> : <PlugZap data-icon="inline-start" />}
              {pinging ? '测试中...' : '测试连通性'}
            </Button>
            <Button type="submit" disabled={loading || submitting} className="cursor-pointer">
              {submitting ? <Spinner data-icon="inline-start" /> : <Save data-icon="inline-start" />}
              {submitting ? '保存中...' : '保存'}
            </Button>
          </CardFooter>
        </Card>
      </form>
    </PageContainer>
  )
}
