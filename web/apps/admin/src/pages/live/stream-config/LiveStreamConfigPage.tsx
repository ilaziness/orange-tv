import { useEffect, useState, type SyntheticEvent } from 'react'
import type { LiveAliyunSettings, LiveProvider, LiveQiniuSettings, LiveSettings } from '@orange-tv/shared'
import { z } from 'zod'
import { adminApi, errorMessage } from '@/lib/api'
import { PageContainer } from '@/components/shared'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
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
import { Switch } from '@/components/ui/switch'
import { Skeleton } from '@/components/ui/skeleton'
import { Spinner } from '@/components/ui/spinner'
import { ExternalLink, Save } from 'lucide-react'
import { toast } from 'sonner'

const LIVE_DOCS_URL = 'https://github.com/ilaziness/orange-tv/blob/main/docs/live-streaming.md'

const emptyAliyun = (): LiveAliyunSettings => ({
  push_domain: '',
  play_domain: '',
  app_name: '',
  region: 'cn-shanghai',
  push_auth_key: '',
  push_auth_key_configured: false,
  play_auth_key: '',
  play_auth_key_configured: false,
  access_key_id: '',
  access_key_id_configured: false,
  access_key_secret: '',
  access_key_secret_configured: false,
})

const emptyQiniu = (): LiveQiniuSettings => ({
  hub: '',
  push_domain: '',
  play_domain: '',
  access_key: '',
  access_configured: false,
  secret_key: '',
  secret_configured: false,
})

const emptyForm = (): LiveSettings => ({
  provider: 'self',
  auth_enabled: false,
  push_ttl_seconds: 43200,
  play_ttl_seconds: 7200,
  sign_secret_configured: false,
  sign_secret: '',
  hook_secret: '',
  srs_api_base: '',
  srs_hook_url: '',
  play_auth_url: '',
  push_server: '',
  play_hls_template: '',
  play_flv_template: '',
  aliyun: emptyAliyun(),
  qiniu: emptyQiniu(),
})

const domain = z.string().refine((v) => {
  const s = v.trim()
  if (!s) return false
  return !s.includes('://') && !s.includes('/') && !s.includes(' ')
}, '域名不能包含协议或路径')

const ttl = z.number().int().min(600, '有效期至少 600 秒').max(86400, '有效期最多 86400 秒')

const pushURL = z.string().refine((v) => {
  const s = v.trim()
  if (!s) return true
  try {
    const u = new URL(s)
    return u.protocol === 'http:' || u.protocol === 'https:' || u.protocol === 'rtmp:'
  } catch {
    return false
  }
}, '推流服务器须为合法 URL（支持 rtmp/http/https）')

const playURL = z.string().refine((v) => {
  const s = v.trim()
  if (!s) return true
  try {
    const u = new URL(s)
    return u.protocol === 'http:' || u.protocol === 'https:'
  } catch {
    return false
  }
}, '播放模板须为 http 或 https 地址')

const templateSchema = z.string().refine((v) => {
  const s = v.trim()
  if (!s) return true
  return s.includes('{stream_key}')
}, '模板须包含 {stream_key} 占位符')

const liveSettingsSchema = z
  .object({
    provider: z.enum(['self', 'aliyun', 'qiniu']),
    auth_enabled: z.boolean(),
    push_ttl_seconds: ttl,
    play_ttl_seconds: ttl,
    sign_secret_configured: z.boolean(),
    sign_secret: z.string().max(256),
    hook_secret: z.string().max(256),
    srs_api_base: z.string().max(500),
    srs_hook_url: z.string(),
    play_auth_url: z.string(),
    push_server: z.string().max(500).and(pushURL),
    play_hls_template: z.string().max(500).and(templateSchema).and(playURL),
    play_flv_template: z.string().max(500).and(templateSchema).and(playURL),
    aliyun: z.object({
      push_domain: z.string(),
      play_domain: z.string(),
      app_name: z.string(),
      region: z.string(),
      push_auth_key: z.string(),
      push_auth_key_configured: z.boolean(),
      play_auth_key: z.string(),
      play_auth_key_configured: z.boolean(),
      access_key_id: z.string(),
      access_key_id_configured: z.boolean(),
      access_key_secret: z.string(),
      access_key_secret_configured: z.boolean(),
    }),
    qiniu: z.object({
      hub: z.string(),
      push_domain: z.string(),
      play_domain: z.string(),
      access_key: z.string(),
      access_configured: z.boolean(),
      secret_key: z.string(),
      secret_configured: z.boolean(),
    }),
  })
  .superRefine((data, ctx) => {
    if (data.provider === 'self' && data.auth_enabled) {
      if (!data.sign_secret_configured && !data.sign_secret.trim()) {
        ctx.addIssue({ code: 'custom', path: ['sign_secret'], message: '打开鉴权需要签名密钥' })
      }
      if (!data.hook_secret.trim() && !data.srs_hook_url) {
        ctx.addIssue({ code: 'custom', path: ['hook_secret'], message: '打开鉴权需要回调密钥' })
      }
      if (!data.srs_api_base.trim()) {
        ctx.addIssue({ code: 'custom', path: ['srs_api_base'], message: '打开鉴权需要 SRS API 地址' })
      }
      if (data.sign_secret.trim() && data.hook_secret.trim() && data.sign_secret.trim() === data.hook_secret.trim()) {
        ctx.addIssue({ code: 'custom', path: ['hook_secret'], message: '签名密钥与回调密钥不能相同' })
      }
    }
    if (data.provider === 'aliyun') {
      const a = data.aliyun
      if (!domain.safeParse(a.push_domain).success) {
        ctx.addIssue({ code: 'custom', path: ['aliyun', 'push_domain'], message: '请填写推流域名，不要带协议或路径' })
      }
      if (!domain.safeParse(a.play_domain).success) {
        ctx.addIssue({ code: 'custom', path: ['aliyun', 'play_domain'], message: '请填写播放域名，不要带协议或路径' })
      }
      if (!a.app_name.trim()) {
        ctx.addIssue({ code: 'custom', path: ['aliyun', 'app_name'], message: '请填写 AppName' })
      }
      if (!a.push_auth_key_configured && !a.push_auth_key.trim()) {
        ctx.addIssue({ code: 'custom', path: ['aliyun', 'push_auth_key'], message: '请填写推流鉴权 Key' })
      }
      if (!a.play_auth_key_configured && !a.play_auth_key.trim()) {
        ctx.addIssue({ code: 'custom', path: ['aliyun', 'play_auth_key'], message: '请填写播放鉴权 Key' })
      }
      if (!a.access_key_id_configured && !a.access_key_id.trim()) {
        ctx.addIssue({ code: 'custom', path: ['aliyun', 'access_key_id'], message: '请填写 AccessKeyId' })
      }
      if (!a.access_key_secret_configured && !a.access_key_secret.trim()) {
        ctx.addIssue({ code: 'custom', path: ['aliyun', 'access_key_secret'], message: '请填写 AccessKeySecret' })
      }
    }
    if (data.provider === 'qiniu') {
      const q = data.qiniu
      if (!q.hub.trim()) {
        ctx.addIssue({ code: 'custom', path: ['qiniu', 'hub'], message: '请填写 Hub' })
      }
      if (!domain.safeParse(q.push_domain).success) {
        ctx.addIssue({ code: 'custom', path: ['qiniu', 'push_domain'], message: '请填写推流域名，不要带协议或路径' })
      }
      if (!domain.safeParse(q.play_domain).success) {
        ctx.addIssue({ code: 'custom', path: ['qiniu', 'play_domain'], message: '请填写播放域名，不要带协议或路径' })
      }
      if (!q.access_configured && !q.access_key.trim()) {
        ctx.addIssue({ code: 'custom', path: ['qiniu', 'access_key'], message: '请填写 AccessKey' })
      }
      if (!q.secret_configured && !q.secret_key.trim()) {
        ctx.addIssue({ code: 'custom', path: ['qiniu', 'secret_key'], message: '请填写 SecretKey' })
      }
    }
  })

function secretPlaceholder(configured: boolean): string {
  return configured ? '已配置（留空表示不修改）' : '请输入密钥'
}

export default function LiveStreamConfigPage() {
  const [form, setForm] = useState<LiveSettings>(emptyForm)
  const [loading, setLoading] = useState(false)
  const [submitting, setSubmitting] = useState(false)

  async function load(opts?: { silent?: boolean }) {
    if (!opts?.silent) setLoading(true)
    try {
      const res = await adminApi.getLiveSettings()
      const data = res.data
      setForm({
        ...emptyForm(),
        ...data,
        provider: data.provider || 'self',
        sign_secret: '',
        hook_secret: '',
        aliyun: { ...emptyAliyun(), ...data.aliyun, push_auth_key: '', play_auth_key: '', access_key_id: '', access_key_secret: '' },
        qiniu: { ...emptyQiniu(), ...data.qiniu, access_key: '', secret_key: '' },
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

  async function save(e: SyntheticEvent<HTMLFormElement>) {
    e.preventDefault()
    if (submitting) return
    const result = liveSettingsSchema.safeParse(form)
    if (!result.success) {
      toast.error(result.error.issues[0]?.message || '表单校验失败')
      return
    }
    const data = result.data
    setSubmitting(true)
    try {
      await adminApi.updateSettings<LiveSettings>({
        group: 'live',
        data: {
          provider: data.provider,
          auth_enabled: data.auth_enabled,
          push_ttl_seconds: data.push_ttl_seconds,
          play_ttl_seconds: data.play_ttl_seconds,
          sign_secret: data.sign_secret.trim(),
          hook_secret: data.hook_secret.trim(),
          srs_api_base: data.srs_api_base.trim(),
          push_server: data.push_server.trim(),
          play_hls_template: data.play_hls_template.trim(),
          play_flv_template: data.play_flv_template.trim(),
          aliyun: {
            push_domain: data.aliyun.push_domain.trim(),
            play_domain: data.aliyun.play_domain.trim(),
            app_name: data.aliyun.app_name.trim(),
            region: data.aliyun.region.trim() || 'cn-shanghai',
            push_auth_key: data.aliyun.push_auth_key.trim(),
            play_auth_key: data.aliyun.play_auth_key.trim(),
            access_key_id: data.aliyun.access_key_id.trim(),
            access_key_secret: data.aliyun.access_key_secret.trim(),
          },
          qiniu: {
            hub: data.qiniu.hub.trim(),
            push_domain: data.qiniu.push_domain.trim(),
            play_domain: data.qiniu.play_domain.trim(),
            access_key: data.qiniu.access_key.trim(),
            secret_key: data.qiniu.secret_key.trim(),
          },
        },
      })
      toast.success('推流配置已保存')
      await load({ silent: true })
    } catch (err) {
      toast.error(errorMessage(err))
    } finally {
      setSubmitting(false)
    }
  }

  const provider = form.provider

  return (
    <PageContainer>
      <Card>
        <CardHeader>
          <CardTitle>推流配置</CardTitle>
          <CardDescription>
            选择自建 SRS、阿里云直播或七牛直播。自建地址仍使用原来的推流服务器和播放模板，不会另存一份。
            密钥留空表示不修改。有房间正在直播时不能保存。部署说明见{' '}
            <a
              href={LIVE_DOCS_URL}
              target="_blank"
              rel="noopener noreferrer"
              className="inline-flex items-center gap-1 text-primary underline-offset-4 hover:underline cursor-pointer"
            >
              docs/live-streaming.md
              <ExternalLink />
            </a>
            。
          </CardDescription>
        </CardHeader>
        <CardContent>
          {loading ? (
            <div className="flex flex-col gap-4">
              {Array.from({ length: 4 }).map((_, i) => (
                <Skeleton key={i} className="h-10 w-full" />
              ))}
            </div>
          ) : (
            <form onSubmit={save} className="flex flex-col gap-4">
              <FieldGroup>
                <Field data-disabled={submitting ? true : undefined}>
                  <FieldLabel>直播厂商</FieldLabel>
                  <Select
                    items={[
                      { value: 'self', label: '自建 SRS' },
                      { value: 'aliyun', label: '阿里云直播' },
                      { value: 'qiniu', label: '七牛直播' },
                    ]}
                    value={provider}
                    onValueChange={(value) =>
                      setForm((prev) => ({ ...prev, provider: (value || 'self') as LiveProvider }))
                    }
                    disabled={submitting}
                  >
                    <SelectTrigger className="w-full cursor-pointer">
                      <SelectValue placeholder="选择厂商" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="self">自建 SRS</SelectItem>
                      <SelectItem value="aliyun">阿里云直播</SelectItem>
                      <SelectItem value="qiniu">七牛直播</SelectItem>
                    </SelectContent>
                  </Select>
                </Field>
                <Field data-disabled={submitting ? true : undefined}>
                  <FieldLabel htmlFor="push_ttl">推流签名有效期（秒）</FieldLabel>
                  <Input
                    id="push_ttl"
                    type="number"
                    min={600}
                    max={86400}
                    value={form.push_ttl_seconds}
                    onChange={(e) =>
                      setForm((prev) => ({ ...prev, push_ttl_seconds: Number(e.target.value) }))
                    }
                    disabled={submitting}
                  />
                  <FieldDescription>默认 12 小时（43200）。范围 600–86400。</FieldDescription>
                </Field>
                <Field data-disabled={submitting ? true : undefined}>
                  <FieldLabel htmlFor="play_ttl">播放签名有效期（秒）</FieldLabel>
                  <Input
                    id="play_ttl"
                    type="number"
                    min={600}
                    max={86400}
                    value={form.play_ttl_seconds}
                    onChange={(e) =>
                      setForm((prev) => ({ ...prev, play_ttl_seconds: Number(e.target.value) }))
                    }
                    disabled={submitting}
                  />
                  <FieldDescription>默认 2 小时（7200）。观看页会在到期前换地址。</FieldDescription>
                </Field>
              </FieldGroup>

              {provider === 'self' ? (
                <FieldGroup>
                  <Field orientation="horizontal" data-disabled={submitting ? true : undefined}>
                    <Switch
                      id="auth_enabled"
                      checked={form.auth_enabled}
                      onCheckedChange={(checked) =>
                        setForm((prev) => ({ ...prev, auth_enabled: checked }))
                      }
                      disabled={submitting}
                    />
                    <FieldLabel htmlFor="auth_enabled">开启自建鉴权</FieldLabel>
                  </Field>
                  <FieldDescription>
                    鉴权开关不会替你改 SRS。回调没配好时，裸流名仍然能推。未打开鉴权时，继续用下面三个地址拼裸 URL。
                  </FieldDescription>
                  <Field data-disabled={submitting ? true : undefined}>
                    <FieldLabel htmlFor="push_server">推流服务器</FieldLabel>
                    <Input
                      id="push_server"
                      placeholder="rtmp://media.example.com/live"
                      value={form.push_server}
                      onChange={(e) => setForm((prev) => ({ ...prev, push_server: e.target.value }))}
                      disabled={submitting}
                      maxLength={500}
                    />
                  </Field>
                  <Field data-disabled={submitting ? true : undefined}>
                    <FieldLabel htmlFor="play_hls_template">HLS 播放模板</FieldLabel>
                    <Input
                      id="play_hls_template"
                      placeholder="https://media.example.com/live/{stream_key}.m3u8"
                      value={form.play_hls_template}
                      onChange={(e) =>
                        setForm((prev) => ({ ...prev, play_hls_template: e.target.value }))
                      }
                      disabled={submitting}
                      maxLength={500}
                    />
                  </Field>
                  <Field data-disabled={submitting ? true : undefined}>
                    <FieldLabel htmlFor="play_flv_template">FLV 播放模板（可选）</FieldLabel>
                    <Input
                      id="play_flv_template"
                      placeholder="https://media.example.com/live/{stream_key}.flv"
                      value={form.play_flv_template}
                      onChange={(e) =>
                        setForm((prev) => ({ ...prev, play_flv_template: e.target.value }))
                      }
                      disabled={submitting}
                      maxLength={500}
                    />
                  </Field>
                  {form.auth_enabled ? (
                    <>
                      <Field data-disabled={submitting ? true : undefined}>
                        <FieldLabel htmlFor="sign_secret">签名密钥</FieldLabel>
                        <Input
                          id="sign_secret"
                          type="password"
                          autoComplete="new-password"
                          placeholder={secretPlaceholder(form.sign_secret_configured)}
                          value={form.sign_secret}
                          onChange={(e) =>
                            setForm((prev) => ({ ...prev, sign_secret: e.target.value }))
                          }
                          disabled={submitting}
                          maxLength={256}
                        />
                        <FieldDescription>只用于推流和播放 HMAC，接口不回传。</FieldDescription>
                      </Field>
                      <Field data-disabled={submitting ? true : undefined}>
                        <FieldLabel htmlFor="hook_secret">回调密钥</FieldLabel>
                        <Input
                          id="hook_secret"
                          type="password"
                          autoComplete="new-password"
                          placeholder={secretPlaceholder(Boolean(form.srs_hook_url))}
                          value={form.hook_secret}
                          onChange={(e) =>
                            setForm((prev) => ({ ...prev, hook_secret: e.target.value }))
                          }
                          disabled={submitting}
                          maxLength={256}
                        />
                        <FieldDescription>只放在 SRS 回调 URL 里，必须和签名密钥不同。</FieldDescription>
                      </Field>
                      <Field data-disabled={submitting ? true : undefined}>
                        <FieldLabel htmlFor="srs_api_base">SRS API 地址</FieldLabel>
                        <Input
                          id="srs_api_base"
                          placeholder="http://127.0.0.1:1985"
                          value={form.srs_api_base}
                          onChange={(e) =>
                            setForm((prev) => ({ ...prev, srs_api_base: e.target.value }))
                          }
                          disabled={submitting}
                          maxLength={500}
                        />
                        <FieldDescription>1985 只监听内网，不要暴露到公网。</FieldDescription>
                      </Field>
                      <Field>
                        <FieldLabel>SRS 回调 URL</FieldLabel>
                        <Input readOnly value={form.srs_hook_url} placeholder="保存回调密钥后生成" />
                      </Field>
                      <Field>
                        <FieldLabel>Nginx 播放鉴权 URL</FieldLabel>
                        <Input readOnly value={form.play_auth_url} placeholder="保存回调密钥后生成" />
                        <FieldDescription>
                          只给 m3u8 和 flv 使用，并把 $request_uri 放到 X-Original-URI。不要把每个 .ts 打到这里。
                        </FieldDescription>
                      </Field>
                    </>
                  ) : null}
                </FieldGroup>
              ) : null}

              {provider === 'aliyun' ? (
                <FieldGroup>
                  <Field data-disabled={submitting ? true : undefined}>
                    <FieldLabel htmlFor="aliyun_push">推流域名</FieldLabel>
                    <Input
                      id="aliyun_push"
                      placeholder="push.example.com"
                      value={form.aliyun.push_domain}
                      onChange={(e) =>
                        setForm((prev) => ({
                          ...prev,
                          aliyun: { ...prev.aliyun, push_domain: e.target.value },
                        }))
                      }
                      disabled={submitting}
                    />
                  </Field>
                  <Field data-disabled={submitting ? true : undefined}>
                    <FieldLabel htmlFor="aliyun_play">播放域名</FieldLabel>
                    <Input
                      id="aliyun_play"
                      placeholder="play.example.com"
                      value={form.aliyun.play_domain}
                      onChange={(e) =>
                        setForm((prev) => ({
                          ...prev,
                          aliyun: { ...prev.aliyun, play_domain: e.target.value },
                        }))
                      }
                      disabled={submitting}
                    />
                  </Field>
                  <Field data-disabled={submitting ? true : undefined}>
                    <FieldLabel htmlFor="aliyun_app">AppName</FieldLabel>
                    <Input
                      id="aliyun_app"
                      placeholder="live"
                      value={form.aliyun.app_name}
                      onChange={(e) =>
                        setForm((prev) => ({
                          ...prev,
                          aliyun: { ...prev.aliyun, app_name: e.target.value },
                        }))
                      }
                      disabled={submitting}
                    />
                  </Field>
                  <Field data-disabled={submitting ? true : undefined}>
                    <FieldLabel htmlFor="aliyun_region">地域</FieldLabel>
                    <Input
                      id="aliyun_region"
                      placeholder="cn-shanghai"
                      value={form.aliyun.region}
                      onChange={(e) =>
                        setForm((prev) => ({
                          ...prev,
                          aliyun: { ...prev.aliyun, region: e.target.value },
                        }))
                      }
                      disabled={submitting}
                    />
                  </Field>
                  <SecretField
                    id="aliyun_push_key"
                    label="推流鉴权 Key"
                    value={form.aliyun.push_auth_key}
                    configured={form.aliyun.push_auth_key_configured}
                    disabled={submitting}
                    onChange={(value) =>
                      setForm((prev) => ({ ...prev, aliyun: { ...prev.aliyun, push_auth_key: value } }))
                    }
                  />
                  <SecretField
                    id="aliyun_play_key"
                    label="播放鉴权 Key"
                    value={form.aliyun.play_auth_key}
                    configured={form.aliyun.play_auth_key_configured}
                    disabled={submitting}
                    onChange={(value) =>
                      setForm((prev) => ({ ...prev, aliyun: { ...prev.aliyun, play_auth_key: value } }))
                    }
                  />
                  <SecretField
                    id="aliyun_ak"
                    label="AccessKeyId"
                    value={form.aliyun.access_key_id}
                    configured={form.aliyun.access_key_id_configured}
                    disabled={submitting}
                    onChange={(value) =>
                      setForm((prev) => ({ ...prev, aliyun: { ...prev.aliyun, access_key_id: value } }))
                    }
                  />
                  <SecretField
                    id="aliyun_sk"
                    label="AccessKeySecret"
                    value={form.aliyun.access_key_secret}
                    configured={form.aliyun.access_key_secret_configured}
                    disabled={submitting}
                    onChange={(value) =>
                      setForm((prev) => ({
                        ...prev,
                        aliyun: { ...prev.aliyun, access_key_secret: value },
                      }))
                    }
                  />
                  <FieldDescription>
                    控制台鉴权方式选 A，推流和播放各一把 Key。AccessKey 只用于禁推和恢复，RAM 不要授其它直播权限。
                  </FieldDescription>
                </FieldGroup>
              ) : null}

              {provider === 'qiniu' ? (
                <FieldGroup>
                  <Field data-disabled={submitting ? true : undefined}>
                    <FieldLabel htmlFor="qiniu_hub">Hub</FieldLabel>
                    <Input
                      id="qiniu_hub"
                      value={form.qiniu.hub}
                      onChange={(e) =>
                        setForm((prev) => ({ ...prev, qiniu: { ...prev.qiniu, hub: e.target.value } }))
                      }
                      disabled={submitting}
                    />
                  </Field>
                  <Field data-disabled={submitting ? true : undefined}>
                    <FieldLabel htmlFor="qiniu_push">推流域名</FieldLabel>
                    <Input
                      id="qiniu_push"
                      placeholder="push.example.com"
                      value={form.qiniu.push_domain}
                      onChange={(e) =>
                        setForm((prev) => ({
                          ...prev,
                          qiniu: { ...prev.qiniu, push_domain: e.target.value },
                        }))
                      }
                      disabled={submitting}
                    />
                  </Field>
                  <Field data-disabled={submitting ? true : undefined}>
                    <FieldLabel htmlFor="qiniu_play">播放域名</FieldLabel>
                    <Input
                      id="qiniu_play"
                      placeholder="play.example.com"
                      value={form.qiniu.play_domain}
                      onChange={(e) =>
                        setForm((prev) => ({
                          ...prev,
                          qiniu: { ...prev.qiniu, play_domain: e.target.value },
                        }))
                      }
                      disabled={submitting}
                    />
                  </Field>
                  <SecretField
                    id="qiniu_ak"
                    label="AccessKey"
                    value={form.qiniu.access_key}
                    configured={form.qiniu.access_configured}
                    disabled={submitting}
                    onChange={(value) =>
                      setForm((prev) => ({ ...prev, qiniu: { ...prev.qiniu, access_key: value } }))
                    }
                  />
                  <SecretField
                    id="qiniu_sk"
                    label="SecretKey"
                    value={form.qiniu.secret_key}
                    configured={form.qiniu.secret_configured}
                    disabled={submitting}
                    onChange={(value) =>
                      setForm((prev) => ({ ...prev, qiniu: { ...prev.qiniu, secret_key: value } }))
                    }
                  />
                  <FieldDescription>
                    播放使用时间戳防盗链，控制台密钥必须与这里的 SecretKey 相同。推流使用带过期时间的管理凭证。
                  </FieldDescription>
                </FieldGroup>
              ) : null}

              <div>
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

function SecretField({
  id,
  label,
  value,
  configured,
  disabled,
  onChange,
}: {
  id: string
  label: string
  value: string
  configured: boolean
  disabled: boolean
  onChange: (value: string) => void
}) {
  return (
    <Field data-disabled={disabled ? true : undefined}>
      <FieldLabel htmlFor={id}>{label}</FieldLabel>
      <Input
        id={id}
        type="password"
        autoComplete="new-password"
        placeholder={secretPlaceholder(configured)}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        disabled={disabled}
      />
    </Field>
  )
}
