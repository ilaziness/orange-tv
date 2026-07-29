import { useCallback, useEffect, useRef, useState } from 'react'
import { adminApi, errorMessage } from '@/lib/api'
import { DEFAULT_PAGE_SIZE } from '@/lib/constants'
import type { AdminCommentItem, AdminCommentParentItem } from '@orange-tv/shared'
import { toast } from 'sonner'

type StatusFilter = '' | '0' | '1'

export function useComments() {
  const [items, setItems] = useState<AdminCommentItem[]>([])
  const [keyword, setKeyword] = useState('')
  const [status, setStatus] = useState<StatusFilter>('')
  const [videoId, setVideoId] = useState('')
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(false)
  const [deleteId, setDeleteId] = useState<number | null>(null)
  const [deleteLoading, setDeleteLoading] = useState(false)
  const [toggleId, setToggleId] = useState<number | null>(null)
  const [parentId, setParentId] = useState<number | null>(null)
  const [parents, setParents] = useState<AdminCommentParentItem[]>([])
  const [parentsLoading, setParentsLoading] = useState(false)

  const keywordRef = useRef(keyword)
  const statusRef = useRef(status)
  const videoIdRef = useRef(videoId)
  const pageRef = useRef(page)

  useEffect(() => { keywordRef.current = keyword }, [keyword])
  useEffect(() => { statusRef.current = status }, [status])
  useEffect(() => { videoIdRef.current = videoId }, [videoId])
  useEffect(() => { pageRef.current = page }, [page])

  const load = useCallback(async (p = pageRef.current) => {
    setLoading(true)
    try {
      const query: Record<string, string | number | undefined> = {
        keyword: keywordRef.current,
        page: p,
        page_size: DEFAULT_PAGE_SIZE,
      }
      if (statusRef.current !== '') {
        query.status = Number(statusRef.current)
      }
      if (videoIdRef.current) {
        query.video_id = Number(videoIdRef.current)
      }
      const res = await adminApi.listComments(query)
      setItems(res.data.list || [])
      setTotal(res.data.total || 0)
      setPage(res.data.page || p)
    } catch (err) {
      toast.error(errorMessage(err))
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => { void load(1) }, [load])

  async function confirmDelete() {
    if (deleteId === null) return
    setDeleteLoading(true)
    try {
      await adminApi.deleteComment(deleteId)
      toast.success('评论已删除')
      await load(page)
    } catch (err) {
      toast.error(errorMessage(err))
    } finally {
      setDeleteLoading(false)
      setDeleteId(null)
    }
  }

  async function toggleStatus(id: number, nextStatus: number) {
    setToggleId(id)
    try {
      await adminApi.updateCommentStatus(id, nextStatus)
      setItems((prev) => prev.map((it) => it.id === id ? { ...it, status: nextStatus } : it))
      toast.success(nextStatus === 1 ? '评论已显示' : '评论已隐藏')
    } catch (err) {
      toast.error(errorMessage(err))
    } finally {
      setToggleId(null)
    }
  }

  async function loadParents(id: number) {
    setParentId(id)
    setParentsLoading(true)
    try {
      const res = await adminApi.getCommentParents(id)
      setParents(res.data || [])
    } catch (err) {
      toast.error(errorMessage(err))
      setParents([])
    } finally {
      setParentsLoading(false)
    }
  }

  return {
    items,
    keyword,
    setKeyword,
    status,
    setStatus,
    videoId,
    setVideoId,
    page,
    total,
    loading,
    deleteId,
    setDeleteId,
    deleteLoading,
    toggleId,
    parentId,
    setParentId,
    parents,
    parentsLoading,
    load,
    confirmDelete,
    toggleStatus,
    loadParents,
  }
}
