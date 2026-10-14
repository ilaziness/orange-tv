import { useCallback, useEffect, useRef, useState } from 'react'
import type { AdminLiveRoomItem } from '@orange-tv/shared'
import { adminApi, errorMessage } from '@/lib/api'
import { PageContainer, Pagination, ConfirmDialog } from '@/components/shared'
import { RoomStatusBadge } from '@/pages/live/_shared/liveStatus'
import { Card, CardAction, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Skeleton } from '@/components/ui/skeleton'
import { Spinner } from '@/components/ui/spinner'
import { Empty, EmptyHeader, EmptyTitle, EmptyDescription } from '@/components/ui/empty'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Ban, CircleStop, RefreshCw, Search, ShieldCheck } from 'lucide-react'
import { toast } from 'sonner'
import { DEFAULT_PAGE_SIZE } from '@/lib/constants'

type StatusFilter = 'all' | '0' | '1' | '2'
type ActionKind = 'force-stop' | 'ban' | 'unban'

export default function LiveRoomsPage() {
  const [items, setItems] = useState<AdminLiveRoomItem[]>([])
  const [keyword, setKeyword] = useState('')
  const [status, setStatus] = useState<StatusFilter>('all')
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(false)
  const [action, setAction] = useState<{ id: number; kind: ActionKind } | null>(null)
  const [acting, setActing] = useState(false)

  const keywordRef = useRef(keyword)
  const statusRef = useRef(status)
  const pageRef = useRef(page)

  useEffect(() => {
    keywordRef.current = keyword
  }, [keyword])
  useEffect(() => {
    statusRef.current = status
  }, [status])
  useEffect(() => {
    pageRef.current = page
  }, [page])

  const load = useCallback(async (p = pageRef.current) => {
    setLoading(true)
    try {
      const query: Record<string, string | number | undefined> = {
        keyword: keywordRef.current || undefined,
        page: p,
        page_size: DEFAULT_PAGE_SIZE,
      }
      if (statusRef.current !== 'all') {
        query.status = Number(statusRef.current)
      }
      const res = await adminApi.listLiveRooms(query)
      setItems(res.data.list || [])
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

  async function confirmAction() {
    if (!action) return
    setActing(true)
    try {
      if (action.kind === 'force-stop') {
        await adminApi.forceStopLiveRoom(action.id)
        toast.success('已强制下播')
      } else if (action.kind === 'ban') {
        await adminApi.banLiveRoom(action.id)
        toast.success('已禁播')
      } else {
        await adminApi.unbanLiveRoom(action.id)
        toast.success('已解禁')
      }
      setAction(null)
      await load(page)
    } catch (err) {
      toast.error(errorMessage(err))
    } finally {
      setActing(false)
    }
  }

  const statusOptions: { value: StatusFilter; label: string }[] = [
    { value: 'all', label: '全部状态' },
    { value: '0', label: '空闲' },
    { value: '1', label: '直播中' },
    { value: '2', label: '禁播' },
  ]

  const actionTitle =
    action?.kind === 'force-stop'
      ? '强制下播'
      : action?.kind === 'ban'
        ? '禁播直播间'
        : '解禁直播间'
  const actionDesc =
    action?.kind === 'force-stop'
      ? '将结束当前场次，房间回到空闲状态。'
      : action?.kind === 'ban'
        ? '若正在直播会先下播，随后标记为禁播，主播无法开播。'
        : '解禁后房间回到空闲，主播可再次开播。'

  return (
    <PageContainer>
      <Card>
        <CardHeader>
          <CardTitle>直播间</CardTitle>
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
          <div className="mb-4 flex flex-wrap gap-2">
            <Input
              placeholder="标题/主播昵称"
              value={keyword}
              onChange={(e) => setKeyword(e.target.value)}
              className="max-w-xs"
              onKeyDown={(e) => {
                if (e.key === 'Enter') void load(1)
              }}
              disabled={loading}
            />
            <Select value={status} onValueChange={(v) => setStatus((v ?? 'all') as StatusFilter)}>
              <SelectTrigger className="w-32" disabled={loading}>
                <SelectValue placeholder="状态">
                  {statusOptions.find((o) => o.value === status)?.label}
                </SelectValue>
              </SelectTrigger>
              <SelectContent>
                <SelectGroup>
                  {statusOptions.map((o) => (
                    <SelectItem key={o.value} value={o.value}>
                      {o.label}
                    </SelectItem>
                  ))}
                </SelectGroup>
              </SelectContent>
            </Select>
            <Button variant="outline" size="sm" onClick={() => void load(1)} disabled={loading}>
              {loading ? <Spinner data-icon="inline-start" /> : <Search data-icon="inline-start" />}
              查询
            </Button>
          </div>

          {loading && items.length === 0 ? (
            <div className="flex flex-col gap-2">
              {Array.from({ length: 5 }).map((_, i) => (
                <Skeleton key={i} className="h-12 w-full" />
              ))}
            </div>
          ) : items.length > 0 ? (
            <>
              <div className="relative rounded-md border">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead className="w-16">ID</TableHead>
                      <TableHead>标题</TableHead>
                      <TableHead>主播</TableHead>
                      <TableHead>分类</TableHead>
                      <TableHead className="w-24">状态</TableHead>
                      <TableHead className="w-36">开播时间</TableHead>
                      <TableHead className="w-56">操作</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {items.map((item) => (
                      <TableRow key={item.id}>
                        <TableCell>{item.id}</TableCell>
                        <TableCell>
                          <div className="max-w-xs truncate font-medium" title={item.title}>
                            {item.title || '-'}
                          </div>
                          {item.stream_key ? (
                            <div
                              className="max-w-xs truncate font-mono text-xs text-muted-foreground"
                              title={item.stream_key}
                            >
                              key: {item.stream_key}
                            </div>
                          ) : null}
                        </TableCell>
                        <TableCell>
                          <div>{item.nickname || '-'}</div>
                          <div className="text-xs text-muted-foreground">
                            UID {item.user_id}
                          </div>
                        </TableCell>
                        <TableCell>{item.category_name || '-'}</TableCell>
                        <TableCell>
                          <RoomStatusBadge status={item.status} />
                        </TableCell>
                        <TableCell className="text-muted-foreground">
                          {item.started_at || '-'}
                        </TableCell>
                        <TableCell>
                          <div className="flex flex-wrap gap-1">
                            {item.status === 1 ? (
                              <Button
                                size="sm"
                                variant="ghost"
                                onClick={() => setAction({ id: item.id, kind: 'force-stop' })}
                              >
                                <CircleStop data-icon="inline-start" />
                                强制下播
                              </Button>
                            ) : null}
                            {item.status === 2 ? (
                              <Button
                                size="sm"
                                variant="ghost"
                                onClick={() => setAction({ id: item.id, kind: 'unban' })}
                              >
                                <ShieldCheck data-icon="inline-start" />
                                解禁
                              </Button>
                            ) : (
                              <Button
                                size="sm"
                                variant="ghost"
                                onClick={() => setAction({ id: item.id, kind: 'ban' })}
                              >
                                <Ban data-icon="inline-start" />
                                禁播
                              </Button>
                            )}
                          </div>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
                {loading && (
                  <div className="absolute inset-0 flex items-center justify-center bg-background/40">
                    <Spinner className="size-5" />
                  </div>
                )}
              </div>
              <Pagination
                page={page}
                total={total}
                pageSize={DEFAULT_PAGE_SIZE}
                loading={loading}
                hasNext={page * DEFAULT_PAGE_SIZE < total}
                onFirst={() => void load(1)}
                onPrev={() => void load(page - 1)}
                onNext={() => void load(page + 1)}
                onLast={() => void load(Math.ceil(total / DEFAULT_PAGE_SIZE))}
              />
            </>
          ) : (
            <Empty className="py-8">
              <EmptyHeader>
                <EmptyTitle>暂无直播间</EmptyTitle>
                <EmptyDescription>主播资格通过后会自动创建直播间</EmptyDescription>
              </EmptyHeader>
            </Empty>
          )}
        </CardContent>
      </Card>

      <ConfirmDialog
        open={action !== null}
        onOpenChange={(open) => {
          if (!open && !acting) setAction(null)
        }}
        title={actionTitle}
        description={actionDesc}
        destructive={action?.kind !== 'unban'}
        loading={acting}
        onConfirm={() => void confirmAction()}
      />
    </PageContainer>
  )
}
