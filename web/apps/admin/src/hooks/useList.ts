import { useCallback, useEffect, useRef, useState } from 'react'
import type { PageData } from '@orange-tv/shared'

type ListFetcher<T> = (params: { page: number; page_size: number; keyword?: string }) => Promise<{ data: PageData<T> }>

export function useList<T>(fetcher: ListFetcher<T>, pageSize = 20) {
  const [items, setItems] = useState<T[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [keyword, setKeyword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const fetcherRef = useRef(fetcher)
  const pageSizeRef = useRef(pageSize)
  const pageRef = useRef(page)
  const keywordRef = useRef(keyword)

  useEffect(() => { fetcherRef.current = fetcher }, [fetcher])
  useEffect(() => { pageSizeRef.current = pageSize }, [pageSize])
  useEffect(() => { pageRef.current = page }, [page])
  useEffect(() => { keywordRef.current = keyword }, [keyword])

  const load = useCallback(async (p: number = pageRef.current, k: string = keywordRef.current) => {
    setLoading(true)
    setError('')
    try {
      const res = await fetcherRef.current({ page: p, page_size: pageSizeRef.current, keyword: k || undefined })
      setItems(res.data.list || [])
      setTotal(res.data.total || 0)
      setPage(res.data.page || p)
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void load(1, '')
  }, [fetcher, pageSize, load])

  return {
    items,
    total,
    page,
    keyword,
    error,
    loading,
    setPage,
    setKeyword,
    load,
    refresh: () => load(page, keyword),
  }
}
