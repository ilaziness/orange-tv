import { useEffect, useState, type SyntheticEvent } from 'react'
import type { LiveSettings } from '@orange-tv/shared'
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
import { Field, FieldGroup, FieldLabel } from '@/components/ui/field'
import { Skeleton } from '@/components/ui/skeleton'
import { Spinner } from '@/components/ui/spinner'
import { ExternalLink, Save } from 'lucide-react'
import { toast } from 'sonner'

const LIVE_DOCS_URL = 'https://github.com/ilaziness/orange-tv/blob/main/docs/live-streaming.md'

const emptyForm = (): LiveSettings => ({
  push_server: '',
  play_hls_template: '',
  play_flv_template: '',
})

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
    push_server: z.string().max(500).and(pushURL),
    play_hls_template: z.string().max(500).and(templateSchema).and(playURL),
    play_flv_template: z.string().max(500).and(templateSchema).and(playURL),
  })
  .superRefine((data, ctx) => {
    const push = data.push_server.trim()
    const hls = data.play_hls_template.trim()
    const flv = data.play_flv_template.trim()
    if (push && !hls && !flv) {
      ctx.addIssue({
        code: 'custom',
        path: ['play_hls_template'],
        message: '配置推流服务器时，至少填写一个播放模板（HLS 或 FLV）',
      })
    }
  })

export default function LiveStreamConfigPage() {
  const [form, setForm] = useState<LiveSettings>(emptyForm)
  const [loading, setLoading] = useState(false)
  const [submitting, setSubmitting] = useState(false)

  async function load(opts?: { silent?: boolean }) {
    if (!opts?.silent) setLoading(true)
    try {
      const res = await adminApi.getLiveSettings()
      setForm({
        push_server: res.data.push_server || '',
        play_hls_template: res.data.play_hls_template || '',
        play_flv_template: res.data.play_flv_template || '',
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
    setSubmitting(true)
    try {
      await adminApi.updateSettings<LiveSettings>({
        group: 'live',
        data: {
          push_server: result.data.push_server.trim(),
          play_hls_template: result.data.play_hls_template.trim(),
          play_flv_template: result.data.play_flv_template.trim(),
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

  return (
    <PageContainer>
      <Card>
        <CardHeader>
          <CardTitle>推流配置</CardTitle>
          <CardDescription>
            配置外置媒体服务器地址与播放模板。占位符仅支持{' '}
            <code className="rounded bg-muted px-1 py-0.5 text-xs">{'{stream_key}'}</code>。
            部署说明见{' '}
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
              {Array.from({ length: 3 }).map((_, i) => (
                <Skeleton key={i} className="h-10 w-full" />
              ))}
            </div>
          ) : (
            <form onSubmit={save} className="flex flex-col gap-4">
              <FieldGroup>
                <Field data-disabled={submitting ? true : undefined}>
                  <FieldLabel htmlFor="push_server">推流服务器</FieldLabel>
                  <Input
                    id="push_server"
                    placeholder="rtmp://media.example.com/live"
                    value={form.push_server}
                    onChange={(e) =>
                      setForm((prev) => ({ ...prev, push_server: e.target.value }))
                    }
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
              </FieldGroup>
              <div>
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
