import { useCallback, useEffect, useRef, useState } from 'react'
import { adminApi, errorMessage } from '@/lib/api'
import { PageContainer, Pagination } from '@/components/shared'
import type { AppLogItem, SystemLogItem } from '@orange-tv/shared'
import { Card, CardAction, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
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
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs'
import { Search, RefreshCw, ChevronDown } from 'lucide-react'
import { toast } from 'sonner'
import { DEFAULT_PAGE_SIZE } from '@/lib/constants'

const APP_LOG_PAGE_SIZE = 50

function levelVariant(level: string): 'destructive' | 'secondary' | 'default' {
  const l = level.toLowerCase()
  if (l === 'error' || l === 'dpanic' || l === 'panic' || l === 'fatal') return 'destructive'
  if (l === 'warn' || l === 'warning') return 'secondary'
  return 'default'
}

function levelLabel(level: string): string {
  const l = level.toLowerCase()
  if (l === 'error') return '错误'
  if (l === 'warn' || l === 'warning') return '警告'
  if (l === 'info') return '信息'
  if (l === 'debug') return '调试'
  if (l === 'dpanic' || l === 'panic' || l === 'fatal') return '严重'
  return level || '-'
}

function formatFieldValue(v: unknown): string {
  if (v === null) return 'null'
  if (typeof v === 'string') return v
  if (typeof v === 'number' || typeof v === 'boolean') return String(v)
  try {
    return JSON.stringify(v)
  } catch {
    return String(v)
  }
}

function OperationLogTab() {
  const [list, setList] = useState<SystemLogItem[]>([])
  const [module, setModule] = useState('')
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [loading, setLoading] = useState(false)
  const moduleRef = useRef(module)
  const pageRef = useRef(page)

  useEffect(() => {
    moduleRef.current = module
  }, [module])
  useEffect(() => {
    pageRef.current = page
  }, [page])

  const load = useCallback(async (p = pageRef.current, m = moduleRef.current) => {
    setLoading(true)
    try {
      const res = await adminApi.listSystemLogs({
        page: p,
        page_size: DEFAULT_PAGE_SIZE,
        module: m || undefined,
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

  useEffect(() => {
    void load(1)
  }, [load])

  const hasNext = page * DEFAULT_PAGE_SIZE < total

  return (
    <Card>
      <CardHeader>
        <CardTitle>操作日志</CardTitle>
        <CardAction>
          <Button size="sm" variant="outline" onClick={() => void load(page)} disabled={loading}>
            {loading ? (
              <Spinner data-icon="inline-start" />
            ) : (
              <RefreshCw data-icon="inline-start" />
            )}
            刷新
          </Button>
        </CardAction>
      </CardHeader>
      <CardContent>
        <div className="mb-4 flex gap-2">
          <Input
            placeholder="模块筛选"
            value={module}
            onChange={(e) => setModule(e.target.value)}
            className="max-w-xs"
            onKeyDown={(e) => {
              if (e.key === 'Enter') void load(1)
            }}
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
                    <TableHead className="w-20">级别</TableHead>
                    <TableHead className="w-24">模块</TableHead>
                    <TableHead className="w-24">操作</TableHead>
                    <TableHead className="w-20">管理员</TableHead>
                    <TableHead className="w-32">IP</TableHead>
                    <TableHead className="w-40">时间</TableHead>
                    <TableHead>内容</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {list.map((l) => (
                    <TableRow key={l.id}>
                      <TableCell>{l.id}</TableCell>
                      <TableCell>
                        <Badge variant={l.level >= 3 ? 'destructive' : 'secondary'}>
                          {l.level >= 3 ? '错误' : l.level === 2 ? '警告' : '信息'}
                        </Badge>
                      </TableCell>
                      <TableCell>{l.module || '-'}</TableCell>
                      <TableCell>{l.action || '-'}</TableCell>
                      <TableCell>{l.admin_id}</TableCell>
                      <TableCell className="text-xs text-muted-foreground">
                        {l.ip_address || '-'}
                      </TableCell>
                      <TableCell className="text-xs text-muted-foreground">
                        {l.created_at || '-'}
                      </TableCell>
                      <TableCell className="max-w-xs truncate text-sm">
                        {l.content || '-'}
                      </TableCell>
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
              <EmptyDescription>暂无系统日志记录</EmptyDescription>
            </EmptyHeader>
          </Empty>
        )}
      </CardContent>
    </Card>
  )
}

function AppLogTab() {
  const [list, setList] = useState<AppLogItem[]>([])
  const [hasMore, setHasMore] = useState(false)
  const [nextOffset, setNextOffset] = useState(0)
  const [loading, setLoading] = useState(false)
  const [loadingMore, setLoadingMore] = useState(false)

  const loadInitial = useCallback(async () => {
    setLoading(true)
    try {
      const res = await adminApi.listAppLogs({ limit: APP_LOG_PAGE_SIZE })
      setList(res.data.list || [])
      setHasMore(res.data.has_more)
      setNextOffset(res.data.next_offset)
    } catch (err) {
      toast.error(errorMessage(err))
    } finally {
      setLoading(false)
    }
  }, [])

  const loadMore = useCallback(async () => {
    if (!hasMore || loadingMore) return
    setLoadingMore(true)
    try {
      const res = await adminApi.listAppLogs({ offset: nextOffset, limit: APP_LOG_PAGE_SIZE })
      setList((prev) => [...prev, ...(res.data.list || [])])
      setHasMore(res.data.has_more)
      setNextOffset(res.data.next_offset)
    } catch (err) {
      toast.error(errorMessage(err))
    } finally {
      setLoadingMore(false)
    }
  }, [hasMore, loadingMore, nextOffset])

  useEffect(() => {
    void loadInitial()
  }, [loadInitial])

  return (
    <Card>
      <CardHeader>
        <CardTitle>应用日志</CardTitle>
        <CardAction>
          <Button size="sm" variant="outline" onClick={() => void loadInitial()} disabled={loading}>
            {loading ? (
              <Spinner data-icon="inline-start" />
            ) : (
              <RefreshCw data-icon="inline-start" />
            )}
            刷新
          </Button>
        </CardAction>
      </CardHeader>
      <CardContent>
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
                    <TableHead className="w-44">时间</TableHead>
                    <TableHead className="w-20">级别</TableHead>
                    <TableHead className="w-48">消息</TableHead>
                    <TableHead>详情</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {list.map((l, i) => (
                    <TableRow key={i}>
                      <TableCell className="text-xs text-muted-foreground">
                        {l.time || '-'}
                      </TableCell>
                      <TableCell>
                        <Badge variant={levelVariant(l.level)}>{levelLabel(l.level)}</Badge>
                      </TableCell>
                      <TableCell className="text-sm">{l.msg || '-'}</TableCell>
                      <TableCell className="text-xs text-muted-foreground">
                        {l.fields && Object.keys(l.fields).length > 0 ? (
                          <div className="flex flex-col gap-0.5">
                            {Object.entries(l.fields).map(([k, v]) => (
                              <span key={k}>
                                <span className="text-foreground/70">{k}</span>:{' '}
                                {formatFieldValue(v)}
                              </span>
                            ))}
                          </div>
                        ) : (
                          '-'
                        )}
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
            {hasMore && (
              <div className="mt-4 flex justify-center">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => void loadMore()}
                  disabled={loadingMore}
                >
                  {loadingMore ? (
                    <Spinner data-icon="inline-start" />
                  ) : (
                    <ChevronDown data-icon="inline-start" />
                  )}
                  加载更多
                </Button>
              </div>
            )}
          </>
        ) : (
          <Empty className="py-8">
            <EmptyHeader>
              <EmptyTitle>暂无数据</EmptyTitle>
              <EmptyDescription>暂无应用日志记录</EmptyDescription>
            </EmptyHeader>
          </Empty>
        )}
      </CardContent>
    </Card>
  )
}

export default function SystemLogPage() {
  return (
    <PageContainer>
      <Tabs defaultValue="operation">
        <TabsList>
          <TabsTrigger value="operation">操作日志</TabsTrigger>
          <TabsTrigger value="app">应用日志</TabsTrigger>
        </TabsList>
        <TabsContent value="operation" className="mt-4">
          <OperationLogTab />
        </TabsContent>
        <TabsContent value="app" className="mt-4">
          <AppLogTab />
        </TabsContent>
      </Tabs>
    </PageContainer>
  )
}
