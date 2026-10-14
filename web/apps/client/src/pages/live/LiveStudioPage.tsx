import { useCallback, useEffect, useState } from 'react'
import { Link, Navigate } from 'react-router'
import type { ClientLiveCategory, ClientMyLiveRoom } from '@orange-tv/shared'
import { LiveRoomStatusBanned, LiveRoomStatusLiving } from '@orange-tv/shared'
import { useAuthStore } from '@/store/auth'
import { useSettingsStore } from '@/store/settings'
import { clientApi, errorMessage } from '@/lib/api'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Field, FieldDescription, FieldGroup, FieldLabel } from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import {
  InputGroup,
  InputGroupAddon,
  InputGroupButton,
  InputGroupInput,
} from '@/components/ui/input-group'
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Separator } from '@/components/ui/separator'
import { Skeleton } from '@/components/ui/skeleton'
import { Spinner } from '@/components/ui/spinner'
import { Textarea } from '@/components/ui/textarea'
import { AlertCircleIcon, CopyIcon, RadioIcon } from 'lucide-react'
import { toast } from 'sonner'
import { usePageTitle } from '@/hooks/usePageTitle'
import { usePageSeo } from '@/hooks/usePageSeo'

const roomStatusLabel: Record<number, string> = {
  0: '空闲',
  1: '直播中',
  2: '已禁播',
}

async function copyText(label: string, value: string) {
  if (!value) {
    toast.error(`${label}为空`)
    return
  }
  try {
    await navigator.clipboard.writeText(value)
    toast.success(`已复制${label}`)
  } catch {
    toast.error('复制失败，请手动选择文本')
  }
}

export function Component() {
  const profile = useAuthStore((s) => s.profile)
  const liveEnabled = useSettingsStore((s) => s.feature.live_enabled)

  const [room, setRoom] = useState<ClientMyLiveRoom | null>(null)
  const [categories, setCategories] = useState<ClientLiveCategory[]>([])
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [starting, setStarting] = useState(false)
  const [stopping, setStopping] = useState(false)
  const [regenerating, setRegenerating] = useState(false)
  const [error, setError] = useState('')
  const [form, setForm] = useState({
    title: '',
    cover: '',
    notice: '',
    category_id: 0,
  })

  usePageTitle('主播工作室')
  usePageSeo({ title: '主播工作室', path: '/live/studio', noindex: true })

  const load = useCallback(async () => {
    setLoading(true)
    setError('')
    try {
      const [roomRes, catRes] = await Promise.all([
        clientApi.myLiveRoom(),
        clientApi.liveCategories(),
      ])
      const data = roomRes.data
      setRoom(data || null)
      setCategories(catRes.data || [])
      if (data) {
        setForm({
          title: data.title || '',
          cover: data.cover || '',
          notice: data.notice || '',
          category_id: data.category_id || 0,
        })
      }
    } catch (err) {
      setRoom(null)
      setError(errorMessage(err))
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    if (profile?.is_streamer) void load()
  }, [profile?.is_streamer, load])

  if (!liveEnabled) {
    return <Navigate to="/" replace />
  }

  if (profile && !profile.is_streamer) {
    return <Navigate to="/profile" replace />
  }

  const update = (field: keyof typeof form, value: string | number) => {
    setForm((prev) => ({ ...prev, [field]: value }))
  }

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault()
    const title = form.title.trim()
    if (!title) {
      toast.error('请填写直播间标题')
      return
    }
    if (title.length > 100) {
      toast.error('标题最长100个字符')
      return
    }
    if (form.cover.trim().length > 500) {
      toast.error('封面URL最长500个字符')
      return
    }
    if (form.notice.trim().length > 500) {
      toast.error('公告最长500个字符')
      return
    }
    if (!form.category_id) {
      toast.error('请选择直播分类')
      return
    }

    setSaving(true)
    try {
      const res = await clientApi.updateMyLiveRoom({
        title,
        cover: form.cover.trim(),
        notice: form.notice.trim(),
        category_id: form.category_id,
      })
      if (res.data) {
        setRoom(res.data)
        toast.success('直播间资料已保存')
      } else {
        toast.error(res.message || '保存失败')
      }
    } catch (err) {
      toast.error(errorMessage(err))
    } finally {
      setSaving(false)
    }
  }

  const handleUploadCover = async (file: File | undefined) => {
    if (!file) return
    if (!file.type.startsWith('image/')) {
      toast.error('请选择图片文件')
      return
    }
    try {
      const res = await clientApi.uploadMedia(file)
      if (res.data?.url) {
        update('cover', res.data.url)
        toast.success('封面已上传')
      } else {
        toast.error(res.message || '上传失败')
      }
    } catch (err) {
      toast.error(errorMessage(err))
    }
  }

  const handleStart = async () => {
    setStarting(true)
    try {
      const body: { category_id?: number; title?: string } = {}
      if (form.category_id) body.category_id = form.category_id
      const title = form.title.trim()
      if (title) body.title = title
      const res = await clientApi.startLive(body)
      if (res.data) {
        toast.success('已开播，请在 OBS 中推流')
        await load()
      } else {
        toast.error(res.message || '开播失败')
      }
    } catch (err) {
      toast.error(errorMessage(err))
    } finally {
      setStarting(false)
    }
  }

  const handleStop = async () => {
    setStopping(true)
    try {
      await clientApi.stopLive()
      toast.success('已停播')
      await load()
    } catch (err) {
      toast.error(errorMessage(err))
    } finally {
      setStopping(false)
    }
  }

  const handleRegenerate = async () => {
    if (!window.confirm('重置后旧密钥将失效，确定继续？')) return
    setRegenerating(true)
    try {
      const res = await clientApi.regenerateLiveStreamKey()
      if (res.data) {
        setRoom(res.data)
        toast.success('推流密钥已重置')
      } else {
        toast.error(res.message || '重置失败')
      }
    } catch (err) {
      toast.error(errorMessage(err))
    } finally {
      setRegenerating(false)
    }
  }

  const isLiving = room?.status === LiveRoomStatusLiving
  const isBanned = room?.status === LiveRoomStatusBanned

  return (
    <div className="mx-auto flex max-w-2xl flex-col gap-6">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div className="flex flex-col gap-1">
          <h1 className="text-2xl font-bold">主播工作室</h1>
          <p className="text-sm text-muted-foreground">编辑直播间资料，配置 OBS 推流并开播</p>
        </div>
        <Button variant="outline" size="sm" nativeButton={false} render={<Link to="/live" />}>
          观看直播
        </Button>
      </div>

      {loading ? (
        <div className="flex flex-col gap-4">
          <Skeleton className="h-16 w-full rounded-lg" />
          <Skeleton className="h-40 w-full rounded-lg" />
          <Skeleton className="h-64 w-full rounded-lg" />
        </div>
      ) : error ? (
        <Alert variant="destructive">
          <AlertCircleIcon />
          <AlertTitle>加载失败</AlertTitle>
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      ) : !room ? (
        <Alert>
          <AlertCircleIcon />
          <AlertTitle>暂无直播间</AlertTitle>
          <AlertDescription>主播资格已通过，但直播间尚未创建，请联系管理员。</AlertDescription>
        </Alert>
      ) : (
        <>
          <div className="flex flex-wrap items-center gap-2 rounded-lg border p-4">
            <RadioIcon className="size-5 text-muted-foreground" />
            <span className="text-sm font-medium">房间状态</span>
            <Badge variant={isLiving ? 'destructive' : isBanned ? 'secondary' : 'outline'}>
              {roomStatusLabel[room.status] || '未知'}
            </Badge>
            {isLiving && room.started_at ? (
              <span className="text-xs text-muted-foreground">开播于 {room.started_at}</span>
            ) : null}
            <div className="ml-auto flex flex-wrap gap-2">
              {isLiving ? (
                <Button
                  variant="destructive"
                  size="sm"
                  disabled={stopping}
                  onClick={() => void handleStop()}
                >
                  {stopping ? <Spinner data-icon="inline-start" /> : null}
                  {stopping ? '停播中...' : '停播'}
                </Button>
              ) : (
                <Button
                  size="sm"
                  disabled={starting || isBanned || !room.push_server}
                  onClick={() => void handleStart()}
                >
                  {starting ? <Spinner data-icon="inline-start" /> : null}
                  {starting ? '开播中...' : '开始直播'}
                </Button>
              )}
            </div>
          </div>

          {isBanned ? (
            <Alert variant="destructive">
              <AlertCircleIcon />
              <AlertTitle>直播间已禁播</AlertTitle>
              <AlertDescription>当前无法开播，请联系管理员解禁。</AlertDescription>
            </Alert>
          ) : null}

          {!isBanned && !room.push_server ? (
            <Alert>
              <AlertCircleIcon />
              <AlertTitle>推流服务器未配置</AlertTitle>
              <AlertDescription>
                管理员尚未配置推流地址，暂时无法开播。密钥可先查看，待配置完成后再推流。
              </AlertDescription>
            </Alert>
          ) : null}

          <section className="flex flex-col gap-4 rounded-lg border p-4">
            <h2 className="text-lg font-semibold">推流信息</h2>
            <p className="text-sm text-muted-foreground">
              在 OBS 中填写服务器与推流密钥；开播前可先配置。请勿泄露密钥。
            </p>
            <FieldGroup>
              <Field>
                <FieldLabel>推流服务器</FieldLabel>
                <InputGroup>
                  <InputGroupInput value={room.push_server || ''} readOnly />
                  <InputGroupAddon align="inline-end">
                    <InputGroupButton
                      type="button"
                      size="icon-sm"
                      onClick={() => void copyText('推流服务器', room.push_server)}
                    >
                      <CopyIcon />
                    </InputGroupButton>
                  </InputGroupAddon>
                </InputGroup>
              </Field>
              <Field>
                <FieldLabel>推流密钥</FieldLabel>
                <InputGroup>
                  <InputGroupInput value={room.stream_key || ''} readOnly />
                  <InputGroupAddon align="inline-end">
                    <InputGroupButton
                      type="button"
                      size="icon-sm"
                      onClick={() => void copyText('推流密钥', room.stream_key)}
                    >
                      <CopyIcon />
                    </InputGroupButton>
                  </InputGroupAddon>
                </InputGroup>
                <FieldDescription>
                  <Button
                    type="button"
                    variant="link"
                    className="h-auto p-0"
                    disabled={regenerating || isLiving}
                    onClick={() => void handleRegenerate()}
                  >
                    {regenerating ? <Spinner data-icon="inline-start" /> : null}
                    {regenerating ? '重置中...' : '重置密钥'}
                  </Button>
                  {isLiving ? '（直播中不可重置）' : null}
                </FieldDescription>
              </Field>
            </FieldGroup>
          </section>

          <Separator />

          <section className="flex flex-col gap-4">
            <h2 className="text-lg font-semibold">直播间资料</h2>
            <form onSubmit={(e) => void handleSave(e)} noValidate>
              <FieldGroup>
                <Field>
                  <FieldLabel htmlFor="live-title">标题</FieldLabel>
                  <Input
                    id="live-title"
                    value={form.title}
                    onChange={(e) => update('title', e.target.value)}
                    placeholder="直播间标题"
                    maxLength={100}
                  />
                </Field>
                <Field>
                  <FieldLabel>分类</FieldLabel>
                  <Select
                    value={form.category_id || null}
                    onValueChange={(v) => update('category_id', Number(v))}
                  >
                    <SelectTrigger className="w-full">
                      <SelectValue placeholder="选择分类" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectGroup>
                        {categories.map((cat) => (
                          <SelectItem key={cat.id} value={cat.id}>
                            {cat.name}
                          </SelectItem>
                        ))}
                      </SelectGroup>
                    </SelectContent>
                  </Select>
                </Field>
                <Field>
                  <FieldLabel htmlFor="live-cover">封面 URL</FieldLabel>
                  <Input
                    id="live-cover"
                    value={form.cover}
                    onChange={(e) => update('cover', e.target.value)}
                    placeholder="https://"
                    maxLength={500}
                  />
                  <FieldDescription>
                    <label className="cursor-pointer text-primary underline-offset-4 hover:underline">
                      上传图片
                      <input
                        type="file"
                        accept="image/*"
                        className="hidden"
                        onChange={(e) => void handleUploadCover(e.target.files?.[0])}
                      />
                    </label>
                  </FieldDescription>
                </Field>
                <Field>
                  <FieldLabel htmlFor="live-notice">公告</FieldLabel>
                  <Textarea
                    id="live-notice"
                    value={form.notice}
                    onChange={(e) => update('notice', e.target.value)}
                    placeholder="直播间公告（可选）"
                    maxLength={500}
                  />
                </Field>
              </FieldGroup>
              <div className="mt-6">
                <Button type="submit" disabled={saving}>
                  {saving ? <Spinner data-icon="inline-start" /> : null}
                  {saving ? '保存中...' : '保存资料'}
                </Button>
              </div>
            </form>
          </section>
        </>
      )}
    </div>
  )
}
