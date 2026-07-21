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
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Field, FieldGroup, FieldLabel } from '@/components/ui/field'
import { Textarea } from '@/components/ui/textarea'
import { Skeleton } from '@/components/ui/skeleton'
import { Save } from 'lucide-react'
import { toast } from 'sonner'

const siteSettingsSchema = z.object({
  name: z.string().min(1, '站点名称不能为空'),
  logo: z.string(),
  copyright: z.string(),
  icp: z.string(),
  seo_keywords: z.string(),
  description: z.string(),
})

export default function SiteSettingsPage() {
  const [form, setForm] = useState({
    name: '', logo: '', copyright: '', icp: '', seo_keywords: '', description: '',
  })
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  async function load() {
    setLoading(true)
    setError('')
    try {
      const res = await adminApi.getSettings()
      const site = res.data.site
      setForm({
        name: site.name || '',
        logo: site.logo || '',
        copyright: site.copyright || '',
        icp: site.icp || '',
        seo_keywords: site.seo_keywords || '',
        description: site.description || '',
      })
    } catch (err) {
      setError(errorMessage(err))
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { void load() }, [])

  async function save(e: React.SyntheticEvent<HTMLFormElement>) {
    e.preventDefault()
    setError('')
    const result = siteSettingsSchema.safeParse(form)
    if (!result.success) {
      setError(result.error.issues[0]?.message || '表单校验失败')
      return
    }
    try {
      await adminApi.updateSettings({ site: result.data })
      toast.success('设置已保存')
      await load()
    } catch (err) {
      setError(errorMessage(err))
    }
  }

  return (
    <PageContainer>
      <Card>
        <CardHeader>
          <CardTitle>站点设置</CardTitle>
        </CardHeader>
        <CardContent>
          {error && (
            <Alert variant="destructive" className="mb-4">
              <AlertTitle>出错了</AlertTitle>
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}
          {loading ? (
            <div className="flex flex-col gap-4">
              {Array.from({ length: 6 }).map((_, i) => (
                <Skeleton key={i} className="h-10 w-full" />
              ))}
            </div>
          ) : (
            <form onSubmit={save} className="flex flex-col gap-4">
              <FieldGroup>
                <Field>
                  <FieldLabel htmlFor="name">站点名称</FieldLabel>
                  <Input id="name" value={form.name} onChange={(e) => setForm((prev) => ({ ...prev, name: e.target.value }))} />
                </Field>
                <Field>
                  <FieldLabel htmlFor="logo">Logo URL</FieldLabel>
                  <Input id="logo" value={form.logo} onChange={(e) => setForm((prev) => ({ ...prev, logo: e.target.value }))} />
                </Field>
                <Field>
                  <FieldLabel htmlFor="copyright">版权信息</FieldLabel>
                  <Input id="copyright" value={form.copyright} onChange={(e) => setForm((prev) => ({ ...prev, copyright: e.target.value }))} />
                </Field>
                <Field>
                  <FieldLabel htmlFor="icp">备案号</FieldLabel>
                  <Input id="icp" value={form.icp} onChange={(e) => setForm((prev) => ({ ...prev, icp: e.target.value }))} />
                </Field>
                <Field>
                  <FieldLabel htmlFor="seo_keywords">SEO 关键词</FieldLabel>
                  <Input id="seo_keywords" value={form.seo_keywords} onChange={(e) => setForm((prev) => ({ ...prev, seo_keywords: e.target.value }))} />
                </Field>
                <Field>
                  <FieldLabel htmlFor="description">站点描述</FieldLabel>
                  <Textarea id="description" rows={3} value={form.description} onChange={(e) => setForm((prev) => ({ ...prev, description: e.target.value }))} />
                </Field>
              </FieldGroup>
              <div className="flex justify-end">
                <Button type="submit">
                  <Save data-icon="inline-start" />
                  保存
                </Button>
              </div>
            </form>
          )}
        </CardContent>
      </Card>
    </PageContainer>
  )
}
