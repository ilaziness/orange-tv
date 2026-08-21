import { useEffect, useRef, useState } from 'react'
import { useNavigate, Link } from 'react-router'
import { isValidEmail, sanitizeEmailInput } from '@orange-tv/shared'
import { clientApi, errorMessage } from '@/lib/api'
import { useAuthStore } from '@/store/auth'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Field, FieldGroup, FieldLabel, FieldError } from '@/components/ui/field'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Spinner } from '@/components/ui/spinner'
import { AlertCircleIcon } from 'lucide-react'
import { usePageTitle } from '@/hooks/usePageTitle'

export function Component() {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [nickname, setNickname] = useState('')
  const [error, setError] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const navigate = useNavigate()
  const token = useAuthStore((s) => s.token)

  usePageTitle('注册')

  const hasRedirected = useRef(false)
  useEffect(() => {
    if (hasRedirected.current) return
    hasRedirected.current = true
    if (token) {
      navigate('/', { replace: true })
    }
  }, [token, navigate])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    const em = email.trim()
    const p = password.trim()
    const cp = confirmPassword.trim()
    const nick = nickname.trim()
    if (!isValidEmail(em)) {
      setError('邮箱格式不正确')
      return
    }
    if (p.length < 5 || p.length > 30) {
      setError('密码长度应为5-30位')
      return
    }
    if (p !== cp) {
      setError('两次密码不一致')
      return
    }
    if (nick && (nick.length < 3 || nick.length > 15)) {
      setError('昵称长度应为3-15位')
      return
    }
    setSubmitting(true)
    try {
      await clientApi.register(em, p, nick || undefined)
      navigate('/login')
    } catch (err) {
      setError(errorMessage(err))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="flex min-h-[60vh] items-center justify-center">
      <Card className="w-full max-w-sm">
        <CardHeader>
          <CardTitle>注册</CardTitle>
          <CardDescription>使用邮箱注册账号</CardDescription>
        </CardHeader>
        <CardContent>
          {error ? (
            <Alert variant="destructive" className="mb-4">
              <AlertCircleIcon />
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          ) : null}
          <form onSubmit={handleSubmit}>
            <FieldGroup>
              <Field>
                <FieldLabel htmlFor="email">邮箱</FieldLabel>
                <Input
                  id="email"
                  type="email"
                  placeholder="请输入邮箱"
                  maxLength={128}
                  value={email}
                  onChange={(e) => setEmail(sanitizeEmailInput(e.target.value))}
                  required
                />
              </Field>
              <Field>
                <FieldLabel htmlFor="password">密码</FieldLabel>
                <Input
                  id="password"
                  type="password"
                  placeholder="5-30 位"
                  minLength={5}
                  maxLength={30}
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  required
                />
              </Field>
              <Field>
                <FieldLabel htmlFor="confirmPassword">确认密码</FieldLabel>
                <Input
                  id="confirmPassword"
                  type="password"
                  placeholder="请再次输入密码"
                  minLength={5}
                  maxLength={30}
                  value={confirmPassword}
                  onChange={(e) => setConfirmPassword(e.target.value)}
                  required
                />
                {password !== confirmPassword && confirmPassword ? (
                  <FieldError>两次密码不一致</FieldError>
                ) : null}
              </Field>
              <Field>
                <FieldLabel htmlFor="nickname">昵称（可选）</FieldLabel>
                <Input
                  id="nickname"
                  placeholder="3-15 位，不填则使用邮箱前缀"
                  minLength={3}
                  maxLength={15}
                  value={nickname}
                  onChange={(e) => setNickname(e.target.value)}
                />
              </Field>
              <Button type="submit" disabled={submitting} className="w-full">
                {submitting ? <Spinner data-icon="inline-start" /> : null}
                注册
              </Button>
            </FieldGroup>
          </form>
          <p className="mt-4 text-center text-sm text-muted-foreground">
            已有账号？{' '}
            <Link to="/login" className="text-primary underline-offset-4 hover:underline">
              登录
            </Link>
          </p>
        </CardContent>
      </Card>
    </div>
  )
}
