import { useEffect, useState } from 'react'
import type * as React from 'react'
import type { PaymentAlipayConfig, PaymentSettings, PaymentWechatConfig } from '@orange-tv/shared'
import { z } from 'zod'
import { adminApi, errorMessage } from '@/lib/api'
import { PageContainer, PasswordInput } from '@/components/shared'
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from '@/components/ui/accordion'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Field, FieldDescription, FieldGroup, FieldLabel } from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Skeleton } from '@/components/ui/skeleton'
import { Spinner } from '@/components/ui/spinner'
import { Switch } from '@/components/ui/switch'
import { Textarea } from '@/components/ui/textarea'
import { Save } from 'lucide-react'
import { toast } from 'sonner'

const emptyAlipay = (): PaymentAlipayConfig => ({
  enabled: false,
  sandbox: false,
  app_id: '',
  sign_mode: 'key',
  private_key: '',
  private_key_configured: false,
  alipay_public_key: '',
  alipay_public_key_configured: false,
  app_cert: '',
  app_cert_configured: false,
  alipay_public_cert: '',
  alipay_public_cert_configured: false,
  alipay_root_cert: '',
  alipay_root_cert_configured: false,
  notify_url: '',
  return_url: '',
  pc_web_enabled: false,
  app_enabled: false,
})

const emptyWechat = (): PaymentWechatConfig => ({
  enabled: false,
  mch_id: '',
  mch_serial_no: '',
  api_v3_key: '',
  api_v3_key_configured: false,
  private_key: '',
  private_key_configured: false,
  app_id_web: '',
  app_id_app: '',
  notify_url: '',
  pc_web_enabled: false,
  app_enabled: false,
})

const optionalURL = z.string().refine((v) => {
  const s = v.trim()
  if (!s) return true
  try {
    const u = new URL(s)
    return u.protocol === 'http:' || u.protocol === 'https:'
  } catch {
    return false
  }
}, '须为 http(s) URL')

const optionalHTTPS = z.string().refine((v) => {
  const s = v.trim()
  if (!s) return true
  try {
    const u = new URL(s)
    return u.protocol === 'https:'
  } catch {
    return false
  }
}, '须为 https URL')

const alipaySchema = z.object({
  enabled: z.boolean(),
  sandbox: z.boolean(),
  app_id: z.string().max(64, 'AppID 过长'),
  sign_mode: z.enum(['key', 'cert']),
  private_key: z.string(),
  private_key_configured: z.boolean(),
  alipay_public_key: z.string(),
  alipay_public_key_configured: z.boolean(),
  app_cert: z.string(),
  app_cert_configured: z.boolean(),
  alipay_public_cert: z.string(),
  alipay_public_cert_configured: z.boolean(),
  alipay_root_cert: z.string(),
  alipay_root_cert_configured: z.boolean(),
  notify_url: optionalURL,
  return_url: optionalURL,
  pc_web_enabled: z.boolean(),
  app_enabled: z.boolean(),
})

const wechatSchema = z.object({
  enabled: z.boolean(),
  mch_id: z.string().max(32, '商户号过长'),
  mch_serial_no: z.string().max(64, '证书序列号过长'),
  api_v3_key: z.string(),
  api_v3_key_configured: z.boolean(),
  private_key: z.string(),
  private_key_configured: z.boolean(),
  app_id_web: z.string().max(64, '公众号/网站 AppID 过长'),
  app_id_app: z.string().max(64, '开放平台 AppID 过长'),
  notify_url: optionalHTTPS,
  pc_web_enabled: z.boolean(),
  app_enabled: z.boolean(),
})

function applyAlipay(src: PaymentAlipayConfig): PaymentAlipayConfig {
  return {
    ...emptyAlipay(),
    ...src,
    sign_mode: src.sign_mode === 'cert' ? 'cert' : 'key',
    private_key: '',
    alipay_public_key: '',
    app_cert: '',
    alipay_public_cert: '',
    alipay_root_cert: '',
  }
}

function applyWechat(src: PaymentWechatConfig): PaymentWechatConfig {
  return {
    ...emptyWechat(),
    ...src,
    api_v3_key: '',
    private_key: '',
  }
}

export default function PaymentProvidersPage() {
  const [alipay, setAlipay] = useState<PaymentAlipayConfig>(emptyAlipay)
  const [wechat, setWechat] = useState<PaymentWechatConfig>(emptyWechat)
  const [loading, setLoading] = useState(true)
  const [savingAlipay, setSavingAlipay] = useState(false)
  const [savingWechat, setSavingWechat] = useState(false)

  async function load(opts?: { silent?: boolean }) {
    if (!opts?.silent) setLoading(true)
    try {
      const res = await adminApi.getPaymentSettings()
      const data: PaymentSettings = res.data
      setAlipay(applyAlipay(data.alipay || emptyAlipay()))
      setWechat(applyWechat(data.wechat || emptyWechat()))
    } catch (err) {
      toast.error(errorMessage(err))
    } finally {
      if (!opts?.silent) setLoading(false)
    }
  }

  useEffect(() => {
    void load()
  }, [])

  async function saveAlipay(e: React.SyntheticEvent<HTMLFormElement>) {
    e.preventDefault()
    if (savingAlipay) return
    const result = alipaySchema.safeParse(alipay)
    if (!result.success) {
      toast.error(result.error.issues[0]?.message || '表单校验失败')
      return
    }
    setSavingAlipay(true)
    try {
      const res = await adminApi.updateSettings<PaymentSettings>({
        group: 'payment',
        data: { alipay: result.data },
      })
      toast.success('支付宝配置已保存')
      if (res.data?.alipay) setAlipay(applyAlipay(res.data.alipay))
    } catch (err) {
      toast.error(errorMessage(err))
    } finally {
      setSavingAlipay(false)
    }
  }

  async function saveWechat(e: React.SyntheticEvent<HTMLFormElement>) {
    e.preventDefault()
    if (savingWechat) return
    const result = wechatSchema.safeParse(wechat)
    if (!result.success) {
      toast.error(result.error.issues[0]?.message || '表单校验失败')
      return
    }
    setSavingWechat(true)
    try {
      const res = await adminApi.updateSettings<PaymentSettings>({
        group: 'payment',
        data: { wechat: result.data },
      })
      toast.success('微信支付配置已保存')
      if (res.data?.wechat) setWechat(applyWechat(res.data.wechat))
    } catch (err) {
      toast.error(errorMessage(err))
    } finally {
      setSavingWechat(false)
    }
  }

  if (loading) {
    return (
      <PageContainer>
        <Card>
          <CardHeader>
            <CardTitle>支付商管理</CardTitle>
            <CardDescription>配置支付宝与微信支付商户参数</CardDescription>
          </CardHeader>
          <CardContent>
            <Skeleton className="h-48 w-full" />
          </CardContent>
        </Card>
      </PageContainer>
    )
  }

  return (
    <PageContainer>
      <Card>
        <CardHeader>
            <CardTitle>支付商管理</CardTitle>
          <CardDescription>
            按支付商分别配置并保存。密钥已配置时留空不覆盖。可同时启用多家。
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Accordion defaultValue={[]} multiple>
            <AccordionItem value="alipay">
              <AccordionTrigger>
                <span className="flex items-center gap-2">
                  支付宝
                  {alipay.enabled ? <Badge>已启用</Badge> : <Badge variant="secondary">未启用</Badge>}
                </span>
              </AccordionTrigger>
              <AccordionContent>
                <form onSubmit={saveAlipay}>
                  <FieldGroup>
                    <Field orientation="horizontal">
                      <Switch
                        id="alipay-enabled"
                        checked={alipay.enabled}
                        onCheckedChange={(checked) => setAlipay((s) => ({ ...s, enabled: checked }))}
                      />
                      <FieldLabel htmlFor="alipay-enabled">启用</FieldLabel>
                    </Field>
                    <Field orientation="horizontal">
                      <Switch
                        id="alipay-sandbox"
                        checked={alipay.sandbox}
                        onCheckedChange={(checked) => setAlipay((s) => ({ ...s, sandbox: checked }))}
                      />
                      <FieldLabel htmlFor="alipay-sandbox">沙箱环境</FieldLabel>
                    </Field>
                    <Field>
                      <FieldLabel htmlFor="alipay-app-id">AppID</FieldLabel>
                      <Input
                        id="alipay-app-id"
                        value={alipay.app_id}
                        onChange={(e) => setAlipay((s) => ({ ...s, app_id: e.target.value }))}
                        maxLength={64}
                      />
                    </Field>
                    <Field>
                      <FieldLabel htmlFor="alipay-sign-mode">签名模式</FieldLabel>
                      <Select
                        items={[
                          { value: 'key', label: '公钥' },
                          { value: 'cert', label: '证书' },
                        ]}
                        value={alipay.sign_mode}
                        onValueChange={(v) =>
                          setAlipay((s) => ({ ...s, sign_mode: v === 'cert' ? 'cert' : 'key' }))
                        }
                      >
                        <SelectTrigger id="alipay-sign-mode" className="w-full max-w-sm">
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectGroup>
                            <SelectItem value="key">公钥</SelectItem>
                            <SelectItem value="cert">证书</SelectItem>
                          </SelectGroup>
                        </SelectContent>
                      </Select>
                    </Field>
                    <Field>
                      <FieldLabel htmlFor="alipay-private-key">应用私钥</FieldLabel>
                      <FieldDescription>
                        {alipay.private_key_configured ? '已配置，留空则不修改' : 'PEM 格式'}
                      </FieldDescription>
                      <Textarea
                        id="alipay-private-key"
                        value={alipay.private_key}
                        onChange={(e) => setAlipay((s) => ({ ...s, private_key: e.target.value }))}
                        rows={4}
                      />
                    </Field>
                    {alipay.sign_mode === 'key' ? (
                      <Field>
                        <FieldLabel htmlFor="alipay-public-key">支付宝公钥</FieldLabel>
                        <FieldDescription>
                          {alipay.alipay_public_key_configured ? '已配置，留空则不修改' : ''}
                        </FieldDescription>
                        <Textarea
                          id="alipay-public-key"
                          value={alipay.alipay_public_key}
                          onChange={(e) =>
                            setAlipay((s) => ({ ...s, alipay_public_key: e.target.value }))
                          }
                          rows={4}
                        />
                      </Field>
                    ) : (
                      <>
                        <Field>
                          <FieldLabel htmlFor="alipay-app-cert">应用公钥证书</FieldLabel>
                          <Textarea
                            id="alipay-app-cert"
                            value={alipay.app_cert}
                            onChange={(e) => setAlipay((s) => ({ ...s, app_cert: e.target.value }))}
                            rows={3}
                            placeholder={alipay.app_cert_configured ? '已配置，留空则不修改' : ''}
                          />
                        </Field>
                        <Field>
                          <FieldLabel htmlFor="alipay-public-cert">支付宝公钥证书</FieldLabel>
                          <Textarea
                            id="alipay-public-cert"
                            value={alipay.alipay_public_cert}
                            onChange={(e) =>
                              setAlipay((s) => ({ ...s, alipay_public_cert: e.target.value }))
                            }
                            rows={3}
                            placeholder={
                              alipay.alipay_public_cert_configured ? '已配置，留空则不修改' : ''
                            }
                          />
                        </Field>
                        <Field>
                          <FieldLabel htmlFor="alipay-root-cert">支付宝根证书</FieldLabel>
                          <Textarea
                            id="alipay-root-cert"
                            value={alipay.alipay_root_cert}
                            onChange={(e) =>
                              setAlipay((s) => ({ ...s, alipay_root_cert: e.target.value }))
                            }
                            rows={3}
                            placeholder={
                              alipay.alipay_root_cert_configured ? '已配置，留空则不修改' : ''
                            }
                          />
                        </Field>
                      </>
                    )}
                    <Field>
                      <FieldLabel htmlFor="alipay-notify-url">异步通知 URL</FieldLabel>
                      <Input
                        id="alipay-notify-url"
                        value={alipay.notify_url}
                        onChange={(e) => setAlipay((s) => ({ ...s, notify_url: e.target.value }))}
                      />
                    </Field>
                    <Field>
                      <FieldLabel htmlFor="alipay-return-url">同步跳转 URL</FieldLabel>
                      <Input
                        id="alipay-return-url"
                        value={alipay.return_url}
                        onChange={(e) => setAlipay((s) => ({ ...s, return_url: e.target.value }))}
                      />
                    </Field>
                    <Field orientation="horizontal">
                      <Switch
                        id="alipay-pc-web"
                        checked={alipay.pc_web_enabled}
                        onCheckedChange={(checked) =>
                          setAlipay((s) => ({ ...s, pc_web_enabled: checked }))
                        }
                      />
                      <FieldLabel htmlFor="alipay-pc-web">PC 网页支付</FieldLabel>
                    </Field>
                    <Field orientation="horizontal">
                      <Switch
                        id="alipay-app"
                        checked={alipay.app_enabled}
                        onCheckedChange={(checked) =>
                          setAlipay((s) => ({ ...s, app_enabled: checked }))
                        }
                      />
                      <FieldLabel htmlFor="alipay-app">APP 支付</FieldLabel>
                    </Field>
                    <Button type="submit" disabled={savingAlipay}>
                      {savingAlipay ? (
                        <Spinner data-icon="inline-start" />
                      ) : (
                        <Save data-icon="inline-start" />
                      )}
                      保存支付宝
                    </Button>
                  </FieldGroup>
                </form>
              </AccordionContent>
            </AccordionItem>

            <AccordionItem value="wechat">
              <AccordionTrigger>
                <span className="flex items-center gap-2">
                  微信支付
                  {wechat.enabled ? <Badge>已启用</Badge> : <Badge variant="secondary">未启用</Badge>}
                </span>
              </AccordionTrigger>
              <AccordionContent>
                <form onSubmit={saveWechat}>
                  <FieldGroup>
                    <Field orientation="horizontal">
                      <Switch
                        id="wechat-enabled"
                        checked={wechat.enabled}
                        onCheckedChange={(checked) => setWechat((s) => ({ ...s, enabled: checked }))}
                      />
                      <FieldLabel htmlFor="wechat-enabled">启用</FieldLabel>
                    </Field>
                    <Field>
                      <FieldLabel htmlFor="wechat-mch-id">商户号</FieldLabel>
                      <Input
                        id="wechat-mch-id"
                        value={wechat.mch_id}
                        onChange={(e) => setWechat((s) => ({ ...s, mch_id: e.target.value }))}
                        maxLength={32}
                      />
                    </Field>
                    <Field>
                      <FieldLabel htmlFor="wechat-serial">商户证书序列号</FieldLabel>
                      <Input
                        id="wechat-serial"
                        value={wechat.mch_serial_no}
                        onChange={(e) => setWechat((s) => ({ ...s, mch_serial_no: e.target.value }))}
                        maxLength={64}
                      />
                    </Field>
                    <Field>
                      <FieldLabel htmlFor="wechat-api-v3">APIv3 密钥</FieldLabel>
                      <FieldDescription>
                        {wechat.api_v3_key_configured ? '已配置，留空则不修改' : ''}
                      </FieldDescription>
                      <PasswordInput
                        id="wechat-api-v3"
                        value={wechat.api_v3_key}
                        onChange={(e) => setWechat((s) => ({ ...s, api_v3_key: e.target.value }))}
                      />
                    </Field>
                    <Field>
                      <FieldLabel htmlFor="wechat-private-key">商户私钥</FieldLabel>
                      <FieldDescription>
                        {wechat.private_key_configured ? '已配置，留空则不修改' : 'PEM 格式'}
                      </FieldDescription>
                      <Textarea
                        id="wechat-private-key"
                        value={wechat.private_key}
                        onChange={(e) => setWechat((s) => ({ ...s, private_key: e.target.value }))}
                        rows={4}
                      />
                    </Field>
                    <Field>
                      <FieldLabel htmlFor="wechat-app-id-web">网站 / 公众号 AppID</FieldLabel>
                      <FieldDescription>用于 PC Native 扫码</FieldDescription>
                      <Input
                        id="wechat-app-id-web"
                        value={wechat.app_id_web}
                        onChange={(e) => setWechat((s) => ({ ...s, app_id_web: e.target.value }))}
                        maxLength={64}
                      />
                    </Field>
                    <Field>
                      <FieldLabel htmlFor="wechat-app-id-app">开放平台 AppID</FieldLabel>
                      <FieldDescription>用于 APP 支付</FieldDescription>
                      <Input
                        id="wechat-app-id-app"
                        value={wechat.app_id_app}
                        onChange={(e) => setWechat((s) => ({ ...s, app_id_app: e.target.value }))}
                        maxLength={64}
                      />
                    </Field>
                    <Field>
                      <FieldLabel htmlFor="wechat-notify-url">异步通知 URL</FieldLabel>
                      <FieldDescription>须为 https</FieldDescription>
                      <Input
                        id="wechat-notify-url"
                        value={wechat.notify_url}
                        onChange={(e) => setWechat((s) => ({ ...s, notify_url: e.target.value }))}
                      />
                    </Field>
                    <Field orientation="horizontal">
                      <Switch
                        id="wechat-pc-web"
                        checked={wechat.pc_web_enabled}
                        onCheckedChange={(checked) =>
                          setWechat((s) => ({ ...s, pc_web_enabled: checked }))
                        }
                      />
                      <FieldLabel htmlFor="wechat-pc-web">PC 网页支付</FieldLabel>
                    </Field>
                    <Field orientation="horizontal">
                      <Switch
                        id="wechat-app"
                        checked={wechat.app_enabled}
                        onCheckedChange={(checked) =>
                          setWechat((s) => ({ ...s, app_enabled: checked }))
                        }
                      />
                      <FieldLabel htmlFor="wechat-app">APP 支付</FieldLabel>
                    </Field>
                    <Button type="submit" disabled={savingWechat}>
                      {savingWechat ? (
                        <Spinner data-icon="inline-start" />
                      ) : (
                        <Save data-icon="inline-start" />
                      )}
                      保存微信支付
                    </Button>
                  </FieldGroup>
                </form>
              </AccordionContent>
            </AccordionItem>
          </Accordion>
        </CardContent>
      </Card>
    </PageContainer>
  )
}
