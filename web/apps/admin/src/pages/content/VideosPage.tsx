import { useCallback, useEffect, useRef, useState } from 'react'
import { adminApi, errorMessage } from '../../lib/api'
import { ErrorAlert, PageCard, PageHeader, StatusBadge } from '../../components/ui'
import { Link } from 'react-router'
import type { VideoListItem } from '@orange-tv/shared'

export default function VideosPage() {
  const [items, setItems] = useState<VideoListItem[]>([])
  const [keyword, setKeyword] = useState('')
  const [error, setError] = useState('')
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)
  const [selected, setSelected] = useState<Set<number>>(new Set())
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

  async function batchPublish(status: number) {
    if (selected.size === 0) return
    if (!confirm(`确认批量${status === 1 ? '上架' : '下架'} ${selected.size} 条影视？`)) return
    try {
      await adminApi.batchUpdatePublishStatus(Array.from(selected), status)
      await load(page)
    } catch (err) {
      setError(errorMessage(err))
    }
  }

  async function batchDelete() {
    if (selected.size === 0) return
    if (!confirm(`确认批量删除 ${selected.size} 条影视？`)) return
    try {
      await adminApi.batchDeleteVideos(Array.from(selected))
      await load(page)
    } catch (err) {
      setError(errorMessage(err))
    }
  }

  return (
    <PageCard className="stack">
      <PageHeader title="影视管理"><Link to="/content/videos/new"><button className="primary">新增影视</button></Link></PageHeader>
      <ErrorAlert>{error}</ErrorAlert>
      <div className="toolbar">
        <input placeholder="关键词" value={keyword} onChange={(e) => setKeyword(e.target.value)} />
        <button onClick={() => void load(1)}>搜索</button>
      </div>
      {selected.size > 0 ? (
        <div className="toolbar">
          <span className="muted">已选 {selected.size} 项</span>
          <button onClick={() => batchPublish(1)}>批量上架</button>
          <button onClick={() => batchPublish(0)}>批量下架</button>
          <button className="danger" onClick={batchDelete}>批量删除</button>
        </div>
      ) : null}
      <table className="table">
        <thead>
          <tr>
            <th><input type="checkbox" checked={selected.size === items.length && items.length > 0} onChange={toggleSelectAll} /></th>
            <th>ID</th><th>标题</th><th>年份</th><th>评分</th><th>状态</th><th>操作</th>
          </tr>
        </thead>
        <tbody>
          {items.map((item) => (
            <tr key={item.id}>
              <td><input type="checkbox" checked={selected.has(item.id)} onChange={() => toggleSelect(item.id)} /></td>
              <td>{item.id}</td>
              <td>{item.title}</td>
              <td>{item.year || '-'}</td>
              <td>{item.rating}</td>
              <td><StatusBadge status={item.publish_status} onText="上架" offText="下架" /></td>
              <td className="actions">
                <Link to={`/content/videos/${item.id}`}><button>编辑</button></Link>
                <button className="danger" onClick={async () => {
                  if (!confirm('确认删除该影视？')) return
                  try {
                    await adminApi.deleteVideo(item.id)
                    await load(page)
                  } catch (err) {
                    setError(errorMessage(err))
                  }
                }}>删除</button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
      <div className="toolbar">
        <span className="muted">共 {total} 条 · 第 {page} 页</span>
        <button disabled={page <= 1} onClick={() => void load(page - 1)}>上一页</button>
        <button disabled={items.length < 20} onClick={() => void load(page + 1)}>下一页</button>
      </div>
    </PageCard>
  )
}
