import { useCallback, useEffect, useRef, useState } from 'react'
import { adminApi, errorMessage } from '@/lib/api'
import type { VideoListItem } from '@orange-tv/shared'
import { toast } from 'sonner'

export function useVideos() {
  const [items, setItems] = useState<VideoListItem[]>([])
  const [keyword, setKeyword] = useState('')
  const [error, setError] = useState('')
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)
  const [selected, setSelected] = useState<Set<number>>(new Set())
  const [batchAction, setBatchAction] = useState<{ type: 'publish' | 'unpublish' | 'delete'; status?: number } | null>(null)
  const [deleteId, setDeleteId] = useState<number | null>(null)
  const keywordRef = useRef(keyword)
  const pageRef = useRef(page)

  useEffect(() => { keywordRef.current = keyword }, [keyword])
  useEffect(() => { pageRef.current = page }, [page])

  const load = useCallback(async (p = pageRef.current) => {
    setError('')
    try {
      const res = await adminApi.listVideos({ keyword: keywordRef.current, page: p, page_size: 20 })
      setItems(res.data.list || [])
      setTotal(res.data.total || 0)
      setPage(res.data.page || p)
      setSelected(new Set())
    } catch (err) {
      setError(errorMessage(err))
    }
  }, [])

  useEffect(() => { void load(1) }, [load])

  function toggleSelect(id: number) {
    setSelected((prev) => {
      const next = new Set(prev)
      if (next.has(id)) next.delete(id)
      else next.add(id)
      return next
    })
  }

  function toggleSelectAll() {
    setSelected((prev) => {
      if (prev.size === items.length) return new Set()
      return new Set(items.map((i) => i.id))
    })
  }

  async function confirmBatch() {
    if (!batchAction || selected.size === 0) return
    try {
      if (batchAction.type === 'delete') {
        await adminApi.batchDeleteVideos(Array.from(selected))
        toast.success(`已删除 ${selected.size} 条影视`)
      } else if (batchAction.status !== undefined) {
        await adminApi.batchUpdatePublishStatus(Array.from(selected), batchAction.status)
        toast.success(`已${batchAction.status === 1 ? '上架' : '下架'} ${selected.size} 条影视`)
      }
      await load(page)
    } catch (err) {
      toast.error(errorMessage(err))
    } finally {
      setBatchAction(null)
    }
  }

  async function confirmDelete() {
    if (deleteId === null) return
    try {
      await adminApi.deleteVideo(deleteId)
      toast.success('影视已删除')
      await load(page)
    } catch (err) {
      toast.error(errorMessage(err))
    } finally {
      setDeleteId(null)
    }
  }

  return {
    items,
    keyword,
    setKeyword,
    error,
    page,
    total,
    selected,
    batchAction,
    setBatchAction,
    deleteId,
    setDeleteId,
    load,
    toggleSelect,
    toggleSelectAll,
    confirmBatch,
    confirmDelete,
  }
}
