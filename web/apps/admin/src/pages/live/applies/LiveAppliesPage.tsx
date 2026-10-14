import { useCallback, useEffect, useRef, useState } from 'react'
import type { AdminLiveApplyItem } from '@orange-tv/shared'
import { adminApi, errorMessage } from '@/lib/api'
import { PageContainer, Pagination, PromptDialog } from '@/components/shared'
import { ApplyStatusBadge } from '@/pages/live/_shared/liveStatus'
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
import { Check, X, RefreshCw, Search } from 'lucide-react'
import { toast } from 'sonner'
import { DEFAULT_PAGE_SIZE } from '@/lib/constants'

type StatusFilter = 'all' | '0' | '1' | '2'
type ReviewAction = { id: number; kind: 'approve' | 'reject' }

export default function LiveAppliesPage() {
  const [items, setItems] = useState<AdminLiveApplyItem[]>([])
  const [keyword, setKeyword] = useState('')
  const [status, setStatus] = useState<StatusFilter>('all')
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(false)
  const [review, setReview] = useState<ReviewAction | null>(null)
  const [reviewLoading, setReviewLoading] = useState(false)

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
      const res = await adminApi.listLiveApplies(query)
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

  async function submitReview(remark: string) {
    if (!review) return
    setReviewLoading(true)
    try {
      if (review.kind === 'approve') {
        await adminApi.approveLiveApply(review.id, remark)
        toast.success('已通过申请')
      } else {
        await adminApi.rejectLiveApply(review.id, remark)
        toast.success('已拒绝申请')
      }
      setReview(null)
      await load(page)
    } catch (err) {
      toast.error(errorMessage(err))
    } finally {
      setReviewLoading(false)
    }
  }

  const statusOptions: { value: StatusFilter; label: string }[] = [
    { value: 'all', label: '全部状态' },
    { value: '0', label: '待审' },
    { value: '1', label: '通过' },
    { value: '2', label: '拒绝' },
  ]

  return (
    <PageContainer>
      <Card>
        <CardHeader>
          <CardTitle>主播申请</CardTitle>
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
              placeholder="昵称/邮箱/联系方式"
              value={keyword}
              onChange={(e) => setKeyword(e.target.value)}
              className="max-w-xs"
              onKeyDown={(e) => {
                if (e.key === 'Enter') void load(1)
              }}
              disabled={loading}
            />
            <Select
              value={status}
              onValueChange={(v) => setStatus((v ?? 'all') as StatusFilter)}
            >
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
                      <TableHead>用户</TableHead>
                      <TableHead>申请理由</TableHead>
                      <TableHead>联系方式</TableHead>
                      <TableHead className="w-20">状态</TableHead>
                      <TableHead>审核备注</TableHead>
                      <TableHead className="w-36">申请时间</TableHead>
                      <TableHead className="w-44">操作</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {items.map((item) => (
                      <TableRow key={item.id}>
                        <TableCell>{item.id}</TableCell>
                        <TableCell>
                          <div className="font-medium">{item.nickname || '-'}</div>
                          <div className="text-xs text-muted-foreground">
                            {item.email || item.user_str_id || `UID ${item.user_id}`}
                          </div>
                        </TableCell>
                        <TableCell className="max-w-xs">
                          <div className="line-clamp-2" title={item.reason}>
                            {item.reason || '-'}
                          </div>
                        </TableCell>
                        <TableCell>{item.contact || '-'}</TableCell>
                        <TableCell>
                          <ApplyStatusBadge status={item.status} />
                        </TableCell>
                        <TableCell className="max-w-[10rem] truncate" title={item.review_remark}>
                          {item.review_remark || '-'}
                        </TableCell>
                        <TableCell className="text-muted-foreground">
                          {item.created_at || '-'}
                        </TableCell>
                        <TableCell>
                          {item.status === 0 ? (
                            <div className="flex gap-1">
                              <Button
                                size="sm"
                                variant="ghost"
                                onClick={() => setReview({ id: item.id, kind: 'approve' })}
                              >
                                <Check data-icon="inline-start" />
                                通过
                              </Button>
                              <Button
                                size="sm"
                                variant="ghost"
                                onClick={() => setReview({ id: item.id, kind: 'reject' })}
                              >
                                <X data-icon="inline-start" />
                                拒绝
                              </Button>
                            </div>
                          ) : (
                            <span className="text-xs text-muted-foreground">
                              {item.reviewed_at || '-'}
                            </span>
                          )}
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
                <EmptyTitle>暂无申请</EmptyTitle>
                <EmptyDescription>还没有主播资格申请记录</EmptyDescription>
              </EmptyHeader>
            </Empty>
          )}
        </CardContent>
      </Card>

      <PromptDialog
        open={review !== null}
        onOpenChange={(open) => {
          if (!open && !reviewLoading) setReview(null)
        }}
        title={review?.kind === 'approve' ? '通过申请' : '拒绝申请'}
        description={
          review?.kind === 'approve'
            ? '可选填写审核备注。通过后将授予主播资格并自动创建直播间（若尚无）。'
            : '建议填写拒绝原因，用户可再次申请。'
        }
        label="审核备注（可选）"
        placeholder="备注"
        confirmText={review?.kind === 'approve' ? '确认通过' : '确认拒绝'}
        loading={reviewLoading}
        onConfirm={submitReview}
      />
    </PageContainer>
  )
}
