import { useEffect, useState } from 'react'
import { useAuthStore } from '@/store/auth'
import { clientApi, errorMessage } from '@/lib/api'
import { Button } from '@/components/ui/button'
import { Field, FieldDescription, FieldGroup, FieldLabel } from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import { Separator } from '@/components/ui/separator'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { ToggleGroup, ToggleGroupItem } from '@/components/ui/toggle-group'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { Badge } from '@/components/ui/badge'
import { toast } from 'sonner'
import type { LoginHistoryItem, UserProfile } from '@orange-tv/shared'
import { usePageTitle } from '@/hooks/usePageTitle'
import { usePageSeo } from '@/hooks/usePageSeo'
import { PasswordInput } from '@/components/shared'

const statusMap: Record<number, string> = {
  1: '正常',
  0: '禁用',
}

const loginStatusMap: Record<number, string> = {
  1: '成功',
  2: '失败',
}

type Tab = 'basic' | 'password' | 'history'

export function Component() {
  const profile = useAuthStore((s) => s.profile)
  const loadProfile = useAuthStore((s) => s.loadProfile)
  const [tab, setTab] = useState<Tab>('basic')

  usePageTitle('个人中心')
  usePageSeo({ title: '个人中心', path: '/profile', noindex: true })

  useEffect(() => {
    void loadProfile()
  }, [loadProfile])

  return (
    <main className="container mx-auto p-4 py-6 md:p-8">
      <div className="grid gap-6 md:grid-cols-[12rem_1fr]">
        <ToggleGroup
          value={[tab]}
          onValueChange={(v) => v[0] && setTab(v[0] as Tab)}
          orientation="vertical"
          className="w-full flex-col"
        >
          <ToggleGroupItem value="basic" className="w-full justify-start">
            基本资料
          </ToggleGroupItem>
          <ToggleGroupItem value="password" className="w-full justify-start">
            修改密码
          </ToggleGroupItem>
          <ToggleGroupItem value="history" className="w-full justify-start">
            登录历史
          </ToggleGroupItem>
        </ToggleGroup>

        <div>
          {tab === 'basic' && <BasicInfo profile={profile} loadProfile={loadProfile} />}
          {tab === 'password' && <ChangePassword />}
          {tab === 'history' && <LoginHistory />}
        </div>
      </div>
    </main>
  )
}

function BasicInfo({
  profile,
  loadProfile,
}: {
  profile: UserProfile | null
  loadProfile: () => Promise<void>
}) {
  const [form, setForm] = useState({
    nickname: '',
    email: '',
    avatar: '',
  })
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    if (profile) {
      setForm({
        nickname: profile.nickname || '',
        email: profile.email || '',
        avatar: profile.avatar || '',
      })
    }
  }, [profile])

  const update = (field: keyof typeof form, value: string) => {
    setForm((prev) => ({ ...prev, [field]: value }))
  }

  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
  const isValidURL = (v: string) => {
    try {
      new URL(v)
      return true
    } catch {
      return false
    }
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    const email = form.email.trim()
    const avatar = form.avatar.trim()

    if (email) {
      if (email.length > 20) {
        toast.error('邮箱最长20个字符')
        return
      }
      if (!emailRegex.test(email)) {
        toast.error('邮箱格式不正确')
        return
      }
    }
    if (avatar) {
      if (avatar.length > 120) {
        toast.error('头像URL最长120个字符')
        return
      }
      if (!isValidURL(avatar)) {
        toast.error('头像URL格式不正确')
        return
      }
    }

    setSaving(true)
    try {
      const body: { nickname?: string; email?: string; avatar?: string } = {}
      const nickname = form.nickname.trim()
      if (nickname) body.nickname = nickname
      if (email) body.email = email
      if (avatar) body.avatar = avatar

      const res = await clientApi.updateProfile(body)
      if (res.data) {
        toast.success('资料已更新')
        await loadProfile()
      } else {
        toast.error(res.message || '更新失败')
      }
    } catch (err) {
      toast.error(errorMessage(err))
    } finally {
      setSaving(false)
    }
  }

  const displayName = profile?.nickname || profile?.email || ''
  const fallback = (
    (profile?.nickname?.[0] || profile?.email?.[0] || 'U') as string
  ).toUpperCase()

  return (
    <div className="flex flex-col gap-6">
      {profile ? (
        <>
          <div className="flex items-center gap-4">
            <Avatar className="size-20">
              {profile.avatar ? <AvatarImage src={profile.avatar} alt={displayName} /> : null}
              <AvatarFallback>{fallback}</AvatarFallback>
            </Avatar>
            <div className="flex flex-col gap-1">
              <p className="text-lg font-semibold">{displayName}</p>
              <p className="text-sm text-muted-foreground">邮箱：{profile.email}</p>
              <p className="text-sm text-muted-foreground">用户ID：{profile.str_id}</p>
              <Badge variant={profile.status === 1 ? 'default' : 'secondary'}>
                {statusMap[profile.status] || '未知'}
              </Badge>
            </div>
          </div>

          <form onSubmit={handleSubmit} noValidate>
            <FieldGroup>
              <Field>
                <FieldLabel htmlFor="nickname">昵称</FieldLabel>
                <Input
                  id="nickname"
                  value={form.nickname}
                  onChange={(e) => update('nickname', e.target.value)}
                  placeholder="昵称"
                  maxLength={15}
                />
                <FieldDescription>3-15 个字符，优先展示</FieldDescription>
              </Field>
              <Field>
                <FieldLabel htmlFor="email">邮箱</FieldLabel>
                <Input
                  id="email"
                  type="email"
                  value={form.email}
                  onChange={(e) => update('email', e.target.value)}
                  placeholder="未设置"
                  maxLength={20}
                />
                <FieldDescription>最长20个字符</FieldDescription>
              </Field>
              <Field>
                <FieldLabel htmlFor="avatar">头像 URL</FieldLabel>
                <Input
                  id="avatar"
                  value={form.avatar}
                  onChange={(e) => update('avatar', e.target.value)}
                  placeholder="https://"
                  maxLength={120}
                />
                <FieldDescription>最长120个字符，格式为 http(s)://</FieldDescription>
              </Field>
            </FieldGroup>
            <div className="mt-6">
              <Button type="submit" disabled={saving}>
                {saving ? '保存中...' : '保存资料'}
              </Button>
            </div>
          </form>
        </>
      ) : (
        <p className="text-sm text-muted-foreground">加载中...</p>
      )}
    </div>
  )
}

function ChangePassword() {
  const [form, setForm] = useState({
    current_password: '',
    new_password: '',
    confirm_password: '',
  })
  const [saving, setSaving] = useState(false)

  const update = (field: keyof typeof form, value: string) => {
    setForm((prev) => ({ ...prev, [field]: value }))
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (form.current_password.length < 5 || form.current_password.length > 30) {
      toast.error('当前密码长度应为5-30个字符')
      return
    }
    if (form.new_password.length < 5 || form.new_password.length > 30) {
      toast.error('新密码长度应为5-30个字符')
      return
    }
    if (form.new_password !== form.confirm_password) {
      toast.error('两次输入的新密码不一致')
      return
    }
    setSaving(true)
    try {
      const res = await clientApi.changePassword({
        current_password: form.current_password,
        new_password: form.new_password,
      })
      if (res.code === 0) {
        toast.success('密码已修改')
        setForm({ current_password: '', new_password: '', confirm_password: '' })
      } else {
        toast.error(res.message || '修改失败')
      }
    } catch (err) {
      toast.error(errorMessage(err))
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="flex flex-col gap-6">
      <h2 className="text-lg font-semibold">修改密码</h2>
      <form onSubmit={handleSubmit} noValidate>
        <FieldGroup>
          <Field>
            <FieldLabel htmlFor="current_password">当前密码</FieldLabel>
            <PasswordInput
              id="current_password"
              value={form.current_password}
              onChange={(e) => update('current_password', e.target.value)}
              placeholder="请输入当前密码"
              minLength={5}
              maxLength={30}
            />
          </Field>
          <Field>
            <FieldLabel htmlFor="new_password">新密码</FieldLabel>
            <PasswordInput
              id="new_password"
              value={form.new_password}
              onChange={(e) => update('new_password', e.target.value)}
              placeholder="5-30 位"
              minLength={5}
              maxLength={30}
            />
          </Field>
          <Field>
            <FieldLabel htmlFor="confirm_password">确认新密码</FieldLabel>
            <PasswordInput
              id="confirm_password"
              value={form.confirm_password}
              onChange={(e) => update('confirm_password', e.target.value)}
              placeholder="再次输入新密码"
              minLength={5}
              maxLength={30}
            />
          </Field>
        </FieldGroup>
        <div className="mt-6">
          <Button type="submit" disabled={saving}>
            {saving ? '保存中...' : '修改密码'}
          </Button>
        </div>
      </form>
    </div>
  )
}

function LoginHistory() {
  const [list, setList] = useState<LoginHistoryItem[]>([])
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [total, setTotal] = useState(0)
  const [totalPages, setTotalPages] = useState(0)
  const [loading, setLoading] = useState(false)

  const load = async (p: number, ps: number) => {
    setLoading(true)
    try {
      const res = await clientApi.loginHistory(p, ps)
      if (res.data) {
        setList(res.data.list || [])
        setTotal(res.data.total || 0)
        setTotalPages(res.data.total_pages || 0)
      } else {
        toast.error(res.message || '加载失败')
      }
    } catch (err) {
      toast.error(errorMessage(err))
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    void load(page, pageSize)
  }, [page, pageSize])

  const canPrev = page > 1
  const canNext = page < totalPages

  return (
    <div className="flex flex-col gap-6">
      <h2 className="text-lg font-semibold">最近 3 个月登录历史</h2>
      {loading ? (
        <p className="text-sm text-muted-foreground">加载中...</p>
      ) : list.length === 0 ? (
        <p className="text-sm text-muted-foreground">暂无登录记录</p>
      ) : (
        <div className="flex flex-col gap-4">
          <div className="flex flex-col gap-3">
            {list.map((item) => (
              <div
                key={item.id}
                className="flex flex-col gap-1 rounded-lg border p-3 sm:flex-row sm:items-center sm:justify-between"
              >
                <div className="flex flex-col gap-0.5">
                  <p className="text-sm font-medium">{item.created_at}</p>
                  <p className="text-xs text-muted-foreground">
                    {item.ip} · {item.user_agent.slice(0, 60)}
                  </p>
                </div>
                <Badge variant={item.status === 1 ? 'default' : 'destructive'}>
                  {loginStatusMap[item.status] || '未知'}
                </Badge>
              </div>
            ))}
          </div>

          <Separator />

          <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
            <p className="text-sm text-muted-foreground">
              共 {total} 条，{totalPages} 页
            </p>
            <div className="flex items-center gap-2">
              <span className="text-sm text-muted-foreground">每页</span>
              <Select
                value={pageSize}
                onValueChange={(v) => {
                  setPageSize(Number(v))
                  setPage(1)
                }}
              >
                <SelectTrigger className="h-9 w-auto min-w-16">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value={5}>5</SelectItem>
                  <SelectItem value={10}>10</SelectItem>
                  <SelectItem value={20}>20</SelectItem>
                  <SelectItem value={50}>50</SelectItem>
                </SelectContent>
              </Select>
              <Button
                variant="outline"
                size="sm"
                onClick={() => setPage((p) => p - 1)}
                disabled={!canPrev || loading}
              >
                上一页
              </Button>
              <span className="text-sm">
                {page} / {totalPages}
              </span>
              <Button
                variant="outline"
                size="sm"
                onClick={() => setPage((p) => p + 1)}
                disabled={!canNext || loading}
              >
                下一页
              </Button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
