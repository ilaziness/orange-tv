import { useState } from 'react'
import { useNavigate, Link } from 'react-router'
import { isValidUsername, sanitizeUsernameInput } from '@orange-tv/shared'
import { clientApi, errorMessage } from '@/lib/api'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Field, FieldGroup, FieldLabel, FieldError } from '@/components/ui/field'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Spinner } from '@/components/ui/spinner'
import { AlertCircleIcon } from 'lucide-react'

export function Component() {
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [error, setError] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const navigate = useNavigate()

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    const u = username.trim()
    const p = password.trim()
    const cp = confirmPassword.trim()
    if (!isValidUsername(u)) {
      setError('用户名只能由 2-15 位字母或数字组成')
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
    setSubmitting(true)
    try {
      await clientApi.register(u, p)
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
          <CardDescription>创建账号以使用完整功能</CardDescription>
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
