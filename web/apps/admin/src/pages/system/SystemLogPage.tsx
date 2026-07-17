import { useCallback, useEffect, useRef, useState } from 'react'
import { adminApi, errorMessage } from '@/lib/api'
import { PageContainer } from '@/components/shared'
import type { LoginLogItem, SystemLogItem } from '@orange-tv/shared'
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
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Search } from 'lucide-react'

export default function SystemLogPage() {
  const [tab, setTab] = useState<'system' | 'login'>('system')
  const [systemLogs, setSystemLogs] = useState<SystemLogItem[]>([])
  const [loginLogs, setLoginLogs] = useState<LoginLogItem[]>([])
  const [module, setModule] = useState('')
  const [username, setUsername] = useState('')
  const [error, setError] = useState('')
  const [total, setTotal] = useState(0)
  const tabRef = useRef(tab)
  const moduleRef = useRef(module)
  const usernameRef = useRef(username)

  useEffect(() => { tabRef.current = tab }, [tab])
  useEffect(() => { moduleRef.current = module }, [module])
  useEffect(() => { usernameRef.current = username }, [username])

  const load = useCallback(async () => {
    setError('')
    try {
      if (tabRef.current === 'system') {
        const res = await adminApi.listSystemLogs({ page: 1, page_size: 50, module: moduleRef.current || undefined })
        setSystemLogs(res.data.list || [])
        setTotal(res.data.total)
      } else {
        const res = await adminApi.listLoginLogs({ page: 1, page_size: 50, username: usernameRef.current || undefined })
        setLoginLogs(res.data.list || [])
        setTotal(res.data.total)
      }
    } catch (err) {
      setError(errorMessage(err))
    }
  }, [])

  useEffect(() => { void load() }, [tab, load])

  return (
    <PageContainer>
      <Card>
        <CardHeader>
          <CardTitle>系统日志</CardTitle>
        </CardHeader>
        <CardContent>
          {error && (
            <Alert variant="destructive" className="mb-4">
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}
          <Tabs value={tab} onValueChange={(v) => setTab(v as 'system' | 'login')}>
            <TabsList className="mb-4">
              <TabsTrigger value="system">操作日志</TabsTrigger>
              <TabsTrigger value="login">登录日志</TabsTrigger>
            </TabsList>
          </Tabs>

          {tab === 'system' ? (
            <>
              <div className="mb-4 flex gap-2">
                <Input
                  placeholder="模块筛选"
                  value={module}
                  onChange={(e) => setModule(e.target.value)}
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
                {systemLogs.map((l) => (
                  <div key={l.id} className="rounded-lg border p-4">
                    <div className="flex items-center gap-2">
                      <Badge variant={String(l.level) === 'error' ? 'destructive' : 'secondary'}>[{l.level}]</Badge>
                      <span className="font-medium">{l.module}/{l.action}</span>
                    </div>
                    <div className="mt-1 text-xs text-muted-foreground">
                      admin={l.admin_id} · {l.ip_address} · {l.created_at}
                    </div>
                    <div className="mt-1 text-sm">{l.content}</div>
                  </div>
                ))}
              </div>
            </>
          ) : (
            <>
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
            </>
          )}
        </CardContent>
      </Card>
    </PageContainer>
  )
}
