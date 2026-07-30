import { useCallback, useEffect, useRef, useState } from 'react'
import { adminApi, errorMessage } from '@/lib/api'
import { PageContainer, Pagination } from '@/components/shared'
import type { AdminLoginLogItem, ApiResponse, PageData, UserLoginLogItem } from '@orange-tv/shared'
import {
  Card,
  CardAction,
  CardContent,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs'
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

type LoginLogRow = AdminLoginLogItem | UserLoginLogItem

export default function LoginLogsPage() {
  return (
    <PageContainer>
      <Tabs defaultValue="admin">
        <TabsList>
          <TabsTrigger value="admin">管理员登录日志</TabsTrigger>
          <TabsTrigger value="user">用户登录日志</TabsTrigger>
        </TabsList>
        <TabsContent value="admin" className="mt-4">
          <LoginLogTabContent
            title="管理员登录日志"
            emptyDescription="暂无管理员登录日志记录"
            fetchFn={(query) => adminApi.listAdminLoginLogs(query)}
          />
        </TabsContent>
        <TabsContent value="user" className="mt-4">
          <LoginLogTabContent
            title="用户登录日志"
            emptyDescription="暂无用户登录日志记录"
            fetchFn={(query) => adminApi.listUserLoginLogs(query)}
          />
        </TabsContent>
      </Tabs>
    </PageContainer>
  )
}

function LoginLogTabContent({
  title,
  emptyDescription,
  fetchFn,
}: {
  title: string
  emptyDescription: string
  fetchFn: (query: Record<string, string | number | undefined>) => Promise<ApiResponse<PageData<LoginLogRow>>>
}) {
  const [list, setList] = useState<LoginLogRow[]>([])
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
      const res = await fetchFn({
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
  }, [fetchFn])

  useEffect(() => { void load(1) }, [load])

  const hasNext = page * DEFAULT_PAGE_SIZE < total

  return (
    <Card>
      <CardHeader>
        <CardTitle>{title}</CardTitle>
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
                      <TableCell className="font-medium">{l.username}</TableCell>
                      <TableCell>
                        <Badge variant={l.status === 1 ? 'default' : 'destructive'}>
                          {l.status === 1 ? '成功' : '失败'}
                        </Badge>
                      </TableCell>
                      <TableCell className="text-xs text-muted-foreground">{l.ip || '-'}</TableCell>
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
              <EmptyDescription>{emptyDescription}</EmptyDescription>
            </EmptyHeader>
          </Empty>
        )}
      </CardContent>
    </Card>
  )
}
