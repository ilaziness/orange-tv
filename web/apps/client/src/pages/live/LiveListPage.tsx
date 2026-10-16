import { useCallback, useEffect, useMemo, useState } from 'react'
import { Link, useSearchParams } from 'react-router'
import type { ClientLiveCategory, ClientLiveRoomListItem } from '@orange-tv/shared'
import { formatLiveCategoryName } from '@orange-tv/shared'
import { clientApi, errorMessage } from '@/lib/api'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { Empty, EmptyDescription, EmptyHeader, EmptyTitle } from '@/components/ui/empty'
import { Skeleton } from '@/components/ui/skeleton'
import { ToggleGroup, ToggleGroupItem } from '@/components/ui/toggle-group'
import { AlertCircleIcon, RadioIcon } from 'lucide-react'
import { usePageTitle } from '@/hooks/usePageTitle'
import { usePageSeo } from '@/hooks/usePageSeo'

const PAGE_SIZE = 24

type CategorySelection = {
  l1Id: number
  l2Id: number
  children: ClientLiveCategory[]
  /** False when category_id is set but missing from the tree (do not highlight 全部). */
  matched: boolean
}

/** Missing or 0 means 全部. Any other non-id does not count as 全部. */
function parseCategoryQuery(raw: string | null): { id: number; invalid: boolean } {
  if (raw == null || raw.trim() === '') return { id: 0, invalid: false }
  if (!/^\d+$/.test(raw.trim())) return { id: 0, invalid: true }
  const n = Number(raw)
  if (!Number.isSafeInteger(n)) return { id: 0, invalid: true }
  return { id: n, invalid: false }
}

function resolveCategorySelection(
  categories: ClientLiveCategory[],
  categoryId: number,
): CategorySelection {
  if (categoryId <= 0) {
    return { l1Id: 0, l2Id: 0, children: [], matched: true }
  }
  for (const cat of categories) {
    if (cat.id === categoryId) {
      return { l1Id: cat.id, l2Id: 0, children: cat.children || [], matched: true }
    }
    const child = (cat.children || []).find((c) => c.id === categoryId)
    if (child) {
      return { l1Id: cat.id, l2Id: child.id, children: cat.children || [], matched: true }
    }
  }
  return { l1Id: 0, l2Id: 0, children: [], matched: false }
}

function LiveRoomCard({ item }: { item: ClientLiveRoomListItem }) {
  const [coverError, setCoverError] = useState(false)
  const hasCover = Boolean(item.cover) && !coverError
  const nickname = item.streamer_nickname || item.streamer_str_id || '主播'
  const fallback = (nickname[0] || 'L').toUpperCase()
  const categoryLabel = formatLiveCategoryName(item.category_name, item.sub_category_name)

  return (
    <Link to={`/live/${item.id}`} className="cursor-pointer">
      <Card className="gap-0 overflow-hidden pb-2 pt-0 transition-all hover:ring-primary/40">
        <div className="relative flex aspect-video w-full items-center justify-center bg-muted">
          {hasCover ? (
            <img
              src={item.cover}
              alt={item.title}
              className="absolute inset-0 h-full w-full object-cover"
              onError={() => setCoverError(true)}
            />
          ) : (
            <RadioIcon className="size-10 text-muted-foreground/40" />
          )}
          <div className="absolute top-2 left-2">
            <Badge variant="destructive" className="gap-1">
              <span className="size-1.5 rounded-full bg-white" />
              直播中
            </Badge>
          </div>
        </div>
        <div className="flex flex-col gap-2 p-2">
          <h3 className="truncate text-sm font-medium">{item.title || '未命名直播间'}</h3>
          <div className="flex items-center gap-2">
            <Avatar size="sm">
              {item.streamer_avatar ? <AvatarImage src={item.streamer_avatar} alt={nickname} /> : null}
              <AvatarFallback>{fallback}</AvatarFallback>
            </Avatar>
            <div className="min-w-0 flex-1">
              <p className="truncate text-xs text-muted-foreground">{nickname}</p>
              {categoryLabel ? (
                <p className="truncate text-xs text-muted-foreground">{categoryLabel}</p>
              ) : null}
            </div>
          </div>
        </div>
      </Card>
    </Link>
  )
}

export function Component() {
  const [searchParams, setSearchParams] = useSearchParams()
  const categoryQuery = parseCategoryQuery(searchParams.get('category_id'))
  const categoryId = categoryQuery.id
  const page = Number(searchParams.get('page') || 1) || 1

  const [categories, setCategories] = useState<ClientLiveCategory[]>([])
  const [rooms, setRooms] = useState<ClientLiveRoomListItem[]>([])
  const [totalPages, setTotalPages] = useState(0)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  usePageTitle('直播')
  usePageSeo({ title: '直播', path: '/live' })

  const selection = useMemo(() => {
    if (categoryQuery.invalid) return { l1Id: 0, l2Id: 0, children: [], matched: false }
    return resolveCategorySelection(categories, categoryId)
  }, [categories, categoryId, categoryQuery.invalid])

  const loadCategories = useCallback(async () => {
    try {
      const res = await clientApi.liveCategories()
      setCategories(res.data || [])
    } catch {
      setCategories([])
    }
  }, [])

  const loadRooms = useCallback(async () => {
    if (categoryQuery.invalid) {
      setRooms([])
      setTotalPages(0)
      setError('')
      setLoading(false)
      return
    }
    setLoading(true)
    setError('')
    try {
      const res = await clientApi.liveRooms({
        page,
        page_size: PAGE_SIZE,
        category_id: categoryId || undefined,
      })
      setRooms(res.data?.list || [])
      setTotalPages(res.data?.total_pages || 0)
    } catch (err) {
      setRooms([])
      setTotalPages(0)
      setError(errorMessage(err))
    } finally {
      setLoading(false)
    }
  }, [page, categoryId, categoryQuery.invalid])

  useEffect(() => {
    void loadCategories()
  }, [loadCategories])

  useEffect(() => {
    void loadRooms()
  }, [loadRooms])

  const setCategory = (id: number) => {
    const next = new URLSearchParams(searchParams)
    if (id > 0) next.set('category_id', String(id))
    else next.delete('category_id')
    next.delete('page')
    setSearchParams(next)
  }

  const setPage = (p: number) => {
    const next = new URLSearchParams(searchParams)
    if (p > 1) next.set('page', String(p))
    else next.delete('page')
    setSearchParams(next)
  }

  return (
    <div className="flex flex-col gap-6">
      <div className="flex flex-col gap-2">
        <h1 className="text-2xl font-bold">直播</h1>
        <p className="text-sm text-muted-foreground">观看正在直播的主播</p>
      </div>

      {categories.length > 0 ? (
        <div className="flex flex-col gap-3">
          <ToggleGroup
            variant="outline"
            size="sm"
            value={selection.matched ? [String(selection.l1Id)] : []}
            onValueChange={(v) => {
              if (!v[0]) return
              // Switching L1 clears L2 by writing L1 id (or 0 for 全部)
              setCategory(Number(v[0]) || 0)
            }}
            className="flex w-full flex-wrap justify-start"
          >
            <ToggleGroupItem value="0">全部</ToggleGroupItem>
            {categories.map((cat) => (
              <ToggleGroupItem key={cat.id} value={String(cat.id)}>
                {cat.name}
              </ToggleGroupItem>
            ))}
          </ToggleGroup>

          {selection.l1Id > 0 && selection.children.length > 0 ? (
            <ToggleGroup
              variant="outline"
              size="sm"
              value={[String(selection.l2Id)]}
              onValueChange={(v) => {
                if (!v[0]) return
                const next = Number(v[0]) || 0
                // L2 "全部" keeps L1 filter
                setCategory(next > 0 ? next : selection.l1Id)
              }}
              className="flex w-full flex-wrap justify-start"
            >
              <ToggleGroupItem value="0">全部</ToggleGroupItem>
              {selection.children.map((cat) => (
                <ToggleGroupItem key={cat.id} value={String(cat.id)}>
                  {cat.name}
                </ToggleGroupItem>
              ))}
            </ToggleGroup>
          ) : null}
        </div>
      ) : null}

      {error ? (
        <Alert variant="destructive">
          <AlertCircleIcon />
          <AlertTitle>加载失败</AlertTitle>
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      ) : null}

      {loading ? (
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5">
          {Array.from({ length: 8 }).map((_, i) => (
            <Card key={i} className="gap-0 overflow-hidden pb-2 pt-0">
              <Skeleton className="aspect-video w-full rounded-t-xl" />
              <div className="flex flex-col gap-2 p-2">
                <Skeleton className="h-4 w-3/4" />
                <Skeleton className="h-3 w-1/2" />
              </div>
            </Card>
          ))}
        </div>
      ) : rooms.length === 0 ? (
        <Empty>
          <EmptyHeader>
            <EmptyTitle>暂无直播</EmptyTitle>
            <EmptyDescription>当前没有正在直播的房间</EmptyDescription>
          </EmptyHeader>
        </Empty>
      ) : (
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5">
          {rooms.map((item) => (
            <LiveRoomCard key={item.id} item={item} />
          ))}
        </div>
      )}

      {totalPages > 1 ? (
        <div className="flex items-center justify-center gap-2">
          <Button
            variant="outline"
            size="sm"
            disabled={page <= 1 || loading}
            onClick={() => setPage(page - 1)}
          >
            上一页
          </Button>
          <span className="text-sm text-muted-foreground">
            {page} / {totalPages}
          </span>
          <Button
            variant="outline"
            size="sm"
            disabled={page >= totalPages || loading}
            onClick={() => setPage(page + 1)}
          >
            下一页
          </Button>
        </div>
      ) : null}
    </div>
  )
}
