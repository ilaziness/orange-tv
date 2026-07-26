import { useCallback, useEffect, useRef, useState } from 'react'
import { useSearchParams } from 'react-router'
import { adminApi, errorMessage } from '@/lib/api'
import type { VideoListItem } from '@orange-tv/shared'
import { toast } from 'sonner'
import { DEFAULT_PAGE_SIZE } from '@/lib/constants'

export function useVideos() {
  const [searchParams] = useSearchParams()
  const directorId = searchParams.get('director_id')
  const actorId = searchParams.get('actor_id')
  const tagId = searchParams.get('tag_id')

  const [items, setItems] = useState<VideoListItem[]>([])
  const [keyword, setKeyword] = useState('')
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)
  const [selected, setSelected] = useState<Set<number>>(new Set())
  const [batchAction, setBatchAction] = useState<{ type: 'publish' | 'unpublish' | 'delete'; status?: number } | null>(null)
  const [deleteId, setDeleteId] = useState<number | null>(null)
  const [toggleId, setToggleId] = useState<number | null>(null)
  const [detailId, setDetailId] = useState<number | null>(null)
  const [loading, setLoading] = useState(false)
  const [batchLoading, setBatchLoading] = useState(false)
  const [deleteLoading, setDeleteLoading] = useState(false)
  const keywordRef = useRef(keyword)
  const pageRef = useRef(page)

  useEffect(() => { keywordRef.current = keyword }, [keyword])
  useEffect(() => { pageRef.current = page }, [page])

  const load = useCallback(async (p = pageRef.current) => {
    setLoading(true)
    try {
      const res = await adminApi.listVideos({
        keyword: keywordRef.current,
        page: p,
        page_size: DEFAULT_PAGE_SIZE,
        director_id: directorId ? Number(directorId) : undefined,
        actor_id: actorId ? Number(actorId) : undefined,
        tag_id: tagId ? Number(tagId) : undefined,
      })
      setItems(res.data.list || [])
      setTotal(res.data.total || 0)
      setPage(res.data.page || p)
      setSelected(new Set())
    } catch (err) {
      toast.error(errorMessage(err))
    } finally {
      setLoading(false)
    }
  }, [directorId, actorId, tagId])

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
    setBatchLoading(true)
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
      setBatchLoading(false)
      setBatchAction(null)
    }
  }

  async function confirmDelete() {
    if (deleteId === null) return
    setDeleteLoading(true)
    try {
      await adminApi.deleteVideo(deleteId)
      toast.success('影视已删除')
      await load(page)
    } catch (err) {
      toast.error(errorMessage(err))
    } finally {
      setDeleteLoading(false)
      setDeleteId(null)
    }
  }

  async function togglePublish(id: number) {
    const target = items.find((i) => i.id === id)
    if (!target) return
    const next = target.publish_status === 1 ? 0 : 1
    setToggleId(id)
    try {
      await adminApi.batchUpdatePublishStatus([id], next)
      setItems((prev) => prev.map((it) => it.id === id ? { ...it, publish_status: next } : it))
      toast.success(`已${next === 1 ? '上架' : '下架'}`)
    } catch (err) {
      toast.error(errorMessage(err))
    } finally {
      setToggleId(null)
    }
  }

  return {
    items,
    keyword,
    setKeyword,
    page,
    total,
    selected,
    batchAction,
    setBatchAction,
    deleteId,
    setDeleteId,
    toggleId,
    detailId,
    setDetailId,
    loading,
    batchLoading,
    deleteLoading,
    load,
    toggleSelect,
    toggleSelectAll,
    confirmBatch,
    confirmDelete,
    togglePublish,
    directorId,
    actorId,
    tagId,
  }
}
