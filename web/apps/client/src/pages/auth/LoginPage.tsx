import { useEffect, useRef, useState } from 'react'
import { useNavigate, Link } from 'react-router'
import { isValidEmail, sanitizeEmailInput } from '@orange-tv/shared'
import { clientApi, errorMessage } from '@/lib/api'
import { useAuthStore } from '@/store/auth'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Field, FieldGroup, FieldLabel } from '@/components/ui/field'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Spinner } from '@/components/ui/spinner'
import { AlertCircleIcon } from 'lucide-react'
import { usePageTitle } from '@/hooks/usePageTitle'
import { CaptchaInput } from '@/components/auth/CaptchaInput'

export function Component() {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [captchaId, setCaptchaId] = useState('')
  const [captcha, setCaptcha] = useState('')
  const [error, setError] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const navigate = useNavigate()
  const setToken = useAuthStore((s) => s.setToken)
  const loadProfile = useAuthStore((s) => s.loadProfile)
  const token = useAuthStore((s) => s.token)
  const captchaRefreshRef = useRef<() => void>(() => {})

  usePageTitle('登录')

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
    if (!isValidEmail(em) || p.length < 5 || p.length > 30) {
      setError('邮箱或密码格式不正确')
      return
    }
    if (!captchaId || !captcha) {
      setError('请输入验证码')
      return
    }
    setSubmitting(true)
    try {
      const res = await clientApi.login(em, p, captchaId, captcha)
      setToken(res.data.access_token)
      await loadProfile()
      navigate('/', { replace: true })
    } catch (err) {
      setError(errorMessage(err))
      captchaRefreshRef.current()
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="flex min-h-[60vh] items-center justify-center">
      <Card className="w-full max-w-sm">
        <CardHeader>
          <CardTitle>登录</CardTitle>
          <CardDescription>登录后享受更多功能</CardDescription>
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
                  placeholder="请输入密码"
                  minLength={5}
                  maxLength={30}
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  required
                />
              </Field>
              <CaptchaInput
                scene="login"
                idPrefix="login-captcha"
                onCaptchaChange={(id, ans) => {
                  setCaptchaId(id)
                  setCaptcha(ans)
                }}
                refreshRef={captchaRefreshRef}
                disabled={submitting}
              />
              <Button type="submit" disabled={submitting} className="w-full">
                {submitting ? <Spinner data-icon="inline-start" /> : null}
                登录
              </Button>
            </FieldGroup>
          </form>
          <p className="mt-4 text-center text-sm text-muted-foreground">
            还没有账号？{' '}
            <Link to="/register" className="text-primary underline-offset-4 hover:underline">
              注册
            </Link>
          </p>
        </CardContent>
      </Card>
    </div>
  )
}
