import { useEffect, useState } from 'react'
import type * as React from 'react'
import { z } from 'zod'
import { adminApi, errorMessage } from '@/lib/api'
import { PageContainer } from '@/components/shared'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Field, FieldGroup, FieldLabel } from '@/components/ui/field'
import { Textarea } from '@/components/ui/textarea'
import { Skeleton } from '@/components/ui/skeleton'
import { Spinner } from '@/components/ui/spinner'
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs'
import { Switch } from '@/components/ui/switch'
import { Save } from 'lucide-react'
import { toast } from 'sonner'

const siteSettingsSchema = z.object({
  name: z.string().min(1, '站点名称不能为空'),
  logo: z.string(),
  copyright: z.string(),
  icp: z.string(),
  seo_keywords: z.string(),
  description: z.string(),
  analytics_code: z.string().max(2048, '统计代码长度不能超过 2048 个字符'),
})

function FeatureSettingsTab() {
  const [form, setForm] = useState({
    livetv_enabled: false,
    comment_enabled: true,
    comment_review: true,
    rating_enabled: true,
  })
  const [loading, setLoading] = useState(false)
  const [submitting, setSubmitting] = useState(false)

  async function load(opts?: { silent?: boolean }) {
    if (!opts?.silent) setLoading(true)
    try {
      const res = await adminApi.getFeatureSettings()
      const f = res.data
      setForm({
        livetv_enabled: !!f.livetv_enabled,
        comment_enabled: !!f.comment_enabled,
        comment_review: !!f.comment_review,
        rating_enabled: !!f.rating_enabled,
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
    setSubmitting(true)
    try {
      await adminApi.updateSettings({ group: 'feature', data: form })
      toast.success('功能设置已保存')
      await load({ silent: true })
    } catch (err) {
      toast.error(errorMessage(err))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>功能设置</CardTitle>
      </CardHeader>
      <CardContent>
        {loading ? (
          <div className="flex flex-col gap-4">
            {Array.from({ length: 3 }).map((_, i) => (
              <Skeleton key={i} className="h-10 w-full" />
            ))}
          </div>
        ) : (
          <form onSubmit={save} className="flex flex-col gap-4">
            <FieldGroup>
              <Field data-disabled={submitting ? true : undefined}>
                <FieldLabel htmlFor="livetv_enabled">电视直播</FieldLabel>
                <div className="flex items-center gap-2 text-sm">
                  <Switch
                    id="livetv_enabled"
                    checked={form.livetv_enabled}
                    onCheckedChange={(checked) =>
                      setForm((prev) => ({ ...prev, livetv_enabled: checked }))
                    }
                    disabled={submitting}
                  />
                  <span>开启用户端电视直播功能</span>
                </div>
              </Field>
              <Field data-disabled={submitting ? true : undefined}>
                <FieldLabel htmlFor="comment_enabled">视频评论</FieldLabel>
                <div className="flex items-center gap-2 text-sm">
                  <Switch
                    id="comment_enabled"
                    checked={form.comment_enabled}
                    onCheckedChange={(checked) =>
                      setForm((prev) => ({
                        ...prev,
                        comment_enabled: checked,
                        ...(!checked ? { comment_review: false } : {}),
                      }))
                    }
                    disabled={submitting}
                  />
                  <span>开启用户端视频评论功能</span>
                </div>
              </Field>
              {form.comment_enabled && (
                <Field data-disabled={submitting ? true : undefined}>
                  <FieldLabel htmlFor="comment_review">评论审核</FieldLabel>
                  <div className="flex items-center gap-2 text-sm">
                    <Switch
                      id="comment_review"
                      checked={form.comment_review}
                      onCheckedChange={(checked) =>
                        setForm((prev) => ({ ...prev, comment_review: checked }))
                      }
                      disabled={submitting}
                    />
                    <span>评论需要审核后才能显示</span>
                  </div>
                </Field>
              )}
              <Field data-disabled={submitting ? true : undefined}>
                <FieldLabel htmlFor="rating_enabled">视频评分</FieldLabel>
                <div className="flex items-center gap-2 text-sm">
                  <Switch
                    id="rating_enabled"
                    checked={form.rating_enabled}
                    onCheckedChange={(checked) =>
                      setForm((prev) => ({ ...prev, rating_enabled: checked }))
                    }
                    disabled={submitting}
                  />
                  <span>开启用户端视频评分功能</span>
                </div>
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
  )
}

export default function SiteSettingsPage() {
  const [form, setForm] = useState({
    name: '',
    logo: '',
    copyright: '',
    icp: '',
    seo_keywords: '',
    description: '',
    analytics_code: '',
  })
  const [loading, setLoading] = useState(false)
  const [submitting, setSubmitting] = useState(false)

  async function load(opts?: { silent?: boolean }) {
    if (!opts?.silent) setLoading(true)
    try {
      const res = await adminApi.getSiteSettings()
      const site = res.data
      setForm({
        name: site.name || '',
        logo: site.logo || '',
        copyright: site.copyright || '',
        icp: site.icp || '',
        seo_keywords: site.seo_keywords || '',
        description: site.description || '',
        analytics_code: site.analytics_code || '',
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
    const result = siteSettingsSchema.safeParse(form)
    if (!result.success) {
      toast.error(result.error.issues[0]?.message || '表单校验失败')
      return
    }
    setSubmitting(true)
    try {
      await adminApi.updateSettings({ group: 'site', data: result.data })
      toast.success('设置已保存')
      await load({ silent: true })
    } catch (err) {
      toast.error(errorMessage(err))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <PageContainer>
      <Tabs defaultValue="site">
        <TabsList>
          <TabsTrigger value="site">站点设置</TabsTrigger>
          <TabsTrigger value="feature">功能设置</TabsTrigger>
        </TabsList>
        <TabsContent value="site">
          <Card>
            <CardHeader>
              <CardTitle>站点设置</CardTitle>
            </CardHeader>
            <CardContent>
              {loading ? (
                <div className="flex flex-col gap-4">
                  {Array.from({ length: 7 }).map((_, i) => (
                    <Skeleton key={i} className="h-10 w-full" />
                  ))}
                </div>
              ) : (
                <form onSubmit={save} className="flex flex-col gap-4">
                  <FieldGroup>
                    <Field data-disabled={submitting ? true : undefined}>
                      <FieldLabel htmlFor="name">站点名称</FieldLabel>
                      <Input
                        id="name"
                        placeholder="请输入站点名称"
                        value={form.name}
                        onChange={(e) => setForm((prev) => ({ ...prev, name: e.target.value }))}
                        disabled={submitting}
                      />
                    </Field>
                    <Field data-disabled={submitting ? true : undefined}>
                      <FieldLabel htmlFor="logo">Logo URL</FieldLabel>
                      <Input
                        id="logo"
                        placeholder="请输入 Logo 图片地址（可选）"
                        value={form.logo}
                        onChange={(e) => setForm((prev) => ({ ...prev, logo: e.target.value }))}
                        disabled={submitting}
                      />
                    </Field>
                    <Field data-disabled={submitting ? true : undefined}>
                      <FieldLabel htmlFor="copyright">版权信息</FieldLabel>
                      <Input
                        id="copyright"
                        placeholder="请输入版权信息，如 © 2024 YourSite"
                        value={form.copyright}
                        onChange={(e) =>
                          setForm((prev) => ({ ...prev, copyright: e.target.value }))
                        }
                        disabled={submitting}
                      />
                    </Field>
                    <Field data-disabled={submitting ? true : undefined}>
                      <FieldLabel htmlFor="icp">备案号</FieldLabel>
                      <Input
                        id="icp"
                        placeholder="请输入 ICP 备案号（可选）"
                        value={form.icp}
                        onChange={(e) => setForm((prev) => ({ ...prev, icp: e.target.value }))}
                        disabled={submitting}
                      />
                    </Field>
                    <Field data-disabled={submitting ? true : undefined}>
                      <FieldLabel htmlFor="seo_keywords">SEO 关键词</FieldLabel>
                      <Input
                        id="seo_keywords"
                        placeholder="请输入 SEO 关键词，逗号分隔"
                        value={form.seo_keywords}
                        onChange={(e) =>
                          setForm((prev) => ({ ...prev, seo_keywords: e.target.value }))
                        }
                        disabled={submitting}
                      />
                    </Field>
                    <Field data-disabled={submitting ? true : undefined}>
                      <FieldLabel htmlFor="analytics_code">统计代码</FieldLabel>
                      <Textarea
                        id="analytics_code"
                        rows={5}
                        placeholder="请输入百度统计、Google Analytics 等网页统计代码（可选）"
                        value={form.analytics_code}
                        onChange={(e) =>
                          setForm((prev) => ({ ...prev, analytics_code: e.target.value }))
                        }
                        disabled={submitting}
                      />
                    </Field>
                    <Field data-disabled={submitting ? true : undefined}>
                      <FieldLabel htmlFor="description">站点描述</FieldLabel>
                      <Textarea
                        id="description"
                        rows={3}
                        placeholder="请输入站点描述（可选）"
                        value={form.description}
                        onChange={(e) =>
                          setForm((prev) => ({ ...prev, description: e.target.value }))
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
        </TabsContent>
        <TabsContent value="feature">
          <FeatureSettingsTab />
        </TabsContent>
      </Tabs>
    </PageContainer>
  )
}
