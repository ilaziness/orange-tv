import { useCallback, useEffect, useRef, useState } from 'react'
import type { AdminLiveStreamerItem } from '@orange-tv/shared'
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
import { RefreshCw, Search, UserMinus, UserPlus } from 'lucide-react'
import { toast } from 'sonner'
import { DEFAULT_PAGE_SIZE } from '@/lib/constants'
import { UserPickerDialog } from './UserPickerDialog'

type RoomStatusFilter = 'all' | '0' | '1' | '2'

export default function LiveStreamersPage() {
  const [items, setItems] = useState<AdminLiveStreamerItem[]>([])
  const [keyword, setKeyword] = useState('')
  const [roomStatus, setRoomStatus] = useState<RoomStatusFilter>('all')
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(false)
  const [revokeId, setRevokeId] = useState<number | null>(null)
  const [revoking, setRevoking] = useState(false)
  const [pickerOpen, setPickerOpen] = useState(false)

  const keywordRef = useRef(keyword)
  const roomStatusRef = useRef(roomStatus)
  const pageRef = useRef(page)

  useEffect(() => {
    keywordRef.current = keyword
  }, [keyword])
  useEffect(() => {
    roomStatusRef.current = roomStatus
  }, [roomStatus])
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
      if (roomStatusRef.current !== 'all') {
        query.room_status = Number(roomStatusRef.current)
      }
      const res = await adminApi.listLiveStreamers(query)
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

  async function confirmRevoke() {
    if (revokeId === null) return
    setRevoking(true)
    try {
      await adminApi.revokeLiveStreamer(revokeId)
      toast.success('已撤销主播资格')
      setRevokeId(null)
      await load(page)
    } catch (err) {
      toast.error(errorMessage(err))
    } finally {
      setRevoking(false)
    }
  }

  const roomOptions: { value: RoomStatusFilter; label: string }[] = [
    { value: 'all', label: '全部房间状态' },
    { value: '0', label: '空闲' },
    { value: '1', label: '直播中' },
    { value: '2', label: '禁播' },
  ]

  return (
    <PageContainer>
      <Card>
        <CardHeader>
          <CardTitle>主播管理</CardTitle>
          <CardAction>
            <div className="flex flex-wrap gap-2">
              <Button size="sm" onClick={() => setPickerOpen(true)} disabled={loading}>
                <UserPlus data-icon="inline-start" />
                添加主播
              </Button>
              <Button
                size="sm"
                variant="outline"
                onClick={() => void load(page)}
                disabled={loading}
              >
                {loading ? (
                  <Spinner data-icon="inline-start" />
                ) : (
                  <RefreshCw data-icon="inline-start" />
                )}
                刷新
              </Button>
            </div>
          </CardAction>
        </CardHeader>
        <CardContent>
          <div className="mb-4 flex flex-wrap gap-2">
            <Input
              placeholder="昵称/邮箱"
              value={keyword}
              onChange={(e) => setKeyword(e.target.value)}
              className="max-w-xs"
              onKeyDown={(e) => {
                if (e.key === 'Enter') void load(1)
              }}
              disabled={loading}
            />
            <Select
              value={roomStatus}
              onValueChange={(v) => setRoomStatus((v ?? 'all') as RoomStatusFilter)}
            >
              <SelectTrigger className="w-40" disabled={loading}>
                <SelectValue placeholder="房间状态">
                  {roomOptions.find((o) => o.value === roomStatus)?.label}
                </SelectValue>
              </SelectTrigger>
              <SelectContent>
                <SelectGroup>
                  {roomOptions.map((o) => (
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
                      <TableHead>主播</TableHead>
                      <TableHead>直播间</TableHead>
                      <TableHead>分类</TableHead>
                      <TableHead className="w-24">房间状态</TableHead>
                      <TableHead className="w-36">成为主播</TableHead>
                      <TableHead className="w-28">操作</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {items.map((item) => (
                      <TableRow key={item.user_id}>
                        <TableCell>
                          <div className="font-medium">{item.nickname || '-'}</div>
                          <div className="text-xs text-muted-foreground">
                            {item.email || item.user_str_id || `UID ${item.user_id}`}
                          </div>
                        </TableCell>
                        <TableCell>
                          <div className="max-w-xs truncate" title={item.room_title}>
                            {item.room_title || '-'}
                          </div>
                          <div className="text-xs text-muted-foreground">
                            房间 #{item.room_id || '-'}
                          </div>
                        </TableCell>
                        <TableCell>{item.category_name || '-'}</TableCell>
                        <TableCell>
                          {item.room_id ? (
                            <RoomStatusBadge status={item.room_status} />
                          ) : (
                            <span className="text-muted-foreground">-</span>
                          )}
                        </TableCell>
                        <TableCell className="text-muted-foreground">
                          {item.streamer_at || '-'}
                        </TableCell>
                        <TableCell>
                          <Button
                            size="sm"
                            variant="ghost"
                            onClick={() => setRevokeId(item.user_id)}
                          >
                            <UserMinus data-icon="inline-start" />
                            撤销
                          </Button>
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
                <EmptyTitle>暂无主播</EmptyTitle>
                <EmptyDescription>可通过审核申请或「添加主播」加入列表</EmptyDescription>
              </EmptyHeader>
            </Empty>
          )}
        </CardContent>
      </Card>

      <UserPickerDialog
        open={pickerOpen}
        onOpenChange={setPickerOpen}
        onGranted={() => void load(1)}
      />

      <ConfirmDialog
        open={revokeId !== null}
        onOpenChange={(open) => {
          if (!open && !revoking) setRevokeId(null)
        }}
        title="撤销主播资格"
        description="将结束进行中的直播并取消主播身份，房间与历史场次保留，用户可再次申请。"
        destructive
        loading={revoking}
        onConfirm={() => void confirmRevoke()}
      />
    </PageContainer>
  )
}
