import { useCallback, useEffect, useState } from 'react'
import { adminApi, errorMessage } from '@/lib/api'
import { PageContainer } from '@/components/shared'
import type { LoginLogItem } from '@orange-tv/shared'
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Search } from 'lucide-react'

export default function LoginLogsPage() {
  const [loginLogs, setLoginLogs] = useState<LoginLogItem[]>([])
  const [username, setUsername] = useState('')
  const [error, setError] = useState('')
  const [total, setTotal] = useState(0)

  const load = useCallback(async () => {
    setError('')
    try {
      const res = await adminApi.listLoginLogs({
        page: 1,
        page_size: 50,
        username: username || undefined,
      })
      setLoginLogs(res.data.list || [])
      setTotal(res.data.total)
    } catch (err) {
      setError(errorMessage(err))
    }
  }, [username])

  useEffect(() => { void load() }, [load])

  return (
    <PageContainer>
      <Card>
        <CardHeader>
          <CardTitle>登录日志</CardTitle>
        </CardHeader>
        <CardContent>
          {error && (
            <Alert variant="destructive" className="mb-4">
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}
          <div className="mb-4 flex gap-2">
            <Input
              placeholder="用户名"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              className="max-w-xs"
              onKeyDown={(e) => { if (e.key === 'Enter') void load() }}
            />
            <Button variant="outline" size="sm" onClick={() => void load()}>
              <Search data-icon="inline-start" />
              查询
            </Button>
          </div>
          <p className="mb-2 text-sm text-muted-foreground">共 {total} 条</p>
          <div className="flex flex-col gap-3">
            {loginLogs.map((l) => (
              <div key={l.id} className="rounded-lg border p-4">
                <div className="flex items-center gap-2">
                  <span className="font-medium">{l.username}</span>
                  <Badge variant={l.status === 1 ? 'default' : 'destructive'}>
                    {l.status === 1 ? '成功' : '失败'}
                  </Badge>
                </div>
                <div className="mt-1 text-xs text-muted-foreground">
                  {l.ip_address} · {l.created_at}
                </div>
                <div className="mt-1 text-xs text-muted-foreground">{l.user_agent}</div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    </PageContainer>
  )
}
