import { useCallback, useEffect, useRef, useState } from 'react'
import type { AdminLiveSessionItem } from '@orange-tv/shared'
import { adminApi, errorMessage } from '@/lib/api'
import { PageContainer, Pagination } from '@/components/shared'
import { endReasonLabel } from '@/pages/live/_shared/liveLabels'
import { Card, CardAction, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
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
import { RefreshCw, Search } from 'lucide-react'
import { toast } from 'sonner'
import { DEFAULT_PAGE_SIZE } from '@/lib/constants'

export default function LiveSessionsPage() {
  const [items, setItems] = useState<AdminLiveSessionItem[]>([])
  const [roomId, setRoomId] = useState('')
  const [userId, setUserId] = useState('')
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(false)

  const roomIdRef = useRef(roomId)
  const userIdRef = useRef(userId)
  const pageRef = useRef(page)

  useEffect(() => {
    roomIdRef.current = roomId
  }, [roomId])
  useEffect(() => {
    userIdRef.current = userId
  }, [userId])
  useEffect(() => {
    pageRef.current = page
  }, [page])

  const load = useCallback(async (p = pageRef.current) => {
    setLoading(true)
    try {
      const query: Record<string, string | number | undefined> = {
        page: p,
        page_size: DEFAULT_PAGE_SIZE,
      }
      if (roomIdRef.current) query.room_id = Number(roomIdRef.current)
      if (userIdRef.current) query.user_id = Number(userIdRef.current)
      const res = await adminApi.listLiveSessions(query)
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

  return (
    <PageContainer>
      <Card>
        <CardHeader>
          <CardTitle>直播场次</CardTitle>
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
              placeholder="房间 ID"
              value={roomId}
              onChange={(e) => setRoomId(e.target.value.replace(/[^0-9]/g, ''))}
              className="w-32"
              onKeyDown={(e) => {
                if (e.key === 'Enter') void load(1)
              }}
              disabled={loading}
            />
            <Input
              placeholder="用户 ID"
              value={userId}
              onChange={(e) => setUserId(e.target.value.replace(/[^0-9]/g, ''))}
              className="w-32"
              onKeyDown={(e) => {
                if (e.key === 'Enter') void load(1)
              }}
              disabled={loading}
            />
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
                      <TableHead>标题（快照）</TableHead>
                      <TableHead>主播</TableHead>
                      <TableHead>分类</TableHead>
                      <TableHead className="w-20">房间</TableHead>
                      <TableHead className="w-36">开始</TableHead>
                      <TableHead className="w-36">结束</TableHead>
                      <TableHead className="w-28">结束原因</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {items.map((item) => (
                      <TableRow key={item.id}>
                        <TableCell>{item.id}</TableCell>
                        <TableCell className="max-w-xs truncate" title={item.title}>
                          {item.title || '-'}
                        </TableCell>
                        <TableCell>
                          <div>{item.streamer_nickname || '-'}</div>
                          <div className="text-xs text-muted-foreground">UID {item.user_id}</div>
                        </TableCell>
                        <TableCell>{item.category_name || '-'}</TableCell>
                        <TableCell>{item.room_id}</TableCell>
                        <TableCell className="text-muted-foreground">
                          {item.started_at || '-'}
                        </TableCell>
                        <TableCell className="text-muted-foreground">
                          {item.ended_at || (
                            <Badge variant="outline">进行中</Badge>
                          )}
                        </TableCell>
                        <TableCell>{endReasonLabel(item.end_reason)}</TableCell>
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
                <EmptyTitle>暂无场次</EmptyTitle>
                <EmptyDescription>主播开播后会产生场次记录</EmptyDescription>
              </EmptyHeader>
            </Empty>
          )}
        </CardContent>
      </Card>
    </PageContainer>
  )
}
