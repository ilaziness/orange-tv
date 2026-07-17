import { useEffect, useState } from 'react'
import { adminApi, errorMessage } from '@/lib/api'
import { PageContainer } from '@/components/shared'
import type { ThemeItem } from '@orange-tv/shared'
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Textarea } from '@/components/ui/textarea'
import { Field, FieldLabel } from '@/components/ui/field'
import { Pencil, Check } from 'lucide-react'
import { toast } from 'sonner'

export default function ThemesPage() {
  const [list, setList] = useState<ThemeItem[]>([])
  const [error, setError] = useState('')
  const [selected, setSelected] = useState<ThemeItem | null>(null)
  const [configText, setConfigText] = useState('{}')
  const [customCss, setCustomCss] = useState('')
  const [customJs, setCustomJs] = useState('')

  async function load() {
    setError('')
    try {
      const res = await adminApi.listThemes()
      setList(res.data || [])
    } catch (err) {
      setError(errorMessage(err))
    }
  }

  useEffect(() => { void load() }, [])

  function pick(item: ThemeItem) {
    setSelected(item)
    setConfigText(JSON.stringify(item.config || {}, null, 2))
    setCustomCss(item.custom_css || '')
    setCustomJs(item.custom_js || '')
  }

  async function save() {
    if (!selected) return
    try {
      const config = JSON.parse(configText || '{}')
      await adminApi.updateTheme(selected.id, { config, custom_css: customCss, custom_js: customJs })
      toast.success('主题配置已保存')
      await load()
    } catch (err) {
      toast.error(errorMessage(err))
    }
  }

  async function activate(id: number) {
    try {
      await adminApi.activateTheme(id)
      toast.success('主题已激活')
      await load()
    } catch (err) {
      toast.error(errorMessage(err))
    }
  }

  return (
    <PageContainer>
      <Card>
        <CardHeader>
          <CardTitle>主题管理</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="mb-4 text-sm text-muted-foreground">
            最小可用：切换激活主题、覆盖 config / custom_css / custom_js。上传第三方主题包不在本阶段。
          </p>
          {error && (
            <Alert variant="destructive" className="mb-4">
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}
          <div className="flex flex-col gap-3">
            {list.map((item) => (
              <div key={item.id} className="flex items-center justify-between rounded-lg border p-4">
                <div>
                  <div className="font-medium">{item.name}</div>
                  <div className="text-sm text-muted-foreground">{item.identifier} · v{item.version}</div>
                </div>
                <div className="flex items-center gap-2">
                  <Badge variant={item.is_active ? 'default' : 'secondary'}>
                    {item.is_active ? '使用中' : '未激活'}
                  </Badge>
                  <Button size="sm" variant="outline" onClick={() => pick(item)}>
                    <Pencil data-icon="inline-start" />
                    编辑
                  </Button>
                  {!item.is_active && (
                    <Button size="sm" onClick={() => void activate(item.id)}>
                      <Check data-icon="inline-start" />
                      激活
                    </Button>
                  )}
                </div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>

      {selected && (
        <Card>
          <CardHeader>
            <CardTitle>编辑：{selected.name}</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex flex-col gap-4">
              <Field>
                <FieldLabel>Config JSON</FieldLabel>
                <Textarea rows={8} value={configText} onChange={(e) => setConfigText(e.target.value)} className="font-mono" />
              </Field>
              <Field>
                <FieldLabel>Custom CSS</FieldLabel>
                <Textarea rows={4} value={customCss} onChange={(e) => setCustomCss(e.target.value)} className="font-mono" />
              </Field>
              <Field>
                <FieldLabel>Custom JS</FieldLabel>
                <Textarea rows={3} value={customJs} onChange={(e) => setCustomJs(e.target.value)} className="font-mono" />
              </Field>
              <div className="flex justify-end">
                <Button onClick={() => void save()}>
                  保存配置
                </Button>
              </div>
            </div>
          </CardContent>
        </Card>
      )}
    </PageContainer>
  )
}
