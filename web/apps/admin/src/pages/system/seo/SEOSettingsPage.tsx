import { useEffect, useState } from 'react'
import type * as React from 'react'
import { Link } from 'react-router'
import { z } from 'zod'
import { adminApi, errorMessage } from '@/lib/api'
import { PageContainer } from '@/components/shared'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Skeleton } from '@/components/ui/skeleton'
import { Spinner } from '@/components/ui/spinner'
import { Field, FieldGroup, FieldLabel } from '@/components/ui/field'
import { Switch } from '@/components/ui/switch'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Save, AlertTriangle } from 'lucide-react'
import { toast } from 'sonner'

const seoSettingsSchema = z.object({
  public_base_url: z
    .string()
    .max(500, '站点根地址过长')
    .refine(
      (v) => {
        const s = v.trim()
        if (!s) return true
        try {
          const u = new URL(s)
          if (!(u.protocol === 'http:' || u.protocol === 'https:') || !u.host) return false
          if (u.username || u.password) return false
          const path = (u.pathname || '').replace(/\/+$/, '')
          return path === '' || path === '/'
        } catch {
          return false
        }
      },
      { message: '站点根地址须为 http(s) 域名（含端口），不能含路径或账号' },
    ),
  default_og_image: z
    .string()
    .max(500, '默认分享图地址过长')
    .refine(
      (v) => {
        const s = v.trim()
        if (!s) return true
        try {
          const u = new URL(s)
          return (u.protocol === 'http:' || u.protocol === 'https:') && !!u.host
        } catch {
          return false
        }
      },
      { message: '默认分享图须为 http(s) 绝对地址' },
    ),
  sitemap_enabled: z.boolean(),
  llms_enabled: z.boolean(),
  llms_intro: z.string().max(2000, 'llms 简介不能超过 2000 个字符'),
  allow_ai_search: z.boolean(),
  allow_ai_training: z.boolean(),
  google_site_verification: z.string().max(255),
  baidu_site_verification: z.string().max(255),
  bing_site_verification: z.string().max(255),
})

type FormState = z.infer<typeof seoSettingsSchema>

const EMPTY_FORM: FormState = {
  public_base_url: '',
  default_og_image: '',
  sitemap_enabled: true,
  llms_enabled: true,
  llms_intro: '',
  allow_ai_search: true,
  allow_ai_training: false,
  google_site_verification: '',
  baidu_site_verification: '',
  bing_site_verification: '',
}

export default function SEOSettingsPage() {
  const [form, setForm] = useState<FormState>(EMPTY_FORM)
  const [loading, setLoading] = useState(false)
  const [submitting, setSubmitting] = useState(false)

  async function load(opts?: { silent?: boolean }) {
    if (!opts?.silent) setLoading(true)
    try {
      const res = await adminApi.getSEOSettings()
      const s = res.data
      setForm({
        public_base_url: s.public_base_url || '',
        default_og_image: s.default_og_image || '',
        sitemap_enabled: !!s.sitemap_enabled,
        llms_enabled: !!s.llms_enabled,
        llms_intro: s.llms_intro || '',
        allow_ai_search: !!s.allow_ai_search,
        allow_ai_training: !!s.allow_ai_training,
        google_site_verification: s.google_site_verification || '',
        baidu_site_verification: s.baidu_site_verification || '',
        bing_site_verification: s.bing_site_verification || '',
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
    const result = seoSettingsSchema.safeParse(form)
    if (!result.success) {
      toast.error(result.error.issues[0]?.message || '配置校验失败')
      return
    }
    setSubmitting(true)
    try {
      await adminApi.updateSettings({
        group: 'seo',
        data: { ...result.data },
      })
      toast.success('SEO 设置已保存')
      await load({ silent: true })
    } catch (err) {
      toast.error(errorMessage(err))
    } finally {
      setSubmitting(false)
    }
  }

  const missingBaseURL = !form.public_base_url.trim()

  return (
    <PageContainer>
      <Card>
        <CardHeader>
          <CardTitle>SEO 设置</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="mb-4 text-sm text-muted-foreground">
            用于站点根地址、社交分享默认图、robots / sitemap / llms，以及站长验证。站点名称、描述、关键词请在
            <Link to="/system/site" className="mx-1 text-primary underline-offset-4 hover:underline">
              站点设置
            </Link>
            中修改。
          </p>

          {missingBaseURL && !loading ? (
            <Alert className="mb-4">
              <AlertTriangle />
              <AlertTitle>尚未配置站点根地址</AlertTitle>
              <AlertDescription>
                未填写公开站点根地址时，sitemap.xml 与 llms.txt 将返回 404，前端 canonical / OG
                也无法生成绝对链接。
              </AlertDescription>
            </Alert>
          ) : null}

          {loading ? (
            <div className="flex flex-col gap-4">
              <Skeleton className="h-10 w-full" />
              <Skeleton className="h-10 w-full" />
              <Skeleton className="h-24 w-full" />
            </div>
          ) : (
            <form onSubmit={save} className="flex flex-col gap-6">
              <FieldGroup>
                <Field data-disabled={submitting ? true : undefined}>
                  <FieldLabel htmlFor="public_base_url">公开站点根地址</FieldLabel>
                  <Input
                    id="public_base_url"
                    placeholder="https://www.example.com"
                    value={form.public_base_url}
                    onChange={(e) => setForm((p) => ({ ...p, public_base_url: e.target.value }))}
                    disabled={submitting}
                  />
                </Field>
                <Field data-disabled={submitting ? true : undefined}>
                  <FieldLabel htmlFor="default_og_image">默认分享图 URL</FieldLabel>
                  <Input
                    id="default_og_image"
                    placeholder="https://www.example.com/og.png"
                    value={form.default_og_image}
                    onChange={(e) => setForm((p) => ({ ...p, default_og_image: e.target.value }))}
                    disabled={submitting}
                  />
                </Field>
              </FieldGroup>

              <FieldGroup>
                <Field data-disabled={submitting ? true : undefined}>
                  <FieldLabel htmlFor="sitemap_enabled">输出 sitemap.xml</FieldLabel>
                  <Switch
                    id="sitemap_enabled"
                    checked={form.sitemap_enabled}
                    onCheckedChange={(checked) =>
                      setForm((p) => ({ ...p, sitemap_enabled: checked }))
                    }
                    disabled={submitting}
                  />
                </Field>
                <Field data-disabled={submitting ? true : undefined}>
                  <FieldLabel htmlFor="llms_enabled">输出 llms.txt</FieldLabel>
                  <Switch
                    id="llms_enabled"
                    checked={form.llms_enabled}
                    onCheckedChange={(checked) => setForm((p) => ({ ...p, llms_enabled: checked }))}
                    disabled={submitting}
                  />
                </Field>
                <Field data-disabled={submitting ? true : undefined}>
                  <FieldLabel htmlFor="llms_intro">llms.txt 简介</FieldLabel>
                  <Textarea
                    id="llms_intro"
                    rows={3}
                    value={form.llms_intro}
                    onChange={(e) => setForm((p) => ({ ...p, llms_intro: e.target.value }))}
                    disabled={submitting}
                    placeholder="留空则使用站点描述"
                  />
                </Field>
              </FieldGroup>

              <FieldGroup>
                <Field data-disabled={submitting ? true : undefined}>
                  <FieldLabel htmlFor="allow_ai_search">允许 AI 检索爬虫</FieldLabel>
                  <Switch
                    id="allow_ai_search"
                    checked={form.allow_ai_search}
                    onCheckedChange={(checked) =>
                      setForm((p) => ({ ...p, allow_ai_search: checked }))
                    }
                    disabled={submitting}
                  />
                </Field>
                <Field data-disabled={submitting ? true : undefined}>
                  <FieldLabel htmlFor="allow_ai_training">允许 AI 训练爬虫</FieldLabel>
                  <Switch
                    id="allow_ai_training"
                    checked={form.allow_ai_training}
                    onCheckedChange={(checked) =>
                      setForm((p) => ({ ...p, allow_ai_training: checked }))
                    }
                    disabled={submitting}
                  />
                </Field>
              </FieldGroup>

              <FieldGroup>
                <Field data-disabled={submitting ? true : undefined}>
                  <FieldLabel htmlFor="google_site_verification">Google 验证码</FieldLabel>
                  <Input
                    id="google_site_verification"
                    value={form.google_site_verification}
                    onChange={(e) =>
                      setForm((p) => ({ ...p, google_site_verification: e.target.value }))
                    }
                    disabled={submitting}
                  />
                </Field>
                <Field data-disabled={submitting ? true : undefined}>
                  <FieldLabel htmlFor="baidu_site_verification">百度验证码</FieldLabel>
                  <Input
                    id="baidu_site_verification"
                    value={form.baidu_site_verification}
                    onChange={(e) =>
                      setForm((p) => ({ ...p, baidu_site_verification: e.target.value }))
                    }
                    disabled={submitting}
                  />
                </Field>
                <Field data-disabled={submitting ? true : undefined}>
                  <FieldLabel htmlFor="bing_site_verification">Bing 验证码</FieldLabel>
                  <Input
                    id="bing_site_verification"
                    value={form.bing_site_verification}
                    onChange={(e) =>
                      setForm((p) => ({ ...p, bing_site_verification: e.target.value }))
                    }
                    disabled={submitting}
                  />
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
