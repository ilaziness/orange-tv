import { useEffect, useState } from 'react'
import type * as React from 'react'
import { adminApi, errorMessage } from '@/lib/api'
import { flattenCategories } from '@/lib/categories'
import { PageContainer, StatusBadge, ConfirmDialog } from '@/components/shared'
import type { Category, CollectCategoryMap, CollectLog, CollectSource, PlaySource } from '@orange-tv/shared'
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Play, Square, Trash2, RefreshCw } from 'lucide-react'
import { toast } from 'sonner'

const emptyForm = { name: '', type: '2', collect_url: '', api_key: '', cron_expr: '', play_source_id: '0', status: '1', config: '' }

export default function CollectPage() {
  const [sources, setSources] = useState<CollectSource[]>([])
  const [playSources, setPlaySources] = useState<PlaySource[]>([])
  const [categories, setCategories] = useState<Category[]>([])
  const [logs, setLogs] = useState<CollectLog[]>([])
  const [maps, setMaps] = useState<CollectCategoryMap[]>([])
  const [error, setError] = useState('')
  const [selectedId, setSelectedId] = useState(0)
  const [form, setForm] = useState(emptyForm)
  const [mapText, setMapText] = useState('[]')
  const [deleteId, setDeleteId] = useState<number | null>(null)

  async function load() {
    setError('')
    try {
      const [s, p, c, l] = await Promise.all([
        adminApi.listCollectSources(),
        adminApi.listPlaySources(),
        adminApi.listCategories(),
        adminApi.listCollectLogs({ page: 1, page_size: 30 }),
      ])
      setSources(s.data.list || [])
      setPlaySources(p.data.list || [])
      setCategories(c.data || [])
      setLogs(l.data.list || [])
    } catch (err) {
      setError(errorMessage(err))
    }
  }

  useEffect(() => { void load() }, [])

  const flatCats = flattenCategories(categories)

  async function onSubmit(e: React.SyntheticEvent<HTMLFormElement>) {
    e.preventDefault()
    try {
      const body = {
        name: form.name,
        type: Number(form.type),
        collect_url: form.collect_url,
        api_key: form.api_key,
        cron_expr: form.cron_expr,
        play_source_id: Number(form.play_source_id),
        status: Number(form.status),
        config: form.config,
      }
      if (selectedId) {
        await adminApi.updateCollectSource(selectedId, body)
        toast.success('采集源已更新')
      } else {
        await adminApi.createCollectSource(body)
        toast.success('采集源已创建')
      }
      setSelectedId(0)
      setForm(emptyForm)
      await load()
    } catch (err) {
      setError(errorMessage(err))
    }
  }

  async function editSource(item: CollectSource) {
    setSelectedId(item.id)
    setForm({
      name: item.name,
      type: String(item.type),
      collect_url: item.collect_url,
      api_key: '',
      cron_expr: item.cron_expr || '',
      play_source_id: String(item.play_source_id),
      status: String(item.status),
      config: item.config || '',
    })
    try {
      const res = await adminApi.getCollectCategories(item.id)
      setMaps(res.data || [])
      setMapText(JSON.stringify((res.data || []).map((m) => ({ external_category: m.external_category, category_id: m.category_id })), null, 2))
    } catch (err) {
      setError(errorMessage(err))
    }
  }

  async function saveMaps() {
    if (!selectedId) return
    try {
      const items = JSON.parse(mapText || '[]')
      const res = await adminApi.setCollectCategories(selectedId, { items })
      setMaps(res.data || [])
      setMapText(JSON.stringify((res.data || []).map((m) => ({
        external_category: m.external_category,
        category_id: m.category_id,
      })), null, 2))
      toast.success('映射已保存')
      await load()
    } catch (err) {
      toast.error(errorMessage(err))
    }
  }

  async function start(id: number) {
    try {
      await adminApi.startCollect(id)
      toast.success('采集已启动')
      await load()
    } catch (err) {
      toast.error(errorMessage(err))
    }
  }

  async function stop(id: number) {
    try {
      await adminApi.stopCollect(id)
      toast.success('采集已停止')
      await load()
    } catch (err) {
      toast.error(errorMessage(err))
    }
  }

  async function confirmDelete() {
    if (deleteId === null) return
    try {
      await adminApi.deleteCollectSource(deleteId)
      toast.success('采集源已删除')
      await load()
    } catch (err) {
      toast.error(errorMessage(err))
    } finally {
      setDeleteId(null)
    }
  }

  return (
    <PageContainer>
      <Card>
        <CardHeader>
          <CardTitle>数据采集</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="mb-4 text-sm text-muted-foreground">
            支持默认 JSON 与苹果 CMS；手动触发异步执行，可配置 cron 定时采集。请先配置分类映射并绑定播放源。
          </p>
          {error && (
            <Alert variant="destructive" className="mb-4">
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}
          <form onSubmit={onSubmit} className="mb-6 flex flex-col gap-4">
            <div className="flex flex-wrap gap-2">
              <Input
                placeholder="源名称"
                value={form.name}
                onChange={(e) => setForm({ ...form, name: e.target.value })}
                required
                className="max-w-xs"
              />
              <Select value={form.type} onValueChange={(v) => setForm({ ...form, type: v ?? '2' })}>
                <SelectTrigger className="w-32">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="1">默认格式</SelectItem>
                  <SelectItem value="2">苹果CMS</SelectItem>
                </SelectContent>
              </Select>
              <Input
                placeholder="采集地址"
                value={form.collect_url}
                onChange={(e) => setForm({ ...form, collect_url: e.target.value })}
                required
                className="min-w-[280px] flex-1"
              />
            </div>
            <div className="flex flex-wrap items-center gap-2">
              <Input
                placeholder="API Key"
                value={form.api_key}
                onChange={(e) => setForm({ ...form, api_key: e.target.value })}
                className="max-w-xs"
              />
              <Input
                placeholder="cron 表达式(空=不定时)"
                value={form.cron_expr}
                onChange={(e) => setForm({ ...form, cron_expr: e.target.value })}
                className="max-w-xs"
              />
              <Select value={form.play_source_id} onValueChange={(v) => setForm({ ...form, play_source_id: v ?? '0' })}>
                <SelectTrigger className="w-40">
                  <SelectValue placeholder="绑定播放源" />
                </SelectTrigger>
                <SelectContent>
                  {playSources.map((ps) => (
                    <SelectItem key={ps.id} value={String(ps.id)}>{ps.name}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
              <Select value={form.status} onValueChange={(v) => setForm({ ...form, status: v ?? '1' })}>
                <SelectTrigger className="w-28">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="1">启用</SelectItem>
                  <SelectItem value="0">禁用</SelectItem>
                </SelectContent>
              </Select>
              <Button type="submit" size="sm">{selectedId ? '保存源' : '新增源'}</Button>
              <Button type="button" variant="outline" size="sm" onClick={() => void load()}>
                <RefreshCw data-icon="inline-start" />
                刷新
              </Button>
            </div>
          </form>

          <div className="rounded-md border">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="w-16">ID</TableHead>
                  <TableHead>名称</TableHead>
                  <TableHead className="w-24">类型</TableHead>
                  <TableHead className="w-20">状态</TableHead>
                  <TableHead className="w-40">最后采集</TableHead>
                  <TableHead className="w-48">操作</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {sources.map((s) => (
                  <TableRow key={s.id}>
                    <TableCell>{s.id}</TableCell>
                    <TableCell>
                      <div className="font-medium">{s.name}</div>
                      <div className="max-w-[320px] truncate text-xs text-muted-foreground">{s.collect_url}</div>
                    </TableCell>
                    <TableCell>{s.type === 2 ? '苹果CMS' : '默认'}</TableCell>
                    <TableCell><StatusBadge status={s.status} /></TableCell>
                    <TableCell className="text-xs text-muted-foreground">{s.last_collect_at || '-'}</TableCell>
                    <TableCell>
                      <div className="flex flex-wrap gap-1">
                        <Button size="sm" variant="ghost" onClick={() => void editSource(s)}>
                          编辑/映射
                        </Button>
                        <Button size="sm" variant="ghost" onClick={() => void start(s.id)}>
                          <Play data-icon="inline-start" />
                          开始
                        </Button>
                        <Button size="sm" variant="ghost" onClick={() => void stop(s.id)}>
                          <Square data-icon="inline-start" />
                          停止
                        </Button>
                        <Button size="sm" variant="ghost" onClick={() => setDeleteId(s.id)}>
                          <Trash2 data-icon="inline-start" />
                          删除
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        </CardContent>
      </Card>

      {selectedId > 0 && (
        <Card>
          <CardHeader>
            <CardTitle>分类映射（JSON 数组）— 源 #{selectedId}</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="mb-2 text-sm text-muted-foreground">
              示例：{`[{"external_category":"1","category_id":11}]`}。外部分类键对应苹果 type_id 或默认 category 字段。
            </p>
            <p className="mb-4 text-sm text-muted-foreground">
              可用系统分类：{flatCats.map((c) => `${c.id}:${c.name}`).join('， ') || '暂无'}
            </p>
            <Textarea
              rows={8}
              value={mapText}
              onChange={(e) => setMapText(e.target.value)}
              className="mb-4 font-mono"
            />
            <Button size="sm" onClick={() => void saveMaps()}>保存映射</Button>
            {maps.length > 0 && (
              <span className="ml-4 text-sm text-muted-foreground">当前映射 {maps.length} 条</span>
            )}
          </CardContent>
        </Card>
      )}

      <Card>
        <CardHeader>
          <CardTitle>采集日志</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="rounded-md border">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="w-16">ID</TableHead>
                  <TableHead className="w-20">源</TableHead>
                  <TableHead className="w-20">状态</TableHead>
                  <TableHead className="w-20">总数</TableHead>
                  <TableHead className="w-20">成功</TableHead>
                  <TableHead className="w-20">失败</TableHead>
                  <TableHead className="w-24">耗时ms</TableHead>
                  <TableHead>时间</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {logs.map((l) => (
                  <TableRow key={l.id}>
                    <TableCell>{l.id}</TableCell>
                    <TableCell>{l.source_id}</TableCell>
                    <TableCell>{l.status}</TableCell>
                    <TableCell>{l.total_count}</TableCell>
                    <TableCell>{l.success_count}</TableCell>
                    <TableCell>{l.failed_count}</TableCell>
                    <TableCell>{l.duration_ms}</TableCell>
                    <TableCell className="text-xs text-muted-foreground">{l.created_at || '-'}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        </CardContent>
      </Card>

      <ConfirmDialog
        open={deleteId !== null}
        onOpenChange={(open) => { if (!open) setDeleteId(null) }}
        title="删除采集源"
        description="确认删除该采集源？此操作不可撤销。"
        destructive
        onConfirm={confirmDelete}
      />
    </PageContainer>
  )
}
