import { useCallback, useEffect, useRef, useState } from 'react'
import type { NamedItem, PageData } from '@orange-tv/shared'
import { DEFAULT_PAGE_SIZE } from '@/lib/constants'
import { errorMessage } from '@/lib/api'
import { toast } from 'sonner'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Checkbox } from '@/components/ui/checkbox'
import { Spinner } from '@/components/ui/spinner'
import { Empty, EmptyDescription, EmptyHeader, EmptyTitle } from '@/components/ui/empty'
import { Plus, Search, X } from 'lucide-react'

interface NamedItemPickerProps {
  title: string
  selected: NamedItem[]
  onChange: (items: NamedItem[]) => void
  searchFn: (query: {
    keyword: string
    page: number
    page_size: number
  }) => Promise<{ data: PageData<NamedItem> }>
}

export function NamedItemPicker({ title, selected, onChange, searchFn }: NamedItemPickerProps) {
  const [open, setOpen] = useState(false)
  const [keyword, setKeyword] = useState('')
  const [results, setResults] = useState<NamedItem[]>([])
  const [page, setPage] = useState(1)
  const [totalPages, setTotalPages] = useState(1)
  const [loading, setLoading] = useState(false)
  const [loadingMore, setLoadingMore] = useState(false)
  const [tempSelected, setTempSelected] = useState<Map<number, NamedItem>>(new Map())

  const keywordRef = useRef(keyword)
  const searchFnRef = useRef(searchFn)
  const selectedRef = useRef(selected)
  const reqIdRef = useRef(0)

  useEffect(() => {
    keywordRef.current = keyword
  }, [keyword])
  useEffect(() => {
    searchFnRef.current = searchFn
  }, [searchFn])
  useEffect(() => {
    selectedRef.current = selected
  }, [selected])

  const loadFirst = useCallback(async (kw: string) => {
    const reqId = ++reqIdRef.current
    setLoading(true)
    try {
      const res = await searchFnRef.current({ keyword: kw, page: 1, page_size: DEFAULT_PAGE_SIZE })
      if (reqId !== reqIdRef.current) return
      setResults(res.data.list || [])
      setTotalPages(res.data.total_pages || 1)
      setPage(1)
    } catch (err) {
      if (reqId !== reqIdRef.current) return
      toast.error(errorMessage(err))
      setResults([])
      setTotalPages(1)
    } finally {
      if (reqId === reqIdRef.current) {
        setLoading(false)
        setLoadingMore(false)
      }
    }
  }, [])

  const loadMore = useCallback(async () => {
    const nextPage = page + 1
    const reqId = ++reqIdRef.current
    setLoadingMore(true)
    try {
      const res = await searchFnRef.current({
        keyword: keywordRef.current,
        page: nextPage,
        page_size: DEFAULT_PAGE_SIZE,
      })
      if (reqId !== reqIdRef.current) return
      setResults((prev) => [...prev, ...(res.data.list || [])])
      setTotalPages(res.data.total_pages || 1)
      setPage(nextPage)
    } catch (err) {
      if (reqId !== reqIdRef.current) return
      toast.error(errorMessage(err))
    } finally {
      if (reqId === reqIdRef.current) {
        setLoading(false)
        setLoadingMore(false)
      }
    }
  }, [page])

  // 打开时重置状态并加载首页
  useEffect(() => {
    if (!open) return
    setKeyword('')
    setTempSelected(new Map(selectedRef.current.map((x) => [x.id, x])))
    void loadFirst('')
  }, [open, loadFirst])

  // 关键词变化 debounce 300ms 重新搜索
  useEffect(() => {
    if (!open) return
    const t = setTimeout(() => {
      void loadFirst(keyword)
    }, 300)
    return () => clearTimeout(t)
  }, [keyword, open, loadFirst])

  function removeItem(id: number) {
    onChange(selected.filter((x) => x.id !== id))
  }

  function toggleTemp(item: NamedItem) {
    setTempSelected((prev) => {
      const next = new Map(prev)
      if (next.has(item.id)) next.delete(item.id)
      else next.set(item.id, item)
      return next
    })
  }

  function handleConfirm() {
    onChange(Array.from(tempSelected.values()))
    setOpen(false)
  }

  const hasMore = page < totalPages

  return (
    <div className="rounded-lg border p-4">
      <div className="flex flex-col gap-3">
        <div className="flex items-center justify-between">
          <h3 className="font-medium">{title}</h3>
          <Button variant="outline" size="sm" onClick={() => setOpen(true)}>
            <Plus data-icon="inline-start" />
            添加
          </Button>
        </div>
        {selected.length > 0 ? (
          <div className="flex flex-wrap gap-2">
            {selected.map((item) => (
              <Badge key={item.id} variant="secondary" className="gap-1 pr-1">
                {item.name}
                <button
                  type="button"
                  className="inline-flex size-4 items-center justify-center rounded-full text-muted-foreground hover:bg-foreground/10 hover:text-foreground"
                  onClick={() => removeItem(item.id)}
                  aria-label={`移除 ${item.name}`}
                >
                  <X className="size-3" />
                </button>
              </Badge>
            ))}
          </div>
        ) : (
          <p className="text-sm text-muted-foreground">暂未选择，点击「添加」搜索并选择{title}。</p>
        )}
      </div>

      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent className="sm:max-w-lg">
          <DialogHeader>
            <DialogTitle>选择{title}</DialogTitle>
            <DialogDescription className="sr-only">搜索并选择关联的{title}</DialogDescription>
          </DialogHeader>

          <div className="relative mb-3">
            <Search className="pointer-events-none absolute left-2.5 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              placeholder={`搜索${title}名称`}
              value={keyword}
              onChange={(e) => setKeyword(e.target.value)}
              className="pl-8"
            />
          </div>

          {loading && results.length === 0 ? (
            <div className="flex items-center justify-center py-8">
              <Spinner />
            </div>
          ) : results.length > 0 ? (
            <div className="relative max-h-80 overflow-auto rounded-md border">
              {loading && (
                <div className="absolute inset-0 z-10 flex items-center justify-center bg-background/50">
                  <Spinner />
                </div>
              )}
              <ul className="divide-y">
                {results.map((item) => {
                  const checked = tempSelected.has(item.id)
                  return (
                    <li key={item.id}>
                      <label className="flex cursor-pointer items-center gap-2 px-3 py-2 text-sm hover:bg-muted">
                        <Checkbox checked={checked} onCheckedChange={() => toggleTemp(item)} />
                        <span>{item.name}</span>
                      </label>
                    </li>
                  )
                })}
                {hasMore && (
                  <li className="p-2">
                    <Button
                      variant="ghost"
                      size="sm"
                      className="w-full"
                      onClick={() => void loadMore()}
                      disabled={loadingMore}
                    >
                      {loadingMore ? <Spinner data-icon="inline-start" /> : null}
                      加载更多
                    </Button>
                  </li>
                )}
              </ul>
            </div>
          ) : (
            <Empty className="py-8">
              <EmptyHeader>
                <EmptyTitle>暂无数据</EmptyTitle>
                <EmptyDescription>请尝试搜索其他关键词</EmptyDescription>
              </EmptyHeader>
            </Empty>
          )}

          <DialogFooter>
            <Button variant="outline" onClick={() => setOpen(false)}>
              取消
            </Button>
            <Button onClick={handleConfirm}>确定（已选 {tempSelected.size}）</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
