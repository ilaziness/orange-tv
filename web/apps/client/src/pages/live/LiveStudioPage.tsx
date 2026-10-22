import { useCallback, useEffect, useState } from 'react'
import { Link, Navigate } from 'react-router'
import type { ClientLiveCategory, ClientMyLiveRoom } from '@orange-tv/shared'
import {
  ApiError,
  isLiveMediaDisconnectError,
  LiveRoomStatusBanned,
  LiveRoomStatusLiving,
} from '@orange-tv/shared'
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

/** Keep sub only when it is an enabled child of the selected L1; otherwise 0. */
function enabledSubCategoryId(
  categories: ClientLiveCategory[],
  categoryId: number,
  subCategoryId: number,
): number {
  if (!categoryId || !subCategoryId) return 0
  const children = categories.find((c) => c.id === categoryId)?.children || []
  return children.some((c) => c.id === subCategoryId) ? subCategoryId : 0
}

function bareStreamName(key: string): string {
  const i = key.indexOf('?')
  return (i === -1 ? key : key.slice(0, i)).trim()
}

/** Keep the key already shown for OBS unless the stream name, server, or signature shape changed. */
function keepDisplayedPushKey(prev: ClientMyLiveRoom | null, next: ClientMyLiveRoom): ClientMyLiveRoom {
  if (!prev?.stream_key || !next.stream_key) return next
  if (bareStreamName(prev.stream_key) !== bareStreamName(next.stream_key)) return next
  if ((prev.push_server || '') !== (next.push_server || '')) return next
  if (prev.stream_key.includes('?') !== next.stream_key.includes('?')) return next
  return { ...next, stream_key: prev.stream_key, push_expires_at: prev.push_expires_at }
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
  const [refreshingPush, setRefreshingPush] = useState(false)
  const [nowMs, setNowMs] = useState(() => Date.now())
  const [error, setError] = useState('')
  const [form, setForm] = useState({
    title: '',
    cover: '',
    notice: '',
    category_id: 0,
    sub_category_id: 0,
  })

  usePageTitle('主播工作室')
  usePageSeo({ title: '主播工作室', path: '/live/studio', noindex: true })

  const syncFormFromRoom = (data: ClientMyLiveRoom, cats: ClientLiveCategory[]) => {
    const storedL1 = data.category_id || 0
    const categoryId = cats.some((c) => c.id === storedL1) ? storedL1 : 0
    setForm({
      title: data.title || '',
      cover: data.cover || '',
      notice: data.notice || '',
      category_id: categoryId,
      sub_category_id: enabledSubCategoryId(cats, categoryId, data.sub_category_id || 0),
    })
  }

  const load = useCallback(async () => {
    setLoading(true)
    setError('')
    try {
      const [roomRes, catRes] = await Promise.all([
        clientApi.myLiveRoom(),
        clientApi.liveCategories(),
      ])
      const data = roomRes.data
      setRoom((prev) => (data ? keepDisplayedPushKey(prev, data) : null))
      const cats = catRes.data || []
      setCategories(cats)
      if (data) syncFormFromRoom(data, cats)
    } catch (err) {
      setRoom(null)
      setError(errorMessage(err))
    } finally {
      setLoading(false)
    }
  }, [])

  /** Reload room without full-page skeleton; used after stop. */
  const refreshRoom = async () => {
    const res = await clientApi.myLiveRoom()
    const data = res.data
    if (!data) return
    setRoom(data)
    syncFormFromRoom(data, categories)
  }

  useEffect(() => {
    if (profile?.is_streamer) void load()
  }, [profile?.is_streamer, load])

  useEffect(() => {
    if (!room?.push_expires_at) return
    const id = window.setInterval(() => setNowMs(Date.now()), 15000)
    return () => window.clearInterval(id)
  }, [room?.push_expires_at])

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
    const selectedL1 = categories.find((c) => c.id === form.category_id)
    const l2Children = selectedL1?.children || []
    const subCategoryId = enabledSubCategoryId(categories, form.category_id, form.sub_category_id)
    if (l2Children.length > 0 && !subCategoryId) {
      toast.error('请选择二级分类')
      return
    }

    setSaving(true)
    try {
      const res = await clientApi.updateMyLiveRoom({
        title,
        cover: form.cover.trim(),
        notice: form.notice.trim(),
        category_id: form.category_id,
        sub_category_id: subCategoryId,
      })
      const saved = res.data
      if (saved) {
        setRoom((prev) => keepDisplayedPushKey(prev, saved))
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
    if (!form.category_id) {
      toast.error('请选择直播分类')
      return
    }
    const selectedL1 = categories.find((c) => c.id === form.category_id)
    const l2Children = selectedL1?.children || []
    const subCategoryId = enabledSubCategoryId(categories, form.category_id, form.sub_category_id)
    if (l2Children.length > 0 && !subCategoryId) {
      toast.error('请选择二级分类')
      return
    }
    setStarting(true)
    try {
      const body: { category_id: number; sub_category_id: number; title?: string } = {
        category_id: form.category_id,
        sub_category_id: subCategoryId,
      }
      const title = form.title.trim()
      if (title) body.title = title
      const res = await clientApi.startLive(body)
      if (res.data) {
        const started = res.data
        const l1 = categories.find((c) => c.id === form.category_id)
        const l2 = l1?.children?.find((c) => c.id === subCategoryId)
        setRoom((prev) =>
          prev
            ? {
                ...prev,
                status: started.status,
                stream_key: started.stream_key,
                push_server: started.push_server || prev.push_server,
                push_expires_at: started.push_expires_at,
                current_session_id: started.session_id,
                started_at: started.started_at,
                title: title || prev.title,
                category_id: form.category_id || prev.category_id,
                category_name: l1?.name || prev.category_name,
                sub_category_id: subCategoryId,
                sub_category_name: l2?.name || '',
              }
            : prev,
        )
        toast.success('已开播，请在 OBS 中推流')
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
      await refreshRoom()
    } catch (err) {
      if (isLiveMediaDisconnectError(err)) {
        toast.warning(err instanceof ApiError ? err.message : '直播已结束，请到媒体服务器手动断开')
        await refreshRoom()
      } else {
        toast.error(errorMessage(err))
      }
    } finally {
      setStopping(false)
    }
  }

  const handleRegenerate = async () => {
    if (!window.confirm('重置后旧流名将失效，确定继续？')) return
    setRegenerating(true)
    try {
      const res = await clientApi.regenerateLiveStreamKey()
      if (res.data) {
        setRoom(res.data)
        toast.success('流名已重置，开播后可获取推流密钥')
      } else {
        toast.error(res.message || '重置失败')
      }
    } catch (err) {
      toast.error(errorMessage(err))
    } finally {
      setRegenerating(false)
    }
  }

  const handleRefreshPush = async () => {
    setRefreshingPush(true)
    try {
      const res = await clientApi.myLiveRoom()
      if (res.data) {
        setRoom(res.data)
        toast.success('已刷新推流密钥。当前这条推流不会被切断，OBS 重连时再使用新密钥')
      } else {
        toast.error(res.message || '刷新失败')
      }
    } catch (err) {
      toast.error(errorMessage(err))
    } finally {
      setRefreshingPush(false)
    }
  }

  const isLiving = room?.status === LiveRoomStatusLiving
  const isBanned = room?.status === LiveRoomStatusBanned
  const selectedL1 = categories.find((c) => c.id === form.category_id)
  const l2Children = selectedL1?.children || []
  const l1Items: { label: string; value: string | null }[] = [
    { label: '选择一级分类', value: null },
    ...categories.map((cat) => ({
      value: String(cat.id),
      label: cat.name,
    })),
  ]
  const l2Items: { label: string; value: string | null }[] = [
    { label: '选择二级分类', value: null },
    ...l2Children.map((cat) => ({
      value: String(cat.id),
      label: cat.name,
    })),
  ]

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
                管理员尚未配置推流地址，暂时无法开播。
              </AlertDescription>
            </Alert>
          ) : null}

          <section className="flex flex-col gap-4 rounded-lg border p-4">
            <h2 className="text-lg font-semibold">推流信息</h2>
            <p className="text-sm text-muted-foreground">
              先点「开始直播」，再把服务器和推流密钥填进 OBS。未开播时不会给出推流密钥。
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
                  {room.push_expires_at ? (
                    <span className="block">
                      推流签名到期：{new Date(room.push_expires_at * 1000).toLocaleString()}。
                      当前这条推流不会因为到期被切断，只有 OBS 重连才需要新密钥。
                      {room.push_expires_at * 1000 - nowMs <= 60_000
                        ? ' 即将过期，请手动刷新后再重连。'
                        : ''}
                    </span>
                  ) : null}
                  {room.push_expires_at ? (
                    <Button
                      type="button"
                      variant="link"
                      className="h-auto p-0"
                      disabled={refreshingPush}
                      onClick={() => void handleRefreshPush()}
                    >
                      {refreshingPush ? <Spinner data-icon="inline-start" /> : null}
                      {refreshingPush ? '刷新中...' : '刷新推流密钥'}
                    </Button>
                  ) : null}
                  <Button
                    type="button"
                    variant="link"
                    className="h-auto p-0"
                    disabled={regenerating || isLiving}
                    onClick={() => void handleRegenerate()}
                  >
                    {regenerating ? <Spinner data-icon="inline-start" /> : null}
                    {regenerating ? '重置中...' : '重置流名'}
                  </Button>
                  {isLiving ? '（直播中不可重置流名）' : null}
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
                  <FieldLabel htmlFor="live-category-l1">一级分类</FieldLabel>
                  <Select
                    items={l1Items}
                    value={form.category_id ? String(form.category_id) : null}
                    onValueChange={(v) => {
                      const nextL1 = v ? Number(v) : 0
                      setForm((prev) => ({
                        ...prev,
                        category_id: nextL1,
                        sub_category_id: 0,
                      }))
                    }}
                  >
                    <SelectTrigger id="live-category-l1" className="w-full">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectGroup>
                        {l1Items.map((item) => (
                          <SelectItem key={item.value ?? 'none'} value={item.value}>
                            {item.label}
                          </SelectItem>
                        ))}
                      </SelectGroup>
                    </SelectContent>
                  </Select>
                </Field>
                {l2Children.length > 0 ? (
                  <Field>
                    <FieldLabel htmlFor="live-category-l2">二级分类</FieldLabel>
                    <Select
                      items={l2Items}
                      value={form.sub_category_id ? String(form.sub_category_id) : null}
                      onValueChange={(v) => update('sub_category_id', v ? Number(v) : 0)}
                    >
                      <SelectTrigger id="live-category-l2" className="w-full">
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectGroup>
                          {l2Items.map((item) => (
                            <SelectItem key={item.value ?? 'none'} value={item.value}>
                              {item.label}
                            </SelectItem>
                          ))}
                        </SelectGroup>
                      </SelectContent>
                    </Select>
                  </Field>
                ) : null}
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
