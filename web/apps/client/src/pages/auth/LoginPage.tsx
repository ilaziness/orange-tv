import { useEffect, useRef, useState } from 'react'
import { useNavigate, Link } from 'react-router'
import { isValidUsername, sanitizeUsernameInput } from '@orange-tv/shared'
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

export function Component() {
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const navigate = useNavigate()
  const setToken = useAuthStore((s) => s.setToken)
  const loadProfile = useAuthStore((s) => s.loadProfile)
  const token = useAuthStore((s) => s.token)

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
    const u = username.trim()
    const p = password.trim()
    if (!isValidUsername(u) || p.length < 5 || p.length > 30) {
      setError('用户名或密码格式不正确')
      return
    }
    setSubmitting(true)
    try {
      const res = await clientApi.login(u, p)
      setToken(res.data.access_token)
      await loadProfile()
      navigate('/', { replace: true })
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
                <FieldLabel htmlFor="username">用户名</FieldLabel>
                <Input
                  id="username"
                  placeholder="请输入用户名"
                  maxLength={15}
                  pattern="[a-zA-Z0-9]{2,15}"
                  value={username}
                  onChange={(e) => setUsername(sanitizeUsernameInput(e.target.value))}
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
