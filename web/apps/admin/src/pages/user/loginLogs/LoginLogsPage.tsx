import { useCallback, useEffect, useRef, useState } from 'react'
import { adminApi, errorMessage } from '@/lib/api'
import { PageContainer, Pagination } from '@/components/shared'
import type { LoginLogItem } from '@orange-tv/shared'
import {
  Card,
  CardAction,
  CardContent,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import { Empty, EmptyHeader, EmptyTitle, EmptyDescription } from '@/components/ui/empty'
import { Spinner } from '@/components/ui/spinner'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Search, RefreshCw } from 'lucide-react'
import { toast } from 'sonner'
import { DEFAULT_PAGE_SIZE } from '@/lib/constants'

export default function LoginLogsPage() {
  const [list, setList] = useState<LoginLogItem[]>([])
  const [username, setUsername] = useState('')
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [loading, setLoading] = useState(false)
  const usernameRef = useRef(username)
  const pageRef = useRef(page)

  useEffect(() => { usernameRef.current = username }, [username])
  useEffect(() => { pageRef.current = page }, [page])

  const load = useCallback(async (p = pageRef.current, u = usernameRef.current) => {
    setLoading(true)
    try {
      const res = await adminApi.listLoginLogs({
        page: p,
        page_size: DEFAULT_PAGE_SIZE,
        username: u || undefined,
      })
      setList(res.data.list || [])
      setTotal(res.data.total || 0)
      setPage(res.data.page || p)
    } catch (err) {
      toast.error(errorMessage(err))
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => { void load(1) }, [load])

  const hasNext = page * DEFAULT_PAGE_SIZE < total

  return (
    <PageContainer>
      <Card>
        <CardHeader>
          <CardTitle>登录日志</CardTitle>
          <CardAction>
            <Button size="sm" variant="outline" onClick={() => void load(page)} disabled={loading}>
              {loading ? <Spinner data-icon="inline-start" /> : <RefreshCw data-icon="inline-start" />}
              刷新
            </Button>
          </CardAction>
        </CardHeader>
        <CardContent>
          <div className="mb-4 flex gap-2">
            <Input
              placeholder="用户名"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              className="max-w-xs"
              onKeyDown={(e) => { if (e.key === 'Enter') void load(1) }}
            />
            <Button variant="outline" size="sm" onClick={() => void load(1)} disabled={loading}>
              {loading ? <Spinner data-icon="inline-start" /> : <Search data-icon="inline-start" />}
              查询
            </Button>
          </div>
          {loading && list.length === 0 ? (
            <div className="flex flex-col gap-2">
              {Array.from({ length: 5 }).map((_, i) => (
                <Skeleton key={i} className="h-12 w-full" />
              ))}
            </div>
          ) : list.length > 0 ? (
            <>
              <div className="relative rounded-md border">
                {loading && (
                  <div className="absolute inset-0 z-10 flex items-center justify-center bg-background/50">
                    <Spinner />
                  </div>
                )}
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead className="w-16">ID</TableHead>
                      <TableHead className="w-20">类型</TableHead>
                      <TableHead>用户名</TableHead>
                      <TableHead className="w-20">状态</TableHead>
                      <TableHead className="w-32">IP</TableHead>
                      <TableHead className="w-40">时间</TableHead>
                      <TableHead>User-Agent</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {list.map((l) => (
                      <TableRow key={l.id}>
                        <TableCell>{l.id}</TableCell>
                        <TableCell>{l.user_type === 1 ? '管理员' : '用户'}</TableCell>
                        <TableCell className="font-medium">{l.username}</TableCell>
                        <TableCell>
                          <Badge variant={l.status === 1 ? 'default' : 'destructive'}>
                            {l.status === 1 ? '成功' : '失败'}
                          </Badge>
                        </TableCell>
                        <TableCell className="text-xs text-muted-foreground">{l.ip_address || '-'}</TableCell>
                        <TableCell className="text-xs text-muted-foreground">{l.created_at || '-'}</TableCell>
                        <TableCell className="max-w-xs truncate text-xs text-muted-foreground">{l.user_agent || '-'}</TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>
              <Pagination
                page={page}
                total={total}
                pageSize={DEFAULT_PAGE_SIZE}
                hasNext={hasNext}
                loading={loading}
                onFirst={() => void load(1)}
                onPrev={() => void load(page - 1)}
                onNext={() => void load(page + 1)}
                onLast={() => void load(Math.ceil(total / DEFAULT_PAGE_SIZE))}
              />
            </>
          ) : (
            <Empty className="py-8">
              <EmptyHeader>
                <EmptyTitle>暂无数据</EmptyTitle>
                <EmptyDescription>暂无登录日志记录</EmptyDescription>
              </EmptyHeader>
            </Empty>
          )}
        </CardContent>
      </Card>
    </PageContainer>
  )
}
