import { useEffect, useState } from 'react'
import type * as React from 'react'
import { z } from 'zod'
import { adminApi, errorMessage } from '@/lib/api'
import { useAuthStore } from '@/store/auth'
import { PageContainer } from '@/components/shared'
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Field, FieldGroup, FieldLabel } from '@/components/ui/field'
import { Skeleton } from '@/components/ui/skeleton'
import { Spinner } from '@/components/ui/spinner'
import { Save, KeyRound } from 'lucide-react'
import { toast } from 'sonner'

const profileSchema = z.object({
  nickname: z.string().max(50, '昵称长度不能超过50'),
  email: z.string().max(100, '邮箱长度不能超过100').refine(
    (v) => v === '' || z.string().email().safeParse(v).success,
    '邮箱格式不正确',
  ),
  avatar: z.string().max(500, '头像URL长度不能超过500'),
})

const passwordSchema = z
  .object({
    old_password: z.string().min(6, '密码至少 6 位'),
    new_password: z.string().min(6, '密码至少 6 位').max(72, '密码长度不能超过72'),
    confirm_password: z.string().min(6, '密码至少 6 位'),
  })
  .refine((data) => data.new_password === data.confirm_password, {
    message: '两次密码不一致',
    path: ['confirm_password'],
  })
  .refine((data) => data.new_password !== data.old_password, {
    message: '新密码不能与旧密码相同',
    path: ['new_password'],
  })

export default function SettingsPage() {
  const profile = useAuthStore((s) => s.profile)
  const updateProfile = useAuthStore((s) => s.updateProfile)

  const [form, setForm] = useState({
    nickname: '',
    email: '',
    avatar: '',
  })
  const [loading, setLoading] = useState(false)
  const [submitting, setSubmitting] = useState(false)

  const [pwdForm, setPwdForm] = useState({
    old_password: '',
    new_password: '',
    confirm_password: '',
  })
  const [pwdSubmitting, setPwdSubmitting] = useState(false)

  async function load() {
    setLoading(true)
    try {
      const res = await adminApi.profile()
      const p = res.data
      setForm({
        nickname: p.nickname || '',
        email: p.email || '',
        avatar: p.avatar || '',
      })
    } catch (err) {
      toast.error(errorMessage(err))
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    void load()
  }, [])

  async function saveProfile(e: React.SyntheticEvent<HTMLFormElement>) {
    e.preventDefault()
    if (submitting) return
    const result = profileSchema.safeParse(form)
    if (!result.success) {
      toast.error(result.error.issues[0]?.message || '表单校验失败')
      return
    }
    setSubmitting(true)
    try {
      await updateProfile(result.data)
      toast.success('资料已保存')
    } catch (err) {
      toast.error(errorMessage(err))
    } finally {
      setSubmitting(false)
    }
  }

  async function changePassword(e: React.SyntheticEvent<HTMLFormElement>) {
    e.preventDefault()
    if (pwdSubmitting) return
    const result = passwordSchema.safeParse(pwdForm)
    if (!result.success) {
      toast.error(result.error.issues[0]?.message || '表单校验失败')
      return
    }
    setPwdSubmitting(true)
    try {
      await adminApi.changePassword({
        old_password: result.data.old_password,
        new_password: result.data.new_password,
      })
      toast.success('密码已修改')
      setPwdForm({ old_password: '', new_password: '', confirm_password: '' })
    } catch (err) {
      toast.error(errorMessage(err))
    } finally {
      setPwdSubmitting(false)
    }
  }

  return (
    <PageContainer>
      <Card>
        <CardHeader>
          <CardTitle>个人资料</CardTitle>
        </CardHeader>
        <CardContent>
          {loading ? (
            <div className="flex flex-col gap-4">
              {Array.from({ length: 4 }).map((_, i) => (
                <Skeleton key={i} className="h-10 w-full" />
              ))}
            </div>
          ) : (
            <form onSubmit={saveProfile} className="flex flex-col gap-4">
              <FieldGroup>
                <Field data-disabled={true}>
                  <FieldLabel htmlFor="username">用户名</FieldLabel>
                  <Input
                    id="username"
                    value={profile?.username || ''}
                    disabled
                  />
                </Field>
                <Field data-disabled={submitting ? true : undefined}>
                  <FieldLabel htmlFor="nickname">昵称</FieldLabel>
                  <Input
                    id="nickname"
                    placeholder="请输入昵称（可选）"
                    value={form.nickname}
                    onChange={(e) => setForm((prev) => ({ ...prev, nickname: e.target.value }))}
                    disabled={submitting}
                  />
                </Field>
                <Field data-disabled={submitting ? true : undefined}>
                  <FieldLabel htmlFor="email">邮箱</FieldLabel>
                  <Input
                    id="email"
                    type="email"
                    placeholder="请输入邮箱（可选）"
                    value={form.email}
                    onChange={(e) => setForm((prev) => ({ ...prev, email: e.target.value }))}
                    disabled={submitting}
                  />
                </Field>
                <Field data-disabled={submitting ? true : undefined}>
                  <FieldLabel htmlFor="avatar">头像 URL</FieldLabel>
                  <Input
                    id="avatar"
                    placeholder="请输入头像图片地址（可选）"
                    value={form.avatar}
                    onChange={(e) => setForm((prev) => ({ ...prev, avatar: e.target.value }))}
                    disabled={submitting}
                  />
                </Field>
              </FieldGroup>
              <div className="flex justify-end">
                <Button type="submit" disabled={submitting}>
                  {submitting ? <Spinner data-icon="inline-start" /> : <Save data-icon="inline-start" />}
                  {submitting ? '保存中...' : '保存'}
                </Button>
              </div>
            </form>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>修改密码</CardTitle>
        </CardHeader>
        <CardContent>
          <form onSubmit={changePassword} className="flex flex-col gap-4">
            <FieldGroup>
              <Field data-disabled={pwdSubmitting ? true : undefined}>
                <FieldLabel htmlFor="old_password">旧密码</FieldLabel>
                <Input
                  id="old_password"
                  type="password"
                  placeholder="请输入旧密码"
                  value={pwdForm.old_password}
                  onChange={(e) => setPwdForm((prev) => ({ ...prev, old_password: e.target.value }))}
                  disabled={pwdSubmitting}
                  required
                  minLength={6}
                />
              </Field>
              <Field data-disabled={pwdSubmitting ? true : undefined}>
                <FieldLabel htmlFor="new_password">新密码</FieldLabel>
                <Input
                  id="new_password"
                  type="password"
                  placeholder="请输入新密码（至少 6 位）"
                  value={pwdForm.new_password}
                  onChange={(e) => setPwdForm((prev) => ({ ...prev, new_password: e.target.value }))}
                  disabled={pwdSubmitting}
                  required
                  minLength={6}
                />
              </Field>
              <Field data-disabled={pwdSubmitting ? true : undefined}>
                <FieldLabel htmlFor="confirm_password">确认新密码</FieldLabel>
                <Input
                  id="confirm_password"
                  type="password"
                  placeholder="请再次输入新密码"
                  value={pwdForm.confirm_password}
                  onChange={(e) => setPwdForm((prev) => ({ ...prev, confirm_password: e.target.value }))}
                  disabled={pwdSubmitting}
                  required
                  minLength={6}
                />
              </Field>
            </FieldGroup>
            <div className="flex justify-end">
              <Button type="submit" disabled={pwdSubmitting}>
                {pwdSubmitting ? <Spinner data-icon="inline-start" /> : <KeyRound data-icon="inline-start" />}
                {pwdSubmitting ? '修改中...' : '修改密码'}
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>
    </PageContainer>
  )
}
